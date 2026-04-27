import { cleanupWorkerAdminSuite } from "./authRuntime";
import {
  isExternalServerHarnessMode,
  usesSharedPlaywrightState,
} from "./harnessState";
import { clearSuiteAdminTotpSecret } from "./helpers";
import { clearWorkerAdminSuiteState } from "./sessionSupport";

export default async function globalTeardown() {
  if (isExternalServerHarnessMode() && usesSharedPlaywrightState()) {
    return;
  }

  try {
    await cleanupWorkerAdminSuite();
  } finally {
    if (!isExternalServerHarnessMode()) {
      clearWorkerAdminSuiteState();
      clearSuiteAdminTotpSecret();
    }
  }
}
