import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type {
  TimelineRelatedRecordCreated,
  TimelineRelatedRecordPort,
} from "../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import { useTimelineCreateRelatedWorkflow } from "./hooks/useTimelineCreateRelatedWorkflow";
import type { WorkbookRow } from "./models/timelineRowModel";

const timeline = requireViewContract(timelineViewSchemaId);
const evidence = requireViewContract(evidenceViewSchemaId);
const createEvidence = timeline.inspectorConfig.featureGroups.find(
  (feature) => feature.featureGroupKey === "create_related.evidence",
);

describe("useTimelineCreateRelatedWorkflow", () => {
  it("discards a stale Evidence continuation without overwriting a reopened workflow", async () => {
    expect(createEvidence).toBeDefined();
    if (createEvidence === undefined) return;
    const createPending =
      deferred<WorkbookOperationOutcome<TimelineRelatedRecordCreated>>();
    const mutationCommands: TimelineRelatedRecordPort = {
      createRelatedRecord: vi.fn(() => createPending.promise),
      linkCreatedEvidence:
        vi.fn<TimelineRelatedRecordPort["linkCreatedEvidence"]>(),
    };
    const applyAcceptedRowMutation = vi.fn();
    const loadRows = vi.fn(async () => undefined);
    const setInspectorMessage = vi.fn();
    const originalRow = committedRow("record-1", 5);
    const nextRow = committedRow("record-2", 7);
    const { result, rerender } = renderHook(
      ({ selectedRow }) =>
        useTimelineCreateRelatedWorkflow({
          actionContext: {
            authorized: true,
            surfaceKey: "view_schema:cartulary.view.timeline.v2",
          },
          applyAcceptedRowMutation,
          currentUserId: null,
          loadRows,
          mutationCommands,
          selectedRow,
          selectedSubject: subject(selectedRow),
          setInspectorMessage,
          targetContracts: new Map([[evidence.viewSchemaId, evidence]]),
        }),
      { initialProps: { selectedRow: originalRow } },
    );

    act(() => result.current.beginWorkflow(createEvidence));
    let completion: Promise<void> | undefined;
    await act(async () => {
      completion = result.current.submitWorkflow();
      await Promise.resolve();
    });
    rerender({ selectedRow: nextRow });
    act(() => result.current.beginWorkflow(createEvidence));
    const reopenedWorkflowId = result.current.workflow?.workflowId;
    const feedbackCallCount = setInspectorMessage.mock.calls.length;

    await act(async () => {
      createPending.resolve({
        kind: "accepted",
        value: {
          changeSetId: "change-create",
          recordId: "evidence-1",
          viewSchemaId: evidenceViewSchemaId,
        },
      });
      await completion;
    });

    expect(mutationCommands.linkCreatedEvidence).not.toHaveBeenCalled();
    expect(applyAcceptedRowMutation).not.toHaveBeenCalled();
    expect(loadRows).not.toHaveBeenCalled();
    expect(setInspectorMessage).toHaveBeenCalledTimes(feedbackCallCount);
    expect(result.current.workflow).toMatchObject({
      phase: "editing",
      subject: subject(nextRow),
      workflowId: reopenedWorkflowId,
    });
  });
});

function subject(row: WorkbookRow) {
  return {
    kind: "live" as const,
    label: "Timeline row",
    recordId: row.recordId ?? "",
    rowVersion: row.rowVersion ?? 0,
    surfaceLabel: timeline.title,
    viewSchemaId: timelineViewSchemaId,
  };
}

function committedRow(recordId: string, rowVersion: number): WorkbookRow {
  const values = {
    dateEnteredText: "",
    analystText: "",
    mitreStageText: "",
    deviceObjectText: "",
    ipAddressText: "",
    activityUTCText: "",
    activityLocalText: "",
    rawActivityText: "",
    activitySynopsisText: "",
    dataSourceText: "",
  };
  return {
    captureState: "rough",
    collectionDrafts: { hostRefs: "", identityRefs: "", tags: "" },
    collectionValues: { hostRefs: [], identityRefs: [], tags: [] },
    committedValues: values,
    key: recordId,
    pendingSignature: null,
    rawRow: {
      cells: {},
      record_id: recordId,
      row_version: rowVersion,
      view_schema_id: timelineViewSchemaId,
    },
    recordId,
    rowVersion,
    values,
    viewSchemaId: timelineViewSchemaId,
  };
}
