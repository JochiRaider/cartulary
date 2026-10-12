import "../observability/tests/test-execution-index.mjs";
import "./test-catalog-metadata-fixtures.mjs";
import { runContractSuite } from "./contract-suite-support.mjs";
import { registerRunObservationTests } from "../diagnostics/tests/run-observation-cases.mjs";
runContractSuite("evidence");
registerRunObservationTests();

import "../observability/tests/test-instrumentation.mjs";
import "../observability/tests/test-foundation-characterization.mjs";

import "../observability/tests/test-collection-engine.mjs";

import "../observability/tests/test-clock-placement.mjs";


import "../observability/tests/test-wrapper-receipt.mjs";

import "../observability/tests/test-qualification.mjs";
