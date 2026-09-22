import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { useCallback, useRef, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createWorkbookPendingMutationAdapter } from "../../adapters/createWorkbookPendingMutationAdapter";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { acceptWorkbookRowObservation } from "../../query/acceptWorkbookRowObservation";
import { createWorkbookMutationRuntime } from "../../runtime/createWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { useTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineCommittedRecordIdle } from "../hooks/useTimelineCommittedRecordIdle";
import { useTimelineCommittedRows } from "../hooks/useTimelineCommittedRows";
import { reconcileCommittedRowsWithLocalDrafts } from "../hooks/useTimelineRowsLoader";
import { inputFocusKey } from "../models/timelineFieldRegistry";
import { timelinePendingSavesRefsFor } from "../models/timelinePendingSaves";
import {
  createDraftRow,
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/timelineRowModel";
import { useTimelineRowMutationCoordinator } from "./useTimelineRowMutationCoordinator";
import { timelineMutationOwnerFor } from "./WorkbookTimelineMutationOwner";

const timelineContract = requireViewContract(timelineViewSchemaId);
const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "11111111-1111-4111-8111-111111111111";
const secondRecordId = "22222222-2222-4222-8222-222222222222";
const thirdRecordId = "33333333-3333-4333-8333-333333333333";

function timelineRow(
  rowVersion: number,
  synopsis: string,
  targetRecordId = recordId,
): WorkbookRow {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timelineContract, targetRecordId, rowVersion, {
        "timeline.activity_synopsis_text": synopsis,
      }),
      "row mutation coordinator fixture",
    ),
  );
}

function timelineApiRow(
  rowVersion: number,
  synopsis: string,
  targetRecordId = recordId,
) {
  const row = timelineRow(rowVersion, synopsis, targetRecordId).rawRow;
  if (row === null) throw new Error("expected API-backed Timeline row");
  return row;
}

function runtimeFixture() {
  return createWorkbookMutationRuntime(
    { clientInstanceId: "client-1", incidentId },
    { create: (prefix) => `${prefix}-txn` },
    createWorkbookPendingMutationAdapter({ apiBase: undefined, incidentId }),
  );
}

function renderCoordinator(
  runtime: WorkbookMutationRuntime,
  initialRows: WorkbookRow[],
  initialCreatedRowPresentationScopeKey = "incident-1:default-query",
) {
  const editorPort = {
    activateEdit: vi.fn(),
    cancelEdit: vi.fn(),
    focus: vi.fn(),
  };
  const completeAcceptedViewportContinuity = vi.fn();
  const rendered = renderHook(
    ({
      createdRowPresentationScopeKey,
    }: {
      createdRowPresentationScopeKey: string;
    }) => {
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
      const [selectedRowId, setSelectedRowId] = useState<string | null>(
        recordId,
      );
      const pending = timelinePendingSavesRefsFor(
        runtime,
        runtime.pendingQueue(),
      );
      const editorDraftRegistry = useTimelineEditorDraftRegistry(
        runtime.localDraftsForSurface(timelineViewSchemaId),
        timelineMutationOwnerFor(runtime).capture,
      );
      const loadRows = async () => undefined;
      const nextDraftIndexRef = useRef(2);
      const committedRows = useTimelineCommittedRows({
        rowsRef,
        mutationRuntime: runtime,
        materializeRow: editorDraftRegistry.materializeRow,
      });
      const coordinator = useTimelineRowMutationCoordinator({
        committedRows: committedRows.commands,
        sheetRef: { kind: "view_schema", id: "cartulary.view.timeline.v2" },
        completeAcceptedViewportContinuity,
        createdRowPresentationScopeKey,
        editorDraftRegistry,
        editorPort,
        mutationRuntime: runtime,
        nextDraftIndex: () => {
          const next = nextDraftIndexRef.current;
          nextDraftIndexRef.current += 1;
          return next;
        },
        pendingSavesRefs: pending,
        rowsRef,
        selectedRowId,
        rowStoreCommands: { replaceRows, updateRows },
        setSelectedRowId,
      });
      const waitForCommittedRecordIdle = useTimelineCommittedRecordIdle({
        conflictQueueRef: coordinator.refs.conflictQueueRef,
        latestCommittedRowVersion:
          coordinator.commands.latestCommittedRowVersion,
        latestCommittedTimelineRow:
          coordinator.commands.latestCommittedTimelineRow,
        loadRows,
        pendingSavesRefs: pending,
      });
      return { coordinator, rows, waitForCommittedRecordIdle };
    },
    {
      initialProps: {
        createdRowPresentationScopeKey: initialCreatedRowPresentationScopeKey,
      },
    },
  );
  return {
    ...rendered,
    completeAcceptedViewportContinuity,
    editorPort,
  };
}

afterEach(() => {
  vi.clearAllTimers();
  vi.restoreAllMocks();
});

describe("useTimelineRowMutationCoordinator", () => {
  it("settles accepted work after unmount without committing or focusing a detached projection", () => {
    const runtime = runtimeFixture();
    const { result, unmount, editorPort, completeAcceptedViewportContinuity } =
      renderCoordinator(runtime, [timelineRow(1, "before")]);
    const apply = result.current.coordinator.commands.applyAcceptedRowMutation;
    unmount();
    expect(
      apply(
        recordId,
        {
          row: timelineApiRow(2, "accepted after navigation"),
          viewSchemaId: timelineViewSchemaId,
        },
        { viewportContinuityToken: 17 },
      ),
    ).toMatchObject({ rowVersion: 2 });
    expect(editorPort.activateEdit).not.toHaveBeenCalled();
    expect(completeAcceptedViewportContinuity).not.toHaveBeenCalled();
    runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("reveals a committed draft row without restoring the pre-create scroll", () => {
    const runtime = runtimeFixture();
    const draft = createDraftRow(1);
    const { completeAcceptedViewportContinuity, editorPort, result, unmount } =
      renderCoordinator(runtime, [draft]);

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

    expect(completeAcceptedViewportContinuity).toHaveBeenCalledWith(17, {
      kind: "fresh_draft",
      focusKey: expect.any(String),
      recordId,
    });
    expect(editorPort.activateEdit).not.toHaveBeenCalled();
    expect(result.current.rows.map((row) => row.recordId)).toEqual([
      recordId,
      null,
    ]);
    expect(
      result.current.coordinator.ports.queryAdmission.currentCreatedRowPresentationRecordId(),
    ).toBe(recordId);
    unmount();
    runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("pins only the latest locally created row within its query presentation scope", () => {
    const runtime = runtimeFixture();
    const draft = createDraftRow(1);
    const { result, rerender, unmount } = renderCoordinator(runtime, [draft]);

    act(() => {
      result.current.coordinator.commands.applyAcceptedRowMutation(draft.key, {
        row: timelineApiRow(1, "first created summary"),
        viewSchemaId: timelineViewSchemaId,
      });
    });
    const nextDraft = result.current.rows.find((row) => row.recordId === null);
    if (nextDraft === undefined) throw new Error("expected next draft row");
    act(() => {
      result.current.coordinator.commands.applyAcceptedRowMutation(
        nextDraft.key,
        {
          row: timelineApiRow(1, "second created summary", secondRecordId),
          viewSchemaId: timelineViewSchemaId,
        },
      );
    });

    expect(
      result.current.coordinator.ports.queryAdmission.currentCreatedRowPresentationRecordId(),
    ).toBe(secondRecordId);
    rerender({ createdRowPresentationScopeKey: "incident-1:filtered-query" });
    expect(
      result.current.coordinator.ports.queryAdmission.currentCreatedRowPresentationRecordId(),
    ).toBeNull();
    unmount();
    runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("retains one query-scoped created row when a refresh omits it", () => {
    const earlierCreated = timelineRow(1, "earlier", recordId);
    const pinnedCreated = timelineRow(1, "pinned", secondRecordId);
    const incoming = timelineRow(1, "incoming", thirdRecordId);
    const reconciled = reconcileCommittedRowsWithLocalDrafts({
      currentRows: [earlierCreated, pinnedCreated, createDraftRow(1)],
      incomingRows: [incoming],
      materializeRow: (row) => row,
      nextDraftIndex: () => 2,
      pinnedCommittedRow: pinnedCreated,
    });

    expect(reconciled.committedRows.map((row) => row.recordId)).toEqual([
      thirdRecordId,
      secondRecordId,
    ]);
    expect(reconciled.rows.filter((row) => row.recordId === null)).toHaveLength(
      1,
    );

    const authoritativePinned = timelineRow(
      2,
      "authoritative pinned",
      secondRecordId,
    );
    const withAuthoritativePin = reconcileCommittedRowsWithLocalDrafts({
      currentRows: reconciled.rows,
      incomingRows: [incoming, authoritativePinned],
      materializeRow: (row) => row,
      nextDraftIndex: () => 3,
      pinnedCommittedRow: pinnedCreated,
    });
    expect(
      withAuthoritativePin.committedRows.filter(
        (row) => row.recordId === secondRecordId,
      ),
    ).toEqual([authoritativePinned]);
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
        baseRowVersion: 1,
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
        baseRowVersion: 3,
      });
    });

    expect(
      result.current.coordinator.ports.queryAdmission.currentMutationEpoch(),
    ).toBeGreaterThan(queryStartEpoch);
    expect(
      result.current.coordinator.ports.queryAdmission.knownTimelineRowVersion(
        recordId,
      ),
    ).toBe(4);
    unmount();
    runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("reconciles same-version query grouping after a lifecycle receipt while retaining local drafts", () => {
    const runtime = runtimeFixture();
    const source = timelineRow(1, "grouped record");
    if (!source.rawRow) throw new Error("Missing source row");
    const scope = {
      actorId: "actor",
      sessionIdentity: "session",
      incidentId,
      epoch: 0,
    };
    const acceptedSource = {
      ...source,
      rawRow: acceptWorkbookRowObservation(source.rawRow, scope),
    };
    const initial = {
      ...acceptedSource,
      rawRow: {
        ...acceptedSource.rawRow,
        group_values: { "timeline.capture_state": "rough" },
      },
    };
    const { result, unmount } = renderCoordinator(runtime, [initial]);
    act(() => {
      result.current.coordinator.commands.acceptCommittedTimelineRow(initial);
      result.current.coordinator.commands.acceptTimelineActionResult({
        captureState: "reviewed",
        changeSetId: "20000000-0000-4000-8000-000000000002",
        incidentId,
        reason: null,
        recordId,
        replacementRecordId: null,
        rowVersion: 2,
        baseRowVersion: 1,
        observation: { recordId, rowVersion: 2, scope },
      });
    });
    const receipt =
      result.current.coordinator.commands.latestCommittedTimelineRow(recordId);
    if (!receipt?.rawRow) throw new Error("Missing accepted lifecycle receipt");
    expect(receipt.captureState).toBe("reviewed");
    expect(receipt.rawRow.group_values).toEqual({
      "timeline.capture_state": "rough",
    });
    const local = {
      ...receipt,
      pendingSignature: "pending-row-edit",
      collectionDrafts: { ...receipt.collectionDrafts, tags: "unsaved tag" },
    };
    const authoritative = rowFromApi({
      ...receipt.rawRow,
      group_values: { "timeline.capture_state": "reviewed" },
    });
    const draft = createDraftRow(1);
    const reconciled = reconcileCommittedRowsWithLocalDrafts({
      currentRows: [local, draft],
      incomingRows: [authoritative],
      materializeRow: (row) => row,
      nextDraftIndex: () => 2,
    });
    expect(reconciled.committedRows[0]?.rawRow?.group_values).toEqual({
      "timeline.capture_state": "reviewed",
    });
    expect(reconciled.committedRows[0]?.rowVersion).toBe(2);
    expect(reconciled.committedRows[0]?.pendingSignature).toBe(
      local.pendingSignature,
    );
    expect(reconciled.committedRows[0]?.collectionDrafts).toBe(
      local.collectionDrafts,
    );
    expect(reconciled.rows).toContain(draft);
    unmount();
    runtime.invalidate({ kind: "runtime_disposed" });
  });

  it("keeps old-scope partial receipts from promoting current saved fields", () => {
    const scope = {
      actorId: "actor",
      sessionIdentity: "current-session",
      incidentId,
      epoch: 2,
    };
    for (const receiptScope of [
      null,
      { ...scope, sessionIdentity: "old-session", epoch: 1 },
    ]) {
      const runtime = runtimeFixture();
      const source = timelineRow(1, "current saved value");
      if (!source.rawRow) throw new Error("Missing source row");
      const acceptedSource = {
        ...source,
        rawRow: acceptWorkbookRowObservation(source.rawRow, scope),
      };
      const { result, unmount } = renderCoordinator(runtime, [acceptedSource]);
      act(() =>
        result.current.coordinator.commands.acceptTimelineActionResult({
          captureState: "reviewed",
          changeSetId: "20000000-0000-4000-8000-000000000002",
          incidentId,
          reason: null,
          recordId,
          replacementRecordId: null,
          rowVersion: 2,
          baseRowVersion: 1,
          ...(receiptScope
            ? { observation: { recordId, rowVersion: 2, scope: receiptScope } }
            : {}),
        }),
      );
      expect(
        result.current.coordinator.commands.latestCommittedTimelineRow(
          recordId,
        ),
      ).toBeNull();
      expect(
        result.current.coordinator.ports.queryAdmission.currentCommittedTimelineRow(
          recordId,
        )?.rawRow,
      ).toEqual(acceptedSource.rawRow);
      expect(
        result.current.coordinator.ports.queryAdmission.knownTimelineRowVersion(
          recordId,
        ),
      ).toBe(2);
      unmount();
      runtime.invalidate({ kind: "runtime_disposed" });
    }
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
      expect(
        committed?.rawRow?.cells["timeline.activity_synopsis_text"]?.value,
      ).toBe("committed value");
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
