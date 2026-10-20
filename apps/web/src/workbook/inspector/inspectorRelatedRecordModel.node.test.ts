import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { buildInspectorRelatedRecordDraft } from "./inspectorRelatedRecordModel";
import type { WorkbookInspectorLiveSubject } from "./workbookInspectorSubject";

const timeline = requireViewContract("cartulary.view.timeline.v2");
const taskRequests = requireViewContract("cartulary.view.task_requests.v1");
const canonicalFeature = timeline.inspectorConfig.featureGroups.find(
  (feature) => feature.featureGroupKey === "create_related.task_request",
);

describe("inspector related-record model", () => {
  it("projects coordination source input seeds", () => {
    for (const target of ["comm_log", "handoff", "status_review", "lesson"]) {
      const feature = timeline.inspectorConfig.featureGroups.find(
        (f) => f.featureGroupKey === `create_related.${target}`,
      );
      if (!feature) throw new Error("Missing coordination feature");
      const contract = requireViewContract(`cartulary.view.${target}.v1`);
      const subject = timelineSubject("source-record", 4);
      const result = buildInspectorRelatedRecordDraft({
        currentUserId: null,
        featureGroup: feature,
        subject: { cells: {}, subject },
        targetContract: contract,
      });
      expect(result.kind).toBe("ready");
      if (result.kind !== "ready") throw new Error("Invalid target");
      expect(result.draft["coordination.source_record_id"]).toBe(
        subject.recordId,
      );
    }
  });

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
          subject: timelineSubject("timeline-1", 4),
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
      subject: { cells: {}, subject: timelineSubject("timeline-1", 4) },
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
        subject: { cells: {}, subject: timelineSubject("timeline-1", 4) },
        targetContract: requireViewContract("cartulary.view.notes.v1"),
      }),
    ).toEqual({ kind: "invalid_target", reason: "semantic_mismatch" });
  });
});

function timelineSubject(
  recordId: string,
  rowVersion: number,
): WorkbookInspectorLiveSubject {
  return {
    kind: "live",
    label: "Timeline row",
    recordId,
    rowVersion,
    surfaceLabel: timeline.title,
    viewSchemaId: timeline.viewSchemaId,
  };
}
