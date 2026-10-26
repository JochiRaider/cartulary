import type { Page, Request } from "@playwright/test";
import { browserScopeFor } from "./cleanup";

/** Keep synchronous observation errors attached to the owning test lifetime. */
export function observePageRequests(
  page: Page,
  observe: (request: Request) => void,
) {
  const scope = browserScopeFor(page);
  let failed = false;
  const listener = (request: Request) => {
    try {
      observe(request);
    } catch {
      failed = true;
    }
  };
  page.on("request", listener);
  const stop = () => {
    page.off("request", listener);
  };
  scope.settleBeforeClosing(async () => {
    stop();
    if (failed) throw new Error("Owned request observation failed");
  }, page.context());
  return stop;
}

/** Fail the owning test during settlement without throwing from an event emitter. */
export function rejectUnexpectedPageErrors(page: Page) {
  const scope = browserScopeFor(page);
  let failed = false;
  const listener = () => {
    failed = true;
  };
  page.on("pageerror", listener);
  scope.settleBeforeClosing(async () => {
    page.off("pageerror", listener);
    if (failed) throw new Error("Unexpected browser page error");
  }, page.context());
}
