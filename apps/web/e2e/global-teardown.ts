import { cleanupWorkerAdminSuite } from "./authRuntime";
import { isExternalServerHarnessMode } from "./harnessState";
import { clearSuiteAdminTotpSecret } from "./helpers";
import { clearWorkerAdminSuiteState } from "./sessionSupport";

export default async function globalTeardown() {
  try {
    await cleanupWorkerAdminSuite();
  } finally {
    if (!isExternalServerHarnessMode()) {
      clearWorkerAdminSuiteState();
      clearSuiteAdminTotpSecret();
    }
  }
}
