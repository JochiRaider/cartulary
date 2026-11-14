import type { FullConfig } from "@playwright/test";
import { withBrowserSetupLock } from "../../../tools/harness/browser/private-state.mjs";
import { prepareWorkerAdminSuite } from "./support/auth/sessions";
import { prepareSuiteAdminState } from "./support/auth/suiteAdmin";
import { verifyOwnedHarnessRuntime } from "./support/runtime/readiness";

export default async function globalSetup(config: FullConfig) {
  const configuredWorkerCount = positiveIntegerEnv(
    "CARTULARY_PLAYWRIGHT_WORKER_COUNT",
  );
  if (config.workers !== 1 || configuredWorkerCount !== 1)
    throw new Error("browser groups require exactly one worker account");
  const workerCount = 1;
  await withBrowserSetupLock(async () => {
    await verifyOwnedHarnessRuntime();
    await prepareSuiteAdminState();
    await prepareWorkerAdminSuite(workerCount);
  });
}

function positiveIntegerEnv(name: string) {
  const value = process.env[name];
  if (value === undefined || value.trim() === "") {
    return null;
  }
  const parsed = Number.parseInt(value, 10);
  if (!Number.isInteger(parsed) || parsed < 1 || String(parsed) !== value) {
    throw new Error(`${name} must be a positive integer`);
  }
  return parsed;
}
