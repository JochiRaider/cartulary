#!/usr/bin/env node
import { collectTargetNames, collectTargetPlanRows } from "./lib/target-plan.mjs";

process.stdout.on("error", (error) => {
  if (error.code === "EPIPE") {
    process.exit(0);
  }
  throw error;
});

function usage() {
  process.stderr.write("usage: print-target-plan.mjs [--json] [--target <target>]\n");
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    json: false,
    target: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--json") {
      options.json = true;
      continue;
    }
    if (arg === "--target") {
      const target = argv[index + 1];
      if (!target) {
        usage();
      }
      options.target = target;
      index += 1;
      continue;
    }
    usage();
  }
  return options;
}

function describeSafety(row) {
  const values = [];
  if (row.check_heavy_safe) {
    values.push("check-heavy");
  }
  if (row.check_service_backed_safe) {
    values.push("check-service-backed");
  }
  if (row.check_isolated_safe) {
    values.push("check-isolated");
  }
  return values.length === 0 ? "direct-only" : values.join(", ");
}

function describeRow(row) {
  const phase = row.manifest_phase === "" ? "raw" : row.manifest_phase;
  const base = `${phase} ${row.section} ${row.coverage}`;
  const dependency =
    row.execution_dependency === "" ? "no manifest dependency" : row.execution_dependency;
  const selector = row.support_only
    ? ` support=${row.support_selector}`
    : row.raw_selector
      ? ` selector=${row.raw_selector}`
      : "";
  return [
    `  - ${row.id}: ${base}; dependency=${dependency}; shared=${row.shared_report}; safety=${describeSafety(row)}${selector}`,
    `    packages: ${row.packages.join(", ")}`,
  ];
}

function renderHuman(rows, target) {
  const targetNames = target ? [target] : [...new Set(rows.map((row) => row.target))];
  const lines = [];
  for (const name of targetNames) {
    const targetRows = rows.filter((row) => row.target === name);
    if (targetRows.length === 0) {
      continue;
    }
    const serviceBacked = targetRows.some((row) => row.service_backed) ? "yes" : "no";
    lines.push(`${name}`);
    lines.push(`  service-backed: ${serviceBacked}`);
    for (const row of targetRows) {
      lines.push(...describeRow(row));
    }
    lines.push("");
  }
  return lines.join("\n").trimEnd();
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const knownTargets = collectTargetNames();
  if (options.target && !knownTargets.includes(options.target)) {
    throw new Error(`unknown target ${options.target}; expected one of: ${knownTargets.join(", ")}`);
  }

  let rows = collectTargetPlanRows(process.cwd());
  if (options.target) {
    rows = rows.filter((row) => row.target === options.target);
  }

  if (options.json) {
    process.stdout.write(`${JSON.stringify(rows, null, 2)}\n`);
    return;
  }

  process.stdout.write(`${renderHuman(rows, options.target)}\n`);
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
