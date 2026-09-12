import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { MentionSubject } from "../actions/timelineMentionOperationModel";
import type { WorkbookRow } from "./timelineRowModel";
import type { InspectorMention } from "./workbookMentionChips";

/** Public mention identity is independent of the opaque collection selector. */
export function timelineMentionSubject(
  mention: InspectorMention,
  row: WorkbookRow | null,
  incidentId: string,
): MentionSubject | null {
  if (
    !row?.recordId ||
    row.viewSchemaId !== timelineViewSchemaId ||
    !row.rowVersion ||
    row.recordId !== mention.rowRecordId ||
    !mention.entityMentionId ||
    !mention.mentionRowVersion
  )
    return null;
  return {
    incidentId,
    mentionId: mention.entityMentionId,
    itemRef: mention.itemRef,
    sourceRecordId: row.recordId,
    sourceRowVersion: row.rowVersion,
    sourceFieldKey: mention.fieldKey,
    entityType: mention.entityType,
    rawText: mention.rawText,
    mentionRowVersion: mention.mentionRowVersion,
    state: mention.status,
    resolvedRecordId: mention.resolvedRecordId,
    resolutionMethod: mention.resolutionMethod,
  };
}
