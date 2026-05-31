#!/usr/bin/env node

import {
  formatPhaseCoverage,
  formatRequirements,
  knownRoles,
  taskGuide,
} from "./lib/task-guidance.mjs";
import { loadFrontendPhaseRegistry } from "./lib/frontend-phase-manifest.mjs";
import { phaseSliceExecutionMap } from "./lib/task-execution-map.mjs";

process.stdout.on("error", (error) => {
  if (error.code === "EPIPE") {
    process.exit(0);
  }
  throw error;
});

function usage() {
  process.stderr.write(
    "usage: print-task-guide.mjs [--role <role>] [--phase <phaseN|FE-PN>] [--phase-namespace <base|frontend>] [--json]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = { role: "", phase: "", phaseNamespace: "base", json: false };
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
    if (arg === "--phase-namespace") {
      options.phaseNamespace = argv[index + 1] ?? "";
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
  return node.target_class === "public" ? `make ${node.id}` : `internal target ${node.id}`;
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
      const targetClass =
        child.target_class && child.target_class !== "public"
          ? ` target_class=${child.target_class}`
          : "";
      lines.push(
        `${indent}  - ${commandForNode(child)}${targetClass} services=${formatRequirements(child.services ?? [])} coverage=${formatPhaseCoverage(child.coverage)} execution=${nodeExecutionSummary(child)} artifacts=${child.artifacts?.latest ?? "none"}`,
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
  if (options.phaseNamespace === "frontend") {
    const registry = loadFrontendPhaseRegistry(process.cwd());
    const phase = registry.phases.find((entry) => entry.phase_id === options.phase);
    if (!phase) {
      throw new Error("unknown frontend phase; expected FE-P0 through FE-P11");
    }
    const guide = {
      schema_id: "cartulary.frontend_task_guide.v1",
      role: options.role || "feature-dev",
      phase_namespace: "frontend",
      phase: phase.phase_id,
      status: phase.status,
      recommendations:
        phase.status === "active"
          ? [
              `make phase-slice PHASE_NAMESPACE=frontend PHASE=${phase.phase_id}`,
              `make service-backed-slice PHASE_NAMESPACE=frontend PHASE=${phase.phase_id}`,
              `make explain-phase PHASE_NAMESPACE=frontend PHASE=${phase.phase_id}`,
              "make phase-ledger-drift",
            ]
          : [
              `make explain-phase PHASE_NAMESPACE=frontend PHASE=${phase.phase_id}`,
              "make phase-ledger-drift",
            ],
    };
    if (options.json) {
      process.stdout.write(`${JSON.stringify(guide, null, 2)}\n`);
      return;
    }
    const lines = [
      "Cartulary task guide",
      `role=${guide.role} phase_namespace=frontend phase=${phase.phase_id}`,
      `status=${phase.status}`,
    ];
    if (phase.status === "active") {
      lines.push(
        `  make phase-slice PHASE_NAMESPACE=frontend PHASE=${phase.phase_id} | run active frontend phase row targets`,
        `  make service-backed-slice PHASE_NAMESPACE=frontend PHASE=${phase.phase_id} | run active browser-backed frontend row targets`,
      );
    }
    lines.push(
      `  make explain-phase PHASE_NAMESPACE=frontend PHASE=${phase.phase_id} | inspect frontend phase rows`,
      "  make phase-ledger-drift | verify frontend ledgers and base ledgers",
    );
    process.stdout.write(`${lines.join("\n")}\n`);
    return;
  }
  if (options.phase.startsWith("FE-P")) {
    throw new Error("frontend phases require PHASE_NAMESPACE=frontend");
  }
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
