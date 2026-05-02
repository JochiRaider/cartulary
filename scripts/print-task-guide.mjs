#!/usr/bin/env node

import {
  formatPhaseCoverage,
  formatRequirements,
  knownRoles,
  taskGuide,
} from "./lib/task-guidance.mjs";
import { phaseSliceExecutionMap } from "./lib/task-execution-map.mjs";

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

function commandForNode(node) {
  return node.classification === "public" ? `make ${node.id}` : `internal target ${node.id}`;
}

function nodeExecutionSummary(node) {
  return Array.from(new Set((node.work_unit_summary ?? []).map((unit) => unit.label))).join(", ") || "none";
}

function renderExecutionSections(lines, executionMap, indent) {
  if (!executionMap?.children?.length) {
    return;
  }
  for (const section of executionMap.children) {
    lines.push(`${indent}${section.label}:`);
    for (const child of section.children ?? []) {
      const classification =
        child.classification && child.classification !== "public"
          ? ` classification=${child.classification}`
          : "";
      lines.push(
        `${indent}  - ${commandForNode(child)}${classification} services=${formatRequirements(child.services ?? [])} coverage=${formatPhaseCoverage(child.coverage)} execution=${nodeExecutionSummary(child)} artifacts=${child.artifacts?.latest ?? "none"}`,
      );
    }
  }
}

function executionMapForRecommendation(guide, item) {
  if (item.phase_relevance === "phase_slice") {
    return guide.execution_map;
  }
  if (item.phase_relevance === "service_backed_slice" && guide.phase) {
    return phaseSliceExecutionMap(guide.phase, { serviceBackedOnly: true });
  }
  return null;
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
    for (const tier of role.recommendation_tiers) {
      lines.push(`  ${tier.name}: ${tier.summary}`);
      for (const item of tier.recommendations) {
        lines.push(
          `    make ${item.target} | ${item.summary} | phase_relevance=${item.phase_relevance} | services=${formatRequirements(item.service_requirements)} | execution=${item.execution_summary || "none"} | latest_artifact=${item.latest_artifact}`,
        );
        if (item.phase_coverage) {
          lines.push(`      phase_coverage: ${formatPhaseCoverage(item.phase_coverage)}`);
        }
        const itemExecutionMap = executionMapForRecommendation(guide, item);
        if (itemExecutionMap?.kind === "phase") {
          renderExecutionSections(lines, itemExecutionMap, "      ");
        }
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
      `unknown task-guide filter; roles=${knownRoles().join(", ")} phase must be registered in tools/phase_registry.json`,
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
