import "../observability/tests/test-execution-index.mjs";
import "../observability/tests/test-owner-evidence-audit.mjs";
import { runContractSuite } from "./contract-suite-support.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { firstActionableLine } from "../output/test-output/runners/shell.mjs";

test("shell failure headlines preserve diagnostics rather than source excerpts", () => {
  assert.equal(firstActionableLine(["file:///workspace/check.mjs:42", "  throw new Error(message);", "       ^", "Error: selector is ambiguous", "    at check (file:///workspace/check.mjs:42:9)"]), "Error: selector is ambiguous");
  assert.equal(firstActionableLine(["node:internal/process:9", "AssertionError [ERR_ASSERTION]: missing owned output"]), "AssertionError [ERR_ASSERTION]: missing owned output");
  assert.equal(firstActionableLine(["", "ok package", "connection refused"]), "connection refused");
});
import { registerRunObservationTests } from "../diagnostics/tests/run-observation-cases.mjs";
runContractSuite("evidence");
registerRunObservationTests();

import "../observability/tests/test-instrumentation.mjs";
import "../observability/tests/test-foundation-characterization.mjs";

import "../observability/tests/test-collection-engine.mjs";

import "../observability/tests/test-clock-placement.mjs";


import "../observability/tests/test-wrapper-receipt.mjs";

import "../observability/tests/test-qualification.mjs";
