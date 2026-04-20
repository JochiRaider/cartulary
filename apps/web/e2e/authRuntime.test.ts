// @vitest-environment node

import { describe, expect, it } from "vitest";

import {
  type UserResource,
  reconcileWorkerAdminManifest,
} from "./authRuntime";
import type { WorkerAdminBlueprint } from "./sessionSupport";

type FakeUser = UserResource & {
  password: string;
};

type FakeControlPlane = {
  controlPlane: Parameters<typeof reconcileWorkerAdminManifest>[0];
  stats: {
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
    const fake = createFakeControlPlane([
      {
        user_id: "worker-0",
        email: blueprints[0].email,
        display_name: "Outdated Worker Admin",
        user_version: 3,
        is_active: false,
        mfa_required: true,
        is_deployment_admin: false,
        password: blueprints[0].password,
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
        email: blueprints[0].email,
        password: blueprints[0].password,
      },
    ]);
  });

  it("resets the canonical password when a reused worker-admin user can no longer log in", async () => {
    const blueprints = buildBlueprints(1);
    const fake = createFakeControlPlane([
      {
        user_id: "worker-0",
        email: blueprints[0].email,
        display_name: blueprints[0].displayName,
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
            email: blueprints[0].email,
            password: blueprints[0].password,
          },
        ],
      },
    );

    expect(fake.stats.createCount).toBe(0);
    expect(fake.stats.patchCount).toBe(0);
    expect(fake.stats.resetPasswordCount).toBe(1);
    expect(manifest.worker_admins[0]?.user_id).toBe("worker-0");
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

function createFakeControlPlane(initialUsers: FakeUser[] = []): FakeControlPlane {
  let nextID = initialUsers.length + 1;
  const users = new Map(initialUsers.map((user) => [user.user_id, user]));
  const stats = {
    createCount: 0,
    patchCount: 0,
    resetPasswordCount: 0,
  };

  return {
    controlPlane: {
      canLogin: async (email, password) => {
        const user = [...users.values()].find((candidate) => candidate.email === email);
        if (!user) {
          return false;
        }
        return (
          user.password === password &&
          user.is_active &&
          !user.mfa_required
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
      listUsers: async () => [...users.values()].map((user) => toUserResource(user)),
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
