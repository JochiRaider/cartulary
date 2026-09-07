import { afterEach, describe, expect, it, vi } from "vitest";
import { listEnterpriseAuthProviders } from "./api/authAccountClient";
import * as api from "./api/deploymentUserClient";
import type { UserResource } from "./api/publicHttpTypes";
import { DeploymentUsersController } from "./deploymentUsersModel";

const user = (
  version = 1,
  name = "Alpha",
  id = "00000000-0000-4000-8000-000000000002",
): UserResource => ({
  user_id: id,
  user_version: version,
  display_name: name,
  email: "alpha@deployment",
  is_active: true,
  is_deployment_admin: false,
  mfa_required: true,
  created_at: "2026-09-06T12:00:00Z",
  updated_at: "2026-09-06T12:00:00Z",
  updated_by_user_id: null,
  last_login_at: null,
  auth_bindings: [],
});
const success = <T>(data: T) => ({
  ok: true as const,
  status: 200,
  payload: { data, meta: { request_id: "test" } },
});
const failure = (code: string, status = 409) => ({
  ok: false as const,
  status,
  payload: { error: { code, status } },
});
const page = (users = [user()], next: string | null = null) => ({
  ...success({ users }),
  payload: {
    data: { users },
    meta: {
      request_id: "test",
      paging: {
        limit: 100,
        ...(next === null
          ? { next_cursor: null, has_more: false as const }
          : { next_cursor: next, has_more: true as const }),
      },
    },
  },
});
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
const flush = async () => {
  for (let i = 0; i < 20; ++i) await Promise.resolve();
};
const controllers: DeploymentUsersController[] = [];
afterEach(() => {
  for (const model of controllers.splice(0)) model.dispose();
  vi.useRealTimers();
});
function fixture() {
  let identity = {
    actor: "00000000-0000-4000-8000-000000000001",
    lifetime: "first",
    admin: true,
  };
  let notify = () => {};
  const refresh = vi.fn<() => Promise<void>>().mockResolvedValue();
  const revoked = vi.fn();
  const lost = vi.fn();
  const ports = {
    ...api,
    listEnterpriseAuthProviders,
    listUsers: vi.fn<typeof api.listUsers>().mockResolvedValue(page()),
    loadUser: vi.fn<typeof api.loadUser>().mockResolvedValue(success(user())),
    patchLocalUser: vi
      .fn<typeof api.patchLocalUser>()
      .mockResolvedValue(success(user(2, "Saved"))),
    createLocalUser: vi
      .fn<typeof api.createLocalUser>()
      .mockResolvedValue(success(user())),
    adminResetPassword: vi
      .fn<typeof api.adminResetPassword>()
      .mockResolvedValue(success(user(2))),
    adminResetTotp: vi
      .fn<typeof api.adminResetTotp>()
      .mockResolvedValue(success(user(2))),
    adminRevokeAllSessions: vi
      .fn<typeof api.adminRevokeAllSessions>()
      .mockResolvedValue(
        success({
          user_id: user().user_id,
          sessions_revoked: true as const,
          revoked_at: "2026-09-06T12:00:00Z",
        }),
      ),
    createEnterpriseAuthBinding: vi
      .fn<typeof api.createEnterpriseAuthBinding>()
      .mockResolvedValue(success(user(2))),
    rotateEnterpriseAuthBinding: vi
      .fn<typeof api.rotateEnterpriseAuthBinding>()
      .mockResolvedValue(success(user(2))),
    retireEnterpriseAuthBinding: vi
      .fn<typeof api.retireEnterpriseAuthBinding>()
      .mockResolvedValue(success(user(2))),
  };
  const model = new DeploymentUsersController(
    {
      identity: () => identity,
      subscribe: (fn) => {
        notify = fn;
        return () => {};
      },
      refresh,
      revoked,
      lost,
    },
    ports,
  );
  controllers.push(model);
  model.open();
  return {
    model,
    ports,
    refresh,
    revoked,
    lost,
    replace: (patch: Partial<typeof identity>) => {
      identity = { ...identity, ...patch };
      notify();
    },
  };
}

describe("deployment user operation ownership", () => {
  it("loads on activation and search while keeping inactive panels quiet", async () => {
    vi.useFakeTimers();
    const { model, ports } = fixture();
    await flush();
    expect(ports.listUsers).toHaveBeenCalledOnce();
    model.open();
    await flush();
    expect(ports.listUsers).toHaveBeenCalledOnce();
    model.changeQuery({ search: "current" });
    await vi.advanceTimersByTimeAsync(180);
    expect(ports.listUsers).toHaveBeenCalledTimes(2);
    expect(ports.listUsers.mock.lastCall?.[0]?.search).toBe("current");
    model.close();
    await vi.advanceTimersByTimeAsync(1000);
    expect(ports.listUsers).toHaveBeenCalledTimes(2);
    model.open();
    await flush();
    expect(ports.listUsers).toHaveBeenCalledTimes(3);
    expect(ports.listUsers.mock.lastCall?.[0]?.search).toBe("current");
  });
  it("requires a fresh successful observation for the captured uncertain operation", async () => {
    const { model, ports } = fixture();
    await model.select(user().user_id);
    model.changeDraft("display_name", "Draft");
    const write = deferred<Awaited<ReturnType<typeof api.patchLocalUser>>>();
    ports.patchLocalUser.mockReturnValueOnce(write.promise);
    const saving = model.save();
    const oldRead = deferred<Awaited<ReturnType<typeof api.loadUser>>>();
    ports.loadUser.mockReturnValueOnce(oldRead.promise);
    const refreshing = model.refreshTarget();
    write.resolve(failure("unavailable", 503));
    await saving;
    oldRead.resolve(success(user()));
    await refreshing;
    expect(model.getSnapshot().operation).toMatchObject({
      kind: "uncertain",
      recovery: { kind: "observe_review", observation: "unobserved" },
    });
    model.review();
    expect(model.getSnapshot().operation.kind).toBe("uncertain");
    ports.loadUser.mockResolvedValueOnce(failure("unavailable", 503));
    await model.refreshTarget();
    expect(model.getSnapshot().operation).toMatchObject({
      recovery: { observation: "failed" },
    });
    await model.refreshTarget();
    expect(model.getSnapshot().operation).toMatchObject({
      kind: "uncertain",
      recovery: { observation: "ready" },
    });
    model.review();
    expect(model.getSnapshot().operation.kind).toBe("idle");
    expect(model.getSnapshot().draft?.changes).toEqual({
      display_name: "Draft",
    });
    expect(ports.patchLocalUser).toHaveBeenCalledOnce();
  });
  it("fences old pages, failed query replacements, duplicate activation and invalid cursors", async () => {
    const { model, ports } = fixture();
    ports.listUsers.mockResolvedValueOnce(page([user()], "cursor"));
    await model.refreshUsers();
    const old = deferred<Awaited<ReturnType<typeof api.listUsers>>>();
    ports.listUsers.mockReturnValueOnce(old.promise);
    const loading = model.loadMore();
    void model.loadMore();
    expect(ports.listUsers).toHaveBeenCalledTimes(2);
    model.changeQuery({ search: "new" });
    expect(model.canPage()).toBe(false);
    ports.listUsers.mockResolvedValueOnce(failure("unavailable", 503));
    await model.refreshUsers();
    expect(model.getSnapshot().users).toEqual([user()]);
    expect(model.getSnapshot().acceptedQuery?.search).toBe("");
    old.resolve(page([user(2, "Obsolete")], "bad"));
    await loading;
    expect(model.getSnapshot().users[0]?.display_name).toBe("Alpha");
    ports.listUsers.mockResolvedValueOnce(page([user()], "fresh"));
    await model.refreshUsers();
    ports.listUsers.mockResolvedValueOnce(
      failure("invalid_pagination_request", 400),
    );
    await model.loadMore();
    expect(model.canPage()).toBe(false);
    expect(model.getSnapshot().paging.invalid).toBe(true);
    ports.listUsers.mockResolvedValueOnce(page());
    await model.refreshUsers();
    expect(ports.listUsers.mock.lastCall?.[0]?.cursorToken).toBeUndefined();
    ports.listUsers.mockResolvedValueOnce(
      page([user(3), user(2), user(3)], "duplicates"),
    );
    await model.refreshUsers();
    expect(model.getSnapshot().users).toEqual([user(3)]);
    ports.listUsers.mockResolvedValueOnce(page([user(5), user(4), user(3)]));
    await model.loadMore();
    expect(model.getSnapshot().users).toEqual([user(5)]);
  });

  it("accepts monotonic resources and rejects inconsistent same-version reads without overwriting drafts", async () => {
    const { model, ports } = fixture();
    await model.select(user().user_id);
    model.changeDraft("display_name", "Draft");
    ports.loadUser.mockResolvedValueOnce(success(user(3, "External")));
    await model.refreshTarget();
    expect(model.getSnapshot().draft).toMatchObject({
      baseVersion: 1,
      changes: { display_name: "Draft" },
    });
    expect(model.canSave()).toBe(false);
    model.review();
    expect(model.canSave()).toBe(true);
    ports.loadUser.mockResolvedValueOnce(success(user(2, "Older")));
    await model.refreshTarget();
    expect(model.getSnapshot().selected?.user_version).toBe(3);
    ports.loadUser.mockResolvedValueOnce(success(user(3, "Inconsistent")));
    await model.refreshTarget();
    expect(model.getSnapshot().selected?.display_name).toBe("External");
    expect(model.getSnapshot().targetStatus).toBe("failed");
  });

  it("suppresses empty patches and acknowledges only the submitted draft revision", async () => {
    const { model, ports } = fixture();
    await model.select(user().user_id);
    await model.save();
    expect(ports.patchLocalUser).not.toHaveBeenCalled();
    model.changeDraft("display_name", "Saved");
    const pending = deferred<Awaited<ReturnType<typeof api.patchLocalUser>>>();
    ports.patchLocalUser.mockReturnValueOnce(pending.promise);
    const saving = model.save();
    void model.save();
    expect(ports.patchLocalUser).toHaveBeenCalledTimes(1);
    expect(ports.patchLocalUser.mock.lastCall?.[0]).toMatchObject({
      baseUserVersion: 1,
      changes: { display_name: "Saved" },
    });
    model.changeDraft("display_name", "Newer draft");
    pending.resolve(success(user(2, "Saved")));
    await saving;
    expect(model.getSnapshot().operation.kind).toBe("confirmed");
    expect(model.getSnapshot().draft).toMatchObject({
      baseVersion: 1,
      changes: { display_name: "Newer draft" },
    });
    expect(model.needsReview()).toBe(true);
    model.review();
    expect(model.getSnapshot().draft?.baseVersion).toBe(2);
  });

  it("keeps Save and leave pending for newer edits and lets Discard leave without erasing an operation", async () => {
    const { model, ports } = fixture();
    await model.select(user().user_id);
    model.changeDraft("display_name", "Saved");
    const leave = model.requestLeave();
    const pending = deferred<Awaited<ReturnType<typeof api.patchLocalUser>>>();
    ports.patchLocalUser.mockReturnValueOnce(pending.promise);
    const saving = model.resolveLeave("save");
    model.changeDraft("email", "new@deployment");
    pending.resolve(success(user(2, "Saved")));
    await saving;
    expect(model.getSnapshot().leavePrompt).toBe(true);
    await model.resolveLeave("discard");
    expect(await leave).toBe(true);
    expect(model.getSnapshot().operation.kind).toBe("confirmed");
    expect(model.hasDirtyDraft()).toBe(false);
    model.changeDraft("display_name", "Other");
    const stay = model.requestLeave();
    await model.resolveLeave("stay");
    expect(await stay).toBe(false);
    expect(model.hasDirtyDraft()).toBe(true);
    ports.patchLocalUser.mockResolvedValueOnce(success(user(3, "Other")));
    const cleanLeave = model.requestLeave();
    await model.resolveLeave("save");
    expect(await cleanLeave).toBe(true);
    model.changeDraft("display_name", "  Other  ");
    const normalizedLeave = model.requestLeave();
    await model.resolveLeave("save");
    expect(await normalizedLeave).toBe(true);
    expect(ports.patchLocalUser).toHaveBeenCalledTimes(2);
  });

  it("preserves a PATCH uncertainty until fresh observation and explicit review, never replaying it", async () => {
    vi.useFakeTimers();
    const { model, ports } = fixture();
    await model.select(user().user_id);
    model.changeDraft("display_name", "Saved");
    const pending = deferred<Awaited<ReturnType<typeof api.patchLocalUser>>>();
    ports.patchLocalUser.mockReturnValueOnce(pending.promise);
    const saving = model.save();
    await vi.advanceTimersByTimeAsync(30_000);
    await saving;
    expect(model.getSnapshot().operation).toMatchObject({
      kind: "uncertain",
      recovery: { kind: "observe_review", observation: "unobserved" },
    });
    model.replay();
    model.review();
    expect(ports.patchLocalUser).toHaveBeenCalledTimes(1);
    pending.resolve(success(user(2, "Saved")));
    await flush();
    expect(model.getSnapshot().operation.kind).toBe("uncertain");
    ports.loadUser.mockResolvedValueOnce(success(user(2, "Saved")));
    await model.refreshTarget();
    expect(model.getSnapshot().operation.kind).toBe("uncertain");
    model.review();
    expect(model.hasDirtyDraft()).toBe(false);
    expect(model.getSnapshot().operation.kind).toBe("idle");
  });

  it("replays exact non-credential actions against captured targets after selection and version changes", async () => {
    for (const action of [
      "totp",
      "revoke",
      "bindingCreate",
      "bindingRotate",
      "bindingRetire",
    ] as const) {
      const { model, ports } = fixture();
      const bound = {
        ...user(),
        auth_bindings: [
          {
            auth_binding_id: "00000000-0000-4000-8000-000000000050",
            provider_type: "oidc" as const,
            provider_key: "corp",
            provider_subject: "original",
            created_at: user().created_at,
            last_auth_at: null,
          },
        ],
      };
      ports.loadUser.mockResolvedValueOnce(success(bound));
      await model.select(bound.user_id);
      model.configure(true);
      model.changeBinding("providerKey", "corp");
      model.changeBinding("subject", "  e\u0301 opaque  ");
      model.changeBinding("newSubject", "  replacement  ");
      model.changeBinding(
        "target",
        bound.auth_bindings[0]?.auth_binding_id ?? "",
      );
      model.changeReason("Original reason");
      const route =
        action === "totp"
          ? ports.adminResetTotp
          : action === "revoke"
            ? ports.adminRevokeAllSessions
            : action === "bindingCreate"
              ? ports.createEnterpriseAuthBinding
              : action === "bindingRotate"
                ? ports.rotateEnterpriseAuthBinding
                : ports.retireEnterpriseAuthBinding;
      route.mockResolvedValueOnce(failure("unavailable", 503));
      model.safeAction(action);
      await flush();
      expect(model.getSnapshot().operation.kind).toBe("uncertain");
      const first = route.mock.calls[0]?.[0];
      expect(first).toBeDefined();
      const other = user(9, "Other", "00000000-0000-4000-8000-000000000003");
      ports.loadUser.mockResolvedValueOnce(success(other));
      await model.select(other.user_id);
      model.changeReason("Changed reason");
      model.changeBinding("subject", "changed");
      model.replay();
      await flush();
      const { signal: _firstSignal, ...original } = first ?? {};
      const { signal: _secondSignal, ...replayed } =
        route.mock.calls[1]?.[0] ?? {};
      expect(replayed).toEqual(original);
      expect(model.getSnapshot().selected?.user_id).toBe(other.user_id);
      expect(model.getSnapshot().operation.kind).toBe("confirmed");
      if (action === "bindingCreate")
        expect(original).toMatchObject({
          providerSubject: "  e\u0301 opaque  ",
        });
      if (action === "bindingRotate")
        expect(original).toMatchObject({
          newProviderSubject: "  replacement  ",
        });
    }
  });

  it("keeps confirmed writes confirmed when refresh fails and retries only propagation", async () => {
    const { model, ports, refresh } = fixture();
    await model.select(user().user_id);
    model.changeDraft("display_name", "Saved");
    refresh.mockRejectedValueOnce(new Error("read unavailable"));
    await model.save();
    expect(model.getSnapshot().operation).toMatchObject({
      kind: "confirmed",
      propagation: "failed",
    });
    await model.retryRefresh();
    expect(model.getSnapshot().operation).toMatchObject({
      kind: "confirmed",
      propagation: "ready",
    });
    expect(ports.patchLocalUser).toHaveBeenCalledTimes(1);
  });

  it("clears submitted credentials and never retains them for recovery or late selection", async () => {
    for (const create of [true, false]) {
      const { model, ports } = fixture();
      await model.select(user().user_id);
      if (create) {
        await model.createDialog(true);
        model.changeCreate("email", "operator@deployment");
        model.changeCreate("displayName", "Created");
        model.changeCreate("password", "  Exact Password!  ");
        ports.createLocalUser.mockResolvedValueOnce(
          failure("unavailable", 503),
        );
        model.create();
      } else {
        model.credentialDialog("password");
        model.changePassword("  Exact Password!  ");
        ports.adminResetPassword.mockResolvedValueOnce(
          failure("unavailable", 503),
        );
        model.resetPassword();
      }
      expect(model.getSnapshot().password).toBe("");
      expect(model.getSnapshot().create.password).toBe("");
      await flush();
      expect(JSON.stringify(model.getSnapshot())).not.toContain(
        "Exact Password",
      );
      expect(model.getSnapshot().operation).toMatchObject({
        kind: "uncertain",
        recovery: { kind: "observe_review", observation: "unobserved" },
      });
      const sent = create
        ? ports.createLocalUser.mock.calls[0]?.[0].initialPassword
        : ports.adminResetPassword.mock.calls[0]?.[0].newPassword;
      expect(sent).toBe("  Exact Password!  ");
      model.close();
      model.open();
      expect(model.getSnapshot().password).toBe("");
    }
  });

  it("makes obsolete success and authorization failures inert after capability loss and identical-account reauthentication", async () => {
    for (const result of [
      success(user(2, "Saved")),
      failure("session_required", 401),
      failure("authorization_denied", 403),
    ]) {
      const { model, ports, refresh, lost, revoked, replace } = fixture();
      await model.select(user().user_id);
      model.changeDraft("display_name", "Saved");
      const pending =
        deferred<Awaited<ReturnType<typeof api.patchLocalUser>>>();
      ports.patchLocalUser.mockReturnValueOnce(pending.promise);
      const saving = model.save();
      const leave = model.requestLeave();
      replace({ admin: false });
      expect(await leave).toBe(false);
      expect(model.getSnapshot().selected).toBeNull();
      expect(model.hasDirtyDraft()).toBe(false);
      replace({ admin: true, lifetime: "replacement" });
      pending.resolve(result);
      await saving;
      expect(model.getSnapshot().operation.kind).toBe("idle");
      expect(refresh).not.toHaveBeenCalled();
      expect(lost).not.toHaveBeenCalled();
      expect(revoked).not.toHaveBeenCalled();
    }
  });

  it("routes self revocation through the session owner and requires review for transaction conflict", async () => {
    const { model, ports, revoked } = fixture();
    const self = user(1, "Self", "00000000-0000-4000-8000-000000000001");
    ports.loadUser.mockResolvedValueOnce(success(self));
    await model.select(self.user_id);
    ports.adminRevokeAllSessions.mockResolvedValueOnce(
      success({
        user_id: self.user_id,
        sessions_revoked: true as const,
        revoked_at: "2026-09-06T12:00:00Z",
      }),
    );
    model.safeAction("revoke");
    await flush();
    expect(revoked).toHaveBeenCalledTimes(1);
    const other = fixture();
    await other.model.select(user().user_id);
    other.ports.adminResetTotp.mockResolvedValueOnce(
      failure("client_txn_conflict"),
    );
    other.model.safeAction("totp");
    await flush();
    expect(other.model.getSnapshot().operation).toMatchObject({
      kind: "rejected",
      review: true,
    });
    expect(other.model.canTargetAction()).toBe(false);
    other.model.review();
    expect(other.model.canTargetAction()).toBe(true);
  });
});
