import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

export type TimelineCaptureAction = "mark-reviewed" | "supersede";
export type TimelineCaptureAuthority = Readonly<{
  actorId: string;
  sessionIdentity: string;
  incidentId: string;
  role: WorkbookIncidentRole;
  closed: boolean;
}>;
export type TimelineCaptureSubject = Readonly<{
  recordId: string;
  incidentId: string;
  rowVersion: number;
  label: string;
  context: string;
  captureState: string;
  replacementRecordId: string | null;
}>;
export type TimelineCaptureReview = Readonly<{
  action: TimelineCaptureAction;
  target: TimelineCaptureSubject;
  replacement: TimelineCaptureSubject | null;
  reason: string | null;
  authority: TimelineCaptureAuthority;
  authorityGeneration: number;
  originKey: string;
}>;

export function timelineCaptureSubject(
  row: WorkbookQueryRow,
  incidentId: string,
): TimelineCaptureSubject {
  const text = (key: string) => {
    const value = row.cells[`timeline.${key}`]?.value;
    return typeof value === "string" ? value : "";
  };
  return Object.freeze({
    recordId: row.record_id,
    incidentId,
    rowVersion: row.row_version,
    label: text("activity_synopsis_text").trim() || "Untitled Timeline row",
    context: [text("activity_utc_text"), text("device_object_text")]
      .filter(Boolean)
      .join(" · "),
    captureState: text("capture_state"),
    replacementRecordId: text("replacement_record_id") || null,
  });
}

export function timelineCaptureIneligibility(
  target: TimelineCaptureSubject | null,
  action: TimelineCaptureAction,
): string | null {
  if (
    !target?.recordId ||
    !Number.isSafeInteger(target.rowVersion) ||
    target.rowVersion < 1
  )
    return "Select a visible saved Timeline row.";
  if (target.captureState === "superseded")
    return "This row is superseded. Use History rollback to correct it.";
  if (action === "mark-reviewed" && target.captureState === "reviewed")
    return "This version is already reviewed.";
  if (!["rough", "enriched", "reviewed"].includes(target.captureState))
    return "The Timeline state is unavailable. Refresh before acting.";
  return null;
}

export function timelineReplacementIneligibility(
  candidate: TimelineCaptureSubject,
  target: TimelineCaptureSubject,
): string | null {
  if (candidate.incidentId !== target.incidentId)
    return "Choose a Timeline row in this incident.";
  if (candidate.recordId === target.recordId)
    return "This is the row being superseded.";
  return timelineCaptureIneligibility(candidate, "supersede");
}

/** Core reason_note_v1 normalization; this action does not own Decision policy. */
export function normalizeTimelineReason(raw: string): string | null {
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

export function timelineCaptureReviewValid(
  review: TimelineCaptureReview,
): boolean {
  return (
    review.target.incidentId === review.authority.incidentId &&
    timelineCaptureIneligibility(review.target, review.action) === null &&
    (review.action === "mark-reviewed"
      ? review.reason === null && review.replacement === null
      : review.reason !== null &&
        normalizeTimelineReason(review.reason) === review.reason &&
        (review.replacement === null ||
          timelineReplacementIneligibility(
            review.replacement,
            review.target,
          ) === null))
  );
}

export const timelineSupersessionConsequence =
  "Superseded rows cannot be edited through ordinary Timeline actions. To correct this change or its replacement, roll back the superseding change in History, then act again.";
