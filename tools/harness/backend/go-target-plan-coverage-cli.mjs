#!/usr/bin/env node

import { collectTargetPlanRows } from "./backend-target-plan.mjs";
import {
  createGoTargetContext,
  inspectAggregateCommand,
} from "./backend-target-execution.mjs";
import { loadTestCatalog, targetForCatalogRow } from "../test-catalog/index.mjs";

function usage() {
  process.stderr.write(
    "usage: check-go-target-plan-coverage.mjs [--commands] [--root <repo-root>] [--quiet]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = { commands: false, quiet: false, root: process.cwd() };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--commands") options.commands = true;
    else if (arg === "--quiet") options.quiet = true;
    else if (arg === "--root") {
      options.root = argv[++index] ?? "";
      if (!options.root) usage();
    } else usage();
  }
  return options;
}

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function validateTargetPlanCoverage(root, rows) {
  const catalogRows = loadTestCatalog(root).rows.filter((row) => row.runner === "go");
  const planRows = rows.filter((row) => row.coverage !== "raw");
  const planByID = new Map();
  for (const row of planRows) {
    if (planByID.has(row.id)) {
      throw new Error(`catalog row ${row.id} appears more than once in the Go target plan`);
    }
    planByID.set(row.id, row);
  }
  for (const catalogRow of catalogRows) {
    const planRow = planByID.get(catalogRow.row_id);
    if (!planRow) {
      throw new Error(`catalog Go row ${catalogRow.row_id} is missing from the target plan`);
    }
    const expectedTarget = targetForCatalogRow(catalogRow);
    if (
      planRow.owner_id !== catalogRow.owner_id ||
      planRow.family_id !== catalogRow.family_id ||
      planRow.execution_family !== catalogRow.family_id ||
      planRow.target !== expectedTarget ||
      planRow.package !== catalogRow.selector.package ||
      JSON.stringify(planRow.symbols) !== JSON.stringify(catalogRow.selector.tests) ||
      planRow.runtime_profile_id !== catalogRow.runtime_profile_id ||
      planRow.resource_profile_id !== catalogRow.resource_profile_id ||
      planRow.fixture_profile_id !== catalogRow.fixture_profile_id ||
      planRow.default_check_required !== catalogRow.default_check
    ) {
      throw new Error(`catalog Go row ${catalogRow.row_id} target-plan projection drift`);
    }
    planByID.delete(catalogRow.row_id);
  }
  if (planByID.size > 0) {
    throw new Error(`unexpected Go target-plan row ${[...planByID.keys()].sort(compareStrings)[0]}`);
  }
  return {
    rows: catalogRows.length,
    selectors: catalogRows.reduce((total, row) => total + row.selector.tests.length, 0),
  };
}

function validateAggregateCommands(root, rows) {
  const ctx = createGoTargetContext({ repoRoot: root });
  const grouped = new Map();
  for (const row of rows.filter((entry) => entry.coverage !== "raw")) {
    const key = `${row.target}\u001f${row.execution_family}`;
    const group = grouped.get(key) ?? {
      target: row.target,
      executionFamily: row.execution_family,
      rows: [],
    };
    group.rows.push(row);
    grouped.set(key, group);
  }
  let aggregateCommands = 0;
  let commandSelectors = 0;
  for (const group of [...grouped.values()].sort(
    (left, right) =>
      compareStrings(left.target, right.target) ||
      compareStrings(left.executionFamily, right.executionFamily),
  )) {
    aggregateCommands += 1;
    const command = inspectAggregateCommand(ctx, group.target, group.executionFamily);
    for (const row of group.rows) {
      for (const symbol of row.symbols) {
        commandSelectors += 1;
        if (!command.includes(symbol)) {
          throw new Error(
            `${group.target} ${group.executionFamily} command omits ${row.id} selector ${symbol}`,
          );
        }
      }
    }
  }
  return { aggregateCommands, commandSelectors };
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const rows = collectTargetPlanRows(options.root);
  const planCounts = validateTargetPlanCoverage(options.root, rows);
  const commandCounts = options.commands
    ? validateAggregateCommands(options.root, rows)
    : null;
  if (!options.quiet) {
    const pieces = [
      `catalog_rows=${planCounts.rows}`,
      `exact_selectors=${planCounts.selectors}`,
    ];
    if (commandCounts) {
      pieces.push(
        `aggregate_commands=${commandCounts.aggregateCommands}`,
        `command_selectors=${commandCounts.commandSelectors}`,
      );
    }
    process.stdout.write(`go target-plan coverage verified: ${pieces.join(" ")}\n`);
  }
}

try {
  main();
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exit(1);
}
