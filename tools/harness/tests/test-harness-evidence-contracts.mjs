import { runContractSuite } from "./contract-suite-support.mjs";
import { registerRunObservationTests } from "../diagnostics/tests/run-observation-cases.mjs";
runContractSuite("evidence");
registerRunObservationTests();

import "../observability/tests/test-instrumentation.mjs";
import "../observability/tests/test-foundation-characterization.mjs";

import "../observability/tests/test-collection-engine.mjs";
