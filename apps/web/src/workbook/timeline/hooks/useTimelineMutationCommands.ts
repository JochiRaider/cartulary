import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import { useCallback, useMemo, useState } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { PendingReplayUnitInput } from "../../utils/workbookPendingQueue";
import { createTimelineScalarGridCommitAdapter } from "../adapters/createTimelineScalarGridCommitAdapter";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { LocalConflictState } from "../models/timelineConflictState";
import type {
  TimelineMutableRef,
  TimelineRowStoreCommands,
  TimelineScalarSaveOptions,
} from "../models/timelineControllerPorts";
import {
  type CollectionDraftKey,
  type CollectionFieldKey,
  type FocusFieldKey,
  inputFocusKey,
  type RowValues,
  type TimelineScalarEditorSurface,
  timelineScalarBindingForValueKey,
} from "../models/timelineFieldRegistry";
import type { TimelinePendingReplayAdmission } from "../models/timelineMutationDriverPlans";
import {
  decideTimelineCollectionCommit,
  planTimelineCollectionMutation,
  planTimelineScalarMutation,
  type TimelineMutationAdmission,
} from "../models/timelineMutationQueueAdmission";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import {
  createDraftRowForKey,
  type WorkbookRow,
} from "../models/timelineRowModel";
import type {
  TimelineRecordActionAccepted,
  TimelineRecordActionPort,
} from "../ports/TimelineRecordActionPort";

type ViewportContinuityRequest =
  | { readonly kind: "input"; readonly focusKey: string }
  | { readonly kind: "row-inspect"; readonly recordId: string }
  | { readonly kind: "scroll-only" };

type LoadRowsForMutation = (options: {
  readonly showLoading: boolean;
  readonly viewportContinuityToken?: number;
}) => Promise<void>;

type CommittedRecordIdle = {
  readonly row: WorkbookRow | null;
  readonly rowVersion: number;
};

function isCollectionDraftKey(
  field: FocusFieldKey,
): field is CollectionDraftKey {
  return field === "hostRefs" || field === "identityRefs" || field === "tags";
}

function resolveScalarSaveSnapshot({
  currentValue,
  editorDraftRegistry,
  focusField,
  rowKey,
  rows,
  surface,
}: {
  readonly currentValue: string | undefined;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly focusField: keyof RowValues;
  readonly rowKey: string;
  readonly rows: readonly WorkbookRow[];
  readonly surface: TimelineScalarEditorSurface;
}) {
  const row =
    rows.find((candidate) => candidate.key === rowKey) ??
    createDraftRowForKey(rowKey);
  if (row === null) return null;
  const focusKey = inputFocusKey(row.key, focusField, surface);
  return {
    focusKey,
    row: editorDraftRegistry.materializeRow(row, {
      field: focusField,
      value:
        currentValue ?? editorDraftRegistry.draftValueForFocusKey(focusKey),
    }),
  };
}

function settleUnadmittedScalarMutation({
  admission,
  deleteDraft,
  onSettled,
}: {
  readonly admission: Exclude<TimelineMutationAdmission, { kind: "admit" }>;
  readonly deleteDraft: () => void;
  readonly onSettled: ((outcome: GridEditCommitOutcome) => void) | undefined;
}) {
  switch (admission.kind) {
    case "rejected":
      onSettled?.(admission.outcome);
      return;
    case "accepted_no_change":
      deleteDraft();
      onSettled?.({ kind: "accepted" });
      return;
    case "accepted_duplicate":
      onSettled?.({ kind: "accepted" });
  }
}

export function useTimelineMutationCommands({
  acceptTimelineActionResult,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  clientInstanceId,
  conflictQueueRef,
  editorDraftRegistry,
  enqueueSaveWork,
  enqueuePendingReplayUnit,
  finishSave,
  incidentId,
  latestCommittedTimelineRow,
  loadRows,
  nextClientTxnId,
  pendingSavesRefs,
  recordActionPort,
  resolvePendingSocketTxn,
  rowsRef,
  rowStoreCommands,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly acceptTimelineActionResult: (
    data: TimelineRecordActionAccepted,
  ) => void;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    request: ViewportContinuityRequest,
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly clientInstanceId: string;
  readonly conflictQueueRef: TimelineMutableRef<
    Record<string, LocalConflictState>
  >;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly enqueueSaveWork: (work: () => Promise<void>) => void;
  readonly enqueuePendingReplayUnit: (
    unit: TimelinePendingReplayAdmission,
    onSettled?: ((outcome: GridEditCommitOutcome) => void) | undefined,
  ) => void;
  readonly finishSave: (result: "Conflict" | "Saved" | "Syncing") => void;
  readonly incidentId: string;
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
  readonly loadRows: LoadRowsForMutation;
  readonly nextClientTxnId: () => string;
  readonly pendingSavesRefs: TimelinePendingSavesRefs;
  readonly recordActionPort: TimelineRecordActionPort;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly rowStoreCommands: TimelineRowStoreCommands;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<CommittedRecordIdle | null>;
}) {
  const { replaceRows } = rowStoreCommands;
  const [replacementDrafts, setReplacementDrafts] = useState<
    Record<string, string>
  >({});
  const changeReplacementDraft = useCallback(
    (rowKey: string, value: string) => {
      setReplacementDrafts((current) => ({ ...current, [rowKey]: value }));
    },
    [],
  );

  const enqueueAutosaveReplayForPendingMutation = useCallback(
    ({
      clientTxnId,
      continueOnFreshDraft,
      detectAutoResolution,
      focusField,
      focusKey,
      mutationSignature,
      payloadIntent,
      promoteToCommittedRowInspect,
      rowKey,
      rowSnapshot,
      surface,
      viewportContinuityToken,
      visibleEdit,
      onSettled,
    }: {
      readonly clientTxnId: string;
      readonly continueOnFreshDraft: boolean;
      readonly detectAutoResolution: boolean;
      readonly focusField: FocusFieldKey;
      readonly focusKey: string;
      readonly mutationSignature: string;
      readonly payloadIntent: PendingReplayUnitInput["payloadIntent"];
      readonly promoteToCommittedRowInspect: boolean;
      readonly rowKey: string;
      readonly rowSnapshot: WorkbookRow;
      readonly surface: TimelineScalarEditorSurface;
      readonly viewportContinuityToken: number;
      readonly visibleEdit?: PendingReplayUnitInput["visibleEdit"];
      readonly onSettled?:
        | ((outcome: GridEditCommitOutcome) => void)
        | undefined;
    }) => {
      pendingSavesRefs.pendingSignaturesRef.current.set(
        rowKey,
        mutationSignature,
      );
      const nextRows = rowsRef.current.map((row) =>
        row.key === rowKey
          ? {
              ...row,
              pendingSignature: mutationSignature,
              rawRow:
                visibleEdit === undefined || row.rawRow === null
                  ? row.rawRow
                  : {
                      ...row.rawRow,
                      cells: {
                        ...row.rawRow.cells,
                        [visibleEdit.fieldKey]: {
                          ...row.rawRow.cells[visibleEdit.fieldKey],
                          value: visibleEdit.value,
                        },
                      },
                    },
              values: isCollectionDraftKey(focusField)
                ? row.values
                : {
                    ...row.values,
                    [focusField]: rowSnapshot.values[focusField],
                  },
            }
          : row,
      );
      rowsRef.current = nextRows;
      replaceRows(nextRows);

      enqueuePendingReplayUnit(
        {
          id: `pending-${clientTxnId}`,
          kind: rowSnapshot.recordId === null ? "create" : "patch",
          source: "autosave",
          incidentId,
          clientInstanceId,
          viewSchemaId: timelineViewSchemaId,
          rowKey,
          recordId: rowSnapshot.recordId,
          focusField,
          focusKey,
          surface,
          payloadIntent,
          clientTxnId,
          mutationSignature,
          coalesceKey:
            rowSnapshot.recordId === null
              ? `draft:${rowKey}`
              : `record:${rowSnapshot.recordId}`,
          enqueueOrder: pendingSavesRefs.pendingReplayOrderRef.current,
          operationClass: "hot_path",
          status: "queued",
          ...(visibleEdit === undefined ? {} : { visibleEdit }),
          rowSnapshot,
          continueOnFreshDraft,
          detectAutoResolution,
          promoteToCommittedRowInspect,
          viewportContinuityToken,
        },
        onSettled,
      );
      pendingSavesRefs.pendingReplayOrderRef.current += 1;
    },
    [
      clientInstanceId,
      enqueuePendingReplayUnit,
      incidentId,
      pendingSavesRefs,
      replaceRows,
      rowsRef,
    ],
  );

  const queueScalarSave = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      options: TimelineScalarSaveOptions,
      currentValue?: string,
      onSettled?: ((outcome: GridEditCommitOutcome) => void) | undefined,
    ) => {
      const resolved = resolveScalarSaveSnapshot({
        currentValue,
        editorDraftRegistry,
        focusField,
        rowKey,
        rows: rowsRef.current,
        surface: options.surface,
      });
      if (resolved === null) {
        onSettled?.({
          kind: "stale_target",
          message: "The timeline row is no longer available.",
        });
        return;
      }
      const { focusKey, row: snapshot } = resolved;
      const effectiveRowKey = snapshot.key;
      const binding = timelineScalarBindingForValueKey(focusField);
      const clientTxnId = nextClientTxnId();
      const admission = planTimelineScalarMutation({
        allowZeroFieldCreate: options.allowZeroFieldCreate === true,
        clientTxnId,
        focusField,
        hasConflict:
          snapshot.recordId !== null &&
          conflictQueueRef.current[
            `${snapshot.recordId}:${binding.fieldKey}`
          ] !== undefined,
        pendingSignature:
          pendingSavesRefs.pendingSignaturesRef.current.get(effectiveRowKey),
        row: snapshot,
      });
      if (admission.kind !== "admit") {
        settleUnadmittedScalarMutation({
          admission,
          deleteDraft: () =>
            editorDraftRegistry.deleteDraftForFocusKey(focusKey),
          onSettled,
        });
        return;
      }
      const viewportContinuityToken = beginViewportContinuity(
        options.preserveInputFocus
          ? {
              kind: "input",
              focusKey: inputFocusKey(
                effectiveRowKey,
                focusField,
                options.surface,
              ),
            }
          : {
              kind: "scroll-only",
            },
      );
      enqueueAutosaveReplayForPendingMutation({
        clientTxnId,
        continueOnFreshDraft: options.continueOnFreshDraft,
        detectAutoResolution: false,
        focusField,
        focusKey,
        mutationSignature: admission.mutationSignature,
        payloadIntent: admission.payloadIntent,
        promoteToCommittedRowInspect: false,
        rowKey: effectiveRowKey,
        surface: options.surface,
        rowSnapshot: snapshot,
        viewportContinuityToken,
        visibleEdit: {
          rowKey: effectiveRowKey,
          ...admission.visibleEdit,
        },
        onSettled,
      });
    },
    [
      beginViewportContinuity,
      conflictQueueRef,
      editorDraftRegistry,
      enqueueAutosaveReplayForPendingMutation,
      nextClientTxnId,
      pendingSavesRefs,
      rowsRef,
    ],
  );

  const commitScalarGridEdit = useMemo(
    () => createTimelineScalarGridCommitAdapter(queueScalarSave),
    [queueScalarSave],
  );

  const queueCollectionSave = useCallback(
    (
      rowKey: string,
      fieldKey: CollectionFieldKey,
      focusField: CollectionDraftKey,
      draftValueOverride?: string,
      source: "keyboard" | "blur" = "blur",
    ) => {
      const focusKey = inputFocusKey(rowKey, focusField, "grid");
      const rowSnapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      if (!rowSnapshot) {
        return;
      }
      const draftValue =
        draftValueOverride ?? rowSnapshot.collectionDrafts[focusField];
      const priorKeyboardCommitValue =
        pendingSavesRefs.collectionKeyboardCommitRef.current.get(focusKey);
      const commitDecision = decideTimelineCollectionCommit({
        draftValue,
        priorKeyboardCommitValue,
        source,
      });
      if (commitDecision.nextKeyboardCommitValue === null) {
        pendingSavesRefs.collectionKeyboardCommitRef.current.delete(focusKey);
      } else {
        pendingSavesRefs.collectionKeyboardCommitRef.current.set(
          focusKey,
          commitDecision.nextKeyboardCommitValue,
        );
      }
      if (!commitDecision.admit) return;
      const snapshot =
        rowSnapshot.recordId === null
          ? rowSnapshot
          : (latestCommittedTimelineRow(rowSnapshot.recordId) ?? rowSnapshot);
      const collectionSnapshot =
        draftValueOverride === undefined
          ? snapshot
          : {
              ...snapshot,
              collectionDrafts: {
                ...snapshot.collectionDrafts,
                [focusField]: draftValue,
              },
            };
      const effectiveSnapshot =
        editorDraftRegistry.materializeRow(collectionSnapshot);
      const clientTxnId = nextClientTxnId();
      const admission = planTimelineCollectionMutation({
        clientTxnId,
        draftValue,
        effectiveRow: effectiveSnapshot,
        fieldKey,
        pendingSignature:
          pendingSavesRefs.pendingSignaturesRef.current.get(rowKey),
      });
      if (admission.kind !== "admit") return;
      const viewportContinuityToken = beginViewportContinuity(
        snapshot.recordId === null
          ? {
              kind: "scroll-only",
            }
          : {
              kind: "row-inspect",
              recordId: snapshot.recordId,
            },
      );
      enqueueAutosaveReplayForPendingMutation({
        clientTxnId,
        continueOnFreshDraft: snapshot.recordId === null,
        detectAutoResolution: true,
        focusField,
        focusKey,
        mutationSignature: admission.mutationSignature,
        payloadIntent: admission.payloadIntent,
        promoteToCommittedRowInspect: snapshot.recordId === null,
        rowKey,
        surface: "grid",
        rowSnapshot: effectiveSnapshot,
        viewportContinuityToken,
        visibleEdit: { rowKey, ...admission.visibleEdit },
      });
    },
    [
      beginViewportContinuity,
      editorDraftRegistry,
      enqueueAutosaveReplayForPendingMutation,
      latestCommittedTimelineRow,
      nextClientTxnId,
      pendingSavesRefs,
      rowsRef,
    ],
  );

  const queueAction = useCallback(
    (rowKey: string, action: "mark-reviewed" | "supersede") => {
      const snapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      const replacementRecordId =
        action === "supersede"
          ? (replacementDrafts[rowKey] ?? "").trim()
          : null;
      if (
        !snapshot ||
        snapshot.recordId === null ||
        snapshot.rowVersion === null ||
        (action === "supersede" && replacementRecordId === "")
      ) {
        return;
      }

      const recordId = snapshot.recordId;
      const clientTxnId = nextClientTxnId();
      const viewportContinuityToken = beginViewportContinuity({
        kind: "row-inspect",
        recordId,
      });
      beginSave();
      enqueueSaveWork(async () => {
        const idleRecord = await waitForCommittedRecordIdle(recordId);
        if (idleRecord === null) {
          clearViewportContinuity(viewportContinuityToken);
          finishSave("Conflict");
          return;
        }
        trackPendingSocketTxn(clientTxnId);
        const result = await recordActionPort.execute({
          action,
          baseRowVersion: idleRecord.rowVersion,
          clientTxnId,
          recordId,
          replacementRecordId,
        });
        if (result.kind === "rejected") {
          resolvePendingSocketTxn(clientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          finishSave("Conflict");
          return;
        }

        acceptTimelineActionResult(result.value);
        await loadRows({
          showLoading: false,
          viewportContinuityToken,
        });
        finishSave("Saved");
      });
    },
    [
      acceptTimelineActionResult,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      finishSave,
      loadRows,
      nextClientTxnId,
      recordActionPort,
      replacementDrafts,
      resolvePendingSocketTxn,
      rowsRef,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  return {
    commands: {
      changeReplacementDraft,
      commitScalarGridEdit,
      queueAction,
      queueCollectionSave,
      queueScalarSave,
    },
    snapshot: { replacementDrafts },
  };
}
