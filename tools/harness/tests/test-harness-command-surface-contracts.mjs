import { runContractSuite } from "./contract-suite-support.mjs";
import "./test-frontend-producer-graph.mjs";
import "./test-frontend-producer-lifecycle.mjs";
import "./test-command-failure.mjs";
runContractSuite("command_surface");
