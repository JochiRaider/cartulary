import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { buildInspectorRelatedRecordDraft } from "./inspectorRelatedRecordModel";

const timeline = requireViewContract("cartulary.view.timeline.v2");
const taskRequests = requireViewContract("cartulary.view.task_requests.v1");
const canonicalFeature = timeline.inspectorConfig.featureGroups.find(
  (feature) => feature.featureGroupKey === "create_related.task_request",
);

describe("inspector related-record model", () => {
  it("builds selected-record, selected-field, and literal seeds", () => {
    expect(canonicalFeature).toBeDefined();
    if (canonicalFeature === undefined) return;
    const feature = {
      ...canonicalFeature,
      seedBindings: [
        {
          source: { kind: "selected_record_id" as const },
          targetFieldKey: "task.timeline_item_id",
        },
        {
          source: {
            kind: "selected_field_value" as const,
            sourceFieldKey: "timeline.title",
          },
          targetFieldKey: "task.title",
        },
        {
          source: { kind: "literal" as const, value: { urgent: true } },
          targetFieldKey: "task.metadata",
        },
      ],
    };

    expect(
      buildInspectorRelatedRecordDraft({
        currentUserId: "user-1",
        featureGroup: feature,
        subject: {
          cells: { "timeline.title": { value: "  Investigate  " } },
          recordId: "timeline-1",
        },
        targetContract: taskRequests,
      }),
    ).toMatchObject({
      kind: "ready",
      draft: {
        "task.metadata": '{"urgent":true}',
        "task.timeline_item_id": "timeline-1",
        "task.title": "Investigate",
      },
    });
  });

  it("leaves defaults intact for a missing selected field", () => {
    expect(canonicalFeature).toBeDefined();
    if (canonicalFeature === undefined) return;
    const feature = {
      ...canonicalFeature,
      seedBindings: [
        {
          source: {
            kind: "selected_field_value" as const,
            sourceFieldKey: "timeline.missing",
          },
          targetFieldKey: "task.title",
        },
      ],
    };
    const result = buildInspectorRelatedRecordDraft({
      currentUserId: null,
      featureGroup: feature,
      subject: { cells: {}, recordId: "timeline-1" },
      targetContract: taskRequests,
    });
    expect(result.kind).toBe("ready");
    if (result.kind === "ready") {
      expect(result.draft["task.title"]).toBeUndefined();
    }
  });

  it("rejects a mismatched target contract", () => {
    expect(canonicalFeature).toBeDefined();
    if (canonicalFeature === undefined) return;
    expect(
      buildInspectorRelatedRecordDraft({
        currentUserId: null,
        featureGroup: canonicalFeature,
        subject: { cells: {}, recordId: "timeline-1" },
        targetContract: requireViewContract("cartulary.view.notes.v1"),
      }),
    ).toEqual({ kind: "invalid_target", reason: "semantic_mismatch" });
  });
});
