import type {
  MentionReceipt,
  MentionReview,
} from "../workbook/timeline/actions/timelineMentionOperationModel";
export function mentionReview(
  changes: Partial<MentionReview> = {},
): MentionReview {
  return {
    subject: {
      incidentId: "10000000-0000-4000-8000-000000000001",
      sourceRecordId: "20000000-0000-4000-8000-000000000001",
      sourceRowVersion: 4,
      mentionId: "60000000-0000-4000-8000-000000000001",
      itemRef: "opaque selector unrelated to mention identity / @",
      sourceFieldKey: "timeline.host_refs",
      entityType: "host",
      rawText: "  Raw.Host\u00a0",
      mentionRowVersion: 2,
      state: "unresolved",
      resolvedRecordId: null,
      resolutionMethod: null,
    },
    intent: {
      action: "resolve_item",
      resolvedRecordId: "70000000-0000-4000-8000-000000000001",
    },
    authority: {
      incidentId: "10000000-0000-4000-8000-000000000001",
      actorId: "50000000-0000-4000-8000-000000000001",
      sessionIdentity: "account-session",
      role: "editor",
      mutationsAvailable: true,
      createTypes: ["host", "identity"],
      actions: ["resolve_item", "dismiss_item", "revert_to_unresolved"],
    },
    ...changes,
  };
}
export function mentionReceipt(review = mentionReview()): MentionReceipt {
  const { subject, intent } = review;
  const resolve = intent.action === "resolve_item";
  return {
    incident_id: subject.incidentId,
    change_set_id: "30000000-0000-4000-8000-000000000001",
    source_record: {
      record_id: subject.sourceRecordId,
      row_version: subject.sourceRowVersion + 1,
    },
    entity_mention: {
      entity_mention_id: subject.mentionId,
      source_record_id: subject.sourceRecordId,
      source_field_key: subject.sourceFieldKey,
      entity_type: subject.entityType,
      raw_text: subject.rawText,
      normalized_text: "raw.host",
      row_version: subject.mentionRowVersion + 1,
      resolution_status: resolve
        ? "resolved"
        : intent.action === "dismiss_item"
          ? "dismissed"
          : "unresolved",
      resolved_record_id: resolve ? intent.resolvedRecordId : null,
      resolution_method: resolve ? "explicit_resolve_route" : null,
      resolved_at: resolve ? "2026-09-12T04:00:00Z" : null,
      resolved_by_user_id: resolve ? review.authority.actorId : null,
    },
    ...(resolve
      ? {
          active_link: {
            dst_record_id: intent.resolvedRecordId,
            link_type:
              subject.entityType === "host"
                ? ("observed_on_host" as const)
                : ("observed_as_identity" as const),
          },
        }
      : {}),
  };
}

import { requireViewContract } from "@cartulary/view-contracts";
import { timelineViewSchemaId } from "../workbook/models/workbookSurfaceRegistry";
import {
  type MentionCreateReview,
  type MentionCreationReceipt,
  mentionEntityContract,
} from "../workbook/timeline/actions/timelineMentionCreationModel";
import {
  normalizeTimelineFullRow,
  rowFromApi,
} from "../workbook/timeline/models/timelineRowModel";
import { buildInspectorMentions } from "../workbook/timeline/models/workbookMentionChips";
import {
  fullWorkbookViewRow,
  workbookCollectionValue,
} from "./timelineWorkbookTestSupport";
export function mentionWorkbookRow(review = mentionReview()) {
  const s = review.subject;
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(
        requireViewContract(timelineViewSchemaId),
        s.sourceRecordId,
        s.sourceRowVersion,
        {
          "timeline.activity_synopsis_text": "Mention source",
          "timeline.capture_state": "enriched",
          [s.sourceFieldKey]: workbookCollectionValue(
            false,
            s.state === "dismissed"
              ? []
              : [
                  {
                    entity_mention_id: s.mentionId,
                    item_ref: s.itemRef,
                    mention_row_version: s.mentionRowVersion,
                    entity_type: s.entityType,
                    item_kind:
                      s.state === "resolved"
                        ? "resolved_ref"
                        : "unresolved_mention",
                    display_text: s.rawText,
                    raw_text: s.rawText,
                    resolved_record_id: s.resolvedRecordId,
                    resolution_method: s.resolutionMethod,
                    auto_resolved: s.resolutionMethod === "auto_match",
                  },
                ],
          ),
        },
      ),
      "Mention fixture",
    ),
  );
}
export function mentionInspector(review = mentionReview()) {
  const mention = buildInspectorMentions(mentionWorkbookRow(review), [])[0];
  if (!mention) throw new Error("Active mention fixture required");
  return mention;
}
export function mentionCreationReceipt(
  review: MentionCreateReview,
): MentionCreationReceipt {
  const contract = mentionEntityContract(review.subject.entityType);
  return {
    data: {
      change_set_id: "30000000-0000-4000-8000-000000000002",
      view_schema_id:
        contract.viewSchemaId as MentionCreationReceipt["data"]["view_schema_id"],
      row: fullWorkbookViewRow(
        contract,
        "70000000-0000-4000-8000-000000000001",
        1,
        {
          [`${review.subject.entityType}.display_name`]: "Reviewed entity",
          [`${review.subject.entityType}.${review.subject.entityType}_state`]:
            "stub",
        },
      ),
    },
    meta: { request_id: "create-mention-fixture" },
  };
}
