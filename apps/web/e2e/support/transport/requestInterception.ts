import type { Page, Request, Route } from "@playwright/test";
import { type BarrierOutcome, browserScopeFor } from "../runtime/cleanup";

/** A cancelled gate must never dispatch or fulfill the held operation. */
export async function awaitResponseBarrier(
  route: Route,
  barrier: Promise<BarrierOutcome> | null,
): Promise<boolean> {
  if (barrier !== null && (await barrier) === "cancelled") {
    await route.abort("aborted");
    return false;
  }
  return true;
}

type RoutePattern = Parameters<Page["route"]>[0];
type RouteHandler = Parameters<Page["route"]>[1];
type OwnedRoute = {
  pattern: RoutePattern;
  handler: RouteHandler;
  dispose: () => Promise<void>;
};
const registrations = new WeakMap<Page, Set<OwnedRoute>>();

/** Track handler failures and share removal/settlement across every disposer. */
export async function installOwnedRoute(
  page: Page,
  pattern: RoutePattern,
  handle: RouteHandler,
  options: Parameters<Page["route"]>[2] | (() => void) = {},
) {
  const scope = browserScopeFor(page);
  const pending = new Set<Promise<void>>();
  let disposal: Promise<void> | undefined;
  let failed = false;
  let disposing = false;
  const handler = (route: Route, request: Request) => {
    const handling = Promise.resolve()
      .then(() => (disposing ? route.abort("aborted") : handle(route, request)))
      .catch(() => {
        failed = true;
        scope.recordRouteFailure();
        throw new Error("Owned browser route handling failed");
      });
    pending.add(handling);
    void handling.then(
      () => pending.delete(handling),
      () => pending.delete(handling),
    );
    return handling;
  };
  await scope.acquireRoute(() =>
    page.route(
      pattern,
      handler,
      typeof options === "function" ? undefined : options,
    ),
  );
  const owned: OwnedRoute = {
    pattern,
    handler: handle,
    dispose: () => {
      if (!disposal) {
        disposing = true;
        if (typeof options === "function") options();
        disposal = (async () => {
          // Selective unroute updates browser interception immediately. Settle
          // cancelled handlers first so that update cannot dispatch their work.
          // Registration is synchronous; no handler can enter between the empty
          // check and unroute's synchronous removal from Playwright's route list.
          while (pending.size > 0) {
            const results = await Promise.allSettled([...pending]);
            failed ||= results.some((result) => result.status === "rejected");
          }
          try {
            if (!page.isClosed()) await page.unroute(pattern, handler);
          } catch {
            failed = true;
            scope.recordRouteFailure();
          }
          registrations.get(page)?.delete(owned);
          if (failed) throw new Error("Owned browser route disposal failed");
        })();
      }
      return disposal;
    },
  };
  let routes = registrations.get(page);
  if (!routes) {
    routes = new Set();
    registrations.set(page, routes);
  }
  routes.add(owned);
  return owned;
}

export async function holdBrowserRequest(
  page: Page,
  options: { method: string; path: string; once?: boolean },
) {
  const scope = browserScopeFor(page);
  const hold = scope.barrier(page.context());
  const hit = scope.barrier(page.context());
  const expectedMethod = options.method.toUpperCase();
  const expectedPath = options.path.startsWith("/")
    ? options.path
    : `/${options.path}`;
  const pattern = (url: URL) => `${url.pathname}${url.search}` === expectedPath;
  let matchingHitCount = 0;
  const route = await installOwnedRoute(
    page,
    pattern,
    async (route) => {
      if (
        route.request().method().toUpperCase() !== expectedMethod ||
        (options.once && matchingHitCount > 0)
      ) {
        await route.fallback();
        return;
      }
      matchingHitCount++;
      hit.release();
      if (!(await awaitResponseBarrier(route, hold.promise))) return;
      await route.continue();
    },
    () => {
      hold.cancel();
      hit.cancel();
    },
  );
  return {
    dispose: route.dispose,
    hitCount: () => matchingHitCount,
    release: hold.release,
    cancel: () => {
      hold.cancel();
      hit.cancel();
    },
    waitForHit: hit.promise,
  };
}

/** Presentation-only error, admitted by exact method/path and bounded hit count. */
export async function installPublicErrorResponse(
  page: Page,
  options: {
    method: string;
    path: string;
    code: string;
    status: number;
    details?: Record<string, unknown>;
    retryable?: boolean;
    message?: string;
    times?: number;
  },
) {
  let hits = 0;
  const route = await installOwnedRoute(
    page,
    (url) => `${url.pathname}${url.search}` === options.path,
    async (route) => {
      if (
        route.request().method() !== options.method.toUpperCase() ||
        hits >= (options.times ?? 1)
      ) {
        await route.fallback();
        return;
      }
      hits++;
      const requestId = "browser-presentation";
      await route.fulfill({
        status: options.status,
        headers: { "x-request-id": requestId },
        json: {
          error: {
            code: options.code,
            status: options.status,
            request_id: requestId,
            retryable: options.retryable ?? false,
            message: options.message ?? options.code,
            details: options.details ?? {},
          },
        },
      });
    },
  );
  return { dispose: route.dispose, hitCount: () => hits };
}

/** Selective removal drains every matching registered handler before detaching it. */
export async function removeOwnedRoute(
  page: Page,
  pattern: RoutePattern,
  handler?: RouteHandler,
) {
  const matches = [...(registrations.get(page) ?? [])].filter(
    (route) =>
      route.pattern === pattern &&
      (handler === undefined || route.handler === handler),
  );
  const results = await Promise.allSettled(
    matches.map((route) => route.dispose()),
  );
  if (results.some((result) => result.status === "rejected"))
    throw new Error("Owned browser route removal failed");
}
