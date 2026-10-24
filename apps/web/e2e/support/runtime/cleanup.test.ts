// @vitest-environment node
import type {
  APIRequestContext,
  BrowserContext,
  Page,
  Route,
} from "@playwright/test";
import { describe, expect, it, vi } from "vitest";
import type { StorageState } from "../auth/storageState";
import { buildWorkerAdminBlueprints } from "../auth/workerAdmin";
import { OwnedJobs } from "../jobs/ownership";
import {
  holdBrowserRequest,
  installOwnedRoute,
  installPublicErrorResponse,
} from "../transport/requestInterception";
import { OwnedBrowserScope } from "./cleanup";
import { OwnedFrontendPublications } from "./frontendBuild";

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: unknown) => void;
  const promise = new Promise<T>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}

function contextResource() {
  const page = { unrouteAll: vi.fn(async () => {}) };
  const request = { dispose: vi.fn(async () => {}) };
  const resource = {
    pages: () => [page],
    request,
    unrouteAll: vi.fn(async () => {}),
    close: vi.fn(async () => {}),
  };
  return { page, resource, context: resource as unknown as BrowserContext };
}

function fixture(
  overrides: Partial<ConstructorParameters<typeof OwnedBrowserScope>[0]> = {},
) {
  const owned = contextResource();
  const browser = { newContext: vi.fn(async () => owned.context) };
  const revokeAllSessions = vi.fn(
    async (_userId: string, _reason: string) => {},
  );
  const verifyRevokedSession = vi.fn(async () => {});
  const scope = new OwnedBrowserScope({
    browser,
    revokeAllSessions,
    verifyRevokedSession,
    ...overrides,
  });
  return { ...owned, browser, revokeAllSessions, verifyRevokedSession, scope };
}

function snapshot(userId: string, cookie: string) {
  return {
    userId,
    email: "private@example.test",
    createdBy: "credential-file=/private",
    purpose: "secret-purpose",
    storageState: storageState(cookie),
  };
}

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

describe("OwnedBrowserScope", () => {
  it("owns actors before login and verifies every distinct acquired session", async () => {
    const { scope, revokeAllSessions, verifyRevokedSession } = fixture();
    scope.registerActor("login-never-completed");
    scope.registerActor("alpha");
    scope.registerSession(snapshot("alpha", "first"));
    scope.registerSession(snapshot("alpha", "first"));
    scope.registerSession(snapshot("alpha", "second"));
    scope.registerSession(snapshot("beta", "third"));
    await scope.dispose();
    expect(revokeAllSessions.mock.calls.map((call) => call[0])).toEqual([
      "beta",
      "alpha",
      "login-never-completed",
    ]);
    expect(verifyRevokedSession).toHaveBeenCalledTimes(3);
  });

  it("delegates context options to the instrumented browser and owns secondary contexts", async () => {
    const first = contextResource(),
      second = contextResource();
    const browser = {
      newContext: vi
        .fn()
        .mockResolvedValueOnce(first.context)
        .mockResolvedValueOnce(second.context),
    };
    const { scope } = fixture({ browser });
    const options = { locale: "fr-FR", storageState: storageState("alias") };
    expect(await scope.newContext(options)).toBe(first.context);
    expect(await scope.newContext()).toBe(second.context);
    expect(browser.newContext.mock.calls).toEqual([[options], [undefined]]);
    await scope.dispose();
    expect(first.resource.close).toHaveBeenCalledOnce();
    expect(second.resource.close).toHaveBeenCalledOnce();
  });

  it("cancels barriers before settling feature work, drains every owner, then closes and revokes", async () => {
    const order: string[] = [];
    const pageDrain = deferred<void>(),
      contextDrain = deferred<void>();
    const { scope, resource, page } = fixture({
      revokeAllSessions: async () => {
        order.push("revoke");
      },
      verifyRevokedSession: async () => {
        order.push("verify");
      },
    });
    await scope.newContext();
    scope.registerSession(snapshot("actor", "session"));
    const request = {
      dispose: vi.fn(async () => {
        order.push("request.dispose");
      }),
    };
    await scope.ownRequestContext(request as unknown as APIRequestContext);
    const gate = scope.barrier();
    scope.settleBeforeClosing(async () => {
      expect(await gate.promise).toBe("cancelled");
      order.push("feature");
    });
    page.unrouteAll.mockImplementation(async () => {
      order.push("page.drain");
      await pageDrain.promise;
    });
    resource.unrouteAll.mockImplementation(async () => {
      order.push("context.drain");
      await contextDrain.promise;
    });
    resource.close.mockImplementation(async () => {
      order.push("context.close");
    });
    const disposal = scope.dispose();
    await vi.waitFor(() =>
      expect(order).toEqual(["feature", "page.drain", "context.drain"]),
    );
    expect(resource.close).not.toHaveBeenCalled();
    pageDrain.resolve();
    await Promise.resolve();
    expect(resource.close).not.toHaveBeenCalled();
    contextDrain.resolve();
    await disposal;
    expect(order).toEqual([
      "feature",
      "page.drain",
      "context.drain",
      "context.close",
      "revoke",
      "verify",
      "request.dispose",
    ]);
  });

  it("uses first-terminal-outcome barriers and never releases cancelled work", async () => {
    const { scope } = fixture();
    const released = scope.barrier(),
      cancelled = scope.barrier(),
      teardown = scope.barrier();
    released.release();
    released.cancel();
    released.release();
    cancelled.cancel();
    cancelled.release();
    await scope.dispose();
    teardown.release();
    expect(
      await Promise.all([
        released.promise,
        cancelled.promise,
        teardown.promise,
      ]),
    ).toEqual(["released", "cancelled", "cancelled"]);
  });

  it("closes a context acquired during disposal before resolving disposal", async () => {
    const late = contextResource();
    const admission = deferred<BrowserContext>();
    const initializeContext = vi.fn(async () => {});
    const { scope } = fixture({
      browser: { newContext: () => admission.promise },
      initializeContext,
    });
    const acquiring = scope.newContext();
    const rejected = expect(acquiring).rejects.toThrow("acquired during");
    const disposal = scope.dispose();
    admission.resolve(late.context);
    await rejected;
    await disposal;
    expect(late.resource.close).toHaveBeenCalledOnce();
    expect(initializeContext).not.toHaveBeenCalled();
  });

  it("owns a late independent client and actor response without exposing either", async () => {
    const requestAdmission = deferred<APIRequestContext>();
    const actorAdmission = deferred<{ user_id: string }>();
    const client = { dispose: vi.fn(async () => {}) };
    const { scope, revokeAllSessions } = fixture();
    const requests = expect(
      scope.ownRequestContext(() => requestAdmission.promise),
    ).rejects.toThrow("acquired during");
    const actor = expect(
      scope.acquireActor(() => actorAdmission.promise),
    ).rejects.toThrow("acquired during");
    const disposal = scope.dispose();
    requestAdmission.resolve(client as unknown as APIRequestContext);
    actorAdmission.resolve({ user_id: "late-actor" });
    await Promise.all([requests, actor, disposal]);
    expect(client.dispose).toHaveBeenCalledOnce();
    expect(revokeAllSessions).toHaveBeenCalledWith(
      "late-actor",
      expect.any(String),
    );
  });

  it("settles in-flight authentication before revoking and verifies its late snapshot", async () => {
    const login = deferred<{ storageState: StorageState }>();
    const { scope, revokeAllSessions, verifyRevokedSession } = fixture();
    const acquiring = scope.acquireSession(
      {
        userId: "actor",
        email: "private@example.test",
        createdBy: "test",
        purpose: "login",
      },
      () => login.promise,
    );
    const rejected = expect(acquiring).rejects.toThrow("acquired during");
    const disposal = scope.dispose();
    expect(revokeAllSessions).not.toHaveBeenCalled();
    login.resolve({ storageState: storageState("late-session") });
    await Promise.all([rejected, disposal]);
    expect(revokeAllSessions).toHaveBeenCalledOnce();
    expect(verifyRevokedSession).toHaveBeenCalledWith(
      expect.objectContaining({
        userId: "actor",
        sessionCookie: "late-session",
      }),
    );
  });

  it("rejects admission synchronously after disposal begins without invoking factories", async () => {
    const { scope, browser } = fixture();
    const disposal = scope.dispose();
    const factory = vi.fn();
    expect(() => scope.newContext()).toThrow("disposal has begun");
    expect(() => scope.ownRequestContext(factory)).toThrow(
      "disposal has begun",
    );
    expect(() => scope.acquireActor(factory)).toThrow("disposal has begun");
    expect(() => scope.registerActor("late")).toThrow("disposal has begun");
    expect(() => scope.registerSession(snapshot("late", "secret"))).toThrow(
      "disposal has begun",
    );
    expect(() => scope.barrier()).toThrow("disposal has begun");
    expect(() => scope.settleBeforeClosing(factory)).toThrow(
      "disposal has begun",
    );
    expect(factory).not.toHaveBeenCalled();
    expect(browser.newContext).not.toHaveBeenCalled();
    await disposal;
  });

  it("returns one completion promise for concurrent and repeated failed disposal", async () => {
    const { scope, revokeAllSessions } = fixture({
      verifyRevokedSession: async () => {
        throw new Error("secret");
      },
    });
    scope.registerSession(snapshot("actor", "private-cookie"));
    const first = scope.dispose(),
      concurrent = scope.dispose();
    expect(concurrent).toBe(first);
    await expect(first).rejects.toThrow("session.verify_revoked");
    expect(scope.dispose()).toBe(first);
    await expect(scope.dispose()).rejects.toThrow("session.verify_revoked");
    expect(revokeAllSessions).toHaveBeenCalledOnce();
  });

  it("continues every cleanup phase after synchronous and asynchronous failures with redacted diagnostics", async () => {
    const secret =
      "private@example.test cookie=secret-token user-secret /private";
    const fail = () => {
      throw new Error(secret);
    };
    const { scope, page, resource } = fixture({
      revokeAllSessions: fail,
      verifyRevokedSession: fail,
    });
    await scope.newContext();
    scope.registerSession(snapshot("user-secret", "secret-token"));
    scope.settleBeforeClosing(fail);
    scope.recordRouteFailure();
    page.unrouteAll.mockImplementation(fail);
    resource.unrouteAll.mockImplementation(fail);
    resource.close.mockImplementation(fail);
    const client = { dispose: vi.fn(fail) };
    await scope.ownRequestContext(client as unknown as APIRequestContext);
    const failure = await scope.dispose().catch((error: unknown) => error);
    expect(failure).toBeInstanceOf(Error);
    expect(String(failure)).toContain(
      "route.handle=1, feature.settle=1, context.drain=2, context.close=1, session.revoke_all=1, session.verify_revoked=1, request.dispose=1",
    );
    expect(String(failure)).not.toMatch(
      /private@example|secret-token|user-secret|\/private/,
    );
    expect(client.dispose).toHaveBeenCalledOnce();
  });

  it("retains acquisition failure while cleaning already acquired actors", async () => {
    const { scope, revokeAllSessions } = fixture({
      browser: {
        newContext: async () => {
          throw new Error("private-acquisition");
        },
      },
    });
    scope.registerActor("actor");
    await expect(scope.newContext()).rejects.toThrow("private-acquisition");
    await expect(scope.dispose()).rejects.toThrow("context.acquire");
    expect(revokeAllSessions).toHaveBeenCalledOnce();
  });

  it("uses the same early-close completion and does not close the context twice", async () => {
    const { scope, context, resource } = fixture();
    await scope.newContext();
    const gate = scope.barrier(context);
    const first = scope.closeContext(context);
    expect(scope.closeContext(context)).toBe(first);
    await first;
    expect(await gate.promise).toBe("cancelled");
    await scope.dispose();
    expect(resource.close).toHaveBeenCalledOnce();
  });

  it("settles context feature work before early closure and shares it with final disposal", async () => {
    const { scope, context, resource } = fixture();
    await scope.newContext();
    const work = deferred<void>();
    const settle = vi.fn(() => work.promise);
    scope.settleBeforeClosing(settle, context);
    const early = scope.closeContext(context);
    const final = scope.dispose();
    expect(scope.closeContext(context)).toBe(early);
    expect(resource.close).not.toHaveBeenCalled();
    work.resolve();
    await Promise.all([early, final]);
    expect(settle).toHaveBeenCalledOnce();
    expect(resource.close).toHaveBeenCalledOnce();
  });

  it("closes and revokes even when page enumeration fails", async () => {
    const { scope, resource, revokeAllSessions } = fixture();
    await scope.newContext();
    scope.registerActor("actor");
    resource.pages = () => {
      throw new Error("private page diagnostic");
    };
    await expect(scope.dispose()).rejects.toThrow("context.drain=1");
    expect(resource.close).toHaveBeenCalledOnce();
    expect(revokeAllSessions).toHaveBeenCalledOnce();
  });

  it("keeps early-close failure visible while still revoking sessions and disposing clients", async () => {
    const { scope, context, resource, revokeAllSessions } = fixture();
    await scope.newContext();
    scope.registerActor("actor");
    resource.close.mockRejectedValue(new Error("private"));
    const closing = scope.closeContext(context);
    await expect(closing).rejects.toThrow("Owned context cleanup failed");
    expect(scope.closeContext(context)).toBe(closing);
    await expect(scope.dispose()).rejects.toThrow("context.close");
    expect(revokeAllSessions).toHaveBeenCalledOnce();
  });
});

describe("owned feature jobs", () => {
  it("owns late admission and settles terminal work before retiring its context", async () => {
    const { scope, context, resource } = fixture();
    await scope.newContext();
    const admitted = deferred<string>();
    const terminal = deferred<void>();
    const observe = vi.fn(() => terminal.promise);
    const jobs = new OwnedJobs(observe);
    scope.settleBeforeClosing(() => jobs.settle(), context);
    const admission = jobs.admit(
      () => admitted.promise,
      async (jobId) => ({ jobId, headers: { cookie: "private-session" } }),
    );
    const closing = scope.dispose();
    admitted.resolve("owned-job");
    await admission;
    const settlement = jobs.settle();
    expect(jobs.settle()).toBe(settlement);
    await Promise.resolve();
    expect(resource.close).not.toHaveBeenCalled();
    terminal.resolve();
    await closing;
    expect(observe).toHaveBeenCalledWith("owned-job", {
      cookie: "private-session",
    });
    expect(resource.close).toHaveBeenCalledOnce();
    expect(() =>
      jobs.admit(
        async () => "late",
        async () => null,
      ),
    ).toThrow("admission has stopped");
  });

  it("retains uncertain admission and settlement failure without private diagnostics", async () => {
    const observed: string[] = [];
    const jobs = new OwnedJobs(async (id) => {
      observed.push(id);
      throw new Error("private-terminal");
    });
    await jobs.admit(
      async () => "first",
      async (jobId) => ({ jobId, headers: {} }),
    );
    await jobs.admit(
      async () => "second",
      async (jobId) => ({ jobId, headers: {} }),
    );
    await expect(
      jobs.admit(
        async () => {
          throw new Error("private-admission");
        },
        async () => null,
      ),
    ).rejects.toThrow("private-admission");
    const completion = jobs.settle();
    await expect(completion).rejects.toThrow(
      /^Feature job settlement failed$/u,
    );
    expect(jobs.settle()).toBe(completion);
    expect(observed).toEqual(["first", "second"]);
  });
});

async function routedFixture() {
  const f = fixture();
  await f.scope.newContext();
  let handle!: (route: Route) => Promise<void>;
  let pattern!: (url: URL) => boolean;
  const page = {
    context: () => f.context,
    isClosed: () => false,
    route: vi.fn(
      async (match: (url: URL) => boolean, handler: typeof handle) => {
        pattern = match;
        handle = handler;
      },
    ),
    unroute: vi.fn(async () => {}),
  };
  const route = {
    request: () => ({ method: () => "PATCH" }),
    continue: vi.fn(async () => {}),
    fallback: vi.fn(async () => {}),
    fulfill: vi.fn(async (_options: unknown) => {}),
    abort: vi.fn(async () => {}),
  };
  return {
    ...f,
    page: page as unknown as Page,
    fakePage: page,
    route,
    invoke: () => handle(route as unknown as Route),
    matches: (url: string) => pattern(new URL(url)),
  };
}

describe("owned request interception", () => {
  it("limits public errors to the exact method and path without consuming other requests", async () => {
    const f = await routedFixture();
    const error = await installPublicErrorResponse(f.page, {
      method: "PATCH",
      path: "/api/v1/users/owned",
      code: "last_deployment_admin",
      status: 409,
    });
    expect(f.matches("https://example.test/api/v1/users/owned")).toBe(true);
    expect(f.matches("https://example.test/api/v1/users/owned?extra=1")).toBe(
      false,
    );
    expect(f.matches("https://example.test/api/v1/users/other")).toBe(false);
    f.route.request = () => ({ method: () => "GET" });
    await f.invoke();
    expect(error.hitCount()).toBe(0);
    f.route.request = () => ({ method: () => "PATCH" });
    await f.invoke();
    await f.invoke();
    expect(error.hitCount()).toBe(1);
    expect(f.route.fallback).toHaveBeenCalledTimes(2);
    expect(f.route.fulfill).toHaveBeenCalledWith({
      status: 409,
      headers: { "x-request-id": "browser-presentation" },
      json: {
        error: {
          code: "last_deployment_admin",
          status: 409,
          request_id: "browser-presentation",
          retryable: false,
          message: "last_deployment_admin",
          details: {},
        },
      },
    });
    await error.dispose();
    await f.scope.dispose();
  });

  it("matches the exact path and query, releases once, and falls through for other methods", async () => {
    const f = await routedFixture();
    const held = await holdBrowserRequest(f.page, {
      method: "PATCH",
      path: "/api/v1/users/owned?mode=one",
      once: true,
    });
    expect(f.matches("https://example.test/api/v1/users/owned?mode=one")).toBe(
      true,
    );
    expect(f.matches("https://example.test/api/v1/users/owned?mode=two")).toBe(
      false,
    );
    expect(
      f.matches("https://example.test/api/v1/users/owned/extra?mode=one"),
    ).toBe(false);
    const first = f.invoke();
    expect(await held.waitForHit).toBe("released");
    expect(f.route.continue).not.toHaveBeenCalled();
    held.release();
    held.release();
    await first;
    await f.invoke();
    expect(f.route.continue).toHaveBeenCalledOnce();
    expect(f.route.fallback).toHaveBeenCalledOnce();
    await held.dispose();
    await f.scope.dispose();
  });

  it("cancels before a handler reaches its wait without dispatching delayed work", async () => {
    const f = await routedFixture();
    const held = await holdBrowserRequest(f.page, {
      method: "PATCH",
      path: "/owned",
    });
    const handling = f.invoke();
    const disposal = held.dispose();
    expect(held.dispose()).toBe(disposal);
    await Promise.all([handling, disposal]);
    expect(await held.waitForHit).toBe("cancelled");
    expect(f.route.abort).toHaveBeenCalledWith("aborted");
    expect(f.route.continue).not.toHaveBeenCalled();
    await f.scope.dispose();
  });

  it("cancels an active handler and waits for its abort to settle", async () => {
    const f = await routedFixture();
    const abort = deferred<void>();
    f.route.abort.mockImplementation(() => abort.promise);
    const held = await holdBrowserRequest(f.page, {
      method: "PATCH",
      path: "/owned",
    });
    const handling = f.invoke();
    await held.waitForHit;
    let settled = false;
    const disposal = held.dispose().then(() => {
      settled = true;
    });
    await vi.waitFor(() => expect(f.route.abort).toHaveBeenCalledOnce());
    expect(settled).toBe(false);
    expect(f.fakePage.unroute).not.toHaveBeenCalled();
    abort.resolve();
    await Promise.all([handling, disposal]);
    expect(f.fakePage.unroute).toHaveBeenCalledOnce();
    expect(f.route.continue).not.toHaveBeenCalled();
    await f.scope.dispose();
  });

  it("retains fulfillment and removal failure without retaining private errors", async () => {
    const f = await routedFixture();
    const installed = await installOwnedRoute(f.page, "**/owned", async () => {
      throw new Error("credential=private-fulfillment");
    });
    await expect(f.invoke()).rejects.toThrow(
      "Owned browser route handling failed",
    );
    f.fakePage.unroute.mockRejectedValue(new Error("cookie=private-removal"));
    const disposal = installed.dispose();
    expect(installed.dispose()).toBe(disposal);
    await expect(disposal).rejects.toThrow(
      "Owned browser route disposal failed",
    );
    expect(installed.dispose()).toBe(disposal);
    const failure = await f.scope.dispose().catch((error: unknown) => error);
    expect(String(failure)).toContain("route.handle");
    expect(String(failure)).not.toContain("private");
  });
});

describe("OwnedFrontendPublications", () => {
  it("closes browser dependencies before settling every independent publication", async () => {
    const owner = fixture();
    await owner.scope.newContext();
    owner.scope.registerActor("ordinary-actor");
    const closed = deferred<void>();
    owner.resource.close.mockImplementation(async () => closed.promise);
    const stop = vi.fn(async () => {
      expect(owner.revokeAllSessions).toHaveBeenCalledWith(
        "ordinary-actor",
        expect.any(String),
      );
    });
    const publications = new OwnedFrontendPublications(owner.scope, () => ({
      finished: Promise.resolve("owned-build"),
      stop,
    }));
    publications.start(false);
    publications.start(true);
    const disposing = publications.dispose();
    expect(publications.dispose()).toBe(disposing);
    expect(() => publications.start(false)).toThrow("disposal has begun");
    await vi.waitFor(() => expect(owner.resource.close).toHaveBeenCalledOnce());
    expect(stop).not.toHaveBeenCalled();
    closed.resolve();
    await disposing;
    expect(stop).toHaveBeenCalledTimes(2);
    expect(owner.resource.close).toHaveBeenCalledOnce();
  });

  it("retains acquired publications after later startup failure", async () => {
    const owner = fixture();
    const stop = vi.fn(async () => {});
    let starts = 0;
    const publications = new OwnedFrontendPublications(owner.scope, () => {
      if (++starts === 2) throw new Error("startup failed");
      return { finished: Promise.resolve("owned-build"), stop };
    });
    publications.start(false);
    expect(() => publications.start(true)).toThrow("startup failed");
    await publications.dispose();
    expect(stop).toHaveBeenCalledOnce();
  });

  it("preserves repeated cleanup failure while attempting every build and redacting private diagnostics", async () => {
    const owner = fixture();
    await owner.scope.newContext();
    owner.resource.close.mockRejectedValue(new Error("private browser value"));
    const failedStop = vi.fn(() => {
      throw new Error("private build value");
    });
    const successfulStop = vi.fn(async () => {});
    const publications = new OwnedFrontendPublications(
      owner.scope,
      (measurement) => ({
        finished: Promise.resolve("owned-build"),
        stop: measurement ? successfulStop : failedStop,
      }),
    );
    publications.start(false);
    publications.start(true);
    const disposing = publications.dispose();
    await expect(disposing).rejects.toThrow(
      "browser.dispose, publication.stop=1",
    );
    expect(publications.dispose()).toBe(disposing);
    await expect(publications.dispose()).rejects.not.toThrow("private");
    expect(failedStop).toHaveBeenCalledOnce();
    expect(successfulStop).toHaveBeenCalledOnce();
  });
});
