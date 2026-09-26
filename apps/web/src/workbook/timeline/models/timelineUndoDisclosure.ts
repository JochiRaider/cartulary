import type { MentionOperation } from "../actions/timelineMentionOperationModel";
import type { AutoResolutionDisclosure } from "../actions/WorkbookTimelineMentionOperationOwner";

/** A retained Undo operation belongs to one exact disclosure, not just its mention. */
export function matchesUndoDisclosure(
  entry: MentionOperation,
  notice: AutoResolutionDisclosure,
): boolean {
  const { subject, intent } = entry.attempt.review;
  return (
    intent.action === "revert_to_unresolved" &&
    subject.mentionId === notice.entityMentionId &&
    subject.sourceRecordId === notice.rowRecordId &&
    subject.sourceFieldKey === notice.fieldKey &&
    subject.itemRef === notice.itemRef &&
    subject.mentionRowVersion === notice.mentionRowVersion &&
    subject.resolvedRecordId === notice.resolvedRecordId &&
    subject.state === "resolved" &&
    subject.resolutionMethod === "auto_match"
  );
}
