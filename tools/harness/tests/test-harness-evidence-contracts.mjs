import { runContractSuite } from "./contract-suite-support.mjs";
import { registerRunObservationTests } from "../diagnostics/tests/run-observation-cases.mjs";
runContractSuite("evidence");
registerRunObservationTests();

import "../observability/tests/test-instrumentation.mjs";
