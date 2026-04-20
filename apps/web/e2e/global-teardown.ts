import { cleanupWorkerAdminSuite } from "./authRuntime";
import { clearSuiteAdminTotpSecret } from "./helpers";

export default async function globalTeardown() {
  try {
    await cleanupWorkerAdminSuite();
  } finally {
    clearSuiteAdminTotpSecret();
  }
}
