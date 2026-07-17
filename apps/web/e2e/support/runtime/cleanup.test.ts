// @vitest-environment node

import { describe, expect, it, vi } from "vitest";
import type { StorageState } from "../../playwrightTypes";
import { buildWorkerAdminBlueprints } from "../auth/workerAdmin";
import { OwnedSessionTracker } from "./cleanup";

function storageState(sessionValue: string, csrfValue = "csrf-token") {
  return {
    cookies: [
      {
        name: "cartulary_session",
        value: sessionValue,
        domain: "127.0.0.1",
        path: "/",
        expires: -1,
        httpOnly: true,
        sameSite: "Lax",
        secure: false,
      },
      {
        name: "cartulary_csrf",
        value: csrfValue,
        domain: "127.0.0.1",
        path: "/",
        expires: -1,
        httpOnly: false,
        sameSite: "Lax",
        secure: false,
      },
    ],
    origins: [],
  } satisfies StorageState;
}

describe("buildWorkerAdminBlueprints", () => {
  it("creates one distinct admin blueprint per parallel index", () => {
    const blueprints = buildWorkerAdminBlueprints(3);

    expect(blueprints).toHaveLength(3);
    expect(blueprints.map((blueprint) => blueprint.parallelIndex)).toEqual([
      0, 1, 2,
    ]);
    expect(new Set(blueprints.map((blueprint) => blueprint.email)).size).toBe(
      3,
    );
    expect(
      new Set(blueprints.map((blueprint) => blueprint.password)).size,
    ).toBe(3);
  });
});

describe("OwnedSessionTracker", () => {
  it("deduplicates repeated registration of the same session cookie", () => {
    const revokeAllSessions = vi.fn(async () => {});
    const verifyRevokedSession = vi.fn(async () => {});
    const tracker = new OwnedSessionTracker({
      label: "dedupe",
      revokeAllSessions,
      verifyRevokedSession,
    });

    tracker.registerSession({
      createdBy: "test",
      email: "user@example.test",
      purpose: "first",
      storageState: storageState("session-a"),
      userId: "user-a",
    });
    tracker.registerSession({
      createdBy: "test",
      email: "user@example.test",
      purpose: "duplicate",
      storageState: storageState("session-a"),
      userId: "user-a",
    });

    expect(tracker.userCount()).toBe(1);
    expect(tracker.sessionCount()).toBe(1);
  });

  it("revokes each owned user once and verifies every tracked snapshot", async () => {
    const revokeAllSessions = vi.fn(async () => {});
    const verifyRevokedSession = vi.fn(async () => {});
    const tracker = new OwnedSessionTracker({
      label: "cleanup",
      revokeAllSessions,
      verifyRevokedSession,
    });

    tracker.registerSession({
      createdBy: "test",
      email: "alpha@example.test",
      purpose: "alpha first",
      storageState: storageState("session-alpha-a"),
      userId: "alpha",
    });
    tracker.registerSession({
      createdBy: "test",
      email: "alpha@example.test",
      purpose: "alpha second",
      storageState: storageState("session-alpha-b"),
      userId: "alpha",
    });
    tracker.registerSession({
      createdBy: "test",
      email: "beta@example.test",
      purpose: "beta only",
      storageState: storageState("session-beta-a"),
      userId: "beta",
    });

    await tracker.cleanup();

    expect(revokeAllSessions).toHaveBeenCalledTimes(2);
    expect(revokeAllSessions).toHaveBeenNthCalledWith(
      1,
      "beta",
      "playwright cleanup: cleanup",
    );
    expect(revokeAllSessions).toHaveBeenNthCalledWith(
      2,
      "alpha",
      "playwright cleanup: cleanup",
    );
    expect(verifyRevokedSession).toHaveBeenCalledTimes(3);
  });

  it("creates alias contexts without registering additional sessions", async () => {
    const tracker = new OwnedSessionTracker({
      label: "alias",
      revokeAllSessions: async () => {},
      verifyRevokedSession: async () => {},
    });
    const fakeContext = { close: vi.fn() };
    const browser = {
      newContext: vi.fn(async () => fakeContext),
    };

    const context = await tracker.newTrackedContext(
      browser,
      storageState("session-alias"),
    );

    expect(context).toBe(fakeContext);
    expect(browser.newContext).toHaveBeenCalledWith({
      storageState: storageState("session-alias"),
    });
    expect(tracker.sessionCount()).toBe(0);
  });

  it("fails cleanup when revocation verification fails", async () => {
    const tracker = new OwnedSessionTracker({
      label: "failure",
      revokeAllSessions: async () => {},
      verifyRevokedSession: async () => {
        throw new Error("still usable");
      },
    });
    tracker.registerSession({
      createdBy: "test",
      email: "user@example.test",
      purpose: "verification",
      storageState: storageState("session-failure"),
      userId: "user-a",
    });

    await expect(tracker.cleanup()).rejects.toThrow("session.verify_revoked");
  });

  it("aggregates failures without exposing identities, cookies, or dependency errors", async () => {
    const tracker = new OwnedSessionTracker({
      label: "failure",
      revokeAllSessions: async (userId) => {
        if (userId === "user-secret") {
          throw new Error("token=control-secret email=user@example.test");
        }
      },
      verifyRevokedSession: async () => {
        throw new Error("session=session-secret");
      },
    });
    tracker.registerSession({
      createdBy: "credential-file=/tmp/secret",
      email: "user@example.test",
      purpose: "session-secret",
      storageState: storageState("session-secret"),
      userId: "user-secret",
    });
    tracker.registerSession({
      createdBy: "test",
      email: "other@example.test",
      purpose: "verification",
      storageState: storageState("other-session-secret"),
      userId: "other-user-secret",
    });

    const error = await tracker.cleanup().catch((caught: unknown) => caught);
    const message = error instanceof Error ? error.message : String(error);

    expect(message).toBe(
      "Owned session cleanup failed (2 operations): session.verify_revoked, session.revoke_all",
    );
    expect(message).not.toMatch(
      /control-secret|user@example\.test|other@example\.test|session-secret|user-secret|\/tmp\/secret/,
    );
  });

  it("consumes cleanup after failure so repeated invocation is a no-op", async () => {
    const revokeAllSessions = vi.fn(async () => {
      throw new Error("failure");
    });
    const tracker = new OwnedSessionTracker({
      label: "idempotent",
      revokeAllSessions,
      verifyRevokedSession: async () => {},
    });
    tracker.registerSession({
      createdBy: "test",
      email: "user@example.test",
      purpose: "idempotent",
      storageState: storageState("session-idempotent"),
      userId: "user-a",
    });

    await expect(tracker.cleanup()).rejects.toThrow(/session\.revoke_all/);
    await expect(tracker.cleanup()).resolves.toBeUndefined();
    expect(revokeAllSessions).toHaveBeenCalledOnce();
  });

  it("does not register or revoke borrowed sessions", async () => {
    const revokeAllSessions = vi.fn(async () => {});
    const tracker = new OwnedSessionTracker({
      label: "borrowed",
      revokeAllSessions,
      verifyRevokedSession: async () => {},
    });
    tracker.registerSession(
      {
        createdBy: "external-server",
        email: "borrowed@example.test",
        purpose: "borrowed",
        storageState: storageState("session-borrowed"),
        userId: "borrowed-user",
      },
      "borrowed",
    );

    await tracker.cleanup();

    expect(tracker.sessionCount()).toBe(0);
    expect(revokeAllSessions).not.toHaveBeenCalled();
  });
});
