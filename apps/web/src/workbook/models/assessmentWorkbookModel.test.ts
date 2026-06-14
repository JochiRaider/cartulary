import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import type { TimelineApiRow } from "../timeline/models/workbookTimelineModel";
import {
  assessmentColumnWidth,
  initialAssessmentDraft,
  isAssessmentConfidenceBand,
  supportRowLabel,
} from "./assessmentWorkbookModel";
import { assessmentsViewSchemaId } from "./workbookSurfaceRegistry";

describe("assessmentWorkbookModel", () => {
  it("uses contract enum defaults for new assessment drafts", () => {
    const contract = requireViewContract(assessmentsViewSchemaId);
    expect(initialAssessmentDraft(contract)).toMatchObject({
      assessedAt: "",
      assessmentState: "unknown",
      confidenceBand: "unset",
      rationale: "",
      subjectRecordId: "",
      subjectType: "host",
      supportRecordIds: [],
    });
    expect(isAssessmentConfidenceBand("high")).toBe(true);
    expect(isAssessmentConfidenceBand("critical")).toBe(false);
  });

  it("keeps assessment column widths and support labels stable", () => {
    const row: TimelineApiRow = {
      view_schema_id: "cartulary.view.timeline.v1",
      record_id: "timeline-1",
      row_version: 3,
      cells: {
        "timeline.summary": { value: "Initial access" },
      },
    };
    const fallback: TimelineApiRow = {
      ...row,
      record_id: "timeline-2",
      cells: {
        "timeline.summary": { value: "" },
      },
    };

    expect(assessmentColumnWidth("assessment.subject_ref")).toBe(300);
    expect(assessmentColumnWidth("assessment.rationale")).toBe(360);
    expect(assessmentColumnWidth("assessment.assessment_state")).toBe(180);
    expect(supportRowLabel(row)).toBe("Initial access");
    expect(supportRowLabel(fallback)).toBe("timeline-2");
  });
});
