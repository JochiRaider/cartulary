#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  DesignPresentationValidationError,
  loadDesignPresentationDocument,
  renderDesignPresentationTypeScript,
  replaceFileAtomically,
} from "./design-presentation.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../../..");
const registryPath = path.join(repoRoot, "contracts/design/presentation.v1.json");
const outputPath = path.join(
  repoRoot,
  "packages/ui-contracts/src/generated/design-presentation.ts",
);

try {
  const check = process.argv.slice(2).includes("--check");
  const output = renderDesignPresentationTypeScript(
    loadDesignPresentationDocument(registryPath),
  );
  if (check) {
    if (readFileSync(outputPath, "utf8") !== output) {
      throw new Error(
        `${path.relative(repoRoot, outputPath)} is stale; run make generate`,
      );
    }
    console.log("design presentation validation passed: generated artifact is current");
  } else {
    replaceFileAtomically(outputPath, output);
    console.log(`generated ${path.relative(repoRoot, outputPath)}`);
  }
} catch (error) {
  const prefix =
    error instanceof DesignPresentationValidationError
      ? "design presentation validation failed"
      : "design presentation generation failed";
  console.error(
    `${prefix}: ${error instanceof Error ? error.message : String(error)}`,
  );
  process.exit(1);
}
