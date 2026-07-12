import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";
import {
  frontendEvidenceAuditInputForTarget,
  frontendEvidenceAuditRootForTarget,
} from "./audit-routing.mjs";
import { sha256File } from "./freshness.mjs";
import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "./registry-loader.mjs";

export const frontendEvidenceAuditSummarySchemaID =
  "cartulary.frontend_evidence_audit_summary.v1";

const rowAccountingSchemaID = "cartulary.frontend_row_accounting.v5";

function relToRepo(repoRoot, file) {
  const relative = path.relative(repoRoot, file).replaceAll("\\", "/");
  return relative.startsWith("../") || path.isAbsolute(relative)
    ? file.replaceAll("\\", "/")
    : relative;
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function resolveInputPath(repoRoot, value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function targetDirForRoot(repoRoot, root, targetName) {
  const resolved = resolveInputPath(repoRoot, root);
  if (existsSync(path.join(resolved, "frontend-row-accounting.json"))) {
    return resolved;
  }
  return path.join(resolved, targetName);
}

function readTargetSummary(repoRoot, targetDir, targetName, failures) {
  for (const filename of ["tool-run-summary.json", "target-summary.json"]) {
    const file = path.join(targetDir, filename);
    if (!existsSync(file)) {
      failures.push(`${relToRepo(repoRoot, file)} is required for ${targetName}`);
      continue;
    }
    const summary = readJSON(file);
    if (summary.status !== "pass") {
      failures.push(
        `${relToRepo(repoRoot, file)} must record status=pass for ${targetName}`,
      );
    }
  }
}

function rowAccountingForTarget({
  repoRoot,
  root,
  targetName,
  digests,
  failures,
}) {
  const targetDir = targetDirForRoot(repoRoot, root, targetName);
  if (!existsSync(targetDir)) {
    failures.push(`${relToRepo(repoRoot, targetDir)} is required for ${targetName}`);
    return null;
  }
  readTargetSummary(repoRoot, targetDir, targetName, failures);
  const file = path.join(targetDir, "frontend-row-accounting.json");
  if (!existsSync(file)) {
    failures.push(`${relToRepo(repoRoot, file)} is required for ${targetName}`);
    return null;
  }
  const accounting = readJSON(file);
  try {
    validateSchemaSync(rowAccountingSchemaID, accounting);
  } catch (error) {
    failures.push(
      `${relToRepo(repoRoot, file)} failed ${rowAccountingSchemaID} validation: ${
        error instanceof Error ? error.message : String(error)
      }`,
    );
    return null;
  }
  if (accounting.target_name !== targetName) {
    failures.push(`${relToRepo(repoRoot, file)} target_name must be ${targetName}`);
  }
  if (accounting.target_status !== "pass") {
    failures.push(`${relToRepo(repoRoot, file)} target_status must be pass`);
  }
  if (accounting.registry_digest !== digests.registry) {
    failures.push(`${relToRepo(repoRoot, file)} registry_digest is stale`);
  }
  if (accounting.guide_digest !== digests.guide) {
    failures.push(`${relToRepo(repoRoot, file)} guide_digest is stale`);
  }
  const phaseMapIndex = accounting.phase_map_refs.indexOf(digests.phaseMapRef);
  if (phaseMapIndex < 0) {
    failures.push(
      `${relToRepo(repoRoot, file)} does not reference ${digests.phaseMapRef}`,
    );
  } else if (accounting.phase_map_digests[phaseMapIndex] !== digests.phaseMap) {
    failures.push(
      `${relToRepo(repoRoot, file)} ${digests.phaseMapRef} digest is stale`,
    );
  }
  return {
    file,
    targetDir,
    accounting,
  };
}

function passedScenarioTitles(accounting, rowID) {
  return new Set(
    (accounting.scenario_results ?? [])
      .filter(
        (scenario) =>
          (scenario.row_ids ?? []).includes(rowID) &&
          scenario.status === "passed",
      )
      .map((scenario) => scenario.scenario_title),
  );
}

function auditRowTarget({
  repoRoot,
  row,
  targetRef,
  accounting,
  accountingFile,
  failures,
}) {
  const rowResult = (accounting.row_results ?? []).find(
    (entry) => entry.row_id === row.id && entry.phase_id === row.phase_id,
  );
  if (!rowResult) {
    failures.push(
      `${relToRepo(repoRoot, accountingFile)} is missing row_result for ${row.id}`,
    );
    return {
      row_id: row.id,
      target_name: targetRef.target_name,
      status: "missing",
      accounting: relToRepo(repoRoot, accountingFile),
    };
  }
  if (rowResult.evidence_class !== row.evidence_class) {
    failures.push(
      `${relToRepo(repoRoot, accountingFile)} ${row.id} evidence_class must be ${row.evidence_class}`,
    );
  }
  if (rowResult.claim_status_at_run !== row.claim_status) {
    failures.push(
      `${relToRepo(repoRoot, accountingFile)} ${row.id} claim_status_at_run must be ${row.claim_status}`,
    );
  }
  if (rowResult.closure_status !== "closed") {
    failures.push(
      `${relToRepo(repoRoot, accountingFile)} ${row.id} closure_status must be closed`,
    );
  }
  if (targetRef.scenario_title_required) {
    const passedTitles = passedScenarioTitles(accounting, row.id);
    for (const title of row.scenario_titles) {
      if (!passedTitles.has(title)) {
        failures.push(
          `${relToRepo(repoRoot, accountingFile)} ${row.id} missing passed scenario title: ${title}`,
        );
      }
    }
  }
  return {
    row_id: row.id,
    target_name: targetRef.target_name,
    status: rowResult.closure_status,
    evidence_class: rowResult.evidence_class,
    accounting: relToRepo(repoRoot, accountingFile),
  };
}

export function buildFrontendEvidenceAuditSummary({
  repoRoot,
  phaseNamespace,
  phaseID,
  roots,
  failures: initialFailures = [],
}) {
  const failures = [...initialFailures];
  let summary = {
    schema_id: frontendEvidenceAuditSummarySchemaID,
    status: "fail",
    phase_namespace: phaseNamespace,
    phase_id: phaseID,
    roots,
    digests: {},
    targets: [],
    rows: [],
    failures,
  };

  if (failures.length === 0) {
    const registry = loadFrontendPhaseRegistry(repoRoot);
    const phase = registry.phases.find((entry) => entry.phase_id === phaseID);
    if (!phase) {
      failures.push(`unknown frontend phase ${phaseID}`);
    } else if (
      phase.status !== "active" ||
      phase.row_rollup_state !== "active_green"
    ) {
      failures.push(`${phaseID} must be active and active_green`);
    }

    if (failures.length === 0) {
      const { manifest, registryEntry } = loadFrontendPhaseMap(repoRoot, phaseID);
      const digests = {
        registry: sha256File(repoRoot, "tools/frontend_phase_registry.json"),
        guide: sha256File(repoRoot, registry.guide_path),
        phaseMapRef: registryEntry.manifest_path,
        phaseMap: sha256File(repoRoot, registryEntry.manifest_path),
      };
      summary.digests = digests;

      const requiredTargets = new Map();
      for (const row of manifest.rows.filter(
        (entry) => entry.claim_status === "implemented",
      )) {
        for (const targetRef of row.targets.filter(
          (entry) => entry.required_for_closure,
        )) {
          requiredTargets.set(targetRef.target_name, targetRef);
        }
      }

      const accountingByTarget = new Map();
      for (const targetName of [...requiredTargets.keys()].sort()) {
        const root = frontendEvidenceAuditRootForTarget(targetName, roots);
        if (!root) {
          const inputName = frontendEvidenceAuditInputForTarget(targetName);
          failures.push(
            inputName
              ? `${inputName} is required because ${targetName} is required for closure`
              : `${targetName} has no retained root input mapping`,
          );
          continue;
        }
        const accounting = rowAccountingForTarget({
          repoRoot,
          root,
          targetName,
          digests,
          failures,
        });
        if (accounting) {
          accountingByTarget.set(targetName, accounting);
          summary.targets.push({
            target_name: targetName,
            root,
            target_dir: relToRepo(repoRoot, accounting.targetDir),
            accounting: relToRepo(repoRoot, accounting.file),
          });
        }
      }

      for (const row of manifest.rows.filter(
        (entry) => entry.claim_status === "implemented",
      )) {
        for (const targetRef of row.targets.filter(
          (entry) => entry.required_for_closure,
        )) {
          const accounting = accountingByTarget.get(targetRef.target_name);
          if (!accounting) {
            summary.rows.push({
              row_id: row.id,
              target_name: targetRef.target_name,
              status: "missing_accounting",
            });
            continue;
          }
          summary.rows.push(
            auditRowTarget({
              repoRoot,
              row,
              targetRef,
              accounting: accounting.accounting,
              accountingFile: accounting.file,
              failures,
            }),
          );
        }
      }
    }
  }

  summary = {
    ...summary,
    status: failures.length === 0 ? "pass" : "fail",
    failures,
  };
  return summary;
}

export function relativeFrontendEvidenceAuditPath(repoRoot, file) {
  return relToRepo(repoRoot, file);
}
