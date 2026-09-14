import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { createTimelineRelatedEvidenceTransport } from "../adapters/createTimelineRelatedEvidenceTransport";
import { TimelineRelatedEvidenceContext } from "../features/evidence/TimelineRelatedEvidenceContext";
import type {
  RelatedEvidenceOutcome,
  RelatedEvidenceTransport,
} from "../features/evidence/timelineRelatedEvidenceOperation";
import { WorkbookTimelineRelatedEvidenceOwner } from "../features/evidence/WorkbookTimelineRelatedEvidenceOwner";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { useTimelineCreateRelatedWorkflow } from "./hooks/useTimelineCreateRelatedWorkflow";
import type { WorkbookRow } from "./models/timelineRowModel";

const timeline = requireViewContract(timelineViewSchemaId);
const evidence = requireViewContract(evidenceViewSchemaId);
const createEvidence = timeline.inspectorConfig.featureGroups.find(
  (feature) => feature.featureGroupKey === "create_related.evidence",
);

describe("useTimelineCreateRelatedWorkflow", () => {
  it("retains accepted Evidence after navigation and refuses to replace its original unsent draft", async () => {
    if (!createEvidence) throw new Error("Missing feature");
    const originalRow = committedRow("record-1", 5),
      nextRow = committedRow("record-2", 7);
    const pending = deferred<RelatedEvidenceOutcome>();
    const authority = {
      actorId: "actor",
      incidentId: "incident",
      role: "editor" as const,
      closed: false,
      sessionIdentity: "session",
    };
    const owner = new WorkbookTimelineRelatedEvidenceOwner(
      "incident",
      { create: () => "create-key" },
      {
        coordinate: async () => ({
          kind: "settled" as const,
          minimumRowVersion: 0,
        }),
        accepted: vi.fn(),
        refresh: async () => {},
        conflict: vi.fn(),
      },
    );
    owner.setAuthority(authority);
    const transport = {
      ...createTimelineRelatedEvidenceTransport(undefined),
      send: vi.fn<RelatedEvidenceTransport["send"]>(() => pending.promise),
    };
    owner.configure(
      {
        availableViews: async () => [
          timelineViewSchemaId,
          evidenceViewSchemaId,
        ],
        verify: async () => {},
        page: async ({ viewSchemaId }) => ({
          kind: "accepted",
          value: {
            candidates: [
              {
                recordId: "record-1",
                viewSchemaId,
                displayText: "Original source",
                row: { record_id: "record-1", row_version: 5, cells: {} },
              },
            ],
            hasMore: false,
            nextCursor: null,
          },
        }),
      },
      async () => authority,
      transport,
    );
    const setInspectorMessage = vi.fn();
    const { result, rerender } = renderHook(
      ({ selectedRow }) =>
        useTimelineCreateRelatedWorkflow({
          selectedRow,
          selectedSubject: subject(selectedRow),
          setInspectorMessage,
        }),
      {
        initialProps: { selectedRow: originalRow },
        wrapper: ({ children }) => (
          <TimelineRelatedEvidenceContext.Provider
            value={{
              owner,
              sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
            }}
          >
            {children}
          </TimelineRelatedEvidenceContext.Provider>
        ),
      },
    );
    act(() => {
      result.current.beginWorkflow(createEvidence);
      owner.update("evidence.title", "Retained metadata");
    });
    const token = result.current.workflow?.workflowId;
    if (!token) throw new Error("Missing attachment");
    const taskFeature = timeline.inspectorConfig.featureGroups.find(
      (feature) => feature.featureGroupKey === "create_related.task_request",
    );
    if (!taskFeature) throw new Error("Missing Task feature");
    act(() => result.current.beginWorkflow(taskFeature));
    expect(owner.getSnapshot().attachment).toBeNull();
    expect(owner.getSnapshot().draft?.values["evidence.title"]).toBe(
      "Retained metadata",
    );
    act(() => owner.resume(token));
    await act(async () => {
      expect(await owner.review()).toBe(true);
    });
    let completion: Promise<void> | undefined;
    await act(async () => {
      completion = owner.submit(token);
    });
    await waitFor(() => expect(transport.send).toHaveBeenCalledOnce());
    rerender({ selectedRow: nextRow });
    act(() => result.current.beginWorkflow(createEvidence));
    expect(owner.getSnapshot().draft?.source.recordId).toBe("record-1");
    expect(owner.getSnapshot().draft?.values["evidence.title"]).toBe(
      "Retained metadata",
    );
    await act(async () => {
      pending.resolve({
        kind: "accepted",
        receipt: {
          meta: { request_id: "request" },
          data: {
            change_set_id: "change-create",
            view_schema_id: evidenceViewSchemaId,
            row: fullWorkbookViewRow(evidence, "evidence-1", 1, {}),
          },
        },
      });
      await completion;
    });
    expect(
      owner.getSnapshot().checkpoints[0]?.create.receipt?.data.row.record_id,
    ).toBe("evidence-1");
    expect(owner.getSnapshot().checkpoints[0]?.links).toEqual([]);
    expect(result.current.workflow).toBeNull();
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
