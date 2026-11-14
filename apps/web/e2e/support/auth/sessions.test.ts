// @vitest-environment node

import type { SafeUserResource } from "@cartulary/protocol-ts/http";
import { type APIRequestContext, request } from "@playwright/test";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { JsonRequestContextLike } from "../transport/publicJsonClient";
import * as browserSession from "./browserSession";
import {
  createDeploymentUser,
  type DeploymentUserCreation,
} from "./deploymentUsers";
import { withOwnedRequestContext } from "./requestLifetime";
import {
  findProvisionedUser,
  listUsers,
  loadUser,
  loginBootstrapControlPlaneContext,
  patchUser,
  reconcileWorkerAdminManifest,
  verifySessionUnauthorized,
} from "./sessions";
import { storageStateFromCookieValues } from "./storageState";
import * as suiteAdmin from "./suiteAdmin";
import type { WorkerAdminBlueprint } from "./workerAdmin";

type FakeUser = Pick<
  SafeUserResource,
  | "display_name"
  | "email"
  | "is_active"
  | "is_deployment_admin"
  | "mfa_required"
  | "user_id"
  | "user_version"
> & { password: string };

type FakeControlPlane = {
  controlPlane: Parameters<typeof reconcileWorkerAdminManifest>[0];
  stats: {
    canLoginCount: number;
    createCount: number;
    patchCount: number;
    resetPasswordCount: number;
  };
  users: Map<string, FakeUser>;
};

describe("createDeploymentUser", () => {
  it("preserves explicit MFA and deployment-admin intent in request bytes", async () => {
    const fetch = vi.fn<JsonRequestContextLike["fetch"]>(
      async (_url, options) => ({
        headers: () => ({}),
        json: async () => {
          const requestBody = options.data as {
            is_deployment_admin: boolean;
            mfa_required: boolean;
          };
          return {
            data: {
              auth_bindings: [
                {
                  created_at: "2026-08-05T00:00:00Z",
                  provider_key: "local",
                  provider_type: "local",
                  username: "explicit@example.test",
                },
              ],
              created_at: "2026-08-05T00:00:00Z",
              user_id: "11111111-1111-4111-8111-111111111111",
              email: "explicit@example.test",
              display_name: "Explicit User",
              user_version: 1,
              is_active: true,
              last_login_at: null,
              updated_at: "2026-08-05T00:00:00Z",
              updated_by_user_id: null,
              is_deployment_admin: requestBody.is_deployment_admin,
              mfa_required: requestBody.mfa_required,
            },
            meta: { request_id: "request-1" },
          };
        },
        ok: () => true,
        status: () => 201,
      }),
    );
    const request = { fetch };
    type IsRequired<Key extends keyof DeploymentUserCreation> =
      Record<never, never> extends Pick<DeploymentUserCreation, Key>
        ? false
        : true;
    const requiredSecurityFields: [
      IsRequired<"is_deployment_admin">,
      IsRequired<"mfa_required">,
    ] = [true, true];
    expect(requiredSecurityFields).toEqual([true, true]);

    await createDeploymentUser(request, {
      email: "explicit-admin@example.test",
      display_name: "Explicit Admin",
      initial_password: "ExplicitAdmin1!",
      mfa_required: true,
      is_deployment_admin: true,
    });
    await createDeploymentUser(request, {
      email: "explicit-member@example.test",
      display_name: "Explicit Member",
      initial_password: "ExplicitMember1!",
      mfa_required: false,
      is_deployment_admin: false,
    });

    expect(fetch).toHaveBeenCalledTimes(2);
    expect(fetch.mock.calls[0]?.[0]).toBe("/api/v1/users");
    expect(fetch.mock.calls[0]?.[1]?.data).toEqual(
      expect.objectContaining({
        auth_kind: "local",
        mfa_required: true,
        is_deployment_admin: true,
      }),
    );
    expect(fetch.mock.calls[1]?.[1]?.data).toEqual(
      expect.objectContaining({
        auth_kind: "local",
        mfa_required: false,
        is_deployment_admin: false,
      }),
    );
  });
});

describe("deployment admin generated operation boundary", () => {
  it("preserves list, load, and patch request bytes while validating resources", async () => {
    const user = safeUserResource("11111111-1111-4111-8111-111111111111");
    const unrelated = {
      ...safeUserResource("22222222-2222-4222-8222-222222222222"),
      email: "other@example.test",
    };
    const fetch = vi.fn<JsonRequestContextLike["fetch"]>(
      async (path, _options) => ({
        headers: () => ({}),
        json: async () =>
          path.startsWith("/api/v1/users?")
            ? {
                data: {
                  users: path.includes("cursor_token=") ? [user] : [unrelated],
                },
                meta: {
                  request_id: "request-1",
                  paging: {
                    limit: 100,
                    has_more: !path.includes("cursor_token="),
                    next_cursor: path.includes("cursor_token=")
                      ? null
                      : "next/+?=",
                  },
                },
              }
            : { data: user, meta: { request_id: "request-1" } },
        ok: () => true,
        status: () => 200,
      }),
    );
    const client = { fetch };

    await expect(findProvisionedUser(client, user.email)).resolves.toBe(
      user.user_id,
    );
    await expect(listUsers(client)).resolves.toEqual([unrelated, user]);
    await expect(loadUser(client, user.user_id)).resolves.toEqual(user);
    await expect(
      patchUser(client, user.user_id, {
        base_user_version: 7,
        is_deployment_admin: false,
      }),
    ).resolves.toEqual(user);

    expect(fetch.mock.calls).toEqual([
      [
        "/api/v1/users?limit=100&search=generated%40example.test",
        { method: "GET" },
      ],
      ["/api/v1/users?cursor_token=next%2F%2B%3F%3D", { method: "GET" }],
      ["/api/v1/users?limit=100", { method: "GET" }],
      ["/api/v1/users?cursor_token=next%2F%2B%3F%3D", { method: "GET" }],
      [`/api/v1/users/${user.user_id}`, { method: "GET" }],
      [
        `/api/v1/users/${user.user_id}`,
        {
          data: {
            base_user_version: 7,
            is_deployment_admin: false,
          },
          method: "PATCH",
        },
      ],
    ]);
  });

  it("fails closed when a successful deployment-user payload drifts", async () => {
    const fetch = vi.fn<JsonRequestContextLike["fetch"]>(async () => ({
      headers: () => ({}),
      json: async () => ({
        data: { users: [{ user_id: "missing-required-fields" }] },
        meta: {
          request_id: "request-1",
          paging: { limit: 100, has_more: false, next_cursor: null },
        },
      }),
      ok: () => true,
      status: () => 200,
    }));

    await expect(
      findProvisionedUser({ fetch }, "generated@example.test"),
    ).rejects.toThrow("list users failed with HTTP 502");

    const user = safeUserResource("11111111-1111-4111-8111-111111111111");
    const terminal = { limit: 100, has_more: false, next_cursor: null };
    for (const fixture of [
      { paging: undefined, message: "HTTP 502" },
      { paging: { ...terminal, has_more: true }, message: "HTTP 502" },
      {
        paging: { ...terminal, next_cursor: "unexpected" },
        message: "HTTP 502",
      },
    ]) {
      const malformed = vi.fn<JsonRequestContextLike["fetch"]>(async () => ({
        headers: () => ({}),
        json: async () => ({
          data: { users: [user] },
          meta: { request_id: "request-1", paging: fixture.paging },
        }),
        ok: () => true,
        status: () => 200,
      }));
      await expect(listUsers({ fetch: malformed })).rejects.toThrow(
        fixture.message,
      );
      expect(malformed).toHaveBeenCalledTimes(1);
    }

    for (const variant of ["cycle", "failure", "ambiguous"] as const) {
      let page = 0;
      const continuation = vi.fn<JsonRequestContextLike["fetch"]>(async () => {
        page += 1;
        const failed = variant === "failure" && page === 2;
        return {
          headers: () => ({}),
          json: async () => ({
            data: {
              users: [
                {
                  ...user,
                  user_id:
                    page === 1
                      ? user.user_id
                      : "22222222-2222-4222-8222-222222222222",
                },
              ],
            },
            meta: {
              request_id: "request-1",
              paging:
                page === 2 && variant === "ambiguous"
                  ? terminal
                  : { limit: 100, has_more: true, next_cursor: "repeat" },
            },
          }),
          ok: () => !failed,
          status: () => (failed ? 503 : 200),
        };
      });
      await expect(
        findProvisionedUser({ fetch: continuation }, user.email),
      ).rejects.toThrow(
        variant === "cycle"
          ? "did not advance"
          : variant === "failure"
            ? "HTTP 503"
            : "identity is ambiguous",
      );
      expect(continuation).toHaveBeenCalledTimes(2);
    }
  });
});

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
        schema_id: "cartulary.playwright_worker_admin_manifest.v1",
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
        schema_id: "cartulary.playwright_worker_admin_manifest.v1",
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

function buildBlueprints(workerCount: number) {
  return Array.from({ length: workerCount }, (_, parallelIndex) => ({
    parallelIndex,
    email: `playwright-worker-admin-${parallelIndex}@example.test`,
    password: `PlaywrightWorker${parallelIndex}Pass!`,
    displayName: `Playwright Worker Admin ${parallelIndex}`,
  })) satisfies WorkerAdminBlueprint[];
}

function safeUserResource(userId: string): SafeUserResource {
  return {
    auth_bindings: [
      {
        created_at: "2026-08-05T00:00:00Z",
        provider_key: "local",
        provider_type: "local",
        username: "generated@example.test",
      },
    ],
    created_at: "2026-08-05T00:00:00Z",
    display_name: "Generated User",
    email: "generated@example.test",
    is_active: true,
    is_deployment_admin: true,
    last_login_at: null,
    mfa_required: false,
    updated_at: "2026-08-05T00:00:00Z",
    updated_by_user_id: null,
    user_id: userId,
    user_version: 7,
  };
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

function toUserResource(user: FakeUser): SafeUserResource {
  return {
    auth_bindings: [],
    created_at: "2026-08-05T00:00:00Z",
    user_id: user.user_id,
    email: user.email,
    display_name: user.display_name,
    user_version: user.user_version,
    is_active: user.is_active,
    mfa_required: user.mfa_required,
    is_deployment_admin: user.is_deployment_admin,
    last_login_at: null,
    updated_at: "2026-08-05T00:00:00Z",
    updated_by_user_id: null,
  };
}

afterEach(() => vi.restoreAllMocks());

describe("authentication request lifetime", () => {
  it("retires before disposal and preserves successful results", async () => {
    const events: string[] = [];
    const context = {
      dispose: vi.fn(async () => {
        events.push("dispose");
      }),
    } as unknown as APIRequestContext;
    const value = { user: "owned" };
    await expect(
      withOwnedRequestContext(
        context,
        async (owned) => {
          expect(owned).toBe(context);
          events.push("operation");
          return value;
        },
        async (owned) => {
          expect(owned).toBe(context);
          events.push("retire");
        },
      ),
    ).resolves.toBe(value);
    expect(events).toEqual(["operation", "retire", "dispose"]);
    expect(context.dispose).toHaveBeenCalledOnce();
  });

  it("attempts every release and preserves primary and cleanup failures separately", async () => {
    for (let failures = 1; failures < 8; failures += 1) {
      const primary = new Error("operation failure");
      const retirement = new Error("retirement failure");
      const disposal = new Error("disposal failure");
      const events: string[] = [];
      const context = {
        dispose: vi.fn(async () => {
          events.push("dispose");
          if (failures & 4) throw disposal;
        }),
      } as unknown as APIRequestContext;
      const outcome = await withOwnedRequestContext(
        context,
        async () => {
          events.push("operation");
          if (failures & 1) throw primary;
          return "completed";
        },
        async () => {
          events.push("retire");
          if (failures & 2) throw retirement;
        },
      ).then(
        () => ({ error: undefined }),
        (error: unknown) => ({ error }),
      );
      expect(events).toEqual(["operation", "retire", "dispose"]);
      expect(context.dispose).toHaveBeenCalledOnce();
      if (failures === 1) {
        expect(outcome.error).toBe(primary);
      } else {
        expect(outcome.error).toBeInstanceOf(AggregateError);
        expect((outcome.error as AggregateError).errors).toEqual([
          ...(failures & 1 ? [primary] : []),
          ...(failures & 2 ? [retirement] : []),
          ...(failures & 4 ? [disposal] : []),
        ]);
        expect((outcome.error as AggregateError).cause).toBe(
          failures & 1 ? primary : undefined,
        );
      }
    }
  });

  it("settles the anonymous predecessor before acquiring a transferable control plane", async () => {
    vi.spyOn(suiteAdmin, "loadSuiteAdminTotpSecret").mockReturnValue(
      "RPS5CE6AQ4OODIJC22FH374UBK33BGWY37H5TKC2UXA2ZQBGE6PA",
    );
    vi.spyOn(browserSession, "loginLocalAPIContext").mockResolvedValue({
      ok: true,
      response: {
        headersArray: () => [
          { name: "set-cookie", value: "cartulary_session=session" },
          { name: "set-cookie", value: "cartulary_csrf=csrf" },
        ],
      },
    } as unknown as Awaited<
      ReturnType<typeof browserSession.loginLocalAPIContext>
    >);
    for (const failDisposal of [false, true]) {
      const events: string[] = [];
      const disposal = new Error("anonymous disposal failure");
      const anonymous = {
        get: vi.fn(async () => ({ ok: () => true })),
        dispose: vi.fn(async () => {
          events.push("anonymous.dispose");
          if (failDisposal) throw disposal;
        }),
      } as unknown as APIRequestContext;
      const authenticated = {
        dispose: vi.fn(async () => {}),
      } as unknown as APIRequestContext;
      const create = vi
        .spyOn(request, "newContext")
        .mockReset()
        .mockImplementationOnce(async () => {
          events.push("anonymous.acquire");
          return anonymous;
        })
        .mockImplementationOnce(async () => {
          events.push("authenticated.acquire");
          return authenticated;
        });
      if (failDisposal) {
        const outcome = await loginBootstrapControlPlaneContext().then(
          () => undefined,
          (error: unknown) => error,
        );
        expect(outcome).toBeInstanceOf(AggregateError);
        expect((outcome as AggregateError).errors).toEqual([disposal]);
        expect(create).toHaveBeenCalledOnce();
      } else {
        const result = await loginBootstrapControlPlaneContext();
        expect(result.request).toBe(authenticated);
        expect(create).toHaveBeenCalledTimes(2);
        expect(authenticated.dispose).not.toHaveBeenCalled();
      }
      expect(anonymous.dispose).toHaveBeenCalledOnce();
      expect(events).toEqual([
        "anonymous.acquire",
        "anonymous.dispose",
        ...(failDisposal ? [] : ["authenticated.acquire"]),
      ]);
    }
  });

  it("leaves borrowed revocation transports with their owner", async () => {
    const transport = {
      get: vi.fn(async () => ({
        status: () => 401,
        json: async () => ({ error: { code: "session_required" } }),
      })),
      dispose: vi.fn(async () => {}),
    } as unknown as APIRequestContext;
    await verifySessionUnauthorized(
      storageStateFromCookieValues("session", "csrf"),
      "borrowed",
      transport,
    );
    expect(transport.get).toHaveBeenCalledOnce();
    expect(transport.dispose).not.toHaveBeenCalled();
  });
});
