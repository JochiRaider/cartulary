// @vitest-environment node

import { describe, expect, it } from "vitest";

import {
  type DeploymentAdminMutationClient,
  reconcileWorkerAdminManifest,
  type UserResource,
  withOnlyActiveDeploymentAdmin,
} from "./sessions";
import type { WorkerAdminBlueprint } from "./workerAdmin";

type FakeUser = UserResource & {
  password: string;
};

type FakeControlPlane = {
  controlPlane: Parameters<typeof reconcileWorkerAdminManifest>[0] &
    DeploymentAdminMutationClient;
  stats: {
    canLoginCount: number;
    createCount: number;
    patchCount: number;
    resetPasswordCount: number;
  };
  users: Map<string, FakeUser>;
};

describe("reconcileWorkerAdminManifest", () => {
  it("reuses the same worker-admin accounts across sequential invocations", async () => {
    const blueprints = buildBlueprints(2);
    const fake = createFakeControlPlane();

    const firstManifest = await reconcileWorkerAdminManifest(
      fake.controlPlane,
      blueprints,
      null,
    );
    const secondManifest = await reconcileWorkerAdminManifest(
      fake.controlPlane,
      blueprints,
      firstManifest,
    );

    expect(fake.stats.createCount).toBe(2);
    expect(fake.stats.patchCount).toBe(0);
    expect(fake.stats.resetPasswordCount).toBe(0);
    expect(secondManifest).toEqual(firstManifest);
  });

  it("rediscovers deterministic worker-admin users when local manifest state is missing", async () => {
    const blueprints = buildBlueprints(1);
    const blueprint = requireBlueprint(blueprints, 0);
    const fake = createFakeControlPlane([
      {
        user_id: "worker-0",
        email: blueprint.email,
        display_name: "Outdated Worker Admin",
        user_version: 3,
        is_active: false,
        mfa_required: true,
        is_deployment_admin: false,
        password: blueprint.password,
      },
    ]);

    const manifest = await reconcileWorkerAdminManifest(
      fake.controlPlane,
      blueprints,
      null,
    );

    expect(fake.stats.createCount).toBe(0);
    expect(fake.stats.patchCount).toBe(1);
    expect(manifest.worker_admins).toEqual([
      {
        parallel_index: 0,
        user_id: "worker-0",
        email: blueprint.email,
        password: blueprint.password,
      },
    ]);
  });

  it("resets the canonical password when a reused worker-admin user can no longer log in", async () => {
    const blueprints = buildBlueprints(1);
    const blueprint = requireBlueprint(blueprints, 0);
    const fake = createFakeControlPlane([
      {
        user_id: "worker-0",
        email: blueprint.email,
        display_name: blueprint.displayName,
        user_version: 1,
        is_active: true,
        mfa_required: false,
        is_deployment_admin: true,
        password: "wrong-password",
      },
    ]);

    const manifest = await reconcileWorkerAdminManifest(
      fake.controlPlane,
      blueprints,
      {
        worker_admins: [
          {
            parallel_index: 0,
            user_id: "worker-0",
            email: blueprint.email,
            password: blueprint.password,
          },
        ],
      },
    );

    expect(fake.stats.createCount).toBe(0);
    expect(fake.stats.patchCount).toBe(0);
    expect(fake.stats.resetPasswordCount).toBe(1);
    expect(manifest.worker_admins[0]?.user_id).toBe("worker-0");
  });

  it("trusts existing shared-harness manifest passwords without creating probe sessions", async () => {
    const blueprints = buildBlueprints(2);
    const first = requireBlueprint(blueprints, 0);
    const second = requireBlueprint(blueprints, 1);
    const fake = createFakeControlPlane([
      fakeUser("worker-0", {
        email: first.email,
        display_name: first.displayName,
        is_deployment_admin: true,
        password: "rotated-password",
      }),
      fakeUser("worker-1", {
        email: second.email,
        display_name: second.displayName,
        is_deployment_admin: true,
        password: "rotated-password",
      }),
    ]);

    const manifest = await reconcileWorkerAdminManifest(
      fake.controlPlane,
      blueprints,
      {
        worker_admins: [
          {
            parallel_index: 0,
            user_id: "worker-0",
            email: first.email,
            password: first.password,
          },
          {
            parallel_index: 1,
            user_id: "worker-1",
            email: second.email,
            password: second.password,
          },
        ],
      },
      { trustExistingManifestPasswords: true },
    );

    expect(fake.stats.canLoginCount).toBe(0);
    expect(fake.stats.resetPasswordCount).toBe(0);
    expect(manifest.worker_admins.map((entry) => entry.user_id)).toEqual([
      "worker-0",
      "worker-1",
    ]);
  });
});

describe("withOnlyActiveDeploymentAdmin", () => {
  it("temporarily demotes active deployment admins except the retained admin", async () => {
    const fake = createFakeControlPlane([
      fakeUser("retained-admin", {
        email: "retained@example.test",
        is_deployment_admin: true,
      }),
      fakeUser("other-admin", {
        email: "other@example.test",
        is_deployment_admin: true,
      }),
      fakeUser("inactive-admin", {
        email: "inactive@example.test",
        is_active: false,
        is_deployment_admin: true,
      }),
      fakeUser("ordinary-user", {
        email: "ordinary@example.test",
      }),
    ]);

    await withOnlyActiveDeploymentAdmin(
      fake.controlPlane,
      "retained-admin",
      async () => {
        expect(
          requireUser(fake.users, "retained-admin").is_deployment_admin,
        ).toBe(true);
        expect(requireUser(fake.users, "other-admin").is_deployment_admin).toBe(
          false,
        );
        expect(
          requireUser(fake.users, "inactive-admin").is_deployment_admin,
        ).toBe(true);
        expect(
          requireUser(fake.users, "ordinary-user").is_deployment_admin,
        ).toBe(false);
      },
    );

    expect(requireUser(fake.users, "other-admin").is_deployment_admin).toBe(
      true,
    );
    expect(requireUser(fake.users, "other-admin").user_version).toBe(3);
  });

  it("restores demoted admins after a guarded block fails", async () => {
    const fake = createFakeControlPlane([
      fakeUser("retained-admin", {
        email: "retained@example.test",
        is_deployment_admin: true,
      }),
      fakeUser("other-admin", {
        email: "other@example.test",
        is_deployment_admin: true,
      }),
    ]);

    await expect(
      withOnlyActiveDeploymentAdmin(
        fake.controlPlane,
        "retained-admin",
        async () => {
          throw new Error("probe failed");
        },
      ),
    ).rejects.toThrow("probe failed");

    expect(requireUser(fake.users, "other-admin").is_deployment_admin).toBe(
      true,
    );
  });
});

function buildBlueprints(workerCount: number) {
  return Array.from({ length: workerCount }, (_, parallelIndex) => ({
    parallelIndex,
    email: `playwright-worker-admin-${parallelIndex}@example.test`,
    password: `PlaywrightWorker${parallelIndex}Pass!`,
    displayName: `Playwright Worker Admin ${parallelIndex}`,
  })) satisfies WorkerAdminBlueprint[];
}

function requireBlueprint(
  blueprints: WorkerAdminBlueprint[],
  index: number,
): WorkerAdminBlueprint {
  const blueprint = blueprints[index];
  if (!blueprint) {
    throw new Error(`missing worker admin blueprint ${index}`);
  }
  return blueprint;
}

function createFakeControlPlane(
  initialUsers: FakeUser[] = [],
): FakeControlPlane {
  let nextID = initialUsers.length + 1;
  const users = new Map(initialUsers.map((user) => [user.user_id, user]));
  const stats = {
    canLoginCount: 0,
    createCount: 0,
    patchCount: 0,
    resetPasswordCount: 0,
  };

  return {
    controlPlane: {
      canLogin: async (email, password) => {
        stats.canLoginCount += 1;
        const user = [...users.values()].find(
          (candidate) => candidate.email === email,
        );
        if (!user) {
          return false;
        }
        return (
          user.password === password && user.is_active && !user.mfa_required
        );
      },
      createUser: async (blueprint) => {
        stats.createCount += 1;
        const created = {
          user_id: `worker-${nextID++}`,
          email: blueprint.email,
          display_name: blueprint.displayName,
          user_version: 1,
          is_active: true,
          mfa_required: false,
          is_deployment_admin: true,
          password: blueprint.password,
        } satisfies FakeUser;
        users.set(created.user_id, created);
        return toUserResource(created);
      },
      listUsers: async () =>
        [...users.values()].map((user) => toUserResource(user)),
      loadUser: async (userId) => toUserResource(requireUser(users, userId)),
      patchUser: async (userId, body) => {
        stats.patchCount += 1;
        const current = requireUser(users, userId);
        const updated = {
          ...current,
          display_name: body.display_name ?? current.display_name,
          is_active: body.is_active ?? current.is_active,
          mfa_required: body.mfa_required ?? current.mfa_required,
          is_deployment_admin:
            body.is_deployment_admin ?? current.is_deployment_admin,
          user_version: current.user_version + 1,
        } satisfies FakeUser;
        users.set(userId, updated);
        return toUserResource(updated);
      },
      resetUserPassword: async (userId, _baseUserVersion, nextPassword) => {
        stats.resetPasswordCount += 1;
        const current = requireUser(users, userId);
        const updated = {
          ...current,
          password: nextPassword,
          user_version: current.user_version + 1,
        } satisfies FakeUser;
        users.set(userId, updated);
        return toUserResource(updated);
      },
      revokeAllSessions: async () => {},
    },
    stats,
    users,
  };
}

function fakeUser(userId: string, overrides: Partial<FakeUser> = {}): FakeUser {
  return {
    user_id: userId,
    email: `${userId}@example.test`,
    display_name: userId,
    user_version: 1,
    is_active: true,
    mfa_required: false,
    is_deployment_admin: false,
    password: "Password1!",
    ...overrides,
  };
}

function requireUser(users: Map<string, FakeUser>, userId: string) {
  const user = users.get(userId);
  if (!user) {
    throw new Error(`missing fake user ${userId}`);
  }
  return user;
}

function toUserResource(user: FakeUser): UserResource {
  return {
    user_id: user.user_id,
    email: user.email,
    display_name: user.display_name,
    user_version: user.user_version,
    is_active: user.is_active,
    mfa_required: user.mfa_required,
    is_deployment_admin: user.is_deployment_admin,
  };
}
