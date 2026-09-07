import type { GetCurrentSessionResponse } from "@cartulary/protocol-ts/http";
import { act, cleanup, renderHook } from "@testing-library/react";
import { StrictMode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { HTTPOperationResult } from "../services/browserApi";
import { sessionResource } from "../testing/appShellTestSupport";
import { deferred } from "../testing/fetchMockTestSupport";
import { useWorkbookAuthorizationState } from "../workbook/hooks/useWorkbookAuthorizationState";
import { AppSessionController } from "./appSessionController";
import { useAppSession } from "./useAppSession";

const accountA = "00000000-0000-4000-8000-000000000001";
const accountB = "00000000-0000-4000-8000-000000000002";
const incident = "00000000-0000-4000-8000-000000000003";
const timestamp = "2026-09-05T12:00:00Z";
type SessionResult = HTTPOperationResult<GetCurrentSessionResponse>;
const controllers: AppSessionController[] = [];
function success(userId = accountA): SessionResult {
  return {
    ok: true,
    status: 200,
    payload: {
      meta: { request_id: "session" },
      data: sessionResource({
        user_id: userId,
        memberships: [{ incident_id: incident, role: "editor" }],
      }),
    },
  };
}
function setup(
  ports: ConstructorParameters<typeof AppSessionController>[0] = {},
) {
  const retireLifetime = vi.fn();
  const replaceAccount = vi.fn();
  const controller = new AppSessionController({
    session: async () => success(),
    preferences: async () => ({
      ok: true,
      status: 200,
      payload: {
        meta: { request_id: "preferences" },
        data: {
          user_id: accountA,
          density_mode: "compact",
          preferences_version: 1,
          created_at: timestamp,
          updated_at: timestamp,
        },
      },
    }),
    extensions: async () => ({
      ok: true,
      status: 200,
      payload: { meta: { request_id: "extensions" }, data: { extensions: [] } },
    }),
    retireLifetime,
    replaceAccount,
    ...ports,
  });
  controllers.push(controller);
  return { controller, retireLifetime, replaceAccount };
}
async function flush() {
  for (let i = 0; i < 8; ++i) await Promise.resolve();
}
afterEach(() => {
  cleanup();
  for (const controller of controllers.splice(0)) controller.dispose();
  vi.useRealTimers();
});

describe("application session lifecycle", () => {
  it("keeps anonymous authentication confirmation inside its current flow and fences replacement", async () => {
    const anonymous: SessionResult = {
      ok: false,
      status: 401,
      payload: { error: { code: "session_required" } },
    };
    const session = vi.fn().mockResolvedValue(anonymous);
    const { controller, retireLifetime } = setup({ session });
    await controller.refreshSession();
    const revision = controller.getSnapshot().revision;
    const retirements = retireLifetime.mock.calls.length;
    expect(
      await controller.confirmAuthentication(
        revision,
        new AbortController().signal,
        () => true,
      ),
    ).toEqual({ kind: "session_lost" });
    expect(controller.getSnapshot().revision).toBe(revision);
    expect(retireLifetime).toHaveBeenCalledTimes(retirements);
    const late = deferred<SessionResult>();
    session.mockReturnValueOnce(late.promise);
    const confirmation = controller.confirmAuthentication(
      revision,
      new AbortController().signal,
      () => true,
    );
    controller.credentialsRevoked();
    late.resolve(success());
    expect(await confirmation).toEqual({ kind: "cancelled" });
    expect(controller.getSnapshot().session).toBeNull();
  });

  it("distinguishes unavailable discovery from confirmed anonymous and permits manual retry", async () => {
    const session = vi
      .fn()
      .mockRejectedValueOnce(new TypeError("offline"))
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        payload: { error: { code: "session_required" } },
      })
      .mockResolvedValue(success());
    const { controller } = setup({ session });
    await controller.refreshSession();
    expect(controller.getSnapshot()).toMatchObject({
      state: "unavailable",
      session: null,
      observing: false,
    });
    await controller.refreshSession();
    expect(controller.getSnapshot()).toMatchObject({
      state: "anonymous",
      session: null,
      ended: false,
    });
    await controller.refreshSession();
    expect(controller.getSnapshot().state).toBe("authenticated");
  });

  it("admits a validated session while preferences and extension discovery fail independently", async () => {
    const preferences = vi
      .fn()
      .mockRejectedValueOnce(new Error("unavailable"))
      .mockResolvedValue({
        ok: true,
        status: 200,
        payload: {
          data: {
            user_id: accountA,
            density_mode: null,
            preferences_version: 1,
            created_at: timestamp,
            updated_at: timestamp,
          },
          meta: { request_id: "preferences" },
        },
      });
    const extensions = vi
      .fn()
      .mockResolvedValueOnce({
        ok: false,
        status: 502,
        payload: { error: { code: "invalid_public_contract_response" } },
      })
      .mockResolvedValue({
        ok: true,
        status: 200,
        payload: {
          data: { extensions: [] },
          meta: { request_id: "extensions" },
        },
      });
    const { controller } = setup({ preferences, extensions });
    await controller.refreshSession();
    await flush();
    expect(controller.getSnapshot()).toMatchObject({
      state: "authenticated",
      preferences: { kind: "failed", failure: "transient" },
      extensions: { kind: "failed", failure: "contract" },
    });
    expect(controller.getSnapshot().extensions).not.toHaveProperty("value");
    await controller.refreshPreferences();
    expect(controller.getSnapshot().preferences.kind).toBe("ready");
    expect(controller.getSnapshot().extensions.kind).toBe("failed");
    await controller.refreshExtensions();
    expect(controller.getSnapshot().extensions).toEqual({
      kind: "ready",
      value: [],
    });
  });

  it("settles session and supporting observations at thirty seconds despite ignored aborts", async () => {
    vi.useFakeTimers();
    const late = deferred<SessionResult>();
    const { controller } = setup({
      session: () => late.promise,
      preferences: () => new Promise(() => {}),
      extensions: () => new Promise(() => {}),
    });
    const read = controller.refreshSession();
    await vi.advanceTimersByTimeAsync(29_999);
    expect(controller.getSnapshot().observing).toBe(true);
    await vi.advanceTimersByTimeAsync(1);
    await read;
    expect(controller.getSnapshot().state).toBe("unavailable");
    late.resolve(success());
    await flush();
    expect(controller.getSnapshot().session).toBeNull();
    const accepted = success();
    if (!accepted.ok) throw new Error("fixture");
    controller.authenticationCompleted(
      accepted.payload.data,
      controller.getSnapshot().revision,
    );
    await vi.advanceTimersByTimeAsync(30_000);
    expect(controller.getSnapshot()).toMatchObject({
      state: "authenticated",
      preferences: { kind: "failed" },
      extensions: { kind: "failed" },
    });
  });

  it("rejects obsolete session successes and failures from overlapping observations", async () => {
    for (const staleFailure of [false, true]) {
      const older = deferred<SessionResult>();
      const newer = deferred<SessionResult>();
      const { controller } = setup({
        session: vi
          .fn()
          .mockReturnValueOnce(older.promise)
          .mockReturnValueOnce(newer.promise),
      });
      const first = controller.refreshSession();
      await flush();
      const second = controller.refreshSession();
      await flush();
      newer.resolve(success(accountB));
      await second;
      if (staleFailure) older.reject(new Error("obsolete failure"));
      else older.resolve(success(accountA));
      await first;
      await flush();
      expect(controller.getSnapshot()).toMatchObject({
        state: "authenticated",
        session: { user_id: accountB },
        error: null,
      });
    }
  });

  it("creates a new lifetime for identical-field reauthentication but preserves ordinary refresh identity", async () => {
    const { controller, retireLifetime } = setup();
    await controller.refreshSession();
    const before = controller.getSnapshot();
    await controller.refreshSession();
    expect(controller.getSnapshot().lifetime).toBe(before.lifetime);
    expect(retireLifetime).toHaveBeenCalledTimes(1);
    if (before.session === null) throw new Error("fixture");
    expect(
      controller.authenticationCompleted(before.session, before.revision),
    ).toBe(true);
    expect(controller.getSnapshot().lifetime).not.toBe(before.lifetime);
    expect(retireLifetime).toHaveBeenCalledTimes(2);
    expect(
      controller.authenticationCompleted(before.session, before.revision),
    ).toBe(false);
  });

  it("preserves same-account work across revocation and clears another account before publication", async () => {
    const events: string[] = [];
    const { controller, replaceAccount } = setup({
      retireLifetime: () => events.push("retire"),
    });
    controller.subscribe(() => {
      if (controller.getSnapshot().session?.user_id === accountB)
        events.push("publish-b");
    });
    await controller.refreshSession();
    controller.credentialsRevoked();
    await controller.refreshSession();
    expect(replaceAccount).not.toHaveBeenCalled();
    controller.logoutConfirmed();
    replaceAccount.mockImplementation(() => events.push("clear-a"));
    const other = success(accountB);
    if (!other.ok) throw new Error("fixture");
    controller.authenticationCompleted(
      other.payload.data,
      controller.getSnapshot().revision,
    );
    expect(replaceAccount).toHaveBeenCalledOnce();
    expect(events.indexOf("clear-a")).toBeLessThan(events.indexOf("publish-b"));
    expect(events.at(events.indexOf("clear-a") - 1)).toBe("retire");
  });

  it("classifies recovery without treating transport or contract failure as membership loss", async () => {
    const session = vi.fn().mockResolvedValue(success());
    const { controller } = setup({ session });
    await controller.refreshSession();
    const port = controller.recoveryPort();
    const recover = () =>
      port.recover({
        incidentId: incident,
        signal: new AbortController().signal,
      });
    session.mockRejectedValueOnce(new TypeError("offline"));
    expect(await recover()).toEqual({
      kind: "unavailable",
      failure: "transient",
    });
    expect(controller.getSnapshot()).toMatchObject({
      error: null,
      observing: false,
      state: "authenticated",
    });
    session.mockResolvedValueOnce({
      ok: false,
      status: 502,
      payload: { error: { code: "invalid_public_contract_response" } },
    });
    expect(await recover()).toEqual({
      kind: "unavailable",
      failure: "contract",
    });
    const absent = success();
    if (!absent.ok) throw new Error("fixture");
    session.mockResolvedValueOnce({
      ...absent,
      payload: {
        ...absent.payload,
        data: { ...absent.payload.data, memberships: [] },
      },
    });
    expect(await recover()).toEqual({ kind: "access_lost" });
    session.mockResolvedValueOnce({
      ok: false,
      status: 401,
      payload: { error: { code: "session_required" } },
    });
    expect(await recover()).toEqual({ kind: "session_lost" });
    expect(controller.getSnapshot().state).toBe("anonymous");
  });

  it("cancels recovery before publication after caller cancellation navigation replacement and disposal", async () => {
    for (const reason of [
      "abort",
      "navigation",
      "replacement",
      "dispose",
    ] as const) {
      for (const reject of [false, true]) {
        const late = deferred<SessionResult>();
        const session = vi
          .fn()
          .mockResolvedValueOnce(success())
          .mockReturnValueOnce(late.promise);
        const { controller } = setup({ session });
        await controller.refreshSession();
        const caller = new AbortController();
        const recovering = controller
          .recoveryPort()
          .recover({ incidentId: incident, signal: caller.signal });
        await flush();
        if (reason === "abort") caller.abort();
        if (reason === "navigation") controller.navigationChanged();
        if (reason === "dispose") controller.dispose();
        if (reason === "replacement") {
          const snapshot = controller.getSnapshot();
          if (snapshot.session)
            controller.authenticationCompleted(
              snapshot.session,
              snapshot.revision,
            );
        }
        const snapshot = controller.getSnapshot();
        if (reject) late.reject(new Error("obsolete"));
        else late.resolve(success(accountB));
        expect(await recovering).toEqual({ kind: "cancelled" });
        expect(controller.getSnapshot().session).toEqual(snapshot.session);
        expect(controller.getSnapshot().error).toEqual(snapshot.error);
      }
    }
  });

  it("settles overlapping recovery callers without publishing the obsolete account", async () => {
    const older = deferred<SessionResult>();
    const session = vi
      .fn()
      .mockResolvedValueOnce(success())
      .mockReturnValueOnce(older.promise)
      .mockResolvedValueOnce(success());
    const { controller } = setup({ session });
    await controller.refreshSession();
    const old = controller
      .recoveryPort()
      .recover({ incidentId: incident, signal: new AbortController().signal });
    await flush();
    expect(
      await controller.recoveryPort().recover({
        incidentId: incident,
        signal: new AbortController().signal,
      }),
    ).toMatchObject({ kind: "authorized" });
    older.resolve(success(accountB));
    expect(await old).toEqual({ kind: "cancelled" });
    expect(controller.getSnapshot().session?.user_id).toBe(accountA);
  });

  it("survives React lifecycle replay and disposes pending observations after unmount", async () => {
    const { controller } = setup();
    const dispose = vi.fn();
    const rendered = renderHook(() => useAppSession(controller, dispose), {
      wrapper: ({ children }) => <StrictMode>{children}</StrictMode>,
    });
    await act(flush);
    expect(rendered.result.current.state).toBe("authenticated");
    expect(dispose).not.toHaveBeenCalled();
    rendered.unmount();
    await flush();
    expect(dispose).toHaveBeenCalledOnce();
    const result = success(accountB);
    if (!result.ok) throw new Error("fixture");
    expect(
      controller.authenticationCompleted(
        result.payload.data,
        controller.getSnapshot().revision,
      ),
    ).toBe(false);
  });

  it("keeps workbook authorization through temporary failure and rejects obsolete hook callbacks", async () => {
    const lost = vi.fn();
    const sessionLost = vi.fn();
    const late = deferred<{ kind: "access_lost" }>();
    const recover = vi
      .fn()
      .mockResolvedValueOnce({ kind: "unavailable", failure: "transient" })
      .mockReturnValueOnce(late.promise);
    const rendered = renderHook(() =>
      useWorkbookAuthorizationState({
        accountUserId: accountA,
        authorizationRecovery: { recover },
        incidentId: incident,
        onIncidentAccessLost: lost,
        onSessionLost: sessionLost,
      }),
    );
    await act(() => rendered.result.current.loadSessionRole());
    expect(lost).not.toHaveBeenCalled();
    expect(rendered.result.current.currentUserId).toBe(accountA);
    const pending = rendered.result.current.loadSessionRole();
    rendered.unmount();
    late.resolve({ kind: "access_lost" });
    await pending;
    expect(lost).not.toHaveBeenCalled();
    expect(sessionLost).not.toHaveBeenCalled();
  });
});
