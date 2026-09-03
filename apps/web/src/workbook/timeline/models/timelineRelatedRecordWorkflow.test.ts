import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import {
  type InspectorRelatedRecordWorkflowState,
  inspectorRelatedRecordWorkflowReducer,
} from "../../inspector/inspectorRelatedRecordModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  planTimelineRelatedSubmission,
  timelineRelatedWorkflowIdentity,
  timelineRelatedWorkflowIsCurrent,
} from "./timelineRelatedRecordWorkflow";
import { normalizeTimelineFullRow, rowFromApi } from "./timelineRowModel";

const timeline = requireViewContract(timelineViewSchemaId);
const task = requireViewContract("cartulary.view.task_requests.v1");
const feature = timeline.inspectorConfig.featureGroups.find(
  (candidate) => candidate.featureGroupKey === "create_related.task_request",
);
const subject = {
  kind: "live" as const,
  label: "Timeline row",
  recordId: "record-1",
  rowVersion: 3,
  surfaceLabel: timeline.title,
  viewSchemaId: timelineViewSchemaId,
};
const row = rowFromApi(
  normalizeTimelineFullRow(
    fullWorkbookViewRow(timeline, subject.recordId, subject.rowVersion, {}),
    "related workflow plan fixture",
  ),
);

describe("Timeline related-record workflow", () => {
  it("admits only the current subject, surface, capability, and version", () => {
    const workflow = editingWorkflow();
    const identity = timelineRelatedWorkflowIdentity(
      workflow,
      "view_schema:timeline",
    );
    const base = {
      context: { authorized: true, surfaceKey: "view_schema:timeline" },
      evidenceViewSchemaId: "cartulary.view.evidence.v1",
      identity,
      selectedRow: row,
      selectedSubject: subject,
      targetContracts: new Map([[task.viewSchemaId, task]]),
      workflow,
    };
    expect(planTimelineRelatedSubmission(base)).toMatchObject({
      kind: "dispatch",
      featureGroupKey: "create_related.task_request",
      sourceRow: row,
    });
    expect(
      planTimelineRelatedSubmission({
        ...base,
        context: { ...base.context, authorized: false },
      }),
    ).toEqual({ kind: "reject", reason: "authorization_lost" });
    expect(
      planTimelineRelatedSubmission({
        ...base,
        selectedRow: { ...row, rowVersion: 4 },
      }),
    ).toEqual({ kind: "reject", reason: "source_invalid" });
    expect(
      planTimelineRelatedSubmission({ ...base, targetContracts: new Map() }),
    ).toEqual({ kind: "reject", reason: "capability_unavailable" });
  });

  it("rejects stale settlement after selection or workflow replacement", () => {
    const workflow = editingWorkflow();
    const identity = timelineRelatedWorkflowIdentity(workflow, "surface-1");
    const base = {
      context: { authorized: true, surfaceKey: "surface-1" },
      identity,
      selectedRow: row,
      selectedSubject: subject,
      workflow,
    };
    expect(timelineRelatedWorkflowIsCurrent(base)).toBe(true);
    expect(
      timelineRelatedWorkflowIsCurrent({
        ...base,
        selectedSubject: { ...subject, recordId: "record-2" },
      }),
    ).toBe(false);
    expect(
      timelineRelatedWorkflowIsCurrent({
        ...base,
        workflow: { ...workflow, workflowId: Symbol("replacement") },
      }),
    ).toBe(false);
  });
});

function editingWorkflow(): InspectorRelatedRecordWorkflowState {
  if (feature === undefined) throw new Error("missing related-record feature");
  const state = inspectorRelatedRecordWorkflowReducer(null, {
    draft: { "task.title": "Investigate" },
    featureGroup: feature,
    subject,
    targetContract: task,
    type: "begin",
    workflowId: Symbol("workflow"),
  });
  if (state === null) throw new Error("missing related workflow state");
  return state;
}
