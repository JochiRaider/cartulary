#!/usr/bin/env node

import {
  formatPhaseCoverage,
  formatRequirements,
  knownRoles,
  taskGuide,
} from "./lib/task-guidance.mjs";

process.stdout.on("error", (error) => {
  if (error.code === "EPIPE") {
    process.exit(0);
  }
  throw error;
});

function usage() {
  process.stderr.write("usage: print-task-guide.mjs [--role <role>] [--phase <phaseN>] [--json]\n");
  process.exit(2);
}

function parseArgs(argv) {
  const options = { role: "", phase: "", json: false };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--role") {
      options.role = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
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
  return options;
}

function renderHuman(guide) {
  const lines = ["Cartulary task guide", `role=${guide.role} phase=${guide.phase || "all"}`, ""];
  if (guide.phase_guidance) {
    lines.push(
      `phase focus: ${guide.phase_guidance.phase} ${guide.phase_guidance.scope || "(scope not declared)"}`,
    );
    lines.push(`phase coverage: ${formatPhaseCoverage(guide.phase_guidance)}`);
    lines.push("");
  }
  for (const role of guide.roles) {
    lines.push(`${role.role}: ${role.summary}`);
    for (const item of role.recommendations) {
      lines.push(
        `  make ${item.target} | ${item.summary} | services=${formatRequirements(item.service_requirements)} | scheduler=${item.scheduler_owner.join(";")} | latest_artifact=${item.latest_artifact}`,
      );
      if (item.phase_coverage) {
        lines.push(`    phase_coverage: ${formatPhaseCoverage(item.phase_coverage)}`);
      }
    }
    lines.push("");
  }
  lines.push(`roles: ${knownRoles().join(", ")}`);
  lines.push("use ROLE=<role> and PHASE=phaseN to narrow this view");
  return lines.join("\n").trimEnd();
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const guide = taskGuide(options);
  if (!guide) {
    throw new Error(
      `unknown task-guide filter; roles=${knownRoles().join(", ")} phase must match tools/phase*_test_map.json`,
    );
  }
  if (options.json) {
    process.stdout.write(`${JSON.stringify(guide, null, 2)}\n`);
    return;
  }
  process.stdout.write(`${renderHuman(guide)}\n`);
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
