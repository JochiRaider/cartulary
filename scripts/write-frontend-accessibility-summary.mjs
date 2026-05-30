#!/usr/bin/env node
import { existsSync, readFileSync, writeFileSync } from "node:fs";
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

function runnerPathForPhaseDir(phaseDir) {
  if (!phaseDir) {
    return "";
  }
  return path.join(path.resolve(phaseDir), "runner.json");
}

function flattenSpecsFromSuite(suite) {
  const specs = Array.isArray(suite?.specs) ? suite.specs : [];
  const suites = Array.isArray(suite?.suites) ? suite.suites : [];
  return [
    ...specs,
    ...suites.flatMap((child) => flattenSpecsFromSuite(child)),
  ];
}

function statusForSpec(spec) {
  if (spec?.ok === true) {
    return "pass";
  }
  const tests = Array.isArray(spec?.tests) ? spec.tests : [];
  if (
    tests.some((test) =>
      Array.isArray(test?.results)
        ? test.results.some((result) => result?.status === "failed")
        : false,
    )
  ) {
    return "fail";
  }
  if (tests.some((test) => test?.status === "skipped")) {
    return "skipped";
  }
  return "fail";
}

function loadScenarioStatuses(phaseDir) {
  const runnerPath = runnerPathForPhaseDir(phaseDir);
  if (!runnerPath || !existsSync(runnerPath)) {
    return new Map();
  }
  const runner = JSON.parse(readFileSync(runnerPath, "utf8"));
  const suites = Array.isArray(runner?.suites) ? runner.suites : [];
  return new Map(
    suites
      .flatMap((suite) => flattenSpecsFromSuite(suite))
      .map((spec) => [String(spec.title ?? ""), statusForSpec(spec)]),
  );
}

function scenarioCoverageTitle(rowId, title) {
  if (rowId === "FE-A11Y-P1-01") {
    return title
      .replace(/^FE-A11Y-P1-01\s+/, "")
      .replace(/\.$/, "");
  }
  return "phase accessibility smoke";
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const scenarioStatuses = loadScenarioStatuses(options.phaseDir);
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
        const scenarioStatus =
          scenarioStatuses.get(title) ?? (options.status === "pass" ? "missing" : "fail");
        scenarios.push({
          row_id: row.id,
          title,
          status: scenarioStatus,
        });
      }
    }
  }
  const blockingScenarios = scenarios.filter(
    (scenario) => scenario.status !== "pass",
  );
  const summaryStatus =
    options.status === "pass" && blockingScenarios.length === 0
      ? "pass"
      : "fail";
  const scenarioChecks = scenarios.map((scenario) => ({
    row_id: scenario.row_id,
    title: scenario.title,
    result: scenario.status === "pass" ? "pass" : "fail",
    coverage: scenarioCoverageTitle(scenario.row_id, scenario.title),
  }));

  const summary = {
    schema_id: "cartulary.frontend_accessibility_summary.v1",
    status: summaryStatus,
    phase_rows: phaseRows,
    scenarios,
    keyboard_matrix: scenarioChecks.map((check) => ({
      ...check,
      coverage: `${check.coverage}: keyboard reachability, visible focus, and focus-trap check`,
    })),
    state_communication_checks: scenarioChecks.map((check) => ({
      ...check,
      coverage: `${check.coverage}: accessible names, live/status exposure, non-color-only cues, and safe public error text`,
    })),
    contrast_checks: scenarioChecks.map((check) => ({
      ...check,
      coverage: `${check.coverage}: visible focus and state-bearing text contrast smoke`,
    })),
    violations: blockingScenarios.map((scenario) => ({
      severity: "blocking",
      row_id: scenario.row_id,
      title: scenario.title,
      message: `Playwright accessibility scenario ${scenario.status}; inspect retained artifacts.`,
    })),
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
