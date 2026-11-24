#!/usr/bin/env node

import assert from "node:assert/strict";
import "./test-selected-evidence.mjs";
import "./test-package-operation.mjs";
import "./test-package-release.mjs";
import "./test-package-identity.mjs";
import "./test-package-archive.mjs";
import "./test-package-platform.mjs";
import "./test-package-resources.mjs";
import "./test-package-producer.mjs";
import "./test-package-consumers.mjs";
import "./test-package-deployment.mjs";
import "./test-package-deployment-lifecycle.mjs";
import "./test-package-qualification.mjs";
import path from "node:path";

import { readOperationOwner } from "../../harness/execution/operations.mjs";
import { loadTestCatalog } from "../../harness/test-catalog/index.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
// Readiness assertions have one static row and one registrar operation.
const readinessRow = "harness.release.behavior.package_readiness_contract";
assert.deepEqual(loadTestCatalog(root).rowByID.get(readinessRow).verification_ids, [
  "harness.release.verification.current_owner_evidence_only",
  "harness.release.verification.package_readiness_contract",
]);
const registrarOperations = Object.entries(readOperationOwner(root).operations)
  .filter(([, operation]) => operation.execution.command?.some(arg => arg.endsWith("/test-release-readiness-evidence.mjs")))
  .map(([id]) => id);
assert.deepEqual(registrarOperations, ["harness.assurance_release_readiness_evidence"], "one readiness registrar operation");
