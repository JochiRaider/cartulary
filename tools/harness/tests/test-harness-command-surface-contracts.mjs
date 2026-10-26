import { registerCommandSurfaceTests as registerWorkGraphCommandSurfaceTests } from "../scheduler/work-graph/tests/test-contracts.mjs";
import { registerCommandSurfaceTests as registerCommandSurfaceCommandSurfaceTests } from "../command-surface/tests/test-contracts.mjs";
import { registerCommandSurfaceTests as registerTestCatalogCommandSurfaceTests } from "../test-catalog/tests/test-contracts.mjs";
import "../contract/tests/test-cleanup.mjs";
import { registerRunStatusCommandTests } from "../diagnostics/tests/run-observation-cases.mjs";
import { registerFrontendProducerGraphTests } from "./test-frontend-producer-graph.mjs";
import { registerFrontendProducerLifecycleTests } from "./test-frontend-producer-lifecycle.mjs";
import { registerCommandFailureTests } from "./test-command-failure.mjs";
import "../command-surface/tests/test-public-inputs.mjs";

registerWorkGraphCommandSurfaceTests();
registerCommandSurfaceCommandSurfaceTests();
registerTestCatalogCommandSurfaceTests();
registerFrontendProducerGraphTests();
registerFrontendProducerLifecycleTests();
registerCommandFailureTests();
registerRunStatusCommandTests();
