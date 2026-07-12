#!/usr/bin/env node
import { existsSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "../phase-accounting/index.mjs";

const accessibilitySummarySchemaID =
  "cartulary.frontend_accessibility_summary.v3";
const contrastThreshold = 4.5;

function parseArgs(argv) {
  const options = {
    output: "",
    status: "pass",
    phaseDir: "",
    contrastDir: "",
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
    if (arg === "--contrast-dir") {
      options.contrastDir = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    throw new Error(`unknown option ${arg}`);
  }
  if (!options.output || !["pass", "fail"].includes(options.status)) {
    throw new Error("usage: write-frontend-accessibility-summary.mjs --output <path> --status <pass|fail> [--phase-dir <path>] [--contrast-dir <path>]");
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

function normalizeScenarioStatus(status) {
  if (status === "pass" || status === "fail" || status === "skipped") {
    return status;
  }
  return "missing";
}

function escapeRegExp(value) {
  return value.replace(/[\\^$.*+?()[\]{}|]/g, "\\$&");
}

function scenarioCoverageTitle(rowId, title) {
  return title
    .replace(new RegExp(`^${escapeRegExp(rowId)}\\s+`), "")
    .replace(/\.$/, "");
}

function scenarioRowsForAccessibility(manifest) {
  return manifest.rows.filter((candidate) => {
    if (!candidate.id.startsWith("FE-A11Y-")) {
      return false;
    }
    return (
      candidate.claim_status === "implemented" &&
      candidate.targets.some(
        (target) => target.target_name === "browser-e2e-a11y",
      )
    );
  });
}

function loadContrastRecords(contrastDir) {
  if (!contrastDir || !existsSync(contrastDir)) {
    return new Map();
  }
  const records = new Map();
  for (const entry of readdirSync(contrastDir, { withFileTypes: true })) {
    if (!entry.isFile() || !entry.name.endsWith(".json")) {
      continue;
    }
    const file = path.join(contrastDir, entry.name);
    const parsed = JSON.parse(readFileSync(file, "utf8"));
    const title = String(parsed.scenario_title ?? "");
    const checks = Array.isArray(parsed.checks) ? parsed.checks : [];
    if (!title) {
      continue;
    }
    records.set(title, checks);
  }
  return records;
}

function contrastSummaryForScenario(scenario, contrastRecords) {
  const checks = contrastRecords.get(scenario.title) ?? [];
  if (scenario.status !== "pass") {
    return {
      row_id: scenario.row_id,
      title: scenario.title,
      result: "fail",
      coverage: `${scenarioCoverageTitle(scenario.row_id, scenario.title)}: contrast not proven because scenario ${scenario.status}`,
      target: "scenario",
      ratio: 0,
      threshold: contrastThreshold,
      foreground: "unknown",
      background: "unknown",
    };
  }
  if (checks.length === 0) {
    return {
      row_id: scenario.row_id,
      title: scenario.title,
      result: "fail",
      coverage: `${scenarioCoverageTitle(scenario.row_id, scenario.title)}: no retained contrast assertion`,
      target: "scenario",
      ratio: 0,
      threshold: contrastThreshold,
      foreground: "unknown",
      background: "unknown",
    };
  }
  const worst = checks
    .map((check) => ({
      target: String(check.target ?? "scenario"),
      ratio: Number(check.ratio ?? 0),
      threshold: Number(check.threshold ?? contrastThreshold),
      foreground: String(check.foreground ?? "unknown"),
      background: String(check.background ?? "unknown"),
    }))
    .sort((left, right) => left.ratio - right.ratio)[0];
  return {
    row_id: scenario.row_id,
    title: scenario.title,
    result: checks.every((check) => check.result === "pass") ? "pass" : "fail",
    coverage: `${scenarioCoverageTitle(scenario.row_id, scenario.title)}: WCAG 2.2 AA contrast assertion`,
    target: worst.target,
    ratio: Number(worst.ratio.toFixed(2)),
    threshold: worst.threshold,
    foreground: worst.foreground,
    background: worst.background,
  };
}

function collectRowsAndScenarios(options) {
  const scenarioStatuses = loadScenarioStatuses(options.phaseDir);
  const registry = loadFrontendPhaseRegistry(process.cwd());
  const phaseRows = [];
  const scenarios = [];
  for (const phase of registry.phases) {
    const { manifest } = loadFrontendPhaseMap(process.cwd(), phase.phase_id);
    for (const row of scenarioRowsForAccessibility(manifest)) {
      phaseRows.push({
        row_id: row.id,
        phase_id: phase.phase_id,
        evidence_class: row.evidence_class,
        claim_status: row.claim_status,
        targets: row.targets,
      });
      for (const title of row.scenario_titles) {
        const scenarioStatus = normalizeScenarioStatus(
          scenarioStatuses.get(title) ??
            (options.status === "pass" ? "missing" : "fail"),
        );
        scenarios.push({
          row_id: row.id,
          title,
          status: scenarioStatus,
        });
      }
    }
  }
  return { phaseRows, scenarios };
}

function artifactRefs(options) {
  const runRoot = path.resolve(path.dirname(options.output), "../..");
  return [
    {
      role: "playwright_phase",
      path_kind: "directory",
      path: path.relative(runRoot, options.phaseDir).replaceAll("\\", "/"),
    },
    {
      role: "contrast_checks",
      path_kind: "directory",
      path: path.relative(runRoot, options.contrastDir).replaceAll("\\", "/"),
    },
  ].filter((entry) => entry.path !== "");
}

function buildEvidenceSummary(options) {
  const { phaseRows, scenarios } = collectRowsAndScenarios(options);
  const blockingScenarios = scenarios.filter(
    (scenario) => scenario.status !== "pass",
  );
  const scenarioChecks = scenarios.map((scenario) => ({
    row_id: scenario.row_id,
    title: scenario.title,
    result: scenario.status === "pass" ? "pass" : "fail",
    coverage: scenarioCoverageTitle(scenario.row_id, scenario.title),
  }));
  const contrastRecords = loadContrastRecords(options.contrastDir);
  const contrastChecks = scenarios.map((scenario) =>
    contrastSummaryForScenario(scenario, contrastRecords),
  );
  const failingContrastChecks = contrastChecks.filter(
    (check) => check.result !== "pass",
  );
  const summaryStatus =
    options.status === "pass" &&
    blockingScenarios.length === 0 &&
    failingContrastChecks.length === 0
      ? "pass"
      : "fail";

  const summary = {
    schema_id: accessibilitySummarySchemaID,
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
    contrast_checks: contrastChecks,
    violations: [
      ...blockingScenarios.map((scenario) => ({
        severity: "blocking",
        row_id: scenario.row_id,
        title: scenario.title,
        message: `Playwright accessibility scenario ${scenario.status}; inspect retained artifacts.`,
      })),
      ...failingContrastChecks.map((check) => ({
        severity: "blocking",
        row_id: check.row_id,
        title: check.title,
        message: `Accessibility contrast check failed for ${check.target}: ratio ${check.ratio} below threshold ${check.threshold}.`,
      })),
    ],
    artifact_refs: artifactRefs(options),
  };
  return summary;
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const summary = buildEvidenceSummary(options);

  validateSchemaSync(summary.schema_id, summary);
  writeFileSync(options.output, `${JSON.stringify(summary, null, 2)}\n`, "utf8");
  if (options.status === "pass" && summary.status !== "pass") {
    throw new Error(`frontend accessibility summary status is ${summary.status}`);
  }
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`frontend accessibility summary failed: ${message}`);
  process.exit(1);
}
