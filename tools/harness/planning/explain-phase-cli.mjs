#!/usr/bin/env node

import {
  formatPhaseCoverage,
  formatRequirements,
  phaseGuidance,
} from "./task-guidance.mjs";
import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "../frontend/frontend-phase-manifest.mjs";

function usage() {
  process.stderr.write(
    "usage: print-explain-phase.mjs --phase <phaseN|FE-PN> [--phase-namespace <base|frontend>] [--json]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = { phase: "", phaseNamespace: "base", json: false };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
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
  if (!options.phase || !["base", "frontend"].includes(options.phaseNamespace)) {
    usage();
  }
  return options;
}

function commandForNode(node) {
  return node.target_class === "public" ? `make ${node.id}` : `internal target ${node.id}`;
}

function renderWorkUnits(lines, units, indent) {
  if (!units?.length) {
    lines.push(`${indent}work units: none`);
    return;
  }
  lines.push(`${indent}work units:`);
  for (const unit of units) {
    const detail = unit.detail ? ` detail=${unit.detail}` : "";
    const stage = unit.stage ? ` stage=${unit.stage}` : "";
    lines.push(`${indent}  - ${unit.label}${stage}${detail}`);
  }
}

function renderExecutionMap(lines, executionMap) {
  lines.push("execution map:");
  if (!executionMap?.children?.length) {
    lines.push("  none");
    return;
  }
  for (const section of executionMap.children) {
    lines.push(`  ${section.label}:`);
    for (const child of section.children ?? []) {
      const targetClass =
        child.target_class && child.target_class !== "public"
          ? ` target_class=${child.target_class}`
          : "";
      lines.push(
        `    ${commandForNode(child)}${targetClass} services=${formatRequirements(child.services ?? [])} coverage=${formatPhaseCoverage(child.coverage)}`,
      );
      renderWorkUnits(lines, child.work_unit_summary, "      ");
      lines.push(`      artifacts: latest=${child.artifacts?.latest ?? "none"}`);
    }
  }
}

function claimStatusCounts(rows = []) {
  const counts = {
    implemented: 0,
    blocked: 0,
    not_applicable: 0,
    unspecified: 0,
  };
  for (const row of rows) {
    if (row.coverage !== "authoritative") {
      continue;
    }
    const status = row.claim_status ?? "";
    if (Object.hasOwn(counts, status)) {
      counts[status] += 1;
    } else {
      counts.unspecified += 1;
    }
  }
  return counts;
}

function aggregateClaimStatus(counts) {
  if (counts.blocked > 0 || counts.unspecified > 0) {
    return "incomplete";
  }
  if (counts.implemented > 0 || counts.not_applicable > 0) {
    return "complete";
  }
  return "not_applicable";
}

function formatClaimStatus(phase) {
  const counts = claimStatusCounts(phase.rows);
  return `${aggregateClaimStatus(counts)} implemented=${counts.implemented} blocked=${counts.blocked} not_applicable=${counts.not_applicable} unspecified=${counts.unspecified}`;
}

function renderHuman(phase) {
  const lines = [
    `Cartulary phase guidance: ${phase.phase}`,
    `scope: ${phase.scope || "(not declared)"}`,
    `owners: ${phase.normative_owners || "(not declared)"}`,
    `manifest: ${phase.manifest_path}`,
    `ledger: ${phase.ledger_path}`,
    `coverage: ${formatPhaseCoverage(phase)}`,
    `claim status: ${formatClaimStatus(phase)}`,
    "",
    "execution dependencies:",
  ];
  for (const dependency of phase.execution_dependencies) {
    lines.push(`  ${dependency}`);
  }
  lines.push("");
  renderExecutionMap(lines, phase.execution_map);
  return lines.join("\n");
}

function frontendPhaseGuidance(phaseID) {
  const registry = loadFrontendPhaseRegistry(process.cwd());
  const entry = registry.phases.find((candidate) => candidate.phase_id === phaseID);
  if (!entry) {
    return null;
  }
  const { manifest } = loadFrontendPhaseMap(process.cwd(), phaseID);
  return {
    schema_id: "cartulary.frontend_phase_guidance.v1",
    phase_namespace: "frontend",
    phase: phaseID,
    status: entry.status,
    manifest_path: entry.manifest_path,
    ledger_path: entry.ledger_path,
    owner_refs: entry.owner_refs,
    depends_on: entry.depends_on,
    rows: manifest.rows,
  };
}

function renderFrontendHuman(phase) {
  const ownerRefs = phase.owner_refs
    .map((owner) => `${owner.path}#${owner.section_ref}`)
    .join("; ");
  const lines = [
    `Cartulary frontend phase guidance: ${phase.phase}`,
    "namespace: frontend",
    `status: ${phase.status}`,
    `manifest: ${phase.manifest_path}`,
    `ledger: ${phase.ledger_path}`,
    `owners: ${ownerRefs}`,
    `depends on: ${phase.depends_on.length === 0 ? "none" : phase.depends_on.join(", ")}`,
    "",
    "rows:",
  ];
  for (const row of phase.rows) {
    const targets = row.targets
      .map((target) => `make ${target.target_name}`)
      .join(", ");
    lines.push(
      `  - ${row.id} evidence_class=${row.evidence_class} claim_status=${row.claim_status} targets=${targets}`,
    );
  }
  return lines.join("\n");
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.phaseNamespace === "frontend") {
    const phase = frontendPhaseGuidance(options.phase);
    if (!phase) {
      throw new Error(`unknown frontend phase ${options.phase}; expected FE-P0 through FE-P11`);
    }
    if (phase.status !== "active") {
      const message = `frontend phase ${options.phase} is ${phase.status}; planned frontend phases are explainable but not executable`;
      if (options.json) {
        process.stdout.write(`${JSON.stringify({ ...phase, executable: false, diagnostic: message }, null, 2)}\n`);
        return;
      }
      process.stdout.write(`${renderFrontendHuman(phase)}\n${message}\n`);
      return;
    }
    if (options.json) {
      process.stdout.write(`${JSON.stringify({ ...phase, executable: true }, null, 2)}\n`);
      return;
    }
    process.stdout.write(`${renderFrontendHuman(phase)}\n`);
    return;
  }
  if (options.phase.startsWith("FE-P")) {
    throw new Error("frontend phases require --phase-namespace frontend");
  }
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
