import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import { useCallback, useMemo } from "react";
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

type ViewportContinuityRequest =
  | { readonly kind: "input"; readonly focusKey: string }
  | { readonly kind: "row-inspect"; readonly recordId: string }
  | { readonly kind: "scroll-only" };

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
    rows.find(
      (candidate) =>
        candidate.key === editorDraftRegistry.resolveRowKey(rowKey),
    ) ?? createDraftRowForKey(rowKey);
  if (row === null) return null;
  const focusKey = inputFocusKey(row.key, focusField, surface);
  return {
    focusKey,
    row: editorDraftRegistry.authoringRow(
      editorDraftRegistry.materializeRow(row, {
        field: focusField,
        surface,
        value:
          currentValue ?? editorDraftRegistry.draftValueForFocusKey(focusKey),
      }),
      surface,
    ),
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
  captureActionBlocksRecord,
  beginViewportContinuity,
  clearViewportContinuity,
  clientInstanceId,
  conflictQueueRef,
  editorDraftRegistry,
  enqueuePendingReplayUnit,
  incidentId,
  latestCommittedTimelineRow,
  nextClientTxnId,
  pendingSavesRefs,
  rowsRef,
  rowStoreCommands,
}: {
  readonly captureActionBlocksRecord: (recordId: string) => boolean;
  readonly beginViewportContinuity: (
    request: ViewportContinuityRequest,
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly clientInstanceId: string;
  readonly conflictQueueRef: TimelineMutableRef<
    Record<string, LocalConflictState>
  >;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly enqueuePendingReplayUnit: (
    unit: TimelinePendingReplayAdmission,
    onSettled?: ((outcome: GridEditCommitOutcome) => void) | undefined,
  ) => void;

  readonly incidentId: string;
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
  readonly nextClientTxnId: () => string;
  readonly pendingSavesRefs: TimelinePendingSavesRefs;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly rowStoreCommands: TimelineRowStoreCommands;
}) {
  const { replaceRows } = rowStoreCommands;
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
      if (
        rowSnapshot.recordId &&
        captureActionBlocksRecord(rowSnapshot.recordId)
      ) {
        clearViewportContinuity(viewportContinuityToken);
        onSettled?.({
          kind: "conflict",
          message:
            "An earlier Timeline action needs to finish or be recovered. Your draft is retained.",
        });
        return;
      }
      pendingSavesRefs.pendingSignaturesRef.current.set(
        rowKey,
        mutationSignature,
      );
      const nextRows = rowsRef.current.map((row) =>
        row.key === rowKey
          ? {
              ...row,
              pendingSignature: mutationSignature,
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
      captureActionBlocksRecord,
      clearViewportContinuity,
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
      let admission = planTimelineScalarMutation({
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
      if (admission.kind === "accepted_duplicate" && onSettled !== undefined) {
        admission = planTimelineScalarMutation({
          allowZeroFieldCreate: options.allowZeroFieldCreate === true,
          clientTxnId,
          focusField,
          hasConflict: false,
          pendingSignature: undefined,
          row: snapshot,
        });
      }
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
        options.preserveInputFocus ||
          (snapshot.recordId === null &&
            editorDraftRegistry.inputElementForFocusKey(focusKey) ===
              document.activeElement)
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
      surface: TimelineScalarEditorSurface = "grid",
      onSettled?: (outcome: GridEditCommitOutcome) => void,
    ) => {
      const focusKey = inputFocusKey(rowKey, focusField, surface);
      const commitKey = inputFocusKey(rowKey, focusField, "grid");
      const rowSnapshot = rowsRef.current.find(
        (candidate) =>
          candidate.key === editorDraftRegistry.resolveRowKey(rowKey),
      );
      if (!rowSnapshot) {
        onSettled?.({
          kind: "stale_target",
          message: "The timeline row is no longer available.",
        });
        return;
      }
      const draftValue =
        draftValueOverride ?? rowSnapshot.collectionDrafts[focusField];
      const priorKeyboardCommitValue =
        pendingSavesRefs.collectionKeyboardCommitRef.current.get(commitKey);
      const commitDecision = decideTimelineCollectionCommit({
        draftValue,
        priorKeyboardCommitValue,
        source,
      });
      if (commitDecision.nextKeyboardCommitValue === null) {
        pendingSavesRefs.collectionKeyboardCommitRef.current.delete(commitKey);
      } else {
        pendingSavesRefs.collectionKeyboardCommitRef.current.set(
          commitKey,
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
      let admission = planTimelineCollectionMutation({
        clientTxnId,
        draftValue,
        effectiveRow: effectiveSnapshot,
        fieldKey,
        pendingSignature:
          pendingSavesRefs.pendingSignaturesRef.current.get(rowKey),
      });
      if (admission.kind === "accepted_duplicate" && onSettled !== undefined) {
        admission = planTimelineCollectionMutation({
          clientTxnId,
          draftValue,
          effectiveRow: effectiveSnapshot,
          fieldKey,
          pendingSignature: undefined,
        });
      }
      if (admission.kind !== "admit") {
        if (admission.kind === "accepted_no_change")
          onSettled?.({ kind: "accepted" });
        return;
      }
      const viewportContinuityToken = beginViewportContinuity(
        snapshot.recordId === null || onSettled !== undefined
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
        continueOnFreshDraft:
          onSettled === undefined && snapshot.recordId === null,
        detectAutoResolution: true,
        focusField,
        focusKey,
        mutationSignature: admission.mutationSignature,
        payloadIntent: admission.payloadIntent,
        promoteToCommittedRowInspect:
          surface === "inspector" && snapshot.recordId === null,
        rowKey: effectiveSnapshot.key,
        onSettled,
        surface,
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

  return {
    commands: {
      commitScalarGridEdit,
      queueCollectionSave,
      queueScalarSave,
    },
  };
}
