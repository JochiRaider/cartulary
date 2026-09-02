import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  buildInspectorRelatedRecordDraft,
  inspectorRelatedRecordWorkflowReducer,
} from "./inspectorRelatedRecordModel";
import type { WorkbookInspectorLiveSubject } from "./workbookInspectorSubject";

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

  it("reduces editing, submission, rejection, retry, and completion exhaustively", () => {
    expect(canonicalFeature).toBeDefined();
    if (canonicalFeature === undefined) return;
    const workflowId = Symbol("workflow");
    const subject = timelineSubject("timeline-1", 4);
    const editing = inspectorRelatedRecordWorkflowReducer(null, {
      type: "begin",
      draft: { "task.title": "Investigate" },
      featureGroup: canonicalFeature,
      subject,
      targetContract: taskRequests,
      workflowId,
    });
    expect(editing).toMatchObject({
      draft: { "task.title": "Investigate" },
      error: null,
      phase: "editing",
      subject,
      workflowId,
    });

    const updated = inspectorRelatedRecordWorkflowReducer(editing, {
      type: "update",
      fieldKey: "task.title",
      value: "Investigate host",
      workflowId,
    });
    expect(updated?.draft["task.title"]).toBe("Investigate host");
    const submitting = inspectorRelatedRecordWorkflowReducer(updated, {
      type: "submit",
      workflowId,
    });
    expect(submitting?.phase).toBe("submitting");
    expect(
      inspectorRelatedRecordWorkflowReducer(submitting, {
        type: "update",
        fieldKey: "task.title",
        value: "Ignored while submitting",
        workflowId,
      }),
    ).toBe(submitting);

    const error = {
      primaryMessage: "Creation failed.",
      technicalFields: [],
    };
    const rejected = inspectorRelatedRecordWorkflowReducer(submitting, {
      type: "reject",
      error,
      workflowId,
    });
    expect(rejected).toMatchObject({ error, phase: "editing" });
    const retrying = inspectorRelatedRecordWorkflowReducer(rejected, {
      type: "submit",
      workflowId,
    });
    expect(retrying).toMatchObject({ error: null, phase: "submitting" });
    expect(
      inspectorRelatedRecordWorkflowReducer(retrying, {
        type: "complete",
        workflowId,
      }),
    ).toBeNull();
  });

  it("ignores obsolete workflow actions after cancel and reopen", () => {
    expect(canonicalFeature).toBeDefined();
    if (canonicalFeature === undefined) return;
    const oldWorkflowId = Symbol("old-workflow");
    const currentWorkflowId = Symbol("current-workflow");
    const subject = timelineSubject("timeline-1", 4);
    const oldState = inspectorRelatedRecordWorkflowReducer(null, {
      type: "begin",
      draft: {},
      featureGroup: canonicalFeature,
      subject,
      targetContract: taskRequests,
      workflowId: oldWorkflowId,
    });
    const canceled = inspectorRelatedRecordWorkflowReducer(oldState, {
      type: "cancel",
      workflowId: oldWorkflowId,
    });
    const currentState = inspectorRelatedRecordWorkflowReducer(canceled, {
      type: "begin",
      draft: { "task.title": "Current" },
      featureGroup: canonicalFeature,
      subject,
      targetContract: taskRequests,
      workflowId: currentWorkflowId,
    });

    for (const obsoleteAction of [
      {
        type: "update" as const,
        fieldKey: "task.title",
        value: "Obsolete",
        workflowId: oldWorkflowId,
      },
      { type: "submit" as const, workflowId: oldWorkflowId },
      {
        type: "reject" as const,
        error: { primaryMessage: "Obsolete", technicalFields: [] },
        workflowId: oldWorkflowId,
      },
      { type: "complete" as const, workflowId: oldWorkflowId },
      { type: "cancel" as const, workflowId: oldWorkflowId },
      {
        type: "retarget" as const,
        subject: null,
        workflowId: oldWorkflowId,
      },
    ]) {
      expect(
        inspectorRelatedRecordWorkflowReducer(currentState, obsoleteAction),
      ).toBe(currentState);
    }
  });

  it("retains an equal subject and clears record, version, surface, or no-row retargets", () => {
    expect(canonicalFeature).toBeDefined();
    if (canonicalFeature === undefined) return;
    const workflowId = Symbol("workflow");
    const subject = timelineSubject("timeline-1", 4);
    const state = inspectorRelatedRecordWorkflowReducer(null, {
      type: "begin",
      draft: {},
      featureGroup: canonicalFeature,
      subject,
      targetContract: taskRequests,
      workflowId,
    });
    expect(
      inspectorRelatedRecordWorkflowReducer(state, {
        type: "retarget",
        subject: { ...subject, label: "Presentation-only rename" },
        workflowId,
      }),
    ).toBe(state);

    for (const retargeted of [
      timelineSubject("timeline-2", 4),
      timelineSubject("timeline-1", 5),
      { ...subject, viewSchemaId: "cartulary.view.notes.v1" },
      null,
    ]) {
      expect(
        inspectorRelatedRecordWorkflowReducer(state, {
          type: "retarget",
          subject: retargeted,
          workflowId,
        }),
      ).toBeNull();
    }
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
