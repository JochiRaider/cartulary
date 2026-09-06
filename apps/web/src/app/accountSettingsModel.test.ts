import { afterEach, describe, expect, it, vi } from "vitest";
import { sessionResource } from "../testing/appShellTestSupport";
import { deferred } from "../testing/fetchMockTestSupport";
import {
  AccountSettingsController,
  accountReviewRequired,
  canSaveAccountEdit,
  validateAccountDisplayName,
} from "./accountSettingsModel";
import type {
  loadAccountProfile,
  patchAccountProfile,
  putAccountPreferences,
} from "./api/authAccountClient";
import { AppSessionController } from "./appSessionController";

const actor = "00000000-0000-4000-8000-000000000001";
const otherActor = "00000000-0000-4000-8000-000000000002";
const timestamp = "2026-09-05T12:00:00Z";
const profile = (display_name = "Operator", user_version = 1) => ({
  user_id: actor,
  display_name,
  user_version,
  email: "operator@example.test",
  created_at: timestamp,
  updated_at: timestamp,
});
const preferences = (
  density_mode: "compact" | "default" | "comfortable" | null = null,
  preferences_version = 1,
) => ({
  user_id: actor,
  density_mode,
  preferences_version,
  created_at: timestamp,
  updated_at: timestamp,
});
const success = <T>(data: T) => ({
  ok: true as const,
  status: 200,
  payload: { data, meta: { request_id: "settings-model" } },
});
const rejection = (code: string, status = 409) => ({
  ok: false as const,
  status,
  payload: { error: { code } },
});
const disposals: (() => void)[] = [];
afterEach(() => {
  for (const dispose of disposals.splice(0)) dispose();
  vi.useRealTimers();
});
async function flush() {
  for (let i = 0; i < 30; ++i) await Promise.resolve();
}
async function setup() {
  let serverName = "Operator";
  const read = vi
    .fn<typeof loadAccountProfile>()
    .mockResolvedValue(success(profile()));
  const patch = vi
    .fn<typeof patchAccountProfile>()
    .mockImplementation(async (request) => {
      serverName = request.displayName;
      return success(profile(serverName, request.baseUserVersion + 1));
    });
  const put = vi
    .fn<typeof putAccountPreferences>()
    .mockImplementation(async (request) =>
      success(
        preferences(request.densityMode, request.basePreferencesVersion + 1),
      ),
    );
  const sessionRead = vi.fn(async () =>
    success(sessionResource({ user_id: actor, display_name: serverName })),
  );
  const preferencesRead = vi.fn(async () => success(preferences()));
  const session = new AppSessionController({
    session: sessionRead,
    preferences: preferencesRead,
    extensions: async () => success({ extensions: [] }),
  });
  await session.refreshSession();
  await flush();
  let txn = 0;
  const controller = new AccountSettingsController(session, {
    readProfile: read,
    patchProfile: patch,
    putPreferences: put,
    newTransactionId: () => `settings-${++txn}`,
  });
  controller.start();
  await controller.refresh("profile");
  disposals.push(() => {
    controller.dispose();
    session.dispose();
  });
  return {
    controller,
    session,
    read,
    patch,
    put,
    sessionRead,
    preferencesRead,
  };
}

describe("account edit lifecycle", () => {
  it("retries initial load and preserves usable data and dirty drafts through refresh failure", async () => {
    const { controller, read, session, preferencesRead } = await setup();
    controller.retireLifetime();
    controller.start();
    // A new lifetime reloads protected resources.
    preferencesRead.mockRejectedValueOnce(
      new TypeError("initial preferences unavailable"),
    );
    session.authenticationCompleted(
      sessionResource({ user_id: actor }),
      session.getSnapshot().revision,
    );
    await flush();
    expect(controller.getSnapshot().appearance).toMatchObject({
      saved: null,
      read: "failed",
    });
    await controller.refresh("appearance");
    expect(controller.getSnapshot().appearance).toMatchObject({
      saved: preferences(),
      read: "ready",
    });
    read.mockResolvedValueOnce(rejection("unavailable", 503));
    await controller.refresh("profile");
    expect(controller.getSnapshot().profile).toMatchObject({
      saved: null,
      read: "failed",
    });
    await controller.refresh("profile");
    controller.change("profile", "Local draft");
    const lifetime = session.getSnapshot().lifetime;
    await session.refreshSession();
    expect(session.getSnapshot().lifetime).toBe(lifetime);
    expect(controller.getSnapshot().profile.draft).toBe("Local draft");
    read.mockResolvedValueOnce(success(profile("External", 2)));
    await controller.refresh("profile");
    expect(controller.getSnapshot().profile.draft).toBe("Local draft");
    expect(accountReviewRequired(controller.getSnapshot().profile)).toBe(true);
    read.mockRejectedValueOnce(new TypeError("offline"));
    await controller.refresh("profile");
    expect(controller.getSnapshot().profile).toMatchObject({
      saved: { user_version: 2 },
      draft: "Local draft",
      read: "failed",
    });
    controller.change("appearance", "comfortable");
    preferencesRead.mockResolvedValueOnce(success(preferences("compact", 2)));
    await controller.refresh("appearance");
    expect(controller.getSnapshot().appearance).toMatchObject({
      draft: "comfortable",
      baseVersion: 1,
      saved: preferences("compact", 2),
    });
    expect(accountReviewRequired(controller.getSnapshot().appearance)).toBe(
      true,
    );
    preferencesRead.mockRejectedValueOnce(new TypeError("offline"));
    await controller.refresh("appearance");
    expect(session.getSnapshot().preferences).toMatchObject({
      kind: "ready",
      value: { density_mode: "compact" },
      refreshError: {},
    });
    expect(controller.getSnapshot().appearance).toMatchObject({
      draft: "comfortable",
      saved: { density_mode: "compact" },
      read: "failed",
    });
  });
  it("locks same-tick submissions independently and acknowledges only the captured revision", async () => {
    const { controller, patch, put } = await setup();
    const pending = deferred<Awaited<ReturnType<typeof patchAccountProfile>>>();
    const appearancePending =
      deferred<Awaited<ReturnType<typeof putAccountPreferences>>>();
    patch.mockReturnValueOnce(pending.promise);
    put.mockReturnValueOnce(appearancePending.promise);
    controller.change("profile", "First");
    controller.submit("profile");
    controller.submit("profile");
    controller.change("profile", "Later");
    controller.change("appearance", "compact");
    controller.submit("appearance");
    controller.submit("appearance");
    controller.change("appearance", "comfortable");
    expect(patch).toHaveBeenCalledTimes(1);
    expect(put).toHaveBeenCalledTimes(1);
    pending.resolve(success(profile("First", 2)));
    appearancePending.resolve(success(preferences("compact", 2)));
    await flush();
    expect(controller.getSnapshot().profile).toMatchObject({
      saved: { display_name: "First" },
      draft: "Later",
      operation: { kind: "confirmed" },
    });
    expect(controller.getSnapshot().appearance.operation.kind).toBe(
      "confirmed",
    );
    expect(controller.getSnapshot().appearance).toMatchObject({
      saved: preferences("compact", 2),
      draft: "comfortable",
    });
  });
  it("requires explicit review for both version conflicts and uses a fresh attempt after review", async () => {
    for (const kind of ["profile", "appearance"] as const) {
      const { controller, patch, put, read, preferencesRead } = await setup();
      if (kind === "profile") {
        patch.mockResolvedValueOnce(rejection("user_version_conflict"));
        read.mockResolvedValue(success(profile("External", 2)));
      } else {
        put.mockResolvedValueOnce(rejection("preferences_version_conflict"));
        preferencesRead.mockResolvedValue(success(preferences("compact", 2)));
      }
      const value = kind === "profile" ? "Local" : "comfortable";
      controller.change(kind, value);
      controller.submit(kind);
      await flush();
      const write = kind === "profile" ? patch : put;
      expect(controller.getSnapshot()[kind]).toMatchObject({
        draft: value,
        operation: { kind: "conflict" },
      });
      controller.submit(kind);
      expect(write).toHaveBeenCalledTimes(1);
      controller.review(kind);
      controller.submit(kind);
      await flush();
      expect(write).toHaveBeenCalledTimes(2);
      expect(write.mock.calls[1]?.[0]).toMatchObject({
        clientTxnId: "settings-2",
        [kind === "profile" ? "baseUserVersion" : "basePreferencesVersion"]: 2,
      });
    }
  });
  it("keeps transaction conflict distinct from version conflict", async () => {
    const { controller, patch } = await setup();
    patch.mockResolvedValueOnce(rejection("client_txn_conflict"));
    controller.change("profile", "Local");
    controller.submit("profile");
    await flush();
    expect(controller.getSnapshot().profile.operation).toEqual({
      kind: "rejected",
      reason: "transaction",
    });
    controller.submit("profile");
    expect(patch).toHaveBeenCalledTimes(1);
    controller.review("profile");
    controller.submit("profile");
    await flush();
    expect(patch.mock.calls[1]?.[0].clientTxnId).toBe("settings-2");
  });
  it("replays an uncertain request exactly while retaining newer edits and explicit discard", async () => {
    for (const kind of ["profile", "appearance"] as const) {
      const { controller, patch, put } = await setup();
      const first = kind === "profile" ? "First" : "compact";
      const later = kind === "profile" ? "Later" : "comfortable";
      if (kind === "profile")
        patch.mockRejectedValueOnce(new TypeError("response lost"));
      else put.mockRejectedValueOnce(new TypeError("response lost"));
      controller.change(kind, first);
      controller.submit(kind);
      await flush();
      controller.change(kind, later);
      controller.submit(kind);
      const write = kind === "profile" ? patch : put;
      expect(write).toHaveBeenCalledTimes(1);
      controller.discard(kind);
      expect(controller.getSnapshot()[kind].operation.kind).toBe("uncertain");
      controller.change(kind, later);
      controller.replay(kind);
      controller.replay(kind);
      await flush();
      const { signal: _firstSignal, ...original } =
        write.mock.calls[0]?.[0] ?? {};
      const { signal: _replaySignal, ...replayed } =
        write.mock.calls[1]?.[0] ?? {};
      expect(replayed).toEqual(original);
      expect(controller.getSnapshot()[kind].draft).toBe(later);
    }
  });
  it("treats timeout and malformed success as uncertain and excludes late ignored-abort completion", async () => {
    vi.useFakeTimers();
    const { controller, patch, session } = await setup();
    const pending = deferred<Awaited<ReturnType<typeof patchAccountProfile>>>();
    patch.mockReturnValueOnce(pending.promise);
    controller.change("profile", "First");
    controller.submit("profile");
    await vi.advanceTimersByTimeAsync(30_000);
    expect(controller.getSnapshot().profile.operation).toMatchObject({
      kind: "uncertain",
      reason: "timeout",
    });
    pending.resolve(success(profile("First", 2)));
    await flush();
    expect(controller.getSnapshot().profile.saved).toEqual(profile());
    patch.mockResolvedValueOnce(
      success({ ...profile("First", 2), user_id: otherActor }),
    );
    controller.replay("profile");
    await flush();
    expect(controller.getSnapshot().profile.operation).toMatchObject({
      kind: "uncertain",
      reason: "contract",
    });
    expect(session.getSnapshot().session?.display_name).toBe("Operator");
  });
  it("does not regress accepted versions from old reads or committed replays", async () => {
    const { controller, read, patch, session, put, preferencesRead } =
      await setup();
    const oldRead = deferred<Awaited<ReturnType<typeof loadAccountProfile>>>();
    read
      .mockReturnValueOnce(oldRead.promise)
      .mockResolvedValueOnce(success(profile("Latest", 5)));
    const firstRead = controller.refresh("profile");
    await controller.refresh("profile");
    oldRead.resolve(success(profile("Old", 2)));
    await firstRead;
    expect(controller.getSnapshot().profile.saved).toEqual(
      profile("Latest", 5),
    );
    patch.mockRejectedValueOnce(new TypeError("lost"));
    controller.change("profile", "Own");
    controller.submit("profile");
    await flush();
    read.mockResolvedValueOnce(success(profile("Newer", 9)));
    await controller.refresh("profile");
    patch.mockResolvedValueOnce(success(profile("Own", 6)));
    controller.replay("profile");
    await flush();
    expect(controller.getSnapshot().profile.saved).toEqual(profile("Newer", 9));
    controller.change("appearance", "compact");
    put.mockRejectedValueOnce(new TypeError("lost"));
    controller.submit("appearance");
    await flush();
    session.preferencesChanged(
      preferences("comfortable", 4),
      session.getSnapshot().lifetime,
    );
    const freshRead = deferred<Awaited<ReturnType<typeof preferencesRead>>>();
    preferencesRead.mockReturnValueOnce(freshRead.promise);
    const refreshing = controller.refresh("appearance");
    put.mockResolvedValueOnce(success(preferences("compact", 2)));
    controller.replay("appearance");
    await flush();
    expect(session.getSnapshot().preferences).toMatchObject({
      kind: "ready",
      value: preferences("comfortable", 4),
    });
    expect(controller.getSnapshot().appearance.saved).toEqual(
      preferences("comfortable", 4),
    );
    expect(controller.getSnapshot().appearance.read).toBe("refreshing");
    freshRead.resolve(success(preferences("default", 5)));
    await refreshing;
    expect(controller.getSnapshot().appearance.saved).toEqual(
      preferences("default", 5),
    );
    expect(
      session.preferencesChanged(
        {
          ...preferences("default", 5),
          updated_at: "2026-09-01T12:00:00Z",
        },
        session.getSnapshot().lifetime,
      ),
    ).toBe("invalid");
    expect(controller.getSnapshot().appearance.saved).toEqual(
      preferences("default", 5),
    );
  });
  it("clears work on logout account change and identical-field reauthentication before obsolete completion", async () => {
    for (const replacement of [
      "logout",
      "account",
      "reauthentication",
    ] as const) {
      const { controller, session, patch, sessionRead } = await setup();
      const pending =
        deferred<Awaited<ReturnType<typeof patchAccountProfile>>>();
      patch.mockReturnValueOnce(pending.promise);
      controller.change("profile", "Private draft");
      controller.submit("profile");
      const stale = controller.bind("profile", session.getSnapshot().lifetime);
      const reads = sessionRead.mock.calls.length;
      if (replacement === "logout") session.logoutConfirmed();
      else
        session.authenticationCompleted(
          sessionResource({
            user_id: replacement === "account" ? otherActor : actor,
          }),
          session.getSnapshot().revision,
        );
      pending.resolve(success(profile("Private draft", 2)));
      await flush();
      expect(controller.getSnapshot().profile).toMatchObject({
        draft: null,
        saved: null,
        operation: { kind: "idle" },
      });
      expect(sessionRead).toHaveBeenCalledTimes(reads);
      stale.change("Obsolete callback");
      stale.submit();
      stale.refresh();
      expect(controller.getSnapshot().profile.draft).toBeNull();
      expect(patch).toHaveBeenCalledTimes(1);
    }
  });
  it("retries failed post-save session publication without repeating the mutation", async () => {
    const { controller, patch, sessionRead } = await setup();
    sessionRead.mockRejectedValueOnce(new TypeError("session unavailable"));
    controller.change("profile", "Saved name");
    controller.submit("profile");
    await flush();
    expect(controller.getSnapshot().profile).toMatchObject({
      saved: { display_name: "Saved name" },
      operation: { kind: "confirmed", publication: "failed" },
    });
    sessionRead.mockResolvedValueOnce(
      success(
        sessionResource({
          user_id: actor,
          display_name: "A later external name",
        }),
      ),
    );
    controller.retryPublication("profile");
    controller.retryPublication("profile");
    await flush();
    expect(patch).toHaveBeenCalledTimes(1);
    expect(controller.getSnapshot().profile.operation).toMatchObject({
      kind: "confirmed",
      publication: "ready",
    });
  });
  it("validates scalar display names and preserves rejected drafts", async () => {
    expect(validateAccountDisplayName("  e\u0301  Analyst  ")).toEqual({
      value: "é  Analyst",
      error: null,
    });
    expect(validateAccountDisplayName("😀".repeat(256)).error).toBeNull();
    for (const value of ["\u0085", "a\u0001", "x".repeat(257), "\ud800"])
      expect(validateAccountDisplayName(value).error).not.toBeNull();
    const { controller, patch } = await setup();
    controller.change("profile", " ");
    controller.submit("profile");
    expect(controller.getSnapshot().profile).toMatchObject({
      draft: " ",
      fieldError: "Enter a display name.",
    });
    expect(patch).not.toHaveBeenCalled();
    patch.mockResolvedValueOnce(rejection("invalid_mutation_payload", 400));
    controller.change("profile", "Valid name");
    controller.submit("profile");
    await flush();
    expect(controller.getSnapshot().profile).toMatchObject({
      draft: "Valid name",
      operation: { kind: "rejected", reason: "validation" },
    });
  });
  it("publishes null and each exact density mode only after confirmation", async () => {
    const { controller, session, put, preferencesRead } = await setup();
    for (const value of ["compact", "default", "comfortable", null] as const) {
      const pending =
        deferred<Awaited<ReturnType<typeof putAccountPreferences>>>();
      put.mockReturnValueOnce(pending.promise);
      const previous = session.getSnapshot().preferences;
      controller.change("appearance", value);
      controller.submit("appearance");
      expect(session.getSnapshot().preferences).toBe(previous);
      const version = controller.getSnapshot().appearance.baseVersion ?? 1;
      pending.resolve(success(preferences(value, version + 1)));
      await flush();
      expect(session.getSnapshot().preferences).toMatchObject({
        kind: "ready",
        value: { density_mode: value },
      });
      expect(canSaveAccountEdit(controller.getSnapshot().appearance)).toBe(
        false,
      );
    }
    preferencesRead.mockRejectedValueOnce(
      new TypeError("post-save refresh unavailable"),
    );
    await controller.refresh("appearance");
    expect(controller.getSnapshot().appearance).toMatchObject({
      operation: { kind: "confirmed", publication: "ready" },
      saved: preferences(null, 5),
      read: "failed",
    });
    expect(session.getSnapshot().preferences).toMatchObject({
      kind: "ready",
      value: preferences(null, 5),
      refreshError: {},
    });
    preferencesRead.mockResolvedValueOnce(success(preferences(null, 5)));
    await controller.refresh("appearance");
    expect(put).toHaveBeenCalledTimes(4);
    put.mockResolvedValueOnce(success(preferences("compact", 0)));
    controller.change("appearance", "compact");
    controller.submit("appearance");
    await flush();
    expect(controller.getSnapshot().appearance.operation).toMatchObject({
      kind: "uncertain",
      reason: "contract",
    });
    expect(session.getSnapshot().preferences).toMatchObject({
      kind: "ready",
      value: preferences(null, 5),
    });
  });
  it("distinguishes authorization denial from session loss and preserves uncertainty on replay denial", async () => {
    const { controller, session, patch } = await setup();
    patch.mockResolvedValueOnce(rejection("authorization_denied", 403));
    controller.change("profile", "Local");
    controller.submit("profile");
    await flush();
    expect(controller.getSnapshot().profile.operation).toEqual({
      kind: "rejected",
      reason: "authorization",
    });
    expect(session.getSnapshot().state).toBe("authenticated");
    await controller.refresh("profile");
    controller.review("profile");
    patch.mockRejectedValueOnce(new TypeError("lost"));
    controller.submit("profile");
    await flush();
    patch.mockResolvedValueOnce(rejection("authorization_denied", 403));
    controller.replay("profile");
    await flush();
    expect(controller.getSnapshot().profile.operation).toMatchObject({
      kind: "uncertain",
      reason: "authorization",
    });
    patch.mockResolvedValueOnce(rejection("session_required", 401));
    controller.replay("profile");
    await flush();
    expect(session.getSnapshot().state).toBe("anonymous");
    expect(controller.getSnapshot().profile.saved).toBeNull();
  });
});
