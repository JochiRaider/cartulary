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

function renderSchedulerPathLines(guidance) {
  const units = guidance.execution_map?.work_unit_summary ?? [];
  if (units.length === 0) {
    return ["scheduler paths: none"];
  }
  const lines = ["scheduler paths:"];
  for (const unit of units) {
    const stage = unit.stage ? ` stage=${unit.stage}` : "";
    const detail = unit.detail ? ` detail=${unit.detail}` : "";
    lines.push(`  - ${unit.label}${stage}${detail}`);
  }
  return lines;
}

function renderExpectedArtifactLines(guidance, limit = 4) {
  const expected = guidance.execution_map?.artifacts?.expected ?? guidance.artifact.expected ?? [];
  const lines = ["expected_artifacts:"];
  if (expected.length === 0) {
    lines.push("  none");
    return lines;
  }
  for (const artifact of expected.slice(0, limit)) {
    lines.push(`  ${artifact}`);
  }
  if (expected.length > limit) {
    lines.push(`  ... ${expected.length - limit} more`);
  }
  return lines;
}

function renderCheckProjectionLine(guidance) {
  const projection = guidance.check_projection;
  if (!projection) {
    return "check_projection: none";
  }
  const mode = projection.mode ? ` mode=${projection.mode}` : "";
  const schedule = projection.schedule ? ` schedule=${projection.schedule}` : "";
  const stage = projection.stage ? ` stage=${projection.stage}` : "";
  const evidence = projection.evidence ? ` evidence=${projection.evidence}` : "";
  const equivalence =
    projection.full_target_equivalent === false
      ? " full_target_equivalent=false"
      : "";
  return `check_projection:${mode}${schedule}${stage}${evidence}${equivalence}`.trimEnd();
}

function renderSummary(guidance) {
  return [
    `Cartulary target guidance: ${guidance.target}`,
    `target_class: ${guidance.target_class}`,
    `help_tier: ${guidance.help_tier ?? "none"}`,
    `default_inclusion_sets: ${guidance.default_inclusion_sets.join(",") || "none"}`,
    renderCheckProjectionLine(guidance),
    `services: ${formatRequirements(guidance.service_requirements)}`,
    `execution: ${guidance.execution_summary || "none"}`,
    ...renderSchedulerPathLines(guidance),
    `latest_artifact: ${guidance.artifact.latest?.path ?? "none"}`,
    ...renderExpectedArtifactLines(guidance),
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

function candidateTargets(target) {
  const needle = target.toLowerCase();
  return allTargetNames()
    .map((candidate) => {
      const lower = candidate.toLowerCase();
      let score = 0;
      if (lower === needle) {
        score += 100;
      }
      if (lower.startsWith(needle) || needle.startsWith(lower)) {
        score += 50;
      }
      if (lower.includes(needle) || needle.includes(lower)) {
        score += 25;
      }
      for (const part of needle.split(/[-_.]+/u).filter(Boolean)) {
        if (lower.includes(part)) {
          score += 5;
        }
      }
      return { candidate, score };
    })
    .filter((entry) => entry.score > 0)
    .sort((left, right) => right.score - left.score || left.candidate.localeCompare(right.candidate))
    .slice(0, 10)
    .map((entry) => entry.candidate);
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const guidance = targetGuidance(options.target);
  if (!guidance) {
    const candidates = candidateTargets(options.target);
    const suffix = candidates.length > 0
      ? ` nearest=${candidates.join(",")}`
      : " nearest=none";
    throw new Error(
      `unknown target ${options.target}; expected="TARGET=<target> [DETAIL=summary|rows|artifacts]"; run make help-all for the complete target list;${suffix}`,
    );
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
