#!/usr/bin/env node

import {
  collectEntries,
  collectSupportGoEntries,
  collectTargetPlanRows,
  entryIsExecutable,
  goEntrySymbols,
  loadManifest,
  phaseManifestNames,
  supportGoEntrySymbols,
} from "../planning/backend-target-plan.mjs";
import { createGoTargetContext, inspectAggregateCommand } from "./go-target-runner.mjs";

function usage() {
  process.stderr.write("usage: check-go-target-plan-coverage.mjs [--commands] [--root <repo-root>] [--quiet]\n");
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    commands: false,
    quiet: false,
    root: process.cwd(),
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--commands") {
      options.commands = true;
      continue;
    }
    if (arg === "--quiet") {
      options.quiet = true;
      continue;
    }
    if (arg === "--root") {
      const root = argv[index + 1];
      if (!root) {
        usage();
      }
      options.root = root;
      index += 1;
      continue;
    }
    usage();
  }
  return options;
}

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function fail(message) {
  throw new Error(message);
}

function collectManifestGoRows(root) {
  const authoritativeRows = [];
  const supportRows = [];
  for (const phase of phaseManifestNames(root)) {
    const { manifest } = loadManifest(root, phase);
    for (const entry of collectEntries(manifest)) {
      if (entry.coverage !== "authoritative" || entry.runner !== "go_test") {
        continue;
      }
      authoritativeRows.push({ ...entry, phase, symbols: goEntrySymbols(entry) });
    }
    for (const entry of collectSupportGoEntries(manifest)) {
      for (const symbol of supportGoEntrySymbols(entry)) {
        supportRows.push({ ...entry, phase, symbol });
      }
    }
  }
  return { authoritativeRows, supportRows };
}

function validateTargetPlanCoverage(root, rows) {
  const { authoritativeRows, supportRows } = collectManifestGoRows(root);

  for (const entry of authoritativeRows) {
    if (!entryIsExecutable(entry)) {
      continue;
    }
    const matches = rows.filter(
      (row) =>
        row.canonical_authoritative === true &&
        row.support_only === false &&
        row.coverage === "authoritative" &&
        row.id === entry.id &&
        row.manifest_phase === entry.phase,
    );
    if (matches.length !== 1) {
      fail(
        `${entry.phase} authoritative ${entry.section} row ${entry.id} must appear in exactly one canonical target-plan row, found ${matches.length}`,
      );
    }
    const row = matches[0];
    if (
      row.execution_dependency !== entry.execution_dependency ||
      row.section !== entry.section ||
      row.execution_family !== entry.execution_family ||
      row.execution_label !== entry.execution_label
    ) {
      fail(
        `${entry.phase} authoritative row ${entry.id} target-plan mismatch: expected ${entry.section}/${entry.execution_dependency}/${entry.execution_family}/${entry.execution_label}, found ${row.section}/${row.execution_dependency}/${row.execution_family}/${row.execution_label}`,
      );
    }
  }

  const authoritativeKeys = new Set(authoritativeRows.map((entry) => `${entry.phase}:${entry.id}`));
  for (const row of rows) {
    if (
      row.canonical_authoritative === true &&
      row.support_only === false &&
      row.coverage === "authoritative" &&
      !authoritativeKeys.has(`${row.manifest_phase}:${row.id}`)
    ) {
      fail(
        `target-plan row ${row.target} ${row.manifest_phase} ${row.id} is not backed by an authoritative backend manifest row`,
      );
    }
  }

  for (const entry of supportRows) {
    const matches = rows.filter(
      (row) =>
        row.support_only === true &&
        row.manifest_phase === entry.phase &&
        row.execution_dependency === entry.target &&
        row.execution_family === entry.execution_family &&
        row.execution_label === entry.execution_label &&
        row.file === entry.file &&
        row.support_selector === entry.selection_pattern &&
        Array.isArray(row.symbols) &&
        row.symbols.includes(entry.symbol),
    );
    if (matches.length !== 1) {
      fail(
        `${entry.phase} support row ${entry.file}::${entry.symbol} must appear in exactly one target-plan support row, found ${matches.length}`,
      );
    }
  }

  return {
    authoritative: authoritativeRows.length,
    support: supportRows.length,
  };
}

function rowsWithManifestSymbols(rows) {
  return rows.filter(
    (row) => row.coverage !== "raw" && Array.isArray(row.symbols) && row.symbols.length > 0,
  );
}

function validateAggregateCommands(root, rows) {
  const ctx = createGoTargetContext({ repoRoot: root });
  const grouped = new Map();
  for (const row of rowsWithManifestSymbols(rows)) {
    const key = `${row.target}\u001f${row.execution_family}`;
    if (!grouped.has(key)) {
      grouped.set(key, {
        target: row.target,
        executionFamily: row.execution_family,
        rows: [],
      });
    }
    grouped.get(key).rows.push(row);
  }

  let aggregateCount = 0;
  let symbolCount = 0;
  for (const group of Array.from(grouped.values()).sort(
    (left, right) =>
      compareStrings(left.target, right.target) ||
      compareStrings(left.executionFamily, right.executionFamily),
  )) {
    aggregateCount += 1;
    const command = inspectAggregateCommand(ctx, group.target, group.executionFamily);
    for (const row of group.rows) {
      for (const symbol of row.symbols) {
        symbolCount += 1;
        if (!command.includes(symbol)) {
          fail(
            `${group.target} ${group.executionFamily} command must include ${row.manifest_phase} ${row.id} symbol ${symbol}`,
          );
        }
      }
    }
  }

  return { aggregateCommands: aggregateCount, commandSymbols: symbolCount };
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const rows = collectTargetPlanRows(options.root);
  const planCounts = validateTargetPlanCoverage(options.root, rows);
  const commandCounts = options.commands ? validateAggregateCommands(options.root, rows) : {};
  if (!options.quiet) {
    const pieces = [
      `authoritative=${planCounts.authoritative}`,
      `support=${planCounts.support}`,
    ];
    if (options.commands) {
      pieces.push(
        `aggregate_commands=${commandCounts.aggregateCommands}`,
        `command_symbols=${commandCounts.commandSymbols}`,
      );
    }
    console.log(`go target-plan coverage verified: ${pieces.join(" ")}`);
  }
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
