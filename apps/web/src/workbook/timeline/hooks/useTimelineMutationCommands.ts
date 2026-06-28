import {
  type Dispatch,
  type SetStateAction,
  startTransition,
  useCallback,
} from "react";
import { apiPath } from "../../../services/browserApi";
import { fetchJSON, readEnvelope } from "../../../services/workbookApi";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  buildStableMutationSignature,
  type PendingReplayUnitInput,
} from "../../utils/workbookPendingQueue";
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
import { buildTimelineRecordActionPayload } from "../services/timelineMutationRequests";
import type { PendingReplayRuntimeMeta } from "./useTimelinePendingReplayController";
import {
  ensureTimelineTabClientInstanceId,
  type TimelineMutableRef,
  type TimelinePendingReplayAdmissionRequest,
  type TimelinePendingSavesRefs,
} from "./useTimelinePendingSaves";

type TimelineActionEnvelope = {
  data: {
    record_id: string;
    incident_id: string;
    row_version: number;
    capture_state: string;
    change_set_id: string;
    reason: string | null;
    replacement_record_id: string | null;
  };
};

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

export type TimelineScalarSaveOptions = {
  readonly allowZeroFieldCreate?: boolean | undefined;
  readonly continueOnFreshDraft: boolean;
  readonly preserveInputFocus: boolean;
  readonly surface: TimelineScalarEditorSurface;
};

export function useTimelineMutationCommands({
  acceptTimelineActionResult,
  apiBase,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  conflictQueueRef,
  enqueuePendingReplayUnit,
  finishSave,
  incidentId,
  latestCommittedTimelineRow,
  loadRowsRef,
  nextClientTxnId,
  pendingSavesRefsRef,
  replacementDrafts,
  resolvePendingSocketTxn,
  rowWithScalarEditorDrafts,
  rowsRef,
  scalarDraftValuesRef,
  setRows,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly acceptTimelineActionResult: (
    data: TimelineActionEnvelope["data"],
  ) => void;
  readonly apiBase?: string | undefined;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    request: ViewportContinuityRequest,
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly conflictQueueRef: TimelineMutableRef<
    Record<string, LocalConflictState>
  >;
  readonly enqueuePendingReplayUnit: (
    unit: TimelinePendingReplayAdmissionRequest<PendingReplayRuntimeMeta>,
  ) => void;
  readonly finishSave: (result: "Conflict" | "Saved" | "Syncing") => void;
  readonly incidentId: string;
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
  readonly loadRowsRef: TimelineMutableRef<LoadRowsForMutation>;
  readonly nextClientTxnId: () => string;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    TimelinePendingSavesRefs<PendingReplayRuntimeMeta>
  >;
  readonly replacementDrafts: Record<string, string>;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
  readonly rowWithScalarEditorDrafts: (
    row: WorkbookRow,
    preferred?: {
      readonly field: keyof RowValues;
      readonly value: string | undefined;
    },
  ) => WorkbookRow;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly scalarDraftValuesRef: TimelineMutableRef<Map<string, string>>;
  readonly setRows: Dispatch<SetStateAction<WorkbookRow[]>>;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<CommittedRecordIdle | null>;
}) {
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
    }) => {
      pendingSavesRefsRef.current.pendingSignaturesRef.current.set(
        rowKey,
        mutationSignature,
      );
      startTransition(() => {
        setRows((current) => {
          const nextRows = current.map((row) =>
            row.key === rowKey
              ? { ...row, pendingSignature: mutationSignature }
              : row,
          );
          rowsRef.current = nextRows;
          return nextRows;
        });
      });

      const targetPath =
        rowSnapshot.recordId === null
          ? apiPath(
              apiBase,
              `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
            )
          : apiPath(apiBase, `/api/v1/records/${rowSnapshot.recordId}`);
      const clientInstanceId = ensureTimelineTabClientInstanceId(
        pendingSavesRefsRef.current.socketClientInstanceIdRef,
      );
      enqueuePendingReplayUnit({
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
        method: rowSnapshot.recordId === null ? "POST" : "PATCH",
        path: targetPath,
        payloadIntent,
        clientTxnId,
        mutationSignature,
        coalesceKey:
          rowSnapshot.recordId === null
            ? `draft:${rowKey}`
            : `record:${rowSnapshot.recordId}`,
        enqueueOrder: pendingSavesRefsRef.current.pendingReplayOrderRef.current,
        operationClass: "hot_path",
        status: "queued",
        ...(visibleEdit === undefined ? {} : { visibleEdit }),
        rowSnapshot,
        continueOnFreshDraft,
        detectAutoResolution,
        promoteToCommittedRowInspect,
        viewportContinuityToken,
      });
      pendingSavesRefsRef.current.pendingReplayOrderRef.current += 1;
    },
    [
      apiBase,
      enqueuePendingReplayUnit,
      incidentId,
      pendingSavesRefsRef,
      rowsRef,
      setRows,
    ],
  );

  const queueScalarSave = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      options: TimelineScalarSaveOptions,
      currentValue?: string,
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
          : rowWithScalarEditorDrafts(rowSnapshot, {
              field: focusField,
              value: scalarDraftValuesRef.current.get(focusKey) ?? currentValue,
            });
      if (!snapshot) {
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
        scalarDraftValuesRef.current.delete(focusKey);
        return;
      }

      const mutationSignature = buildStableMutationSignature(payload);
      if (
        pendingSavesRefsRef.current.pendingSignaturesRef.current.get(
          effectiveRowKey,
        ) === mutationSignature
      ) {
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
      });
    },
    [
      beginViewportContinuity,
      conflictQueueRef,
      enqueueAutosaveReplayForPendingMutation,
      nextClientTxnId,
      pendingSavesRefsRef,
      rowWithScalarEditorDrafts,
      rowsRef,
      scalarDraftValuesRef,
    ],
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
        pendingSavesRefsRef.current.collectionKeyboardCommitRef.current.get(
          focusKey,
        );
      if (source === "blur") {
        pendingSavesRefsRef.current.collectionKeyboardCommitRef.current.delete(
          focusKey,
        );
        if (priorKeyboardCommitValue === draftValue) {
          return;
        }
      } else {
        pendingSavesRefsRef.current.collectionKeyboardCommitRef.current.set(
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
      const effectiveSnapshot = rowWithScalarEditorDrafts(collectionSnapshot);
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
        pendingSavesRefsRef.current.pendingSignaturesRef.current.get(rowKey) ===
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
      enqueueAutosaveReplayForPendingMutation,
      latestCommittedTimelineRow,
      nextClientTxnId,
      pendingSavesRefsRef,
      rowWithScalarEditorDrafts,
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
      pendingSavesRefsRef.current.saveQueueRef.current =
        pendingSavesRefsRef.current.saveQueueRef.current
          .catch(() => undefined)
          .then(async () => {
            const idleRecord = await waitForCommittedRecordIdle(recordId);
            if (idleRecord === null) {
              clearViewportContinuity(viewportContinuityToken);
              finishSave("Conflict");
              return;
            }
            const body = buildTimelineRecordActionPayload({
              action,
              baseRowVersion: idleRecord.rowVersion,
              clientTxnId,
              replacementRecordId,
            });
            trackPendingSocketTxn(clientTxnId);
            const result = await fetchJSON<TimelineActionEnvelope>(
              apiPath(apiBase, `/api/v1/records/${recordId}/${action}`),
              {
                method: "POST",
                body: JSON.stringify(body),
              },
            );
            if (!result.ok) {
              resolvePendingSocketTxn(clientTxnId);
              clearViewportContinuity(viewportContinuityToken);
              finishSave("Conflict");
              return;
            }

            const envelope = readEnvelope<TimelineActionEnvelope>(
              result.payload,
            );
            acceptTimelineActionResult(envelope.data);
            await loadRowsRef.current({
              showLoading: false,
              viewportContinuityToken,
            });
            finishSave("Saved");
          });
    },
    [
      acceptTimelineActionResult,
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      finishSave,
      loadRowsRef,
      nextClientTxnId,
      pendingSavesRefsRef,
      replacementDrafts,
      resolvePendingSocketTxn,
      rowsRef,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  return {
    queueAction,
    queueCollectionSave,
    queueScalarSave,
  };
}
