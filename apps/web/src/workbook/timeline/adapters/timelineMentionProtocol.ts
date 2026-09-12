import type {
  MentionAttempt,
  MentionReceipt,
} from "../actions/timelineMentionOperationModel";

/** Generated decoding owns field shape; this validates the receipt against immutable intent. */
export function validateMentionReceipt(
  attempt: MentionAttempt,
  data: MentionReceipt,
): MentionReceipt | null {
  const { subject, intent, authority } = attempt.review;
  const mention = data.entity_mention;
  if (
    data.incident_id !== subject.incidentId ||
    !data.change_set_id ||
    data.source_record.record_id !== subject.sourceRecordId ||
    !Number.isSafeInteger(data.source_record.row_version) ||
    data.source_record.row_version <= subject.sourceRowVersion ||
    mention.entity_mention_id !== subject.mentionId ||
    mention.source_record_id !== subject.sourceRecordId ||
    mention.source_field_key !== subject.sourceFieldKey ||
    mention.entity_type !== subject.entityType ||
    mention.raw_text !== subject.rawText ||
    typeof mention.normalized_text !== "string" ||
    mention.row_version !== subject.mentionRowVersion + 1
  )
    return null;
  if (intent.action === "resolve_item") {
    const link = data.active_link;
    if (
      mention.resolution_status !== "resolved" ||
      mention.resolved_record_id !== intent.resolvedRecordId ||
      mention.resolved_by_user_id !== authority.actorId ||
      typeof mention.resolved_at !== "string" ||
      !Number.isFinite(Date.parse(mention.resolved_at)) ||
      mention.resolution_method !== "explicit_resolve_route" ||
      !link ||
      link.dst_record_id !== intent.resolvedRecordId ||
      link.link_type !==
        (subject.entityType === "host"
          ? "observed_on_host"
          : "observed_as_identity") ||
      (link.src_record_id !== undefined &&
        link.src_record_id !== subject.sourceRecordId)
    )
      return null;
  } else if (
    mention.resolution_status !==
      (intent.action === "dismiss_item" ? "dismissed" : "unresolved") ||
    mention.resolved_record_id !== null ||
    mention.resolved_by_user_id !== null ||
    mention.resolved_at !== null ||
    mention.resolution_method !== null ||
    "active_link" in data
  )
    return null;
  return data;
}
