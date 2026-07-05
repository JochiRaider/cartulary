import {
  assertObjectKeys,
  assertUnique,
  requireBoolean,
  requireEnum,
  requireInteger,
  requireObjectArray,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "../../contract/json-shape.mjs";
import { frontendEvidenceAuditInputForTarget } from "./audit-routing.mjs";
import { requirePhaseID } from "./common.mjs";
import {
  commandIDPattern,
  frontendPhaseNamespace,
  frontendPhaseTestMapSchemaID,
  mapKeys,
  rowIDPattern,
  rowKeys,
  targetRefKeys,
  validClaimStatuses,
  validDefaultCheckKinds,
  validDefaultCheckReasonCodes,
  validEvidenceClasses,
  validEvidenceRoles,
  validLayers,
  validWarmLocalCostClasses,
} from "./constants.mjs";
import { validateBlocker, validateOwnerRef } from "./owner-refs.mjs";
import {
  validateClaim,
  validateCore03SortingFilteringGroupingOwnerRefs,
  validateFrontendBrowserScenarioTitleOwnership,
  validateRowMetadata,
  validateVisualAccessibilityEvidenceBoundary,
} from "./row-validation.mjs";

function validateTargetRef(target, label, row, targetEntriesByName = null) {
  assertObjectKeys(target, targetRefKeys, label);
  requireString(target.target_name, `${label}.target_name`);
  requireString(target.command_id, `${label}.command_id`, {
    pattern: commandIDPattern,
  });
  requireEnum(target.evidence_role, `${label}.evidence_role`, validEvidenceRoles);
  requireBoolean(target.required_for_closure, `${label}.required_for_closure`);
  requireBoolean(
    target.frontend_row_accounting_required,
    `${label}.frontend_row_accounting_required`,
  );
  requireBoolean(
    target.scenario_title_required,
    `${label}.scenario_title_required`,
  );
  if (targetEntriesByName) {
    const targetEntry = targetEntriesByName.get(target.target_name);
    if (!targetEntry) {
      throw new Error(
        `${label}.target_name must reference a task-surface target: ${target.target_name}`,
      );
    }
    if (targetEntry.command_id !== target.command_id) {
      throw new Error(
        `${label}.command_id must match task-surface ${target.target_name} command_id ${targetEntry.command_id}`,
      );
    }
  }
  if (
    row.claim_status === "implemented" &&
    target.required_for_closure &&
    !frontendEvidenceAuditInputForTarget(target.target_name)
  ) {
    throw new Error(
      `${label}.target_name ${target.target_name} is required for implemented row closure but has no frontend-evidence-audit retained-root route`,
    );
  }
  if (
    row.claim_status === "implemented" &&
    !["implementation_support", "claim_publication_boundary"].includes(
      row.evidence_class,
    ) &&
    target.required_for_closure &&
    target.scenario_title_required !== true
  ) {
    throw new Error(
      `${label}.scenario_title_required must be true for implemented non-support rows`,
    );
  }
  return target;
}

export function validateFrontendPhaseMap(
  manifest,
  label,
  expectedPhaseID = "",
  options = {},
) {
  const root = options.root ?? process.cwd();
  assertObjectKeys(manifest, mapKeys, label);
  requireSchemaID(manifest, frontendPhaseTestMapSchemaID, label);
  requireInteger(manifest.schema_version, `${label}.schema_version`, { min: 3 });
  if (manifest.phase_namespace !== frontendPhaseNamespace) {
    throw new Error(
      `${label}.phase_namespace must be ${frontendPhaseNamespace}`,
    );
  }
  const phaseID = requirePhaseID(manifest.phase_id, `${label}.phase_id`);
  if (expectedPhaseID && phaseID !== expectedPhaseID) {
    throw new Error(`${label}.phase_id must be ${expectedPhaseID}`);
  }
  requireString(manifest.guide_digest, `${label}.guide_digest`);
  const rows = requireObjectArray(manifest.rows, `${label}.rows`, {
    nonEmpty: true,
  });
  const targetEntriesByName = options.targetEntriesByName ?? null;
  const ids = [];
  const frontendBrowserTitleOwners = options.frontendBrowserTitleOwners ?? new Map();
  for (const [index, row] of rows.entries()) {
    const rowLabel = `${label}.rows[${index + 1}]`;
    assertObjectKeys(row, rowKeys, rowLabel);
    ids.push(
      requireString(row.id, `${rowLabel}.id`, { pattern: rowIDPattern }),
    );
    if (requirePhaseID(row.phase_id, `${rowLabel}.phase_id`) !== phaseID) {
      throw new Error(`${rowLabel}.phase_id must be ${phaseID}`);
    }
    requireEnum(row.layer, `${rowLabel}.layer`, validLayers);
    requireEnum(
      row.evidence_class,
      `${rowLabel}.evidence_class`,
      validEvidenceClasses,
    );
    requireBoolean(row.default_check_required, `${rowLabel}.default_check_required`);
    requireEnum(row.default_check_kind, `${rowLabel}.default_check_kind`, validDefaultCheckKinds);
    requireEnum(
      row.default_check_reason_code,
      `${rowLabel}.default_check_reason_code`,
      validDefaultCheckReasonCodes,
    );
    requireString(row.primary_evidence_owner, `${rowLabel}.primary_evidence_owner`);
    requireString(row.duplicate_of, `${rowLabel}.duplicate_of`);
    requireString(row.evidence_delta, `${rowLabel}.evidence_delta`);
    requireEnum(row.warm_local_cost_class, `${rowLabel}.warm_local_cost_class`, validWarmLocalCostClasses);
    if (row.future_default_check_candidate !== undefined) {
      requireBoolean(
        row.future_default_check_candidate,
        `${rowLabel}.future_default_check_candidate`,
      );
    }
    if (row.default_check_reason !== undefined) {
      requireString(row.default_check_reason, `${rowLabel}.default_check_reason`);
    }
    requireObjectArray(row.owner_refs, `${rowLabel}.owner_refs`, {
      nonEmpty: true,
    }).forEach((ownerRef, ownerIndex) => {
      validateOwnerRef(
        root,
        ownerRef,
        `${rowLabel}.owner_refs[${ownerIndex + 1}]`,
        row,
      );
    });
    requireStringArray(row.core_req_ids, `${rowLabel}.core_req_ids`);
    requireStringArray(row.core_ac_ids, `${rowLabel}.core_ac_ids`);
    requireStringArray(
      row.support_or_design_ac_ids,
      `${rowLabel}.support_or_design_ac_ids`,
    );
    requireObjectArray(row.targets, `${rowLabel}.targets`, {
      nonEmpty: true,
    }).forEach((targetRef, targetIndex) => {
      validateTargetRef(
        targetRef,
        `${rowLabel}.targets[${targetIndex + 1}]`,
        row,
        targetEntriesByName,
      );
    });
    requireStringArray(row.scenario_titles, `${rowLabel}.scenario_titles`);
    const claimStatus = requireEnum(
      row.claim_status,
      `${rowLabel}.claim_status`,
      validClaimStatuses,
    );
    if (claimStatus === "blocked" && row.default_check_required === true) {
      throw new Error(
        `${rowLabel} blocked rows must not declare current default_check_required=true; use future_default_check_candidate for planned check placement`,
      );
    }
    const blockers = requireObjectArray(row.blockers, `${rowLabel}.blockers`).map(
      (blocker, blockerIndex) =>
        validateBlocker(blocker, `${rowLabel}.blockers[${blockerIndex + 1}]`),
    );
    if (claimStatus === "blocked" && blockers.length === 0) {
      throw new Error(`${rowLabel} blocked rows must declare blockers[]`);
    }
    if (claimStatus === "implemented" && blockers.length !== 0) {
      throw new Error(`${rowLabel} implemented rows must not declare blockers[]`);
    }
    if (
      row.future_default_check_candidate === true &&
      claimStatus !== "blocked"
    ) {
      throw new Error(
        `${rowLabel}.future_default_check_candidate is only valid for blocked rows`,
      );
    }
    if (
      row.default_check_required === true &&
      row.default_check_kind === "explicit_only"
    ) {
      throw new Error(`${rowLabel} default_check_required=true cannot use default_check_kind=explicit_only`);
    }
    if (
      row.default_check_required === false &&
      row.default_check_kind === "primary_local_evidence"
    ) {
      throw new Error(`${rowLabel} default_check_required=false cannot use primary_local_evidence`);
    }
    if (
      row.default_check_required === true &&
      (typeof row.default_check_reason !== "string" ||
        row.default_check_reason.trim() === "")
    ) {
      throw new Error(
        `${rowLabel} default-check frontend rows must declare default_check_reason`,
      );
    }
    validateClaim(row.claim, `${rowLabel}.claim`, row);
    requireStringArray(row.out_of_scope, `${rowLabel}.out_of_scope`);
    if (!row.id.includes(`-${phaseID.replace("FE-", "")}-`)) {
      throw new Error(`${rowLabel}.id must belong to ${phaseID}`);
    }
    if (
      row.targets.some(
        (target) =>
          target.target_name.startsWith("browser-e2e") &&
          target.scenario_title_required,
      ) &&
      row.scenario_titles.length === 0
    ) {
      throw new Error(
        `${rowLabel}.scenario_titles must be non-empty for scenario-backed browser rows`,
      );
    }
    if (
      row.targets.some((target) => target.scenario_title_required) &&
      claimStatus === "implemented" &&
      row.scenario_titles.length === 0
    ) {
      throw new Error(
        `${rowLabel}.scenario_titles must be non-empty when scenario_title_required=true`,
      );
    }
    if (
      claimStatus === "implemented" &&
      !row.targets.some((target) => target.required_for_closure)
    ) {
      throw new Error(
        `${rowLabel} implemented rows must have at least one required closure target`,
      );
    }
    if (
      row.layer === "accessibility" &&
      claimStatus === "implemented" &&
      row.targets.some(
        (target) =>
          target.target_name === "browser-e2e-a11y-preflight" &&
          (!target.required_for_closure ||
            !target.frontend_row_accounting_required ||
            !target.scenario_title_required),
      )
    ) {
      throw new Error(
        `${rowLabel} implemented accessibility preflight rows must require current frontend row accounting and exact scenario closure`,
      );
    }
    validateRowMetadata(row, rowLabel);
    validateVisualAccessibilityEvidenceBoundary(row, rowLabel);
    validateCore03SortingFilteringGroupingOwnerRefs(row, rowLabel);
    validateFrontendBrowserScenarioTitleOwnership(
      root,
      row,
      rowLabel,
      frontendBrowserTitleOwners,
    );
  }
  assertUnique(ids, `${label}.rows.id`);

}
