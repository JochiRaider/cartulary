#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import { publicExitCodeForFailure } from "../contract/index.mjs";
import { reportCommandFailure } from "../runtime/command-failure.mjs";
import { environmentRegistryPath, inspectEnvironmentRegistry } from "./environment-registry.mjs";

const shown = 50;
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
try {
  const { findings, summary } = inspectEnvironmentRegistry(root);
  if (findings.length > 0) {
    const more = findings.length > shown ? `\n... and ${findings.length - shown} more` : "";
    throw new Error(`${findings.length} findings against ${environmentRegistryPath}:\n${findings.slice(0, shown).join("\n")}${more}`);
  }
  process.stdout.write(`environment registry: ${summary.registered} registered names and ${summary.declared_inputs} declared public inputs cover ${summary.harness_names} harness names\n`);
} catch (error) {
  const failure = reportCommandFailure(root, error, { failure_class: "config", failure_reason: "configuration_error" });
  process.stderr.write(`environment registry rejected: ${error instanceof Error ? error.message : "invalid registry"}\n`);
  process.exitCode = publicExitCodeForFailure(failure);
}
