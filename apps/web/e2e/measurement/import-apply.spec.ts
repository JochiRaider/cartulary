import { test } from "./fixtures";
import { runImportProgress } from "./importProgressScenario";

test.use({ jobSettlementTimeout: 120_000 });

test.use({
  viewport: { width: 1920, height: 1080 },
  deviceScaleFactor: 1,
  actionTimeout: 10_000,
  navigationTimeout: 30_000,
});
test.describe.configure({ mode: "serial", timeout: 10 * 60 * 1_000 });

test("measures CSV apply progress and cancel while Timeline accepts editing", async ({
  browserScope,
  page,
  testActor,
  sessionTracker,
}, info) => {
  await runImportProgress(
    "apply",
    browserScope,
    page,
    testActor,
    sessionTracker,
    info,
    "csv",
  );
});

test("measures XLSX apply progress and cancel while Timeline accepts editing", async ({
  browserScope,
  page,
  testActor,
  sessionTracker,
}, info) => {
  await runImportProgress(
    "apply",
    browserScope,
    page,
    testActor,
    sessionTracker,
    info,
    "xlsx",
  );
});
