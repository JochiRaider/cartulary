import type { TimelineCaptureReview } from "../workbook/timeline/actions/timelineCaptureActionModel";
export function timelineCaptureReview(
  changes: Partial<TimelineCaptureReview> = {},
): TimelineCaptureReview {
  return {
    action: "mark-reviewed",
    target: {
      recordId: "20000000-0000-4000-8000-000000000001",
      incidentId: "10000000-0000-4000-8000-000000000001",
      rowVersion: 4,
      label: "Original observation",
      context: "Device A",
      captureState: "enriched",
      replacementRecordId: null,
    },
    reason: null,
    replacement: null,
    authority: {
      actorId: "50000000-0000-4000-8000-000000000001",
      sessionIdentity: "test-session",
      incidentId: "10000000-0000-4000-8000-000000000001",
      role: "reviewer",
      closed: false,
    },
    authorityGeneration: 1,
    originKey: "history:target",
    ...changes,
  };
}
