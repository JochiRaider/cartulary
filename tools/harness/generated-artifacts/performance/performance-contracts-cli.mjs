#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";

import { replaceFileAtomically } from "../design-tokens/design-tokens.mjs";
import {
  loadPerformanceContracts,
  loadPerformanceFixtureProfiles,
  renderPerformanceContractsTypeScript,
  renderPerformanceFixtureKeyVectors,
  renderPerformanceFixtureProfilesGo,
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
const fixtureCatalogOutputPath = path.join(
  root,
  "internal/gen/performancefixtureprofile/catalog_gen.go",
);
const fixtureKeyVectorsOutputPath = path.join(
  root,
  "tools/performance_fixture_snapshot_key_vectors.json",
);

try {
  const outputs = [
    {
      path: outputPath,
      content: renderPerformanceContractsTypeScript(
        loadPerformanceContracts(performancePath, policyPath),
      ),
    },
    {
      path: fixtureCatalogOutputPath,
      content: renderPerformanceFixtureProfilesGo(
        loadPerformanceFixtureProfiles(root),
      ),
    },
    {
      path: fixtureKeyVectorsOutputPath,
      content: renderPerformanceFixtureKeyVectors(
        loadPerformanceFixtureProfiles(root),
      ),
    },
  ];
  if (process.argv.slice(2).includes("--check")) {
    for (const output of outputs) {
      if (readFileSync(output.path, "utf8") !== output.content) {
        throw new Error(
          `${path.relative(root, output.path)} is stale; run make generate`,
        );
      }
    }
    console.log("performance contract validation passed: generated artifact is current");
  } else {
    for (const output of outputs) {
      replaceFileAtomically(output.path, output.content);
      console.log(`generated ${path.relative(root, output.path)}`);
    }
  }
} catch (error) {
  console.error(
    `performance contract generation failed: ${error instanceof Error ? error.message : String(error)}`,
  );
  process.exit(1);
}
