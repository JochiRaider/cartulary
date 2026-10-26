import path from "node:path";
import test from "node:test";
import { createContractTestContext } from "./contract-test-context.mjs";
import { assertLazyCaseContext, assertImportAndFailureIsolation } from "./contract-initialization-cases.mjs";
import { assertPrivateChildCaptureBoundary } from "./private-child-capture-cases.mjs";
const root = path.resolve(import.meta.dirname, "../../..");
export function registerBoundariesTests() {
  test("private_child_capture_boundary: private child capture owns allocation and exception-safe cleanup", async () => {
    const context = createContractTestContext(root);
    assertPrivateChildCaptureBoundary(context);
  });
  test("lazy_case_context: case contexts load only their own dependencies", async () => {
    const context = createContractTestContext(root);
    assertLazyCaseContext(context);
  });
  test("import_and_failure_isolation: support imports are inert and policy failures remain case-local", async () => {
    const context = createContractTestContext(root);
    assertImportAndFailureIsolation(context);
  });
}
