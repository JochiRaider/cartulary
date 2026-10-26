import { registerEvidenceTests as registerTestCatalogEvidenceTests } from "../test-catalog/tests/test-contracts.mjs";
import { registerEvidenceTests as registerReadinessEvidenceTests } from "../readiness/tests/test-frontend-artifact.mjs";
import "../observability/tests/test-critical-path.mjs";
import "../observability/tests/test-execution-index.mjs";
import "../observability/tests/test-owner-evidence-audit.mjs";
import { registerRunObservationTests } from "../diagnostics/tests/run-observation-cases.mjs";
import "../observability/tests/test-instrumentation.mjs";
import "../observability/tests/test-foundation-characterization.mjs";
import "../observability/tests/test-collection-engine.mjs";
import "../observability/tests/test-clock-placement.mjs";
import "../observability/tests/test-wrapper-receipt.mjs";
import "../output/tests/test-shell-headlines.mjs";

registerTestCatalogEvidenceTests();
registerReadinessEvidenceTests();
registerRunObservationTests();
