import { cleanupWorkerAdminSuite } from "./authRuntime";
import { clearSuiteAdminTotpSecret } from "./helpers";
import { clearWorkerAdminSuiteState } from "./sessionSupport";
import { isExternalServerHarnessMode } from "./harnessState";

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
