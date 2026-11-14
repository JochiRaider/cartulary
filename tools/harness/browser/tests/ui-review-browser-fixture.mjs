// Test composition only. Production resource cleanup belongs to the native owner.
import { boundedCleanup, stopOwnedProcess } from "../ui-review/ownership.mjs";
export async function closeReviewBrowser(browser) {
  if (!browser) return;
  browser.closing = true;
  try { await boundedCleanup(async () => { await browser.browser?.close(); await browser.server?.close(); }, 10000); }
  finally {
    if (browser.process) {
      await stopOwnedProcess(browser.process);
      await browser.onOwnedResource({ kind: "browser_process", target: browser.process, state: "released" });
    }
  }
}

export async function closeReviewBrowserFixture({ browser, viewer, admission, server }) {
  const outcomes = await Promise.allSettled([
    (async () => {
      const closures = await Promise.allSettled([
        closeReviewBrowser(browser),
        Promise.resolve().then(() => viewer?.close()),
      ]);
      const failures = closures.filter(result => result.status === "rejected").map(result => result.reason);
      if (failures.length) throw new AggregateError(failures, "browser closure failed");
      await admission?.release();
    })(),
    new Promise((resolve, reject) => {
      server.closeAllConnections();
      if (!server.listening) { resolve(); return; }
      server.close(error => error ? reject(error) : resolve());
    }),
  ]);
  const failures = outcomes.filter(result => result.status === "rejected").map(result => result.reason);
  if (failures.length) throw new AggregateError(failures, "browser fixture cleanup failed");
}
