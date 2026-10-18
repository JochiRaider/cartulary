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
