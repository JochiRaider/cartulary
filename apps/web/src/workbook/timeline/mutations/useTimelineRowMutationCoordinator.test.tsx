import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { useCallback, useRef, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createWorkbookPendingMutationAdapter } from "../../adapters/createWorkbookPendingMutationAdapter";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { useTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineCommittedRecordIdle } from "../hooks/useTimelineCommittedRecordIdle";
import { useTimelinePendingSaves } from "../hooks/useTimelinePendingSaves";
import type { PendingReplayRuntimeMeta } from "../models/timelineControllerPorts";
import type {
  AutoResolutionNotice,
  DismissedMention,
} from "../models/workbookMentionChips";
import {
  createDraftRow,
  inputFocusKey,
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import { useTimelineRowMutationCoordinator } from "./useTimelineRowMutationCoordinator";

const timelineContract = requireViewContract(timelineViewSchemaId);
const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "11111111-1111-4111-8111-111111111111";

function timelineRow(rowVersion: number, synopsis: string): WorkbookRow {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timelineContract, recordId, rowVersion, {
        "timeline.activity_synopsis_text": synopsis,
      }),
      "row mutation coordinator fixture",
    ),
  );
}

function timelineApiRow(rowVersion: number, synopsis: string) {
  const row = timelineRow(rowVersion, synopsis).rawRow;
  if (row === null) throw new Error("expected API-backed Timeline row");
  return row;
}

function runtimeFixture() {
  return new WorkbookMutationRuntime(
    { clientInstanceId: "client-1", incidentId },
    { create: (prefix) => `${prefix}-txn` },
    createWorkbookPendingMutationAdapter({ apiBase: undefined, incidentId }),
  );
}

function renderCoordinator(
  runtime: WorkbookMutationRuntime,
  initialRows: WorkbookRow[],
) {
  const editorPort = {
    activateEdit: vi.fn(),
    cancelEdit: vi.fn(),
    focus: vi.fn(),
    focusInput: vi.fn(),
    reveal: vi.fn(),
  };
  const advanceViewportContinuity = vi.fn();
  const clearViewportContinuity = vi.fn();
  const rendered = renderHook(() => {
    const [rows, setRows] = useState(initialRows);
    const replaceRows = useCallback((nextRows: WorkbookRow[]) => {
      setRows(nextRows);
    }, []);
    const updateRows = useCallback(
      (updater: (current: WorkbookRow[]) => WorkbookRow[]) => {
        setRows((current) => updater(current));
      },
      [],
    );
    const rowsRef = useRef(rows);
    rowsRef.current = rows;
    const [selectedRowId, setSelectedRowId] = useState<string | null>(recordId);
    const [, setAutoResolutionNotices] = useState<AutoResolutionNotice[]>([]);
    const [, setDismissedMentionsByRow] = useState<
      Record<string, DismissedMention[]>
    >({});
    const pending = useTimelinePendingSaves<PendingReplayRuntimeMeta>({
      mutationRuntime: runtime,
    });
    const editorDraftRegistry =
      useTimelineEditorDraftRegistry(timelineViewSchemaId);
    const loadRows = async () => undefined;
    const nextDraftIndexRef = useRef(2);
    const coordinator = useTimelineRowMutationCoordinator({
      advanceViewportContinuity,
      clearViewportContinuity,
      editorDraftRegistry,
      editorPort,
      mutationRuntime: runtime,
      nextDraftIndex: () => {
        const next = nextDraftIndexRef.current;
        nextDraftIndexRef.current += 1;
        return next;
      },
      pendingQueueSnapshot: pending.snapshot.pendingQueueSnapshot,
      pendingSavesRefs: pending.refs,
      rowsRef,
      selectedRowId,
      clearActiveCollectionInputKey: () => undefined,
      setAutoResolutionNotices,
      setDismissedMentionsByRow,
      setPendingQueueSnapshot: pending.commands.setPendingQueueSnapshot,
      rowStoreCommands: { replaceRows, updateRows },
      setSelectedRowId,
    });
    const waitForCommittedRecordIdle = useTimelineCommittedRecordIdle({
      conflictQueueRef: coordinator.refs.conflictQueueRef,
      latestCommittedRowVersion: coordinator.commands.latestCommittedRowVersion,
      latestCommittedTimelineRow:
        coordinator.commands.latestCommittedTimelineRow,
      loadRows,
      pendingSavesRefs: pending.refs,
    });
    return { coordinator, rows, waitForCommittedRecordIdle };
  });
  return {
    ...rendered,
    advanceViewportContinuity,
    clearViewportContinuity,
    editorPort,
  };
}

afterEach(() => {
  vi.clearAllTimers();
  vi.restoreAllMocks();
});

describe("useTimelineRowMutationCoordinator", () => {
  it("reveals a committed draft row without restoring the pre-create scroll", () => {
    const runtime = runtimeFixture();
    const draft = createDraftRow(1);
    const {
      advanceViewportContinuity,
      clearViewportContinuity,
      editorPort,
      result,
      unmount,
    } = renderCoordinator(runtime, [draft]);

    act(() => {
      result.current.coordinator.commands.applyAcceptedRowMutation(
        draft.key,
        {
          row: timelineApiRow(1, "created summary"),
          viewSchemaId: timelineViewSchemaId,
        },
        { continueOnFreshDraft: true, viewportContinuityToken: 17 },
      );
    });

    expect(clearViewportContinuity).toHaveBeenCalledWith(17);
    expect(advanceViewportContinuity).not.toHaveBeenCalled();
    expect(editorPort.reveal).toHaveBeenCalledWith({
      fieldKey: "timeline.activity_synopsis_text",
      recordId,
    });
    expect(editorPort.focusInput).toHaveBeenCalledOnce();
    expect(result.current.rows.map((row) => row.recordId)).toEqual([
      recordId,
      null,
    ]);
    unmount();
    runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("prevents stale action and mutation results from regressing a live high-water row", () => {
    const runtime = runtimeFixture();
    const initial = timelineRow(3, "live version");
    const { result, unmount } = renderCoordinator(runtime, [
      initial,
      createDraftRow(1),
    ]);
    act(() => {
      result.current.coordinator.commands.acceptCommittedTimelineRow(initial);
      result.current.coordinator.commands.applyAcceptedRowMutation(recordId, {
        row: timelineApiRow(2, "stale mutation"),
        viewSchemaId: timelineViewSchemaId,
      });
      result.current.coordinator.commands.acceptTimelineActionResult({
        captureState: "reviewed",
        changeSetId: "20000000-0000-4000-8000-000000000001",
        incidentId,
        reason: null,
        recordId,
        replacementRecordId: null,
        rowVersion: 2,
      });
    });

    const visible = result.current.rows.find(
      (row) => row.recordId === recordId,
    );
    expect(visible?.rowVersion).toBe(3);
    expect(visible?.values.activitySynopsisText).toBe("live version");
    expect(
      result.current.coordinator.ports.collaborationAdmission.isStaleRecordVersion(
        recordId,
        2,
      ),
    ).toBe(true);
    unmount();
    runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("invalidates an in-flight query generation when an action advances the row", () => {
    const runtime = runtimeFixture();
    const initial = timelineRow(3, "query base");
    const { result, unmount } = renderCoordinator(runtime, [initial]);
    let queryStartEpoch = 0;
    act(() => {
      result.current.coordinator.commands.acceptCommittedTimelineRow(initial);
      queryStartEpoch =
        result.current.coordinator.ports.queryAdmission.beginLoad()
          .queryStartEpoch;
      result.current.coordinator.commands.acceptTimelineActionResult({
        captureState: "reviewed",
        changeSetId: "20000000-0000-4000-8000-000000000002",
        incidentId,
        reason: null,
        recordId,
        replacementRecordId: null,
        rowVersion: 4,
      });
    });

    expect(
      result.current.coordinator.ports.queryAdmission.committedRowsChangedSince(
        queryStartEpoch,
      ),
    ).toBe(true);
    expect(
      result.current.coordinator.ports.queryAdmission.knownTimelineRowVersion(
        recordId,
      ),
    ).toBe(4);
    unmount();
    runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("admits conflict server state without collapsing its local draft", async () => {
    const runtime = runtimeFixture();
    const initial = timelineRow(5, "committed value");
    const pending = {
      ...initial,
      pendingSignature: "pending-signature",
      values: { ...initial.values, activitySynopsisText: "optimistic value" },
    };
    const { result, unmount } = renderCoordinator(runtime, [pending]);
    act(() => {
      const committed =
        result.current.coordinator.commands.latestCommittedTimelineRow(
          recordId,
        );
      expect(committed?.pendingSignature).toBeNull();
      expect(committed?.values.activitySynopsisText).toBe("committed value");
      result.current.coordinator.commands.registerSameFieldConflict(
        {
          base_row_version: 5,
          client_value: "local draft",
          conflict_resolution_class: "text_compare_merge",
          conflict_token: "conflict-token",
          current_row_version: 6,
          field_key: "timeline.activity_synopsis_text",
          record_id: recordId,
          server_value: "server value",
        },
        inputFocusKey(recordId, "activitySynopsisText", "grid"),
        "grid",
      );
    });

    const conflicted = result.current.rows.find(
      (row) => row.recordId === recordId,
    );
    expect(conflicted?.rowVersion).toBe(6);
    expect(conflicted?.pendingSignature).toBeNull();
    expect(conflicted?.values.activitySynopsisText).toBe("server value");
    expect(result.current.coordinator.snapshot.activeConflict?.localValue).toBe(
      "local draft",
    );
    expect(
      result.current.coordinator.ports.queryAdmission.knownTimelineRowVersion(
        recordId,
      ),
    ).toBe(6);
    await expect(
      result.current.waitForCommittedRecordIdle(recordId),
    ).resolves.toBeNull();
    unmount();
    runtime.invalidate({ kind: "runtime_disposed" });
  });
});
