import type { Page, Request } from "@playwright/test";

/**
 * Resolves when the first request matching `matches` finishes or fails.
 * Register it before the action that issues the request.
 */
export function settledRequestMatching(
  page: Page,
  matches: (url: URL, request: Request) => boolean,
): Promise<void> {
  return new Promise<void>((resolve) => {
    const onSettled = (request: Request) => {
      if (!matches(new URL(request.url()), request)) return;
      page.off("requestfinished", onSettled);
      page.off("requestfailed", onSettled);
      resolve();
    };
    page.on("requestfinished", onSettled);
    page.on("requestfailed", onSettled);
  });
}

/** Resolves when the first `method` request whose path matches settles. */
export function settledRequest(
  page: Page,
  method: string,
  pathPattern: RegExp,
): Promise<void> {
  return settledRequestMatching(
    page,
    (url, request) =>
      request.method() === method && pathPattern.test(url.pathname),
  );
}
