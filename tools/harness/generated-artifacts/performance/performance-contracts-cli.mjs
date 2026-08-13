#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";

import { replaceFileAtomically } from "../design-tokens/design-tokens.mjs";
import {
  loadPerformanceContracts,
  renderPerformanceContractsTypeScript,
} from "./performance-contracts.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const performancePath = path.join(
  root,
  "contracts/performance/ac043.v1.json",
);
const policyPath = path.join(root, "tools/measurement_policy_owner.json");
const outputPath = path.join(
  root,
  "packages/ui-contracts/src/generated/performance-contracts.ts",
);

try {
  const output = renderPerformanceContractsTypeScript(
    loadPerformanceContracts(performancePath, policyPath),
  );
  if (process.argv.slice(2).includes("--check")) {
    if (readFileSync(outputPath, "utf8") !== output) {
      throw new Error(
        `${path.relative(root, outputPath)} is stale; run make generate`,
      );
    }
    console.log("performance contract validation passed: generated artifact is current");
  } else {
    replaceFileAtomically(outputPath, output);
    console.log(`generated ${path.relative(root, outputPath)}`);
  }
} catch (error) {
  console.error(
    `performance contract generation failed: ${error instanceof Error ? error.message : String(error)}`,
  );
  process.exit(1);
}
