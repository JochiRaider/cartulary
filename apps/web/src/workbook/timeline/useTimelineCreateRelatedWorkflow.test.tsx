import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type {
  TimelineRelatedEvidenceLinked,
  TimelineRelatedRecordCreated,
  TimelineRelatedRecordPort,
} from "../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import { useTimelineCreateRelatedWorkflow } from "./hooks/useTimelineCreateRelatedWorkflow";
import type { WorkbookRow } from "./models/workbookTimelineModel";

const timeline = requireViewContract(timelineViewSchemaId);
const evidence = requireViewContract(evidenceViewSchemaId);
const createEvidence = timeline.inspectorConfig.featureGroups.find(
  (feature) => feature.featureGroupKey === "create_related.evidence",
);

describe("useTimelineCreateRelatedWorkflow", () => {
  it("finishes captured Evidence effects without overwriting a reopened workflow", async () => {
    expect(createEvidence).toBeDefined();
    if (createEvidence === undefined) return;
    const createPending =
      deferred<WorkbookOperationOutcome<TimelineRelatedRecordCreated>>();
    const linkPending =
      deferred<WorkbookOperationOutcome<TimelineRelatedEvidenceLinked>>();
    const mutationCommands: TimelineRelatedRecordPort = {
      createRelatedRecord: vi.fn(() => createPending.promise),
      linkCreatedEvidence: vi.fn(() => linkPending.promise),
    };
    const applyAcceptedRowMutation = vi.fn();
    const loadRows = vi.fn(async () => undefined);
    const setInspectorMessage = vi.fn();
    const originalRow = committedRow("record-1", 5);
    const nextRow = committedRow("record-2", 7);
    const { result, rerender } = renderHook(
      ({ selectedRow }) =>
        useTimelineCreateRelatedWorkflow({
          applyAcceptedRowMutation,
          currentUserId: null,
          loadRows,
          mutationCommands,
          selectedRow,
          selectedSubjectKey: subjectKey(selectedRow),
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
      await Promise.resolve();
    });
    expect(mutationCommands.linkCreatedEvidence).toHaveBeenCalledWith({
      createdRecordId: "evidence-1",
      sourceRow: originalRow,
    });
    await act(async () => {
      linkPending.resolve({
        kind: "accepted",
        value: {
          changeSetId: "change-link",
          row: {
            cells: {},
            record_id: originalRow.recordId ?? "",
            row_version: 6,
            view_schema_id: timelineViewSchemaId,
          },
          viewSchemaId: timelineViewSchemaId,
        },
      });
      await completion;
    });

    expect(applyAcceptedRowMutation).toHaveBeenCalledOnce();
    expect(applyAcceptedRowMutation).toHaveBeenCalledWith(
      originalRow.key,
      expect.objectContaining({ viewSchemaId: timelineViewSchemaId }),
    );
    expect(loadRows).toHaveBeenCalledOnce();
    expect(loadRows).toHaveBeenCalledWith({ showLoading: false });
    expect(setInspectorMessage).toHaveBeenCalledTimes(feedbackCallCount);
    expect(result.current.workflow).toMatchObject({
      phase: "editing",
      subjectKey: subjectKey(nextRow),
      workflowId: reopenedWorkflowId,
    });
  });
});

function subjectKey(row: WorkbookRow) {
  return {
    recordId: row.recordId ?? "",
    rowVersion: row.rowVersion ?? 0,
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
