import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  assessmentColumnWidth,
  assessmentCreateErrors,
  assessmentSupportCandidate,
  buildAssessmentCreatePayload,
  confidenceScoreFromBand,
  followOnAssessmentDraft,
  initialAssessmentDraft,
  isAssessmentConfidenceBand,
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
    expect(assessmentColumnWidth("assessment.subject_ref")).toBe(300);
    expect(assessmentColumnWidth("assessment.rationale")).toBe(360);
    expect(assessmentColumnWidth("assessment.assessment_state")).toBe(180);
    expect(assessmentSupportCandidate("timeline-1", "Initial access")).toEqual({
      displayText: "Initial access",
      recordId: "timeline-1",
    });
    expect(assessmentSupportCandidate("timeline-2", "")).toEqual({
      displayText: "timeline-2",
      recordId: "timeline-2",
    });
  });

  it("owns confidence mapping and normalized create payload construction", () => {
    expect(confidenceScoreFromBand("unset")).toBeNull();
    expect(confidenceScoreFromBand("low")).toBe(25);
    expect(confidenceScoreFromBand("medium")).toBe(55);
    expect(confidenceScoreFromBand("high")).toBe(85);

    expect(
      buildAssessmentCreatePayload(
        {
          assessedAt: " 2026-04-24T12:00:00Z ",
          assessmentState: " confirmed ",
          confidenceBand: "medium",
          rationale: " Confirmed by support. ",
          subjectRecordId: " host-1 ",
          subjectType: "host",
          supportRecordIds: [" support-1 ", "support-1"],
        },
        "txn-assessment-create",
      ),
    ).toEqual({
      client_txn_id: "txn-assessment-create",
      "assessment.subject_ref": "host-1",
      "assessment.subject_type": "host",
      "assessment.assessment_state": "confirmed",
      "assessment.confidence_score": 55,
      "assessment.rationale": "Confirmed by support.",
      "assessment.assessed_at": "2026-04-24T12:00:00Z",
      "assessment.support_refs": {
        kind: "collection_actions_v1",
        actions: [
          {
            op: "add_record_ref",
            linked_record_id: "support-1",
          },
        ],
      },
    });
  });

  it("requires deliberate subjects and validates timestamp and support boundaries without editing input", () => {
    const contract = requireViewContract(assessmentsViewSchemaId);
    const draft = {
      ...initialAssessmentDraft(contract),
      rationale: "A judgment",
    };
    expect(assessmentCreateErrors(draft).subject).toBe("Select a subject.");
    draft.subjectRecordId = "host-1";
    draft.assessedAt = "2026-04-24 12:00:00";
    expect(buildAssessmentCreatePayload(draft, "txn")).toBeNull();
    expect(draft.assessedAt).toBe("2026-04-24 12:00:00");
    draft.assessedAt = "";
    draft.supportRecordIds = Array.from(
      { length: 65 },
      (_, index) => `support-${index}`,
    );
    expect(buildAssessmentCreatePayload(draft, "txn")).toBeNull();
    draft.supportRecordIds.pop();
    expect(
      buildAssessmentCreatePayload(draft, "txn")?.["assessment.support_refs"]
        ?.actions,
    ).toHaveLength(64);
    draft.supportRecordIds = [];
    const payload = buildAssessmentCreatePayload(draft, "txn");
    expect(payload).not.toHaveProperty("assessment.assessed_at");
    expect(payload).not.toHaveProperty("assessment.assessor");
    expect(payload).not.toHaveProperty("assessment.support_refs");
    expect(
      buildAssessmentCreatePayload(draft, "txn", {
        ...contract,
        createCapable: false,
      }),
    ).toBeNull();
  });

  it("seeds follow-on drafts from subject identity only", () => {
    const contract = requireViewContract(assessmentsViewSchemaId);
    expect(
      followOnAssessmentDraft(contract, {
        record_id: "assessment-1",
        row_version: 4,
        cells: {
          "assessment.subject_ref": { value: "identity-1" },
          "assessment.subject_type": { value: "identity" },
          "assessment.assessment_state": { value: "confirmed" },
          "assessment.confidence_score": { value: 85 },
          "assessment.rationale": { value: "Old rationale" },
          "assessment.assessor": { value: "user-1" },
          "assessment.assessed_at": { value: "2026-04-24T12:00:00Z" },
          "assessment.support_refs": {
            value: {
              kind: "collection_value_v1",
              ordered: false,
              items: [{ linked_record_id: "timeline-1" }],
            },
          },
        },
      }),
    ).toEqual({
      assessedAt: "",
      assessmentState: "unknown",
      confidenceBand: "unset",
      rationale: "",
      subjectRecordId: "identity-1",
      subjectType: "identity",
      supportRecordIds: [],
    });
  });
});
