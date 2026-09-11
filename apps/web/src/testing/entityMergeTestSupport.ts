import type { EntityMergeReceipt } from "../workbook/features/entities/entityMergeOperation";
import type {
  EntityMergeAuthority,
  EntityMergeReview,
} from "../workbook/features/entities/entityMergeReview";
import { entityMergeIdentifierFields } from "../workbook/models/entityIdentifierClasses";
import { buildMergePlan } from "../workbook/models/entityMergePlan";
import {
  type EntityRow,
  entityRowFromApi,
} from "../workbook/models/entityWorkbookModel";
import type { WorkbookQueryRow } from "../workbook/query/WorkbookQueryRow";

export const mergeSurvivorId = "00000000-0000-4000-8000-000000007100";
export const mergeLoserId = "00000000-0000-4000-8000-000000007101";
export const mergeIncidentId = "00000000-0000-4000-8000-000000007103";
export const mergeAuthority: EntityMergeAuthority = {
  actorId: "analyst",
  sessionIdentity: "session",
  incidentId: mergeIncidentId,
  role: "reviewer",
  closed: false,
};
export function mergeEntityRow(
  recordId: string,
  rowVersion: number,
  type: "host" | "identity" = "host",
): EntityRow {
  const cells: WorkbookQueryRow["cells"] = Object.fromEntries(
    entityMergeIdentifierFields[type].map((field) => [
      field.key,
      { value: null },
    ]),
  );
  cells[`${type}.display_name`] = { value: "Duplicate label" };
  cells[`${type}.${type}_state`] = { value: "canonical" };
  cells[`${type}.aliases`] = {
    value: { kind: "collection_value_v1", ordered: false, items: [] },
  };
  cells[`${type}.reusable_identifiers`] = {
    value: { kind: "collection_value_v1", ordered: false, items: [] },
  };
  return entityRowFromApi(
    { cells, record_id: recordId, row_version: rowVersion },
    type,
  );
}
export function mergeReview(
  type: "host" | "identity" = "host",
  generation = 1,
): EntityMergeReview {
  const survivor = mergeEntityRow(mergeSurvivorId, 7, type);
  const loser = mergeEntityRow(mergeLoserId, 2, type);
  return {
    survivor: {
      recordId: survivor.recordId,
      label: survivor.label,
      baseRowVersion: 7,
    },
    loser: { recordId: loser.recordId, label: loser.label, baseRowVersion: 2 },
    authority: mergeAuthority,
    authorityGeneration: generation,
    entityType: type,
    originSurface: type,
    presentationGeneration: 0,
    lifecycleKey: "lifecycle",
    reason: "Merge duplicate entity",
    plan: buildMergePlan(survivor, loser),
  };
}
export function mergeReceipt(
  type: "host" | "identity" = "host",
): EntityMergeReceipt {
  return {
    incident_id: mergeIncidentId,
    record_type: type,
    survivor_record_id: mergeSurvivorId,
    loser_record_id: mergeLoserId,
    survivor_row_version: 8,
    loser_row_version: 3,
    merged_into_record_id: mergeSurvivorId,
    change_set_id: "00000000-0000-4000-8000-000000007200",
    merge_summary: {
      record_type: type,
      repointed_mention_resolution_count: 2,
      repointed_link_count: 1,
      deduped_link_count: 0,
      repointed_tag_count: 1,
      deduped_tag_count: 0,
      repointed_assessment_count: 1,
      suggestion_aliases_copied_count: 1,
      suggestion_alias_duplicate_noop_count: 0,
      provenance_only_retained_count: 2,
      exact_match_classes: entityMergeIdentifierFields[type].map((field) => ({
        identifier_class: field.identifierClass,
        promoted_count: 0,
        carried_count: 0,
        duplicate_noop_count: 0,
        blocked_conflict_count: 0,
        provenance_only_count: 0,
        suggestion_only_count: 0,
      })),
    },
  };
}
