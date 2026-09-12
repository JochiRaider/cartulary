import type { TimelineMentionSourceReader } from "../adapters/createTimelineMentionSourceReader";
import { readCollectionItems } from "../models/workbookMentionChips";
import type { MentionReceipt } from "./timelineMentionOperationModel";
import type {
  MentionReconciliationScope,
  WorkbookTimelineMentionOperationOwner,
} from "./WorkbookTimelineMentionOperationOwner";
export async function reconcileTimelineMentionReceipt(
  owner: WorkbookTimelineMentionOperationOwner,
  readSource: TimelineMentionSourceReader,
  receipt: MentionReceipt,
  scope: MentionReconciliationScope,
) {
  const row = await readSource(receipt.source_record.record_id, scope.signal);
  if (
    !scope.isCurrent() ||
    row.record_id !== receipt.source_record.record_id ||
    row.row_version <
      Math.max(
        receipt.source_record.row_version,
        owner.latestVersion(row.record_id) ?? 0,
      )
  )
    throw new Error("A newer source version still needs refresh.");
  const field =
    receipt.entity_mention.entity_type === "host"
      ? "timeline.host_refs"
      : "timeline.identity_refs";
  const item = readCollectionItems(row, field).find(
    (item) => item.entityMentionId === receipt.entity_mention.entity_mention_id,
  );
  if (
    item &&
    (item.mentionRowVersion ?? 0) < receipt.entity_mention.row_version
  )
    throw new Error("A newer mention version still needs refresh.");
  if (
    row.row_version === receipt.source_record.row_version &&
    (receipt.entity_mention.resolution_status === "dismissed"
      ? !!item
      : !item ||
        item.mentionRowVersion !== receipt.entity_mention.row_version ||
        item.resolvedRecordId !== receipt.entity_mention.resolved_record_id ||
        item.rawText !== receipt.entity_mention.raw_text ||
        item.resolutionMethod !== receipt.entity_mention.resolution_method)
  )
    throw new Error("Mention result projection is incomplete.");
  owner.acceptVersion(row.record_id, row.row_version);
  if (item?.entityMentionId && item.mentionRowVersion)
    owner.observeMention({
      incidentId: owner.incidentId,
      sourceRecordId: row.record_id,
      sourceRowVersion: row.row_version,
      sourceFieldKey: field,
      mentionId: item.entityMentionId,
      itemRef: item.itemRef,
      entityType: item.entityType,
      rawText: item.rawText,
      mentionRowVersion: item.mentionRowVersion,
      state: item.itemKind === "resolved_ref" ? "resolved" : "unresolved",
      resolvedRecordId: item.resolvedRecordId,
      resolutionMethod: item.resolutionMethod,
    });
  await owner.refreshPresentation(row.record_id, row.row_version);
  if (!scope.isCurrent()) throw new Error("Mention reconciliation detached.");
}
