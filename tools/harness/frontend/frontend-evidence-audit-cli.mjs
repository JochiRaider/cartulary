#!/usr/bin/env node
import { createHash } from "node:crypto";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "./frontend-phase-manifest.mjs";
import { validateSchemaSync } from "../core/harness-contract.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const schemaID = "cartulary.frontend_evidence_audit_summary.v1";
const rowAccountingSchemaID = "cartulary.frontend_row_accounting.v3";

const checkRootTargets = new Set([
  "frontend-unit",
  "browser-e2e-webserver-backed",
  "browser-e2e-stateful",
]);

const explicitTargetInputs = new Map([
  ["browser-e2e-support", "BROWSER_SUPPORT_RESULTS_DIR"],
  ["browser-e2e-visual", "BROWSER_VISUAL_RESULTS_DIR"],
  ["browser-e2e-a11y", "BROWSER_A11Y_RESULTS_DIR"],
]);

function relToRepo(file) {
  const relative = path.relative(repoRoot, file).replaceAll("\\", "/");
  return relative.startsWith("../") || path.isAbsolute(relative)
    ? file.replaceAll("\\", "/")
    : relative;
}

function sha256File(relativePath) {
  return createHash("sha256")
    .update(readFileSync(path.join(repoRoot, relativePath)))
    .digest("hex");
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function input(name) {
  return String(process.env[name] ?? "").trim();
}

function requireInput(name, failures) {
  const value = input(name);
  if (value === "") {
    failures.push(`${name} is required`);
  }
  return value;
}

function resolveInputPath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function targetDirForRoot(root, targetName) {
  const resolved = resolveInputPath(root);
  if (existsSync(path.join(resolved, "frontend-row-accounting.json"))) {
    return resolved;
  }
  return path.join(resolved, targetName);
}

function readTargetSummary(targetDir, targetName, failures) {
  for (const filename of ["tool-run-summary.json", "target-summary.json"]) {
    const file = path.join(targetDir, filename);
    if (!existsSync(file)) {
      failures.push(`${relToRepo(file)} is required for ${targetName}`);
      continue;
    }
    const summary = readJSON(file);
    if (summary.status !== "pass") {
      failures.push(`${relToRepo(file)} must record status=pass for ${targetName}`);
    }
  }
}

function rowAccountingForTarget({ root, targetName, digests, failures }) {
  const targetDir = targetDirForRoot(root, targetName);
  if (!existsSync(targetDir)) {
    failures.push(`${relToRepo(targetDir)} is required for ${targetName}`);
    return null;
  }
  readTargetSummary(targetDir, targetName, failures);
  const file = path.join(targetDir, "frontend-row-accounting.json");
  if (!existsSync(file)) {
    failures.push(`${relToRepo(file)} is required for ${targetName}`);
    return null;
  }
  const accounting = readJSON(file);
  try {
    validateSchemaSync(rowAccountingSchemaID, accounting);
  } catch (error) {
    failures.push(
      `${relToRepo(file)} failed ${rowAccountingSchemaID} validation: ${
        error instanceof Error ? error.message : String(error)
      }`,
    );
    return null;
  }
  if (accounting.target_name !== targetName) {
    failures.push(`${relToRepo(file)} target_name must be ${targetName}`);
  }
  if (accounting.target_status !== "pass") {
    failures.push(`${relToRepo(file)} target_status must be pass`);
  }
  if (accounting.registry_digest !== digests.registry) {
    failures.push(`${relToRepo(file)} registry_digest is stale`);
  }
  if (accounting.guide_digest !== digests.guide) {
    failures.push(`${relToRepo(file)} guide_digest is stale`);
  }
  const phaseMapIndex = accounting.phase_map_refs.indexOf(digests.phaseMapRef);
  if (phaseMapIndex < 0) {
    failures.push(`${relToRepo(file)} does not reference ${digests.phaseMapRef}`);
  } else if (accounting.phase_map_digests[phaseMapIndex] !== digests.phaseMap) {
    failures.push(`${relToRepo(file)} ${digests.phaseMapRef} digest is stale`);
  }
  return {
    file,
    targetDir,
    accounting,
  };
}

function rootForTarget(targetName, roots) {
  if (checkRootTargets.has(targetName)) {
    return roots.CHECK_RESULTS_DIR;
  }
  const inputName = explicitTargetInputs.get(targetName);
  return inputName ? roots[inputName] : "";
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

function auditRowTarget({ row, targetRef, accounting, accountingFile, failures }) {
  const rowResult = (accounting.row_results ?? []).find(
    (entry) => entry.row_id === row.id && entry.phase_id === row.phase_id,
  );
  if (!rowResult) {
    failures.push(
      `${relToRepo(accountingFile)} is missing row_result for ${row.id}`,
    );
    return {
      row_id: row.id,
      target_name: targetRef.target_name,
      status: "missing",
      accounting: relToRepo(accountingFile),
    };
  }
  if (rowResult.evidence_class !== row.evidence_class) {
    failures.push(
      `${relToRepo(accountingFile)} ${row.id} evidence_class must be ${row.evidence_class}`,
    );
  }
  if (rowResult.claim_status_at_run !== row.claim_status) {
    failures.push(
      `${relToRepo(accountingFile)} ${row.id} claim_status_at_run must be ${row.claim_status}`,
    );
  }
  if (rowResult.closure_status !== "closed") {
    failures.push(
      `${relToRepo(accountingFile)} ${row.id} closure_status must be closed`,
    );
  }
  if (targetRef.scenario_title_required) {
    const passedTitles = passedScenarioTitles(accounting, row.id);
    for (const title of row.scenario_titles) {
      if (!passedTitles.has(title)) {
        failures.push(
          `${relToRepo(accountingFile)} ${row.id} missing passed scenario title: ${title}`,
        );
      }
    }
  }
  return {
    row_id: row.id,
    target_name: targetRef.target_name,
    status: rowResult.closure_status,
    evidence_class: rowResult.evidence_class,
    accounting: relToRepo(accountingFile),
  };
}

function commandSpecificSummaryPath() {
  const dir =
    process.env.CARTULARY_PHASE_ARTIFACT_DIR ||
    path.join(repoRoot, ".cartulary", "test-results", "frontend-evidence-audit-direct");
  mkdirSync(dir, { recursive: true });
  return path.join(dir, "frontend-evidence-audit-summary.json");
}

function writeSummary(summary) {
  const file = commandSpecificSummaryPath();
  validateSchemaSync(schemaID, summary);
  writeFileSync(file, `${JSON.stringify(summary, null, 2)}\n`);
  return file;
}

function main() {
  const failures = [];
  const phaseNamespace = requireInput("PHASE_NAMESPACE", failures);
  const phaseID = requireInput("PHASE", failures);
  const roots = {
    CHECK_RESULTS_DIR: requireInput("CHECK_RESULTS_DIR", failures),
    BROWSER_SUPPORT_RESULTS_DIR: requireInput("BROWSER_SUPPORT_RESULTS_DIR", failures),
    BROWSER_VISUAL_RESULTS_DIR: requireInput("BROWSER_VISUAL_RESULTS_DIR", failures),
    BROWSER_A11Y_RESULTS_DIR: requireInput("BROWSER_A11Y_RESULTS_DIR", failures),
  };

  if (phaseNamespace && phaseNamespace !== "frontend") {
    failures.push("PHASE_NAMESPACE must be frontend");
  }
  if (phaseID && !/^FE-P(?:0|[1-9][0-9]*)$/.test(phaseID)) {
    failures.push("PHASE must be FE-P<N>");
  }

  let summary = {
    schema_id: schemaID,
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
    } else if (phase.status !== "active" || phase.row_rollup_state !== "active_green") {
      failures.push(`${phaseID} must be active and active_green`);
    }
    const { manifest, registryEntry } = loadFrontendPhaseMap(repoRoot, phaseID);
    const digests = {
      registry: sha256File("tools/frontend_phase_registry.json"),
      guide: sha256File(registry.guide_path),
      phaseMapRef: registryEntry.manifest_path,
      phaseMap: sha256File(registryEntry.manifest_path),
    };
    summary.digests = digests;

    const requiredTargets = new Map();
    for (const row of manifest.rows.filter((entry) => entry.claim_status === "implemented")) {
      for (const targetRef of row.targets.filter((entry) => entry.required_for_closure)) {
        requiredTargets.set(targetRef.target_name, targetRef);
      }
    }

    const accountingByTarget = new Map();
    for (const targetName of [...requiredTargets.keys()].sort()) {
      const root = rootForTarget(targetName, roots);
      if (!root) {
        failures.push(`${targetName} has no retained root input mapping`);
        continue;
      }
      const accounting = rowAccountingForTarget({
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
          target_dir: relToRepo(accounting.targetDir),
          accounting: relToRepo(accounting.file),
        });
      }
    }

    for (const row of manifest.rows.filter((entry) => entry.claim_status === "implemented")) {
      for (const targetRef of row.targets.filter((entry) => entry.required_for_closure)) {
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

  summary = {
    ...summary,
    status: failures.length === 0 ? "pass" : "fail",
    failures,
  };
  const summaryFile = writeSummary(summary);
  if (summary.status === "fail") {
    process.stderr.write(
      `frontend evidence audit failed; summary=${relToRepo(summaryFile)}\n${failures.join("\n")}\n`,
    );
    process.exit(1);
  }
  process.stdout.write(
    `frontend evidence audit passed phase=${phaseID} summary=${relToRepo(summaryFile)}\n`,
  );
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  const summaryFile = writeSummary({
    schema_id: schemaID,
    status: "fail",
    phase_namespace: input("PHASE_NAMESPACE"),
    phase_id: input("PHASE"),
    roots: {},
    digests: {},
    targets: [],
    rows: [],
    failures: [message],
  });
  process.stderr.write(
    `frontend evidence audit failed; summary=${relToRepo(summaryFile)}\n${message}\n`,
  );
  process.exit(1);
}
