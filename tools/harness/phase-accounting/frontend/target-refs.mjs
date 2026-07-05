import {
  assertObjectKeys,
  requireBoolean,
  requireEnum,
  requireString,
} from "../../contract/json-shape.mjs";
import { frontendEvidenceAuditInputForTarget } from "./audit-routing.mjs";
import {
  commandIDPattern,
  targetRefKeys,
  validEvidenceRoles,
} from "./constants.mjs";

export function targetDisplayName(target) {
  return `make ${target.target_name}`;
}

export function targetRefMatches(target, normalizedTarget) {
  return targetDisplayName(target) === normalizedTarget;
}

export function validateTargetRef(target, label, row, targetEntriesByName = null) {
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
