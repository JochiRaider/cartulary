#!/usr/bin/env node

import {
  formatPhaseCoverage,
  formatRequirements,
  phaseGuidance,
} from "./lib/task-guidance.mjs";

function usage() {
  process.stderr.write("usage: print-explain-phase.mjs --phase <phaseN> [--json]\n");
  process.exit(2);
}

function parseArgs(argv) {
  const options = { phase: "", json: false };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--phase") {
      options.phase = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--json") {
      options.json = true;
      continue;
    }
    usage();
  }
  if (!options.phase) {
    usage();
  }
  return options;
}

function renderHuman(phase) {
  const lines = [
    `Cartulary phase guidance: ${phase.phase}`,
    `scope: ${phase.scope || "(not declared)"}`,
    `owners: ${phase.normative_owners || "(not declared)"}`,
    `manifest: ${phase.manifest_path}`,
    `ledger: ${phase.ledger_path}`,
    `coverage: ${formatPhaseCoverage(phase)}`,
    "",
    "execution dependencies:",
  ];
  for (const dependency of phase.execution_dependencies) {
    lines.push(`  ${dependency}`);
  }
  lines.push("");
  lines.push("targets:");
  for (const target of phase.targets) {
    const command =
      target.classification === "public"
        ? `make ${target.target}`
        : `internal target ${target.target}`;
    const classification =
      target.classification && target.classification !== "public"
        ? ` classification=${target.classification}`
        : "";
    lines.push(
      `  ${command}${classification} services=${formatRequirements(target.service_requirements)} scheduler=${target.scheduler_owner.join(";")} coverage=${formatPhaseCoverage({ ...target.counts, phases: [phase.phase], execution_dependencies: target.execution_dependencies })}`,
    );
  }
  return lines.join("\n");
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const phase = phaseGuidance(options.phase);
  if (!phase) {
    throw new Error(`unknown phase ${options.phase}; expected one of tools/phase_registry.json`);
  }
  if (options.json) {
    process.stdout.write(`${JSON.stringify(phase, null, 2)}\n`);
    return;
  }
  process.stdout.write(`${renderHuman(phase)}\n`);
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
