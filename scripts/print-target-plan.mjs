#!/usr/bin/env node
import { collectTargetNames, collectTargetPlanRows } from "./lib/target-plan.mjs";

process.stdout.on("error", (error) => {
  if (error.code === "EPIPE") {
    process.exit(0);
  }
  throw error;
});

function usage() {
  process.stderr.write("usage: print-target-plan.mjs [--json] [--detail] [--target <target>]\n");
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    detail: false,
    json: false,
    target: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--json") {
      options.json = true;
      continue;
    }
    if (arg === "--detail") {
      options.detail = true;
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

function aggregateTargetRows(rows) {
  const byTarget = new Map();
  for (const row of rows) {
    if (!byTarget.has(row.target)) {
      byTarget.set(row.target, {
        target: row.target,
        service_backed: false,
        rows: 0,
        authoritative: 0,
        support: 0,
        raw: 0,
        phases: new Set(),
        shared_reports: new Set(),
        safety: new Set(),
      });
    }
    const aggregate = byTarget.get(row.target);
    aggregate.service_backed = aggregate.service_backed || row.service_backed;
    aggregate.rows += 1;
    if (row.coverage === "authoritative") {
      aggregate.authoritative += 1;
    } else if (row.coverage === "support") {
      aggregate.support += 1;
    } else if (row.coverage === "raw") {
      aggregate.raw += 1;
    }
    if (row.manifest_phase) {
      aggregate.phases.add(row.manifest_phase);
    }
    if (row.shared_report) {
      aggregate.shared_reports.add(row.shared_report);
    }
    if (row.check_heavy_safe) {
      aggregate.safety.add("check-heavy");
    }
    if (row.check_service_backed_safe) {
      aggregate.safety.add("check-service-backed");
    }
    if (row.check_isolated_safe) {
      aggregate.safety.add("check-isolated");
    }
  }
  return Array.from(byTarget.values()).sort((left, right) => left.target.localeCompare(right.target));
}

function renderCompact(rows, target) {
  const aggregates = aggregateTargetRows(rows);
  const targetNames = target ? [target] : aggregates.map((aggregate) => aggregate.target);
  const lines = [];
  for (const name of targetNames) {
    const aggregate = aggregates.find((candidate) => candidate.target === name);
    if (!aggregate) {
      continue;
    }
    const safety = aggregate.safety.size === 0 ? "direct-only" : Array.from(aggregate.safety).sort().join(",");
    lines.push(
      `${aggregate.target} service_backed=${aggregate.service_backed ? 1 : 0} phases=${aggregate.phases.size} rows=${aggregate.rows} authoritative=${aggregate.authoritative} support=${aggregate.support} raw=${aggregate.raw} safety=${safety} shared_reports=${aggregate.shared_reports.size}`,
    );
  }
  return lines.join("\n");
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

  const rendered = options.detail ? renderHuman(rows, options.target) : renderCompact(rows, options.target);
  process.stdout.write(`${rendered}\n`);
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
