import type { FullConfig } from "@playwright/test";
import { prepareWorkerAdminSuite } from "./authRuntime";
import { clearSuiteAdminTotpSecret, prepareSuiteAdminState } from "./helpers";
import { clearWorkerAdminSuiteState } from "./sessionSupport";

export default async function globalSetup(config: FullConfig) {
  clearWorkerAdminSuiteState();
  clearSuiteAdminTotpSecret();
  await prepareSuiteAdminState();
  const workerCount = typeof config.workers === "number" ? config.workers : 1;
  await prepareWorkerAdminSuite(workerCount);
}
