import type { Page, Route } from "@playwright/test";

type HeldBrowserRequest = {
  readonly dispose: () => Promise<void>;
  readonly hitCount: () => number;
  readonly release: () => void;
  readonly waitForHit: Promise<void>;
};

function browserRequestRoute(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `**${normalizedPath}`;
}

export async function holdBrowserRequest(
  page: Page,
  options: {
    method: string;
    path: string;
  },
): Promise<HeldBrowserRequest> {
  const routePattern = browserRequestRoute(options.path);
  const expectedMethod = options.method.toUpperCase();
  let matchingHitCount = 0;
  let hitResolved = false;
  let releaseHold: (() => void) | null = null;
  let resolveHit: (() => void) | null = null;
  const waitForHit = new Promise<void>((resolve) => {
    resolveHit = resolve;
  });
  const hold = new Promise<void>((resolve) => {
    releaseHold = resolve;
  });

  const routeHandler = async (route: Route) => {
    if (route.request().method().toUpperCase() !== expectedMethod) {
      await route.fallback();
      return;
    }
    matchingHitCount += 1;
    if (!hitResolved) {
      hitResolved = true;
      resolveHit?.();
    }
    await hold;
    try {
      await route.continue();
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      if (!message.includes("Route is already handled")) {
        throw error;
      }
    }
  };

  await page.route(routePattern, routeHandler);
  return {
    dispose: async () => {
      releaseHold?.();
      await safelyRemoveRoute(page, routePattern, routeHandler);
    },
    hitCount: () => matchingHitCount,
    release: () => releaseHold?.(),
    waitForHit,
  };
}

export async function safelyRemoveRoute(
  page: Page,
  routePattern: Parameters<Page["unroute"]>[0],
  routeHandler: Parameters<Page["unroute"]>[1],
) {
  if (page.isClosed()) {
    return;
  }
  try {
    await page.unroute(routePattern, routeHandler);
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : String(error);
    if (
      !message.includes("Target page, context or browser has been closed") &&
      !message.includes("Page closed") &&
      !message.includes("BrowserContext closed")
    ) {
      throw error;
    }
  }
}
