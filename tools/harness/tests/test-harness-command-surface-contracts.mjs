import "../browser/tests/test-ui-review-contract.mjs";
import { runContractSuite } from "./contract-suite-support.mjs";
import { registerFrontendProducerGraphTests } from "./test-frontend-producer-graph.mjs";
import { registerFrontendProducerLifecycleTests } from "./test-frontend-producer-lifecycle.mjs";
import { registerCommandFailureTests } from "./test-command-failure.mjs";
registerFrontendProducerGraphTests();
registerFrontendProducerLifecycleTests();
registerCommandFailureTests();
runContractSuite("command_surface");
