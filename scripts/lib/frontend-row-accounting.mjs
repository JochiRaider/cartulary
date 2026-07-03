import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "./frontend-phase-manifest.mjs";
import { flattenPlaywrightSuites } from "../../tools/harness/browser/playwright-report.mjs";
import { frontendRowAccountingSchemaID } from "../../tools/harness/core/test-output/context.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");

export const frontendRowAccountingScopeModes = new Set([
  "active_target",
  "selected_rows",
  "disabled",
]);

const frontendRowIDPattern =
  /^FE-(?:U|I|B|E|V|A11Y|S)-P(?:0|[1-9][0-9]*)-[0-9]{2}$/;

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function resolveArtifactPath(value) {
  if (!value) {
    return "";
  }
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function uniqueSorted(values) {
  return Array.from(new Set(values.filter(Boolean))).sort(compareStrings);
}

function parseRowIDs(value) {
  if (Array.isArray(value)) {
    return uniqueSorted(value.map((item) => String(item).trim()));
  }
  return uniqueSorted(
    String(value ?? "")
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean),
  );
}

function validateRowIDs(rowIDs, label) {
  for (const rowID of rowIDs) {
    if (!frontendRowIDPattern.test(rowID)) {
      throw new Error(`${label} contains invalid frontend row id ${rowID}`);
    }
  }
}

export function defaultFrontendRowAccountingScope() {
  return {
    mode: "active_target",
    invocation_kind: "standalone_target",
    phase_namespace: "",
    phase: "",
    selection_policy: "all_active_rows_for_target",
    selected_row_ids: [],
  };
}

export function normalizeFrontendRowAccountingScope(
  options = {},
  env = process.env,
) {
  const mode =
    String(
      options.mode ?? env.CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE ?? "",
    ).trim() || "active_target";
  if (!frontendRowAccountingScopeModes.has(mode)) {
    throw new Error(`invalid frontend row-accounting scope ${mode}`);
  }

  const phaseNamespace = String(
    options.phaseNamespace ??
      options.phase_namespace ??
      env.CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE_NAMESPACE ??
      "",
  ).trim();
  const phase = String(
    options.phase ?? env.CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE ?? "",
  ).trim();
  const selectedRowIDs = parseRowIDs(
    options.rowIDs ??
      options.selected_row_ids ??
      env.CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS ??
      "",
  );
  validateRowIDs(selectedRowIDs, "frontend row-accounting scope");

  if (mode === "active_target") {
    return defaultFrontendRowAccountingScope();
  }

  if (mode === "disabled") {
    if (phaseNamespace !== "base" || !/^phase[0-9]+$/.test(phase)) {
      throw new Error(
        "disabled frontend row accounting requires base phase-slice scope",
      );
    }
    if (selectedRowIDs.length > 0) {
      throw new Error(
        "disabled frontend row accounting must not declare selected row ids",
      );
    }
    return {
      mode,
      invocation_kind: "base_phase_slice",
      phase_namespace: "base",
      phase,
      selection_policy: "base_phase_no_frontend_rows",
      selected_row_ids: [],
    };
  }

  if (phaseNamespace !== "frontend" || !/^FE-P(?:0|[1-9][0-9]*)$/.test(phase)) {
    throw new Error(
      "selected frontend row accounting requires frontend phase-slice scope",
    );
  }
  return {
    mode,
    invocation_kind: "frontend_phase_slice",
    phase_namespace: "frontend",
    phase,
    selection_policy: "frontend_rows_through_selected_phase",
    selected_row_ids: selectedRowIDs,
  };
}

function frontendMapRef(phaseID) {
  return `tools/frontend_phase_maps/fe_p${phaseID.slice("FE-P".length)}_test_map.json`;
}

function phaseNumber(phaseID) {
  const match = String(phaseID).match(/^FE-P([0-9]+)$/);
  return match ? Number.parseInt(match[1], 10) : Number.NaN;
}

function rowIsInActiveTargetScope(phase, row) {
  if (phase.status === "active") {
    return true;
  }
  if (phase.status !== "planned") {
    return false;
  }
  return row.claim_status === "implemented" || row.claim_status === "stale";
}

function frontendRowsForTarget(target, scope) {
  if (scope.mode === "disabled") {
    return { rows: [], phaseMapRefs: [], explicitScope: true };
  }

  let registry;
  try {
    registry = loadFrontendPhaseRegistry(repoRoot);
  } catch {
    return {
      rows: [],
      phaseMapRefs: [],
      explicitScope: scope.mode !== "active_target",
    };
  }

  const rows = [];
  const selectedIDs = new Set(scope.selected_row_ids);
  const knownSelectedIDs = new Set();
  const mappedSelectedIDs = new Set();
  const selectedPhaseNumber = phaseNumber(scope.phase);
  for (const phase of registry.phases) {
    if (scope.mode === "selected_rows") {
      const currentPhaseNumber = phaseNumber(phase.phase_id);
      if (
        !Number.isFinite(currentPhaseNumber) ||
        currentPhaseNumber > selectedPhaseNumber
      ) {
        continue;
      }
    }

    const { manifest } = loadFrontendPhaseMap(repoRoot, phase.phase_id);
    for (const row of manifest.rows) {
      if (scope.mode === "active_target" && !rowIsInActiveTargetScope(phase, row)) {
        continue;
      }
      if (scope.mode === "selected_rows") {
        if (selectedIDs.has(row.id)) {
          knownSelectedIDs.add(row.id);
        } else {
          continue;
        }
      }

      const targetRefs = row.targets.filter(
        (targetRef) => targetRef.target_name === target,
      );
      const closingTargetRefs = targetRefs.filter(
        (targetRef) =>
          targetRef.required_for_closure ||
          targetRef.frontend_row_accounting_required,
      );
      if (closingTargetRefs.length === 0) {
        continue;
      }
      if (scope.mode === "selected_rows") {
        mappedSelectedIDs.add(row.id);
      }
      rows.push({
        phase_id: phase.phase_id,
        phase_status: phase.status,
        row_rollup_state: phase.row_rollup_state,
        row_id: row.id,
        layer: row.layer,
        evidence_class: row.evidence_class,
        claim_status: row.claim_status,
        claim: row.claim,
        blockers: row.blockers,
        required_for_closure: closingTargetRefs.some(
          (targetRef) => targetRef.required_for_closure,
        ),
        scenario_titles: row.scenario_titles,
      });
    }
  }

  if (scope.mode === "selected_rows") {
    const unknown = [...selectedIDs]
      .filter((rowID) => !knownSelectedIDs.has(rowID))
      .sort(compareStrings);
    if (unknown.length > 0) {
      throw new Error(
        `selected frontend row id(s) not found: ${unknown.join(",")}`,
      );
    }
    const unmapped = [...selectedIDs]
      .filter((rowID) => !mappedSelectedIDs.has(rowID))
      .sort(compareStrings);
    if (unmapped.length > 0) {
      throw new Error(
        `selected frontend row id(s) not mapped to ${target}: ${unmapped.join(",")}`,
      );
    }
  }

  return {
    rows,
    phaseMapRefs: uniqueSorted(rows.map((row) => frontendMapRef(row.phase_id))),
    explicitScope: scope.mode !== "active_target",
  };
}

function sha256RepoFile(relativePath) {
  const absolute = path.join(repoRoot, relativePath);
  if (!existsSync(absolute)) {
    return "";
  }
  return createHash("sha256").update(readFileSync(absolute)).digest("hex");
}

function collectVitestFileResults(report) {
  if (Array.isArray(report.testResults)) {
    return report.testResults;
  }
  if (Array.isArray(report.testResults?.testResults)) {
    return report.testResults.testResults;
  }
  return [];
}

function normalizeVitestOwnerPath(name) {
  const normalized = normalizePath(name);
  if (normalized.startsWith(repoRoot)) {
    return normalizePath(path.relative(repoRoot, normalized));
  }
  return normalized;
}

function collectVitestTitleObservations(reportFile) {
  const byTitle = new Map();
  if (!existsSync(reportFile)) {
    return byTitle;
  }
  const report = JSON.parse(readFileSync(reportFile, "utf8"));
  for (const fileResult of collectVitestFileResults(report)) {
    const ownerPath = normalizeVitestOwnerPath(fileResult.name ?? "");
    for (const assertion of fileResult.assertionResults ?? []) {
      const title = assertion.title ?? "";
      if (title === "") {
        continue;
      }
      const observations = byTitle.get(title) ?? [];
      observations.push({
        file: ownerPath,
        status: assertion.status ?? "unknown",
      });
      byTitle.set(title, observations);
    }
  }
  return byTitle;
}

function collectPhaseSummaryRunnerFiles(targetDir) {
  const runnerFiles = new Set();
  if (!existsSync(targetDir)) {
    return [];
  }
  const stack = [targetDir];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (!entry.isFile() || entry.name !== "phase-summary.json") {
        continue;
      }
      let summary;
      try {
        summary = JSON.parse(readFileSync(next, "utf8"));
      } catch {
        continue;
      }
      const runnerJSON = summary.artifacts?.runner_json;
      if (typeof runnerJSON === "string" && runnerJSON.trim() !== "") {
        runnerFiles.add(resolveArtifactPath(runnerJSON));
      }
    }
  }
  return [...runnerFiles].sort();
}

function normalizePlaywrightFile(file) {
  const normalized = normalizePath(file);
  if (normalized.startsWith("apps/web/")) {
    return normalized;
  }
  return normalizePath(path.join("apps/web", "e2e", normalized));
}

function phaseSummaryTitleObservationFile(value) {
  const file = typeof value === "string" ? value : "";
  if (file === "") {
    return "";
  }
  return file.startsWith("apps/web/")
    ? normalizePath(file)
    : normalizePlaywrightFile(file);
}

function collectPhaseSummaryTitleObservations(targetDir) {
  const byTitle = new Map();
  if (!existsSync(targetDir)) {
    return byTitle;
  }
  const stack = [targetDir];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (!entry.isFile() || entry.name !== "phase-summary.json") {
        continue;
      }
      let summary;
      try {
        summary = JSON.parse(readFileSync(next, "utf8"));
      } catch {
        continue;
      }
      for (const item of summary.inventory ?? []) {
        const title = item.symbol_or_title ?? "";
        if (title === "") {
          continue;
        }
        const observations = byTitle.get(title) ?? [];
        observations.push({
          file: phaseSummaryTitleObservationFile(item.package_or_file),
          status: "passed",
        });
        byTitle.set(title, observations);
      }
      for (const failure of summary.failures ?? []) {
        const title = failure.symbol_or_title ?? "";
        if (title === "") {
          continue;
        }
        const observations = byTitle.get(title) ?? [];
        observations.push({
          file: phaseSummaryTitleObservationFile(failure.package_or_file),
          status: "failed",
        });
        byTitle.set(title, observations);
      }
    }
  }
  return byTitle;
}

function normalizePlaywrightScenarioStatus(status) {
  switch (status) {
    case "passed":
    case "flaky":
      return "passed";
    case "skipped":
      return "skipped";
    case "failed":
    case "timedOut":
    case "interrupted":
      return "failed";
    default:
      return "unknown";
  }
}

function playwrightSpecScenarioStatus(spec) {
  const statuses = [];
  for (const test of spec.tests ?? []) {
    const results = test.results ?? [];
    if (results.length === 0) {
      const testStatus = normalizePlaywrightScenarioStatus(test.status ?? "");
      if (testStatus !== "unknown") {
        statuses.push(testStatus);
      }
      continue;
    }
    for (const result of results) {
      statuses.push(normalizePlaywrightScenarioStatus(result.status ?? ""));
    }
  }
  if (statuses.length === 0) {
    return "unknown";
  }
  if (statuses.some((status) => status === "failed")) {
    return "failed";
  }
  if (statuses.some((status) => status === "passed")) {
    return "passed";
  }
  if (statuses.every((status) => status === "skipped")) {
    return "skipped";
  }
  return "unknown";
}

function collectPlaywrightTitleObservations(reportFile) {
  const byTitle = new Map();
  if (!existsSync(reportFile)) {
    return byTitle;
  }
  const parsed = JSON.parse(readFileSync(reportFile, "utf8"));
  const reports = Array.isArray(parsed) ? parsed : [parsed];
  for (const report of reports) {
    for (const spec of flattenPlaywrightSuites(report.suites ?? [])) {
      const title = spec.title ?? "";
      if (title === "") {
        continue;
      }
      const observations = byTitle.get(title) ?? [];
      observations.push({
        file: normalizePlaywrightFile(spec.file ?? ""),
        status: playwrightSpecScenarioStatus(spec),
      });
      byTitle.set(title, observations);
    }
  }
  return byTitle;
}

function mergeTitleObservations(target, source) {
  for (const [title, observations] of source.entries()) {
    target.set(title, [...(target.get(title) ?? []), ...observations]);
  }
}

function collectPlaywrightTitleObservationsForTarget(targetDir) {
  const byTitle = collectPhaseSummaryTitleObservations(targetDir);
  for (const runnerFile of collectPhaseSummaryRunnerFiles(targetDir)) {
    mergeTitleObservations(
      byTitle,
      collectPlaywrightTitleObservations(runnerFile),
    );
  }
  return byTitle;
}

function frontendScenarioStatus(observations) {
  if (!observations || observations.length === 0) {
    return "missing";
  }
  if (observations.some((entry) => entry.status === "failed")) {
    return "failed";
  }
  if (observations.some((entry) => entry.status === "passed")) {
    return "passed";
  }
  if (observations.every((entry) => entry.status === "skipped")) {
    return "skipped";
  }
  return "unknown";
}

function frontendRowClosureStatus(row, targetStatus) {
  if (row.scenarios.length === 0) {
    if (
      row.claim_status === "implemented" &&
      (row.evidence_class === "implementation_support" ||
        row.evidence_class === "claim_publication_boundary")
    ) {
      return targetStatus === "pass" ? "closed" : "blocked_by_target";
    }
    return "not_evaluable";
  }
  if (row.scenarios.some((scenario) => scenario.status === "failed")) {
    return "failed";
  }
  if (
    row.scenarios.some((scenario) =>
      ["missing", "skipped", "unknown"].includes(scenario.status),
    )
  ) {
    return "missing";
  }
  if (targetStatus !== "pass") {
    return "blocked_by_target";
  }
  return "closed";
}

function frontendRowClosureStatusV3(row) {
  if (row.claim_status === "blocked") {
    return "blocked";
  }
  if (row.claim_status === "stale") {
    return "stale";
  }
  if (!row.required_for_closure) {
    return "not_applicable";
  }
  if (row.closure_status === "closed") {
    return "closed";
  }
  return "not_closed";
}

function frontendRowFailureReason(row) {
  if (row.closure_status === "closed" || row.claim_status === "blocked") {
    return "";
  }
  if (row.closure_status === "missing") {
    return "missing_required_scenario";
  }
  if (row.closure_status === "failed") {
    return "failed_required_scenario";
  }
  if (row.closure_status === "blocked_by_target") {
    return "target_failed";
  }
  if (row.closure_status === "not_evaluable") {
    return "row_not_evaluable";
  }
  return "unknown";
}

export function frontendRowAccountingForTarget(
  target,
  targetStatus,
  targetDir,
  { scope: rawScope = null } = {},
) {
  const scope = normalizeFrontendRowAccountingScope(rawScope ?? {});
  const frontendRows = frontendRowsForTarget(target, scope);
  if (frontendRows.rows.length === 0 && !frontendRows.explicitScope) {
    return null;
  }

  let titleObservations = new Map();
  if (target === "frontend-unit") {
    titleObservations = collectVitestTitleObservations(
      path.join(targetDir, "raw", "frontend-unit", "runner.json"),
    );
  } else if (target.startsWith("browser-e2e")) {
    titleObservations = collectPlaywrightTitleObservationsForTarget(targetDir);
  }

  const rows = frontendRows.rows.map((row) => {
    const scenarios = row.scenario_titles.map((title) => {
      const observations = titleObservations.get(title) ?? [];
      return {
        title,
        status: frontendScenarioStatus(observations),
        files: [...new Set(observations.map((entry) => entry.file))].sort(),
      };
    });
    const output = {
      ...row,
      target,
      target_status: targetStatus,
      scenarios,
    };
    return {
      ...output,
      closure_status: frontendRowClosureStatus(output, targetStatus),
    };
  });

  const scenarioStatuses = rows.flatMap((row) =>
    row.scenarios.map((scenario) => scenario.status),
  );
  const scenarioResults = rows.flatMap((row) =>
    row.scenarios.map((scenario) => ({
      scenario_title: scenario.title,
      status: scenario.status,
      row_ids: [row.row_id],
      artifact_refs: scenario.files,
    })),
  );
  const rowResults = rows.map((row) => ({
    row_id: row.row_id,
    phase_id: row.phase_id,
    evidence_class: row.evidence_class,
    claim_status_at_run: row.claim_status,
    target_mapping_status:
      row.claim_status === "blocked" ? "blocked" : "mapped",
    closure_status: frontendRowClosureStatusV3(row),
    closing_scenario_titles: row.scenarios
      .filter((scenario) => scenario.status === "passed")
      .map((scenario) => scenario.title),
    failure_reason: frontendRowFailureReason(row),
  }));
  return {
    schema_id: frontendRowAccountingSchemaID,
    target_name: target,
    command_id: `cartulary.harness.command.${target.replaceAll("-", "_")}.v1`,
    phase_namespace: "frontend",
    accounting_scope: scope,
    registry_ref: "tools/frontend_phase_registry.json",
    registry_digest: sha256RepoFile("tools/frontend_phase_registry.json"),
    guide_ref: "docs/guides/cartulary_frontend_implementation_testing_guide.md",
    guide_digest: sha256RepoFile(
      "docs/guides/cartulary_frontend_implementation_testing_guide.md",
    ),
    phase_map_refs: frontendRows.phaseMapRefs,
    phase_map_digests: frontendRows.phaseMapRefs.map((phaseMapRef) =>
      sha256RepoFile(phaseMapRef),
    ),
    run_root: normalizePath(path.relative(repoRoot, targetDir)),
    target_status: targetStatus,
    scenario_results: scenarioResults,
    row_results: rowResults,
    rollup: {
      implemented: rows.filter((row) => row.claim_status === "implemented")
        .length,
      blocked: rows.filter((row) => row.claim_status === "blocked").length,
      missing: rowResults.filter((row) => row.closure_status === "not_closed")
        .length,
      stale: rowResults.filter((row) => row.closure_status === "stale").length,
      not_applicable: rowResults.filter(
        (row) => row.closure_status === "not_applicable",
      ).length,
      closed: rowResults.filter((row) => row.closure_status === "closed")
        .length,
      failed: rowResults.filter(
        (row) =>
          row.closure_status === "not_closed" &&
          row.failure_reason !== "missing_required_scenario",
      ).length,
    },
    target,
    rows,
    counts: {
      rows: rows.length,
      scenarios: scenarioStatuses.length,
      closed_rows: rows.filter((row) => row.closure_status === "closed").length,
      blocked_by_target_rows: rows.filter(
        (row) => row.closure_status === "blocked_by_target",
      ).length,
      failed_rows: rows.filter((row) => row.closure_status === "failed").length,
      missing_rows: rows.filter((row) => row.closure_status === "missing")
        .length,
      not_evaluable_rows: rows.filter(
        (row) => row.closure_status === "not_evaluable",
      ).length,
      passed_scenarios: scenarioStatuses.filter((status) => status === "passed")
        .length,
      failed_scenarios: scenarioStatuses.filter((status) => status === "failed")
        .length,
      missing_scenarios: scenarioStatuses.filter(
        (status) => status === "missing",
      ).length,
      skipped_scenarios: scenarioStatuses.filter(
        (status) => status === "skipped",
      ).length,
      unknown_scenarios: scenarioStatuses.filter(
        (status) => status === "unknown",
      ).length,
    },
  };
}

export function frontendRowAccountingFailures(accounting) {
  if (!accounting) {
    return [];
  }
  return accounting.rows
    .filter(
      (row) =>
        row.claim_status === "implemented" &&
        (row.scenario_titles.length > 0 ||
          row.evidence_class === "implementation_support") &&
        row.closure_status !== "closed",
    )
    .map((row) => ({
      failure_class: "harness",
      failure_reason: "frontend_row_accounting",
      kind: "failure",
      source: "frontend-row-accounting",
      target: accounting.target,
      phase: row.phase_id,
      row_id: row.row_id,
      message: `${row.row_id} implemented frontend row did not close: ${row.closure_status}`,
    }));
}

export function appendFrontendRowAccountingFailures(
  section,
  failures,
  { normalizeCounts, failureFieldsForJSON },
) {
  if (failures.length === 0) {
    return;
  }
  section.status = "fail";
  section.counts = normalizeCounts(section.counts);
  section.counts.failed += failures.length;
  section.counts.non_test += failures.length;
  section.counts.non_test_failed += failures.length;
  const failureFields = failureFieldsForJSON(
    [...(section.failures ?? []), ...failures],
    section.counts,
  );
  Object.assign(section, failureFields);
}
