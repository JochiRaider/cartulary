import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import { startTransition, useCallback, useState } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type {
  WorkbookPendingReplayAdmissionRequest,
  WorkbookPendingSavesRefs,
} from "../../runtime/workbookPendingReplayRuntime";
import {
  buildStableMutationSignature,
  type PendingReplayUnitInput,
} from "../../utils/workbookPendingQueue";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type {
  PendingReplayRuntimeMeta,
  TimelineMutableRef,
  TimelineRowStoreCommands,
  TimelineScalarSaveOptions,
} from "../models/timelineControllerPorts";
import {
  buildCollectionPatchIntent,
  buildCreatePayload,
  buildScalarPatchIntent,
  type CollectionDraftKey,
  type CollectionFieldKey,
  createDraftRowForKey,
  type FocusFieldKey,
  inputFocusKey,
  type LocalConflictState,
  type RowValues,
  type TimelineScalarEditorSurface,
  timelineScalarBindings,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
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
    unit: WorkbookPendingReplayAdmissionRequest<PendingReplayRuntimeMeta>,
    onSettled?: ((outcome: GridEditCommitOutcome) => void) | undefined,
  ) => void;
  readonly finishSave: (result: "Conflict" | "Saved" | "Syncing") => void;
  readonly incidentId: string;
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
  readonly loadRows: LoadRowsForMutation;
  readonly nextClientTxnId: () => string;
  readonly pendingSavesRefs: WorkbookPendingSavesRefs<PendingReplayRuntimeMeta>;
  readonly recordActionPort: TimelineRecordActionPort;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly rowStoreCommands: TimelineRowStoreCommands;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<CommittedRecordIdle | null>;
}) {
  const { updateRows } = rowStoreCommands;
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
      startTransition(() => {
        updateRows((current) => {
          const nextRows = current.map((row) =>
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
          return nextRows;
        });
      });

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
      rowsRef,
      updateRows,
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
      const requestedRowSnapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      const rowSnapshot =
        requestedRowSnapshot ?? createDraftRowForKey(rowKey) ?? undefined;
      const effectiveRowKey = rowSnapshot?.key ?? rowKey;
      const focusKey = inputFocusKey(
        effectiveRowKey,
        focusField,
        options.surface,
      );
      const snapshot =
        rowSnapshot === undefined
          ? undefined
          : editorDraftRegistry.materializeRow(rowSnapshot, {
              field: focusField,
              value:
                currentValue ??
                editorDraftRegistry.draftValueForFocusKey(focusKey),
            });
      if (!snapshot) {
        onSettled?.({
          kind: "stale_target",
          message: "The timeline row is no longer available.",
        });
        return;
      }
      const binding = timelineScalarBindings.find(
        (candidate) => candidate.key === focusField,
      );
      if (
        snapshot.recordId !== null &&
        binding &&
        conflictQueueRef.current[`${snapshot.recordId}:${binding.fieldKey}`]
      ) {
        onSettled?.({
          kind: "conflict",
          message: "Resolve the existing field conflict before editing.",
        });
        return;
      }

      const clientTxnId = nextClientTxnId();
      const payload =
        snapshot.recordId === null
          ? buildCreatePayload(snapshot, clientTxnId, {
              allowZeroFieldCreate: options.allowZeroFieldCreate === true,
            })
          : buildScalarPatchIntent(snapshot, clientTxnId);
      if (payload === null) {
        editorDraftRegistry.deleteDraftForFocusKey(focusKey);
        onSettled?.({ kind: "accepted" });
        return;
      }

      const mutationSignature = buildStableMutationSignature(payload);
      if (
        pendingSavesRefs.pendingSignaturesRef.current.get(effectiveRowKey) ===
        mutationSignature
      ) {
        onSettled?.({ kind: "accepted" });
        return;
      }
      const visibleEdit =
        binding === undefined
          ? undefined
          : {
              rowKey: effectiveRowKey,
              fieldKey: binding.fieldKey,
              value: snapshot.values[focusField],
            };
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
        mutationSignature,
        payloadIntent: payload,
        promoteToCommittedRowInspect: false,
        rowKey: effectiveRowKey,
        surface: options.surface,
        rowSnapshot: snapshot,
        viewportContinuityToken,
        visibleEdit,
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

  const commitScalarGridEdit = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      currentValue: string,
    ): Promise<GridEditCommitOutcome> =>
      new Promise((resolve) => {
        queueScalarSave(
          rowKey,
          focusField,
          {
            continueOnFreshDraft: false,
            preserveInputFocus: false,
            surface: "grid",
          },
          currentValue,
          resolve,
        );
      }),
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
      if (source === "blur") {
        pendingSavesRefs.collectionKeyboardCommitRef.current.delete(focusKey);
        if (priorKeyboardCommitValue === draftValue) {
          return;
        }
      } else {
        pendingSavesRefs.collectionKeyboardCommitRef.current.set(
          focusKey,
          draftValue,
        );
      }
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
      const payload =
        snapshot.recordId === null
          ? buildCreatePayload(effectiveSnapshot, clientTxnId)
          : buildCollectionPatchIntent(fieldKey, draftValue, clientTxnId);
      if (payload === null) {
        return;
      }

      const mutationSignature = buildStableMutationSignature(payload);
      if (
        pendingSavesRefs.pendingSignaturesRef.current.get(rowKey) ===
        mutationSignature
      ) {
        return;
      }
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
        mutationSignature,
        payloadIntent: payload,
        promoteToCommittedRowInspect: snapshot.recordId === null,
        rowKey,
        surface: "grid",
        rowSnapshot: effectiveSnapshot,
        viewportContinuityToken,
        visibleEdit: {
          rowKey,
          fieldKey,
          value: draftValue,
        },
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
