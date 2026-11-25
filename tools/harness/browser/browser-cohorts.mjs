import { isDeepStrictEqual } from "node:util";
import { browserCompatibility } from "./browser-compatibility.mjs";
import { semanticJSONDigest } from "../contract/index.mjs";

// The generated group is a compatibility partition. Only selected catalog rows
// become executable cohorts, so selection never inherits an obsolete file split.
// The Go planner packs cohorts; a group runner validates the cohort its
// command names against the partition.
export function browserCohort(partition, rows, policy) {
  const ordered = [...rows].sort((a, b) => a.row_id < b.row_id ? -1 : a.row_id > b.row_id ? 1 : 0);
  const available = new Set(partition.selectedRowIDs);
  const compatibility = browserCompatibility(partition, policy);
  if (ordered.length === 0 || new Set(ordered.map((row) => row.row_id)).size !== ordered.length ||
      (partition.stateScope !== "test_owned" && ordered.length !== 1) ||
      ordered.some((row) => !available.has(row.row_id) || row.runner !== "playwright" ||
        !isDeepStrictEqual(browserCompatibility(partition, policy, row), compatibility) ||
        !partition.specs.includes(row.selector.file))) {
    throw new Error(`browser cohort violates compatibility partition ${partition.name}`);
  }
  const digest = semanticJSONDigest({
    compatibility,
    selections: ordered.map((row) => ({
      row_id: row.row_id, file: row.selector.file,
      cases: row.selector.titles.map((title, index) => ({
        title, scenario_id: row.selector.scenario_ids?.[index] ?? null,
      })).sort((a, b) => a.title < b.title ? -1 : a.title > b.title ? 1 : 0),
    })),
  }).replace(/^sha256:/u, "");
  return {
    ...partition,
    partitionName: partition.name,
    name: `cohort-${digest}`,
    selectedRowIDs: ordered.map((row) => row.row_id),
    specs: [...new Set(ordered.map((row) => row.selector.file))].sort(),
  };
}

// A test-owned cohort runs to completion: each test owns its state, so one
// failure cannot poison the next, and every failing test is reported. A cohort
// whose tests share state stops at its first failure (A27).
export function playwrightFailureArguments(cohort) {
  return cohort.stateScope === "test_owned" ? [] : ["--max-failures=1"];
}
