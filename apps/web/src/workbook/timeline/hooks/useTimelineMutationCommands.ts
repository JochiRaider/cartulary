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
  timelineScalarBindings,
} from "../models/timelineFieldRegistry";
import type { TimelinePendingReplayAdmission } from "../models/timelineMutationDriverPlans";
import {
  planTimelineCollectionMutation,
  planTimelineScalarMutation,
} from "../models/timelineMutationQueueAdmission";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import type { WorkbookRow } from "../models/timelineRowModel";

type ViewportContinuityRequest =
  | { readonly kind: "input"; readonly focusKey: string }
  | { readonly kind: "row-inspect"; readonly recordId: string }
  | { readonly kind: "scroll-only" };

function resolveScalarSaveSnapshot({
  currentValue,
  editorDraftRegistry,
  latestCommittedTimelineRow,
  focusField,
  rowKey,
  rows,
  surface,
}: {
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
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
    ) ?? latestCommittedTimelineRow(editorDraftRegistry.resolveRowKey(rowKey));
  if (!row) return null;
  const focusKey = inputFocusKey(row.key, focusField, surface);
  const retained = editorDraftRegistry.draftValueForFocusKey(focusKey);
  // Also capture callers that submit a DOM value without an input notification.
  if (
    currentValue !== undefined &&
    currentValue !== retained &&
    (retained !== undefined || currentValue !== row.committedValues[focusField])
  ) {
    editorDraftRegistry.setDraft(
      { rowKey: row.key, field: focusField, surface },
      currentValue,
      row,
    );
  }
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
    onAdmissionRefused?: (() => void) | undefined,
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
      onAdmissionRefused,
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
      readonly viewportContinuityToken: number | undefined;
      readonly visibleEdit?: PendingReplayUnitInput["visibleEdit"];
      readonly onSettled?:
        | ((outcome: GridEditCommitOutcome) => void)
        | undefined;
      readonly onAdmissionRefused: () => void;
    }) => {
      if (
        rowSnapshot.recordId &&
        captureActionBlocksRecord(rowSnapshot.recordId)
      ) {
        if (viewportContinuityToken !== undefined)
          clearViewportContinuity(viewportContinuityToken);
        onAdmissionRefused();
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
        onAdmissionRefused,
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
        latestCommittedTimelineRow,
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
      const capturedRevisions = editorDraftRegistry.captureRow(
        effectiveRowKey,
        options.surface,
        snapshot.recordId === null
          ? undefined
          : new Set([
              binding.fieldKey,
              ...timelineScalarBindings
                .filter(
                  (binding) =>
                    snapshot.values[binding.key] !==
                    snapshot.committedValues[binding.key],
                )
                .map((binding) => binding.fieldKey),
            ]),
      );
      const revisions = JSON.stringify([...capturedRevisions]);
      const settlementKey = JSON.stringify([effectiveRowKey, options.surface]);
      const prior = pendingSavesRefs.scalarCommits.get(settlementKey);
      if (prior?.revisions === revisions) {
        if (onSettled) {
          if (prior.outcome) onSettled(prior.outcome);
          else prior.listeners.add(onSettled);
        }
        return;
      }
      const clientTxnId = nextClientTxnId();
      const admission = planTimelineScalarMutation({
        allowZeroFieldCreate:
          options.allowZeroFieldCreate === true ||
          (snapshot.recordId === null &&
            pendingSavesRefs.pendingQueueRef.current.model
              .snapshot()
              .units.some(
                (unit) =>
                  unit.rowKey === effectiveRowKey && unit.kind === "create",
              )),
        clientTxnId,
        focusField,
        hasConflict:
          snapshot.recordId !== null &&
          conflictQueueRef.current[
            `${snapshot.recordId}:${binding.fieldKey}`
          ] !== undefined,
        row: snapshot,
      });
      if (admission.kind !== "admit") {
        if (
          admission.kind === "accepted_no_change" &&
          capturedRevisions.get(focusKey) ===
            editorDraftRegistry.revisionForFocusKey(focusKey)
        ) {
          editorDraftRegistry.deleteDraftForFocusKey(focusKey);
        }
        onSettled?.(
          admission.kind === "rejected"
            ? admission.outcome
            : { kind: "accepted" },
        );
        return;
      }
      const commit: NonNullable<
        ReturnType<typeof pendingSavesRefs.scalarCommits.get>
      > = {
        revisions,
        listeners: new Set(onSettled ? [onSettled] : []),
      };
      pendingSavesRefs.scalarCommits.set(settlementKey, commit);
      const settle = (outcome: GridEditCommitOutcome) => {
        commit.outcome = outcome;
        for (const listener of commit.listeners) listener(outcome);
        commit.listeners.clear();
      };
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
        // Operation identity survives equal-text gestures; retries still use
        // the captured request and transaction owned by the existing driver.
        mutationSignature: JSON.stringify([
          admission.mutationSignature,
          clientTxnId,
        ]),
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
        onSettled: settle,
        onAdmissionRefused: () => {
          // No operation was admitted: a later explicit gesture can retry this
          // authoring revision after the admission restriction is resolved.
          if (pendingSavesRefs.scalarCommits.get(settlementKey) === commit)
            pendingSavesRefs.scalarCommits.delete(settlementKey);
        },
      });
    },
    [
      beginViewportContinuity,
      conflictQueueRef,
      editorDraftRegistry,
      enqueueAutosaveReplayForPendingMutation,
      latestCommittedTimelineRow,
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
      surface: TimelineScalarEditorSurface = "grid",
      onSettled?: (outcome: GridEditCommitOutcome) => void,
    ) => {
      const focusKey = inputFocusKey(
        editorDraftRegistry.resolveRowKey(rowKey),
        focusField,
        surface,
      );
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
        draftValueOverride ??
        editorDraftRegistry.draftValueForFocusKey(focusKey) ??
        (surface === "grid" ? rowSnapshot.collectionDrafts[focusField] : "");
      const prior = pendingSavesRefs.collectionCommits.get(focusKey);
      const currentRevision = editorDraftRegistry.revisionForFocusKey(focusKey);
      if (
        prior &&
        prior.value === draftValue &&
        (prior.revision === currentRevision || currentRevision === 0)
      ) {
        if (onSettled) {
          if (prior.outcome) onSettled(prior.outcome);
          else prior.listeners.add(onSettled);
        }
        return;
      }
      if (draftValue.trim() === "") {
        onSettled?.({ kind: "accepted" });
        return;
      }
      // Capture legacy/recordless callers at the same runtime-owned boundary.
      editorDraftRegistry.setDraft(
        { rowKey: rowSnapshot.key, field: focusField, surface },
        draftValue,
      );
      const commit: NonNullable<
        ReturnType<typeof pendingSavesRefs.collectionCommits.get>
      > = {
        revision: editorDraftRegistry.revisionForFocusKey(focusKey),
        value: draftValue,
        listeners: new Set(onSettled ? [onSettled] : []),
      };
      pendingSavesRefs.collectionCommits.set(focusKey, commit);
      const settle = (outcome: GridEditCommitOutcome) => {
        commit.outcome = outcome;
        for (const listener of commit.listeners) listener(outcome);
        commit.listeners.clear();
      };
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
      const materialized = editorDraftRegistry.materializeRow(
        collectionSnapshot,
        { surface },
      );
      const effectiveSnapshot =
        snapshot.recordId === null
          ? materialized
          : { ...materialized, values: { ...snapshot.committedValues } };
      const clientTxnId = nextClientTxnId();
      const admission = planTimelineCollectionMutation({
        clientTxnId,
        draftValue,
        effectiveRow: effectiveSnapshot,
        fieldKey,
      });
      if (admission.kind !== "admit") {
        settle(
          admission.kind === "rejected"
            ? admission.outcome
            : { kind: "accepted" },
        );
        return;
      }
      if (snapshot.recordId === null)
        editorDraftRegistry.beginCapture(snapshot.key);
      // A caller awaiting settlement owns its captured departure. Starting a
      // second restoration here would race that destination after acceptance.
      const viewportContinuityToken =
        onSettled === undefined
          ? beginViewportContinuity(
              snapshot.recordId === null
                ? { kind: "scroll-only" }
                : { kind: "row-inspect", recordId: snapshot.recordId },
            )
          : undefined;
      enqueueAutosaveReplayForPendingMutation({
        clientTxnId,
        continueOnFreshDraft:
          onSettled === undefined && snapshot.recordId === null,
        detectAutoResolution: true,
        focusField,
        focusKey,
        mutationSignature: JSON.stringify([
          admission.mutationSignature,
          focusKey,
          commit.revision,
        ]),
        payloadIntent: admission.payloadIntent,
        promoteToCommittedRowInspect:
          surface === "inspector" && snapshot.recordId === null,
        rowKey: effectiveSnapshot.key,
        onSettled: settle,
        onAdmissionRefused: () => {
          if (pendingSavesRefs.collectionCommits.get(focusKey) === commit)
            pendingSavesRefs.collectionCommits.delete(focusKey);
        },
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
