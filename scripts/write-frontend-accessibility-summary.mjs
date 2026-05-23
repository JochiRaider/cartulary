#!/usr/bin/env node
import { writeFileSync } from "node:fs";
import path from "node:path";

import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "./lib/frontend-phase-manifest.mjs";

function parseArgs(argv) {
  const options = {
    output: "",
    status: "pass",
    phaseDir: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--output") {
      options.output = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--status") {
      options.status = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--phase-dir") {
      options.phaseDir = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    throw new Error(`unknown option ${arg}`);
  }
  if (!options.output || !["pass", "fail"].includes(options.status)) {
    throw new Error("usage: write-frontend-accessibility-summary.mjs --output <path> --status <pass|fail> [--phase-dir <path>]");
  }
  return options;
}

function relToRepo(value) {
  if (!value) {
    return "";
  }
  return path.relative(process.cwd(), path.resolve(value)).replaceAll(path.sep, "/");
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const registry = loadFrontendPhaseRegistry(process.cwd());
  const phaseRows = [];
  const scenarios = [];
  for (const phase of registry.phases) {
    const { manifest } = loadFrontendPhaseMap(process.cwd(), phase.phase_id);
    for (const row of manifest.rows.filter((candidate) => candidate.id.startsWith("FE-A11Y-"))) {
      phaseRows.push({
        row_id: row.id,
        phase_id: phase.phase_id,
        evidence_class: row.evidence_class,
        claim_status: row.claim_status,
        targets: row.targets,
      });
      for (const title of row.scenario_titles) {
        scenarios.push({
          row_id: row.id,
          title,
          status: options.status,
        });
      }
    }
  }

  const summary = {
    schema_id: "cartulary.frontend_accessibility_summary.v1",
    status: options.status,
    phase_rows: phaseRows,
    scenarios,
    keyboard_matrix: phaseRows.map((row) => ({
      row_id: row.row_id,
      result: options.status,
      coverage: "keyboard reachability and visible focus smoke",
    })),
    state_communication_checks: phaseRows.map((row) => ({
      row_id: row.row_id,
      result: options.status,
      coverage: "accessible names and non-color-only state smoke",
    })),
    contrast_checks: phaseRows.map((row) => ({
      row_id: row.row_id,
      result: options.status,
      coverage: "computed contrast hook reserved for browser assertions",
    })),
    violations:
      options.status === "pass"
        ? []
        : [
            {
              severity: "blocking",
              message: "Playwright accessibility scenario failed; inspect retained artifacts.",
            },
          ],
    artifact_refs: [
      {
        kind: "playwright_phase",
        path: relToRepo(options.phaseDir),
      },
    ].filter((entry) => entry.path !== ""),
  };

  writeFileSync(options.output, `${JSON.stringify(summary, null, 2)}\n`, "utf8");
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`frontend accessibility summary failed: ${message}`);
  process.exit(1);
}
