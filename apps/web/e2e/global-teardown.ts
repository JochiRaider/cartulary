import { cleanupWorkerAdminSuite } from "./support/auth/sessions";
import { clearSuiteAdminTotpSecret } from "./support/auth/suiteAdmin";
import { clearWorkerAdminSuiteState } from "./support/auth/workerAdmin";
import {
  isExternalServerHarnessMode,
  usesSharedPlaywrightState,
} from "./support/runtime/harnessState";

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
