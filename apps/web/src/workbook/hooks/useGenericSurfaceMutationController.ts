import {
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookProtocolPatchRecordRequest } from "../adapters/workbookProtocolTypes";
import { taskViewId } from "../features/coordination/taskLifecycleModel";
import {
  type WorkbookInspectorErrorPresentation,
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
} from "../inspector/workbookInspectorErrorModel";
import type {
  GenericMutationCommandPort,
  GenericViewMutationAccepted,
} from "../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";

type GenericPatchMutationRequest = {
  readonly baseline?: WorkbookQueryRow;
  readonly baseRowVersion: number;
  readonly changes: readonly WorkbookProtocolPatchRecordRequest["changes"][number][];
  readonly purpose: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
};
export type GenericSurfaceMutationController = {
  readonly explicitPatches: WorkbookMutationRuntime["explicitPatches"];
  readonly taskDrafts: WorkbookMutationRuntime["taskDrafts"];
  readonly beginMutation: () => () => void;
  readonly beginMutationReport: () => () => void;
  readonly clearMutationError: () => void;
  readonly completeGenericMutation: () => Promise<void>;
  readonly mutationError: WorkbookInspectorErrorPresentation | null;
  readonly mutationPending: boolean;
  readonly rejectMutationFailure: (failure: WorkbookOperationFailure) => void;
  readonly setValidationError: (message: string) => void;
  readonly submitPatchMutation: (
    request: GenericPatchMutationRequest,
  ) => Promise<GenericViewMutationAccepted | null>;
};

export function useGenericSurfaceMutationController({
  mutationCommands,
  mutationRuntime,
  onRefresh,
  refreshReferenceOptions,
  surfaceLabel,
  sheetRef,
  selectedRecordId = "",
}: {
  readonly mutationCommands: GenericMutationCommandPort;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onRefresh: () => Promise<void> | void;
  readonly refreshReferenceOptions: () => Promise<void> | void;
  readonly surfaceLabel: string;
  readonly sheetRef: SheetRef;
  readonly selectedRecordId?: string;
}): GenericSurfaceMutationController {
  useSyncExternalStore(
    mutationRuntime.explicitPatches.subscribe,
    mutationRuntime.explicitPatches.getSnapshot,
  );
  const generation = useRef(0);
  useLayoutEffect(() => {
    void sheetRef;
    return () => {
      generation.current++;
    };
  }, [sheetRef]);
  const subjectRef = useRef(selectedRecordId);
  subjectRef.current = selectedRecordId;
  const [errorState, setErrorState] = useState<{
    subject: string;
    error: WorkbookInspectorErrorPresentation | null;
  }>({ subject: selectedRecordId, error: null });
  const setMutationError = useCallback(
    (error: WorkbookInspectorErrorPresentation | null) => {
      setErrorState({ subject: subjectRef.current, error });
    },
    [],
  );
  const mutationError =
    errorState.subject === selectedRecordId ? errorState.error : null;
  const [pendingCount, setPendingCount] = useState(0);
  const beginMutationReport = useCallback(
    () => mutationRuntime.beginExplicitMutation(),
    [mutationRuntime],
  );
  const beginMutation = useCallback(() => {
    const finish = beginMutationReport();
    setPendingCount((count) => count + 1);
    setMutationError(null);
    let finished = false;
    return () => {
      if (finished) return;
      finished = true;
      finish();
      setPendingCount((count) => count - 1);
    };
  }, [beginMutationReport, setMutationError]);
  const clearMutationError = useCallback(
    () => setMutationError(null),
    [setMutationError],
  );
  const rejectMutationFailure = useCallback(
    (failure: WorkbookOperationFailure) =>
      setMutationError(workbookInspectorErrorPresentation(failure)),
    [setMutationError],
  );
  const setValidationError = useCallback(
    (message: string) =>
      setMutationError(workbookInspectorLocalErrorPresentation(message)),
    [setMutationError],
  );
  const completeGenericMutation = useCallback(async () => {
    try {
      await onRefresh();
      await refreshReferenceOptions();
    } catch {
      setMutationError(
        workbookInspectorLocalErrorPresentation(
          "The change was accepted, but the workbook could not be refreshed. Previously loaded rows may be stale.",
        ),
      );
    }
  }, [onRefresh, refreshReferenceOptions, setMutationError]);
  const submitPatchMutation = useCallback(
    async (request: GenericPatchMutationRequest) => {
      const started = generation.current;
      const subject = subjectRef.current;
      if (request.viewSchemaId === taskViewId && request.baseline) {
        const result = await mutationRuntime.explicitPatches.submit({
          baseline: request.baseline,
          changes: request.changes,
          purpose: request.purpose,
          sheetRef,
          surfaceLabel,
        });
        if (
          started !== generation.current ||
          subject !== subjectRef.current ||
          !mutationRuntime.explicitPatches.getSnapshot().authority
        )
          return null;
        if (!result) {
          setValidationError(
            "This Task cannot be submitted while an operation needs recovery or authorization is unavailable. Your draft is retained.",
          );
          return null;
        }
        if (result.failure) rejectMutationFailure(result.failure);
        return result.receipt;
      }
      const result = await mutationCommands.patchRecord(request);
      if (result.kind === "rejected") {
        if (result.failure.kind === "same_field_conflict")
          mutationRuntime.registerConflict({
            conflict: result.failure.conflict,
            focusKey: `${request.recordId}:${result.failure.conflict.field_key}`,
            rowLabel: request.recordId,
            surfaceLabel,
            viewSchemaId: request.viewSchemaId,
            sheetRef,
          });
        rejectMutationFailure(result.failure);
        return null;
      }
      return result.value;
    },
    [
      mutationCommands,
      mutationRuntime,
      rejectMutationFailure,
      sheetRef,
      surfaceLabel,
      setValidationError,
    ],
  );
  return {
    explicitPatches: mutationRuntime.explicitPatches,
    taskDrafts: mutationRuntime.taskDrafts,
    beginMutation,
    beginMutationReport,
    clearMutationError,
    completeGenericMutation,
    mutationError,
    mutationPending: pendingCount > 0,
    rejectMutationFailure,
    setValidationError,
    submitPatchMutation,
  };
}
