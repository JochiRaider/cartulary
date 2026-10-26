import { registerGraphTests as registerWorkGraphGraphTests } from "../scheduler/work-graph/tests/test-contracts.mjs";
import "./test-catalog-metadata-fixtures.mjs";
import "./test-command-selections.mjs";
import "../scheduler/work-graph/tests/test-composition.mjs";

registerWorkGraphGraphTests();

import "../diagnostics/tests/test-canonical-target-plan.mjs";
