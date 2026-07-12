import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { frontendEvidenceAuditInputForTarget } from "./frontend/audit-routing.mjs";
import { loadFrontendPhaseRegistry } from "./frontend/registry-loader.mjs";
import {
  frontendTargetHasClosureRows,
  frontendRowsForAccountingTarget,
  parseFrontendRowIDs,
  validateFrontendRowIDs,
} from "./frontend/row-scope.mjs";
import {
  collectPlaywrightTitleObservationsForTarget,
  collectVitestTitleObservations,
  frontendScenarioStatus,
} from "../output/test-output/frontend-row-evidence.mjs";
import { frontendRowAccountingSchemaID } from "../contract/test-output-context.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");

export const frontendRowAccountingScopeModes = new Set([
  "active_target",
  "selected_rows",
  "disabled",
]);

function normalizePath(value) {
  return value.replaceAll("\\", "/");
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
  const selectedRowIDs = parseFrontendRowIDs(
    options.rowIDs ??
      options.selected_row_ids ??
      env.CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS ??
      "",
  );
  validateFrontendRowIDs(
    selectedRowIDs,
    (rowID) => `frontend row-accounting scope contains invalid frontend row id ${rowID}`,
  );

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

function frontendRowAccountingBoundTarget(options = {}, env = process.env) {
  return String(
    options.accountingTarget ??
      options.accounting_target ??
      env.CARTULARY_FRONTEND_ROW_ACCOUNTING_TARGET ??
      "",
  ).trim();
}

function frontendAccountingOwnerDataRequiredForTarget(target, scope) {
  if (scope.mode === "disabled") {
    return false;
  }
  if (scope.mode === "selected_rows") {
    return true;
  }
  return frontendEvidenceAuditInputForTarget(target) !== "";
}

function frontendRowsForTarget(target, scope, root) {
  if (scope.mode === "disabled") {
    return { rows: [], phaseMapRefs: [], explicitScope: true };
  }

  let registry;
  try {
    registry = loadFrontendPhaseRegistry(root);
  } catch (error) {
    if (frontendAccountingOwnerDataRequiredForTarget(target, scope)) {
      const message = error instanceof Error ? error.message : String(error);
      throw new Error(`frontend row accounting owner data failed to load: ${message}`);
    }
    return {
      rows: [],
      phaseMapRefs: [],
      explicitScope: scope.mode !== "active_target",
    };
  }

  return frontendRowsForAccountingTarget({
    root,
    registry,
    target,
    scope,
  });
}

function sha256RepoFile(root, relativePath) {
  const absolute = path.join(root, relativePath);
  if (!existsSync(absolute)) {
    return "";
  }
  return createHash("sha256").update(readFileSync(absolute)).digest("hex");
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

function frontendRowAccountingClosureStatus(row) {
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
  { scope: rawScope = null, root = repoRoot } = {},
) {
  const normalizedRoot = path.resolve(root);
  const scopeOptions = rawScope ?? {};
  const scope = normalizeFrontendRowAccountingScope(scopeOptions);
  const accountingTarget = frontendRowAccountingBoundTarget(scopeOptions);
  if (
    scope.mode === "selected_rows" &&
    accountingTarget !== "" &&
    accountingTarget !== target
  ) {
    const registry = loadFrontendPhaseRegistry(normalizedRoot);
    if (
      frontendTargetHasClosureRows({
        root: normalizedRoot,
        registry,
        target,
      })
    ) {
      throw new Error(
        `selected frontend row-accounting scope is bound to ${accountingTarget} and cannot be consumed by evidence target ${target}`,
      );
    }
    return null;
  }
  const frontendRows = frontendRowsForTarget(target, scope, normalizedRoot);
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
    closure_status: frontendRowAccountingClosureStatus(row),
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
    registry_digest: sha256RepoFile(normalizedRoot, "tools/frontend_phase_registry.json"),
    guide_ref: "docs/guides/cartulary_frontend_implementation_testing_guide.md",
    guide_digest: sha256RepoFile(
      normalizedRoot,
      "docs/guides/cartulary_frontend_implementation_testing_guide.md",
    ),
    phase_map_refs: frontendRows.phaseMapRefs,
    phase_map_digests: frontendRows.phaseMapRefs.map((phaseMapRef) =>
      sha256RepoFile(normalizedRoot, phaseMapRef),
    ),
    run_root: normalizePath(path.relative(normalizedRoot, targetDir)),
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
  };
}

export function frontendRowAccountingFailures(accounting) {
  if (!accounting) {
    return [];
  }
  return (accounting.row_results ?? [])
    .filter(
      (row) =>
        row.claim_status_at_run === "implemented" &&
        row.closure_status !== "closed" &&
        row.closure_status !== "not_applicable",
    )
    .map((row) => ({
      failure_class: "harness",
      failure_reason: "frontend_row_accounting",
      kind: "failure",
      source: "frontend-row-accounting",
      target: accounting.target_name,
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
