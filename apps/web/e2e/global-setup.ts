import type { FullConfig } from "@playwright/test";
import { withBrowserSetupLock } from "../../../tools/harness/browser/private-state.mjs";
import { prepareWorkerAdminSuite } from "./support/auth/sessions";
import { prepareSuiteAdminState } from "./support/auth/suiteAdmin";
import { verifyOwnedHarnessRuntime } from "./support/runtime/readiness";

export default async function globalSetup(config: FullConfig) {
  if (config.workers !== 1)
    throw new Error("browser groups require exactly one worker account");
  const workerCount = 1;
  await withBrowserSetupLock(async () => {
    await verifyOwnedHarnessRuntime();
    await prepareSuiteAdminState();
    await prepareWorkerAdminSuite(workerCount);
  });
}
