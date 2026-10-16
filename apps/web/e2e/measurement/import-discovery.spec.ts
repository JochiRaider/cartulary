import { test } from "./fixtures";
import { runImportProgress } from "./importProgressScenario";

test.use({
  viewport: { width: 1920, height: 1080 },
  deviceScaleFactor: 1,
  actionTimeout: 10_000,
  navigationTimeout: 30_000,
});
test.describe.configure({ mode: "serial", timeout: 10 * 60 * 1_000 });

test("measures CSV discovery progress and cancel while Timeline accepts editing", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runImportProgress(
    "discovery",
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "csv",
  );
});

test("measures XLSX discovery progress and cancel while Timeline accepts editing", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runImportProgress(
    "discovery",
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "xlsx",
  );
});
