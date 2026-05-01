#!/usr/bin/env node

import {
  allTargetNames,
  formatPhaseCoverage,
  formatRequirements,
  targetGuidance,
} from "./lib/task-guidance.mjs";

const validDetails = new Set(["summary", "rows", "artifacts"]);

function usage() {
  process.stderr.write(
    "usage: print-explain-target.mjs --target <target> [--detail summary|rows|artifacts] [--json]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = { target: "", detail: "summary", json: false };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") {
      options.target = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--detail") {
      options.detail = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--json") {
      options.json = true;
      continue;
    }
    usage();
  }
  if (!options.target || !validDetails.has(options.detail)) {
    usage();
  }
  return options;
}

function renderSummary(guidance) {
  return [
    `Cartulary target guidance: ${guidance.target}`,
    `classification: ${guidance.classification}`,
    `help_tier: ${guidance.help_tier ?? "none"}`,
    `included_in: ${guidance.included_in.join(",") || "none"}`,
    `services: ${formatRequirements(guidance.service_requirements)}`,
    `scheduler: ${guidance.scheduler_owner.join(";")}`,
    `latest_artifact: ${guidance.artifact.latest?.path ?? "none"}`,
    `phase_coverage: ${formatPhaseCoverage(guidance.phase_coverage)}`,
  ];
}

function renderRows(guidance) {
  const lines = [...renderSummary(guidance), "", "rows:"];
  const rows = guidance.go_rows.length > 0 ? guidance.go_rows : guidance.rows;
  if (rows.length === 0) {
    lines.push("  none");
    return lines;
  }
  for (const row of rows) {
    const phase = row.manifest_phase ?? row.phase ?? "raw";
    const coverage = row.coverage ?? "raw";
    const section = row.section ?? "";
    const dependency = row.execution_dependency || "none";
    const runner = row.runner_family ?? row.runner ?? "";
    const packages = row.packages?.join(",") || row.package || "";
    const file = row.file ? ` file=${row.file}` : "";
    lines.push(
      `  - ${row.id}: ${phase} ${section} ${coverage} dependency=${dependency} runner=${runner} packages=${packages}${file}`,
    );
  }
  return lines;
}

function renderArtifacts(guidance) {
  const lines = [...renderSummary(guidance), "", "artifacts:"];
  lines.push(`  latest: ${guidance.artifact.latest?.path ?? "none"}`);
  lines.push("  discovered:");
  const candidates = guidance.artifact.candidates ?? [];
  if (candidates.length === 0) {
    lines.push("    none");
  } else {
    for (const artifact of candidates.slice(0, 10)) {
      const label = artifact.label ? ` label=${artifact.label}` : "";
      const status = artifact.status ? ` status=${artifact.status}` : "";
      lines.push(`    ${artifact.kind}: ${artifact.path}${label}${status}`);
    }
  }
  lines.push("  expected:");
  for (const artifact of guidance.artifact.expected) {
    lines.push(`    ${artifact}`);
  }
  return lines;
}

function renderHuman(guidance, detail) {
  if (detail === "rows") {
    return renderRows(guidance).join("\n");
  }
  if (detail === "artifacts") {
    return renderArtifacts(guidance).join("\n");
  }
  return renderSummary(guidance).join("\n");
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const guidance = targetGuidance(options.target);
  if (!guidance) {
    throw new Error(`unknown target ${options.target}; expected one of: ${allTargetNames().join(", ")}`);
  }
  if (options.json) {
    process.stdout.write(`${JSON.stringify(guidance, null, 2)}\n`);
    return;
  }
  process.stdout.write(`${renderHuman(guidance, options.detail)}\n`);
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
