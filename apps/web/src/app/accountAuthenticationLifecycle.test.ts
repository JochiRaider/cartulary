import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { APIError } from "../services/browserApi";
import {
  credentialStateResource,
  sessionResource,
} from "../testing/appShellTestSupport";
import {
  validateAccountEmail,
  validateProvisioningPassword,
} from "./accountInputValidation";
import { accountOperationError } from "./accountOperation";
import { AccountSecurityController } from "./accountSecurityModel";
import * as api from "./api/authAccountClient";
import {
  AppSessionController,
  type SessionObservation,
} from "./appSessionController";
import { AuthenticationController } from "./authenticationModel";

const success = <T>(data: T) => ({
  ok: true as const,
  status: 200,
  payload: { data, meta: { request_id: "test" } },
});
const failure = (
  code: string,
  status = 401,
  details?: APIError["details"],
) => ({
  ok: false as const,
  status,
  payload: { error: { code, status, ...(details ? { details } : {}) } },
});
function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: () => void;
  const promise = new Promise<T>((a, b) => {
    resolve = a;
    reject = b;
  });
  return { promise, resolve, reject };
}
const flush = async () => {
  for (let i = 0; i < 16; ++i) await Promise.resolve();
};
const controllers: { dispose: () => void }[] = [];
const provider = {
  provider_key: "corp",
  provider_type: "oidc" as const,
  display_name: "Corporate",
};
const enrollment = () => ({
  enrollment_id: "00000000-0000-4000-8000-000000000001",
  expires_at: new Date(Date.now() + 60_000).toISOString(),
  totp_setup: {
    secret_base32: "SEED",
    otpauth_uri: "otpauth://test",
    algorithm: "SHA1" as const,
    digits: 6 as const,
    period_seconds: 30 as const,
  },
});
function auth(gate?: AppSessionController) {
  gate ??= new AppSessionController();
  controllers.push(gate);
  let alive = true;
  const authenticated = vi.fn();
  const uncertain = vi
    .fn<
      (
        signal: AbortSignal,
        current: () => boolean,
      ) => Promise<SessionObservation>
    >()
    .mockResolvedValue({ kind: "session_lost" });
  const assign = vi.fn();
  const loginLocal = vi
    .fn<typeof api.loginLocal>()
    .mockResolvedValue(failure("invalid_credentials"));
  const beginEnterpriseAuth = vi
    .fn<typeof api.beginEnterpriseAuth>()
    .mockResolvedValue(
      success({
        ...provider,
        redirect_url: "https://idp.example/start",
        expires_at: new Date(Date.now() + 60_000).toISOString(),
      }),
    );
  const beginTotpEnrollment = vi
    .fn<typeof api.beginTotpEnrollment>()
    .mockResolvedValue(success(enrollment()));
  const completeTotpEnrollment = vi
    .fn<typeof api.completeTotpEnrollment>()
    .mockResolvedValue(
      success({
        user_id: sessionResource().user_id,
        totp: { enrolled_at: new Date().toISOString() },
        sessions_revoked: false,
      }),
    );
  const listEnterpriseAuthProviders = vi
    .fn<typeof api.listEnterpriseAuthProviders>()
    .mockResolvedValue(success({ providers: [provider] }));
  const controller = new AuthenticationController(
    () => ({
      actor: sessionResource().user_id,
      current: () => alive,
      admitTransport: gate.reserveAuthenticationTransport,
      canAuthenticate: () => !gate.getSnapshot().authenticationTransportPending,
      authenticated,
      inspectSession: uncertain,
    }),
    { assign, returnTo: () => "/?deployment_admin=1" },
    {
      loginLocal,
      beginEnterpriseAuth,
      beginTotpEnrollment,
      completeTotpEnrollment,
      listEnterpriseAuthProviders,
    },
  );
  controllers.push(controller);
  controller.open();
  const credentials = () => {
    controller.dispatch({
      type: "field",
      field: "username",
      value: "operator@deployment",
    });
    controller.dispatch({
      type: "field",
      field: "password",
      value: "  Exact e\u0301 Password  ",
    });
  };
  credentials();
  return {
    controller,
    credentials,
    authenticated,
    uncertain,
    assign,
    loginLocal,
    beginEnterpriseAuth,
    beginTotpEnrollment,
    completeTotpEnrollment,
    listEnterpriseAuthProviders,
    obsolete: () => {
      alive = false;
    },
  };
}
function security(gate?: AppSessionController) {
  gate ??= new AppSessionController();
  controllers.push(gate);
  let alive = true;
  const event = vi.fn();
  const loadCredentialState = vi
    .fn<typeof api.loadCredentialState>()
    .mockResolvedValue(success(credentialStateResource()));
  const changePassword = vi.fn<typeof api.changePassword>().mockResolvedValue(
    success({
      user_id: sessionResource().user_id,
      password: { changed_at: new Date().toISOString() },
      sessions_revoked: true,
    }),
  );
  const beginTotpEnrollment = vi
    .fn<typeof api.beginTotpEnrollment>()
    .mockResolvedValue(success(enrollment()));
  const logoutCurrentSession = vi
    .fn<typeof api.logoutCurrentSession>()
    .mockResolvedValue(
      success({
        user_id: sessionResource().user_id,
        logged_out: true,
        sessions_revoked: false,
      }),
    );
  const controller = new AccountSecurityController(
    () => ({
      actor: sessionResource().user_id,
      current: () => alive,
      event,
      admitLogout: gate.reserveAuthenticationTransport,
    }),
    {
      ...api,
      loadCredentialState,
      changePassword,
      beginTotpEnrollment,
      logoutCurrentSession,
    },
  );
  controllers.push(controller);
  controller.open();
  const credentials = () => {
    controller.change("passwordCurrent", "Current Password!");
    controller.change("passwordNext", "  Exact e\u0301 Password  ");
  };
  return {
    controller,
    event,
    logoutCurrentSession,
    changePassword,
    loadCredentialState,
    beginTotpEnrollment,
    credentials,
    obsolete: () => {
      alive = false;
    },
  };
}
beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(new Date("2026-09-07T00:00:00Z"));
});
afterEach(() => {
  for (const c of controllers.splice(0)) c.dispose();
  vi.useRealTimers();
  vi.restoreAllMocks();
});
describe("authentication and Security operation lifetime", () => {
  it("bounds and serializes session inspection without converting observations into write receipts", async () => {
    const s = auth();
    const first = deferred<SessionObservation>();
    const second = deferred<SessionObservation>();
    s.uncertain
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    s.loginLocal.mockRejectedValue(new TypeError());
    s.controller.login();
    await flush();
    await s.controller.retrySession();
    expect(s.uncertain).toHaveBeenCalledOnce();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(s.controller.getSnapshot().sessionRead).toBe("unavailable");
    expect(s.uncertain.mock.calls[0]?.[0].aborted).toBe(true);
    expect(s.uncertain.mock.calls[0]?.[1]()).toBe(false);
    const retry = s.controller.retrySession();
    expect(s.uncertain).toHaveBeenCalledTimes(2);
    first.resolve({ kind: "accepted", session: sessionResource() });
    await flush();
    expect(s.controller.getSnapshot().sessionRead).toBe("pending");
    second.resolve({ kind: "session_lost" });
    await retry;
    expect(s.controller.getSnapshot()).toMatchObject({
      sessionRead: "absent",
      operation: { kind: "uncertain" },
    });
    expect(s.loginLocal).toHaveBeenCalledOnce();
  });
  it("accepts a manually observed session while retaining admission for an outstanding login", async () => {
    const read = vi
      .fn<typeof api.loadSession>()
      .mockResolvedValue(failure("session_required"));
    const gate = new AppSessionController({
      session: (signal) => read({ signal }),
    });
    const s = auth(gate);
    const login = deferred<Awaited<ReturnType<typeof api.loginLocal>>>();
    s.loginLocal.mockReturnValue(login.promise);
    s.uncertain.mockImplementation((signal, current) =>
      gate.confirmAuthentication(0, signal, current),
    );
    s.controller.login();
    await vi.advanceTimersByTimeAsync(30_000);
    read.mockResolvedValue(success(sessionResource()));
    await s.controller.retrySession();
    expect(gate.getSnapshot().session?.user_id).toBe(sessionResource().user_id);
    expect(gate.getSnapshot().authenticationTransportPending).toBe(true);
    expect(gate.reserveAuthenticationTransport()).toBeNull();
    s.controller.retire();
    login.resolve(success(sessionResource()));
    await flush();
    expect(gate.getSnapshot().authenticationTransportPending).toBe(false);
    expect(s.authenticated).not.toHaveBeenCalled();
  });
  it("releases only matching Security transports across retirement and disposal", async () => {
    const s = security();
    const password = deferred<Awaited<ReturnType<typeof api.changePassword>>>();
    const logout =
      deferred<Awaited<ReturnType<typeof api.logoutCurrentSession>>>();
    s.changePassword.mockReturnValue(password.promise);
    s.logoutCurrentSession.mockReturnValue(logout.promise);
    s.credentials();
    s.controller.password();
    s.controller.logout();
    s.controller.retire();
    s.controller.open();
    password.resolve(
      success({
        user_id: sessionResource().user_id,
        password: { changed_at: new Date().toISOString() },
        sessions_revoked: true,
      }),
    );
    await flush();
    expect(s.controller.getSnapshot().transportPending).toBe(true);
    logout.resolve(
      success({
        user_id: sessionResource().user_id,
        logged_out: true,
        sessions_revoked: false,
      }),
    );
    await flush();
    expect(s.controller.getSnapshot().transportPending).toBe(false);
    expect(s.event).not.toHaveBeenCalled();
    const late =
      deferred<Awaited<ReturnType<typeof api.logoutCurrentSession>>>();
    s.logoutCurrentSession.mockReturnValue(late.promise);
    s.controller.logout();
    s.controller.dispose();
    const snapshot = s.controller.getSnapshot();
    late.resolve(
      success({
        user_id: sessionResource().user_id,
        logged_out: true,
        sessions_revoked: false,
      }),
    );
    await flush();
    expect(s.controller.getSnapshot()).toBe(snapshot);
    expect(s.event).not.toHaveBeenCalled();
  });
  it("allows manual session inspection while an uncertain login transport remains outstanding", async () => {
    const s = auth();
    const pending = deferred<Awaited<ReturnType<typeof api.loginLocal>>>();
    s.loginLocal.mockReturnValue(pending.promise);
    s.controller.login();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(s.uncertain).toHaveBeenCalledOnce();
    await s.controller.retrySession();
    expect(s.uncertain).toHaveBeenCalledTimes(2);
    s.controller.login();
    expect(s.loginLocal).toHaveBeenCalledOnce();
    pending.resolve(success(sessionResource()));
    await flush();
    expect(s.authenticated).not.toHaveBeenCalled();
  });
  it("fences session acceptance itself when the initiating authentication flow closes", async () => {
    const pending = deferred<Awaited<ReturnType<typeof api.loadSession>>>();
    const session = new AppSessionController({
      session: () => pending.promise,
    });
    controllers.push(session);
    const s = auth();
    s.uncertain.mockImplementation((signal, current) =>
      session.confirmAuthentication(0, signal, current),
    );
    s.loginLocal.mockRejectedValue(new TypeError());
    s.controller.login();
    await flush();
    expect(s.uncertain).toHaveBeenCalledOnce();
    s.controller.close();
    pending.resolve(success(sessionResource()));
    await flush();
    expect(session.getSnapshot().session).toBeNull();
  });
  it("releases Security transport availability after retirement without publishing its outcome", async () => {
    const s = security();
    const pending = deferred<Awaited<ReturnType<typeof api.changePassword>>>();
    s.changePassword.mockReturnValue(pending.promise);
    s.credentials();
    s.controller.password();
    s.controller.retire();
    s.controller.open();
    expect(s.controller.getSnapshot().transportPending).toBe(true);
    pending.resolve(
      success({
        user_id: sessionResource().user_id,
        password: { changed_at: new Date().toISOString() },
        sessions_revoked: true,
      }),
    );
    await flush();
    expect(s.controller.getSnapshot().transportPending).toBe(false);
    expect(s.controller.getSnapshot().operation.kind).toBe("idle");
    expect(s.event).not.toHaveBeenCalled();
  });
  it("serializes local and enterprise activation while preserving exact login bytes", async () => {
    const s = auth();
    await s.controller.discover();
    const pending = deferred<Awaited<ReturnType<typeof api.loginLocal>>>();
    s.loginLocal.mockReturnValue(pending.promise);
    s.controller.login();
    s.controller.login();
    s.controller.enterprise("corp");
    expect(s.loginLocal).toHaveBeenCalledTimes(1);
    expect(s.beginEnterpriseAuth).not.toHaveBeenCalled();
    expect(s.loginLocal.mock.calls[0]?.[0]).toMatchObject({
      username: "operator@deployment",
      password: "  Exact e\u0301 Password  ",
    });
    pending.resolve(success(sessionResource()));
    await flush();
    expect(s.authenticated).toHaveBeenCalledTimes(1);
    expect(s.controller.getSnapshot().flow).toMatchObject({ password: "" });
  });
  it("bounds MFA password retention from initial dispatch and clears submitted factor codes", async () => {
    const s = auth();
    s.loginLocal
      .mockResolvedValueOnce(failure("mfa_required"))
      .mockResolvedValueOnce(failure("invalid_second_factor"));
    s.controller.login();
    await flush();
    expect(s.controller.getSnapshot().flow.kind).toBe("mfa");
    expect(s.controller.getSnapshot().flow).toMatchObject({
      kind: "mfa",
      password: "  Exact e\u0301 Password  ",
    });
    await vi.advanceTimersByTimeAsync(240_000);
    s.controller.dispatch({
      type: "field",
      field: "totpCode",
      value: "123456",
    });
    s.controller.login();
    expect(s.controller.getSnapshot().flow).toMatchObject({
      kind: "mfa",
      code: "",
    });
    await flush();
    expect(s.controller.getSnapshot().flow).toMatchObject({
      kind: "mfa",
      password: "  Exact e\u0301 Password  ",
    });
    await vi.advanceTimersByTimeAsync(60_000);
    expect(s.controller.getSnapshot()).toMatchObject({
      flow: { kind: "credentials", password: "" },
    });
  });
  it("keeps timeout uncertainty separate from outstanding authentication transport settlement", async () => {
    const session = new AppSessionController();
    const release = session.reserveAuthenticationTransport();
    expect(release).not.toBeNull();
    session.logoutConfirmed();
    expect(session.reserveAuthenticationTransport()).toBeNull();
    expect(session.getSnapshot().authenticationTransportPending).toBe(true);
    release?.();
    release?.();
    expect(session.getSnapshot().authenticationTransportPending).toBe(false);
    session.dispose();
    const s = auth();
    const pending = deferred<Awaited<ReturnType<typeof api.loginLocal>>>();
    s.loginLocal.mockReturnValue(pending.promise);
    s.controller.login();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(s.controller.getSnapshot()).toMatchObject({
      operation: { kind: "uncertain", action: "login" },
      flow: { kind: "credentials", password: "" },
      transportPending: true,
    });
    expect(s.uncertain).toHaveBeenCalledOnce();
    s.credentials();
    s.controller.login();
    expect(s.loginLocal).toHaveBeenCalledOnce();
    pending.resolve(success(sessionResource()));
    await flush();
    expect(s.authenticated).not.toHaveBeenCalled();
    expect(s.controller.getSnapshot().transportPending).toBe(false);
    s.credentials();
    s.controller.login();
    expect(s.loginLocal).toHaveBeenCalledTimes(2);
    const gate = new AppSessionController();
    const exiting = security(gate);
    const logoutPending =
      deferred<Awaited<ReturnType<typeof api.logoutCurrentSession>>>();
    exiting.logoutCurrentSession.mockReturnValue(logoutPending.promise);
    exiting.controller.logout();
    await vi.advanceTimersByTimeAsync(30_000);
    const replacement = auth(gate);
    replacement.controller.login();
    expect(replacement.loginLocal).not.toHaveBeenCalled();
    exiting.controller.retire();
    logoutPending.resolve(
      success({
        user_id: sessionResource().user_id,
        logged_out: true,
        sessions_revoked: false,
      }),
    );
    await flush();
    expect(exiting.event).not.toHaveBeenCalled();
    replacement.controller.login();
    expect(replacement.loginLocal).toHaveBeenCalledOnce();
    gate.dispose();
  });
  it("fences uncertain-session continuation and late login publication after retirement", async () => {
    const s = auth();
    const observed = deferred<SessionObservation>();
    s.uncertain.mockReturnValue(observed.promise);
    s.loginLocal.mockRejectedValue(new TypeError());
    s.controller.login();
    await flush();
    s.controller.close();
    observed.resolve({ kind: "accepted", session: sessionResource() });
    await flush();
    expect(s.controller.getSnapshot().flow).toMatchObject({ password: "" });
    expect(s.authenticated).not.toHaveBeenCalled();
    s.controller.open();
    s.credentials();
    const pending = deferred<Awaited<ReturnType<typeof api.loginLocal>>>();
    s.loginLocal.mockReturnValue(pending.promise);
    s.controller.login();
    s.obsolete();
    pending.resolve(success(sessionResource()));
    await flush();
    expect(s.authenticated).not.toHaveBeenCalled();
  });
  it("retains only active bootstrap material and never signs in from enrollment completion", async () => {
    const s = auth();
    s.loginLocal.mockResolvedValue(
      failure("mfa_setup_required", 401, {
        bootstrap_token: "PRIVATE-TOKEN",
        bootstrap_expires_at: new Date(Date.now() + 120_000).toISOString(),
      }),
    );
    s.controller.login();
    await flush();
    expect(s.controller.getSnapshot()).toMatchObject({
      flow: { kind: "setup", enrollment: { kind: "unissued" } },
      error: { code: "mfa_setup_required" },
    });
    expect(JSON.stringify(s.controller.getSnapshot())).not.toContain(
      "PRIVATE-TOKEN",
    );
    expect(s.controller.getSnapshot().error?.details).not.toHaveProperty(
      "bootstrap_token",
    );
    s.controller.begin();
    s.controller.begin();
    await flush();
    expect(s.beginTotpEnrollment).toHaveBeenCalledOnce();
    expect(s.controller.getSnapshot().flow).toMatchObject({
      kind: "setup",
      enrollment: { kind: "ready", key: "SEED" },
    });
    s.controller.dispatch({
      type: "field",
      field: "bootstrapCompleteCode",
      value: "123456",
    });
    s.controller.complete();
    expect(s.controller.getSnapshot().flow).toMatchObject({
      kind: "setup",
      code: "",
    });
    await flush();
    expect(s.controller.getSnapshot()).toMatchObject({
      flow: { kind: "credentials", password: "" },
    });
    expect(s.authenticated).not.toHaveBeenCalled();
  });
  it("expires enrollment material and excludes obsolete enrollment success", async () => {
    const s = auth();
    s.loginLocal.mockResolvedValue(
      failure("mfa_setup_required", 401, {
        bootstrap_token: "PRIVATE",
        bootstrap_expires_at: new Date(Date.now() + 45_000).toISOString(),
      }),
    );
    s.controller.login();
    await flush();
    s.controller.begin();
    await flush();
    await vi.advanceTimersByTimeAsync(45_000);
    expect(s.controller.getSnapshot()).toMatchObject({
      flow: { kind: "credentials", password: "" },
    });
    s.credentials();
    s.loginLocal.mockResolvedValue(
      failure("mfa_setup_required", 401, {
        bootstrap_token: "NEW",
        bootstrap_expires_at: new Date(Date.now() + 60_000).toISOString(),
      }),
    );
    s.controller.login();
    await flush();
    const pending =
      deferred<Awaited<ReturnType<typeof api.beginTotpEnrollment>>>();
    s.beginTotpEnrollment.mockReturnValue(pending.promise);
    s.controller.begin();
    s.controller.dispatch({ type: "use_different_account" });
    pending.resolve(success(enrollment()));
    await flush();
    expect(s.controller.getSnapshot().flow.kind).toBe("credentials");
    expect(JSON.stringify(s.controller.getSnapshot())).not.toContain("SEED");
  });
  it("separates failed provider discovery from unclaimed discovery and isolates navigation instances", async () => {
    const a = auth();
    const b = auth();
    a.listEnterpriseAuthProviders.mockRejectedValueOnce(new TypeError());
    await a.controller.discover();
    expect(a.controller.getSnapshot().providersStatus).toBe("failed");
    a.listEnterpriseAuthProviders.mockResolvedValueOnce(
      failure("extension_profile_not_claimed", 404),
    );
    await a.controller.discover();
    expect(a.controller.getSnapshot()).toMatchObject({
      providersStatus: "ready",
      enterpriseProviders: [],
    });
    await a.controller.discover();
    await b.controller.discover();
    const pending =
      deferred<Awaited<ReturnType<typeof api.beginEnterpriseAuth>>>();
    a.beginEnterpriseAuth.mockReturnValue(pending.promise);
    a.controller.enterprise("corp");
    a.controller.close();
    b.controller.enterprise("corp");
    await flush();
    expect(b.assign).toHaveBeenCalledOnce();
    expect(a.assign).not.toHaveBeenCalled();
    pending.resolve(
      success({
        ...provider,
        redirect_url: "https://obsolete.example",
        expires_at: new Date(Date.now() + 30_000).toISOString(),
      }),
    );
    await flush();
    expect(a.assign).not.toHaveBeenCalled();
  });
  it("treats malformed authentication success as uncertainty and clears terminal-failure secrets", async () => {
    expect(
      accountOperationError({
        code: "invalid_credentials",
        message: "PRIVATE-PASSWORD",
        details: {
          bootstrap_token: "PRIVATE-TOKEN",
          secret_base32: "PRIVATE-SEED",
        },
      }),
    ).toEqual({ code: "invalid_credentials", details: {} });
    const s = auth();
    s.loginLocal.mockResolvedValueOnce(
      failure("invalid_public_contract_response", 502),
    );
    s.controller.login();
    await flush();
    expect(s.uncertain).toHaveBeenCalledOnce();
    expect(s.controller.getSnapshot().flow).toMatchObject({ password: "" });
    s.credentials();
    s.controller.login();
    await flush();
    expect(s.controller.getSnapshot()).toMatchObject({
      flow: { kind: "credentials", password: "" },
    });
    const confirmed = auth();
    confirmed.loginLocal.mockResolvedValue(success(sessionResource()));
    confirmed.authenticated.mockRejectedValue(
      new Error("publication unavailable"),
    );
    confirmed.controller.login();
    await flush();
    expect(confirmed.controller.getSnapshot().operation).toEqual({
      kind: "confirmed",
      propagation: "failed",
    });
    confirmed.controller.login();
    expect(confirmed.loginLocal).toHaveBeenCalledOnce();
    confirmed.uncertain.mockResolvedValue({
      kind: "accepted",
      session: sessionResource(),
    });
    await confirmed.controller.retrySession();
    expect(confirmed.controller.getSnapshot().operation).toEqual({
      kind: "confirmed",
      propagation: "ready",
    });
    expect(confirmed.loginLocal).toHaveBeenCalledOnce();
  });
  it("serializes Security writes and clears every dispatched credential", async () => {
    const s = security();
    await flush();
    const pending = deferred<Awaited<ReturnType<typeof api.changePassword>>>();
    s.changePassword.mockReturnValue(pending.promise);
    s.credentials();
    s.controller.password();
    s.controller.password();
    expect(s.changePassword).toHaveBeenCalledOnce();
    expect(s.changePassword.mock.calls[0]?.[0].newPassword).toBe(
      "  Exact e\u0301 Password  ",
    );
    expect(s.controller.getSnapshot()).toMatchObject({
      passwordCurrent: "",
      passwordNext: "",
      passwordFactorCode: "",
    });
    s.controller.close();
    s.controller.open();
    pending.resolve(
      success({
        user_id: sessionResource().user_id,
        password: { changed_at: new Date().toISOString() },
        sessions_revoked: true,
      }),
    );
    await flush();
    expect(s.event).not.toHaveBeenCalled();
    expect(s.controller.getSnapshot().operation.kind).toBe("confirmed");
    const logout = security();
    await flush();
    const earlier = deferred<Awaited<ReturnType<typeof api.changePassword>>>();
    logout.changePassword.mockReturnValue(earlier.promise);
    logout.credentials();
    logout.controller.password();
    logout.controller.change("passwordNext", "Private later draft");
    logout.controller.logout();
    expect(logout.controller.getSnapshot().passwordNext).toBe("");
    await flush();
    expect(logout.logoutCurrentSession).toHaveBeenCalledOnce();
    earlier.resolve(
      success({
        user_id: sessionResource().user_id,
        password: { changed_at: new Date().toISOString() },
        sessions_revoked: true,
      }),
    );
    await flush();
    expect(logout.event).toHaveBeenCalledOnce();
    expect(logout.event.mock.calls[0]?.[0].kind).toBe("logout_confirmed");
  });
  it("keeps Security uncertainty after a late success and requires explicit review", async () => {
    const s = security();
    await flush();
    const pending = deferred<Awaited<ReturnType<typeof api.changePassword>>>();
    s.changePassword.mockReturnValue(pending.promise);
    s.credentials();
    s.controller.password();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(s.controller.getSnapshot().operation.kind).toBe("uncertain");
    pending.resolve(
      success({
        user_id: sessionResource().user_id,
        password: { changed_at: new Date().toISOString() },
        sessions_revoked: true,
      }),
    );
    await flush();
    expect(s.event).not.toHaveBeenCalled();
    expect(s.controller.getSnapshot().operation.kind).toBe("uncertain");
    s.credentials();
    s.controller.password();
    expect(s.changePassword).toHaveBeenCalledOnce();
    await s.controller.refreshCredentialState();
    s.controller.review();
    s.controller.password();
    expect(s.changePassword).toHaveBeenCalledTimes(2);
  });
  it("keeps confirmed credential changes separate from failed propagation", async () => {
    const s = security();
    s.event.mockRejectedValueOnce(new Error("read unavailable"));
    s.credentials();
    s.controller.password();
    await flush();
    expect(s.controller.getSnapshot().operation).toMatchObject({
      kind: "confirmed",
      propagation: "failed",
    });
    await s.controller.refresh();
    expect(s.changePassword).toHaveBeenCalledOnce();
    expect(s.controller.getSnapshot().operation).toMatchObject({
      kind: "confirmed",
      propagation: "ready",
    });
  });
  it("retains safe credential resources while showing failed reads and fencing retired reads", async () => {
    const s = security();
    await flush();
    await s.controller.refreshCredentialState();
    expect(s.controller.getSnapshot().credentialState).not.toBeNull();
    s.loadCredentialState.mockRejectedValueOnce(new TypeError());
    await s.controller.refreshCredentialState();
    expect(s.controller.getSnapshot()).toMatchObject({
      credentialRead: "failed",
    });
    expect(s.controller.getSnapshot().credentialState).not.toBeNull();
    s.loadCredentialState.mockResolvedValueOnce(
      success(
        credentialStateResource({
          user_id: "00000000-0000-4000-8000-000000000002",
        }),
      ),
    );
    await s.controller.refreshCredentialState();
    expect(s.controller.getSnapshot().credentialRead).toBe("failed");
    expect(s.controller.getSnapshot().credentialState?.user_id).toBe(
      sessionResource().user_id,
    );
    const pending =
      deferred<Awaited<ReturnType<typeof api.loadCredentialState>>>();
    s.loadCredentialState.mockReturnValue(pending.promise);
    void s.controller.refreshCredentialState();
    s.controller.retire();
    pending.resolve(failure("session_required"));
    await flush();
    expect(s.event).not.toHaveBeenCalled();
    expect(s.controller.getSnapshot().credentialState).toBeNull();
  });
  it("validates owner email and provisioning scalar boundaries without password transformations", () => {
    for (const email of [
      "operator@deployment",
      " café@部署 ",
      `${"a".repeat(318)}@b`,
    ])
      expect(validateAccountEmail(email).error).toBeNull();
    for (const email of [
      "@deployment",
      "a@",
      "a@@b",
      "a b@c",
      "a\u0085b@c",
      `${"a".repeat(319)}@b`,
      "a\ud800@b",
    ])
      expect(validateAccountEmail(email).error).not.toBeNull();
    for (const password of [
      "😀".repeat(12),
      "x".repeat(1024),
      "  Exact e\u0301 Password  ",
    ])
      expect(validateProvisioningPassword(password)).toBeNull();
    for (const password of [
      "😀".repeat(11),
      "x".repeat(1025),
      " ".repeat(12),
      "Password\nWithControl",
      "Password\ud800WithSurrogate",
    ])
      expect(validateProvisioningPassword(password)).not.toBeNull();
  });
});
