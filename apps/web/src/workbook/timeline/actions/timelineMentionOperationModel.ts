import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { WorkbookProtocolMentionReceipt } from "../../adapters/workbookProtocolTypes";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";

export type MentionState = "unresolved" | "resolved" | "dismissed";
export type MentionAction =
  | Readonly<{ action: "resolve_item"; resolvedRecordId: string }>
  | Readonly<{
      action: "dismiss_item" | "revert_to_unresolved";
      resolvedRecordId?: never;
    }>;
export type MentionAuthority = Readonly<{
  incidentId: string;
  actorId: string;
  sessionIdentity: string;
  role: WorkbookIncidentRole;
  mutationsAvailable: boolean;
  actions: readonly MentionAction["action"][];
  createTypes: readonly ("host" | "identity")[];
}>;
export type MentionSubject = Readonly<{
  displayText?: string;
  matchedAliasText?: string | null;
  incidentId: string;
  mentionId: string;
  itemRef: string;
  sourceRecordId: string;
  sourceRowVersion: number;
  sourceFieldKey: "timeline.host_refs" | "timeline.identity_refs";
  entityType: "host" | "identity";
  rawText: string;
  mentionRowVersion: number;
  state: MentionState;
  resolvedRecordId: string | null;
  resolutionMethod: string | null;
}>;
export type MentionReview = Readonly<{
  subject: MentionSubject;
  intent: MentionAction;
  authority: MentionAuthority;
}>;
export type MentionReceipt = Readonly<WorkbookProtocolMentionReceipt>;
export type MentionAttempt = Readonly<{
  id: string;
  review: MentionReview;
  operationID: "resolveEntityMention";
  method: "POST";
  path: string;
  apiBase: string | undefined;
  body: string;
}>;
export type MentionOutcome =
  | Readonly<{ kind: "accepted"; receipt: MentionReceipt }>
  | Readonly<{ kind: "uncertain" }>
  | Readonly<{ kind: "rejected"; failure: WorkbookOperationFailure }>;
export type MentionBinding = Readonly<{
  isCurrent: () => boolean;
  prepare: (signal: AbortSignal) => Promise<MentionSubject | null>;
}>;
export type MentionOperation = Readonly<{
  key: number;
  attempt: MentionAttempt;
  phase:
    | "preparing"
    | "preparation_failed"
    | "submitting"
    | "uncertain"
    | "rejected"
    | "accepted";
  receipt: MentionReceipt | null;
  failure: WorkbookOperationFailure | null;
  refresh: "pending" | "refreshing" | "required" | "complete";
  transportPending: boolean;
}>;
export function mentionTransitionAllowed(
  state: MentionState,
  action: MentionAction["action"],
): boolean {
  return action === "revert_to_unresolved"
    ? state === "resolved" || state === "dismissed"
    : state === "unresolved" || state === "resolved";
}
export function mentionReviewValid(review: MentionReview): boolean {
  const { subject, intent } = review;
  return (
    !!subject.mentionId &&
    !!subject.itemRef &&
    !!subject.sourceRecordId &&
    subject.incidentId === review.authority.incidentId &&
    Number.isSafeInteger(subject.mentionRowVersion) &&
    subject.mentionRowVersion > 0 &&
    Number.isSafeInteger(subject.sourceRowVersion) &&
    subject.sourceRowVersion > 0 &&
    subject.sourceFieldKey ===
      (subject.entityType === "host"
        ? "timeline.host_refs"
        : "timeline.identity_refs") &&
    mentionTransitionAllowed(subject.state, intent.action) &&
    (intent.action === "resolve_item"
      ? !!intent.resolvedRecordId
      : !("resolvedRecordId" in intent))
  );
}
/** Source versions can advance for unrelated preceding edits; mention intent cannot. */
export function sameMentionIntent(
  a: MentionSubject,
  b: MentionSubject,
): boolean {
  return (
    a.incidentId === b.incidentId &&
    a.mentionId === b.mentionId &&
    a.itemRef === b.itemRef &&
    a.sourceRecordId === b.sourceRecordId &&
    a.sourceFieldKey === b.sourceFieldKey &&
    a.entityType === b.entityType &&
    a.rawText === b.rawText &&
    a.mentionRowVersion === b.mentionRowVersion &&
    a.state === b.state &&
    a.resolvedRecordId === b.resolvedRecordId &&
    a.resolutionMethod === b.resolutionMethod
  );
}
export function freezeMention<T>(value: T): T {
  if (value && typeof value === "object") {
    for (const child of Object.values(value)) freezeMention(child);
    Object.freeze(value);
  }
  return value;
}
