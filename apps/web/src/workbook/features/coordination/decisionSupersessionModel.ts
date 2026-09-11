import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

export const decisionViewId = "cartulary.view.decisions.v1";
export type DecisionAuthority = Readonly<{
  actorId: string;
  sessionIdentity: string;
  incidentId: string;
  role: WorkbookIncidentRole;
  closed: boolean;
}>;
export type ReviewedDecision = Readonly<{
  recordId: string;
  incidentId: string;
  label: string;
  baseRowVersion: number;
  status: string;
  ownerId: string | null;
  decidedAt: string | null;
  isSuperseded: boolean | null;
  supersedesRecordId: string | null;
  malformed: boolean;
}>;
export type DecisionSupersessionReview = Readonly<{
  target: ReviewedDecision;
  replacement: ReviewedDecision;
  reason: string;
  authority: DecisionAuthority;
  authorityGeneration: number;
  originSurface: string;
  lifecycleKey: string;
}>;

export function reviewedDecision(
  row: WorkbookQueryRow,
  incidentId: string,
): ReviewedDecision {
  const value = (key: string) => row.cells[`decision.${key}`]?.value;
  const text = (key: string) =>
    typeof value(key) === "string" ? (value(key) as string) : null;
  return Object.freeze({
    recordId: row.record_id,
    incidentId,
    baseRowVersion: row.row_version,
    label: text("summary") || "Untitled Decision",
    status: text("status") ?? "unknown",
    ownerId: text("owner_user_id"),
    decidedAt: text("decided_at"),
    isSuperseded:
      typeof value("is_superseded") === "boolean"
        ? (value("is_superseded") as boolean)
        : null,
    supersedesRecordId: text("supersedes_record_id"),
    malformed:
      value("supersedes_record_id") !== null &&
      typeof value("supersedes_record_id") !== "string",
  });
}

export function decisionIneligibility(
  record: ReviewedDecision,
  purpose: "target" | "replacement",
  target?: ReviewedDecision,
): string | null {
  if (
    record.malformed ||
    !Number.isSafeInteger(record.baseRowVersion) ||
    record.baseRowVersion < 1 ||
    !record.ownerId ||
    !record.decidedAt ||
    !Number.isFinite(Date.parse(record.decidedAt)) ||
    record.isSuperseded === null ||
    !["proposed", "approved", "rejected", "executed", "superseded"].includes(
      record.status,
    ) ||
    (record.status === "superseded" && !record.isSuperseded) ||
    (record.isSuperseded &&
      !["superseded", "executed"].includes(record.status)) ||
    (record.supersedesRecordId !== null &&
      !["approved", "executed"].includes(record.status))
  )
    return "The loaded Decision state is inconsistent or incomplete. Refresh before reviewing.";
  if (target && record.incidentId !== target.incidentId)
    return "Choose a Decision in this incident.";
  if (target?.recordId === record.recordId)
    return "This is the target Decision.";
  if (purpose === "target") {
    if (record.isSuperseded) return "This Decision already has a replacement.";
    if (!["proposed", "approved", "executed"].includes(record.status))
      return "Only proposed, approved, or executed Decisions can be superseded.";
  } else if (!["approved", "executed"].includes(record.status))
    return "The replacement must be approved or executed.";
  return null;
}

/** reason_note_v1: NFC, LF, Unicode whitespace, controls, scalar length. */
export function normalizeDecisionReason(raw: string): string | null {
  const value = raw
    .normalize("NFC")
    .replace(/\r\n?/g, "\n")
    .replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");
  if (
    !value ||
    [...value].length > 4096 ||
    [...value].some(
      (char) => char !== "\n" && char !== "\t" && /[\p{Cc}\p{Cs}]/u.test(char),
    )
  )
    return null;
  return value;
}

export function decisionReviewValid(
  review: DecisionSupersessionReview,
): boolean {
  return (
    review.authority.incidentId === review.target.incidentId &&
    decisionIneligibility(review.target, "target") === null &&
    decisionIneligibility(review.replacement, "replacement", review.target) ===
      null &&
    normalizeDecisionReason(review.reason) === review.reason
  );
}

export function decisionConsequence(status: string): string {
  return status === "executed"
    ? "The target remains executed and is marked as superseded. Its execution is preserved."
    : "The target status changes to superseded because it has not been executed.";
}
