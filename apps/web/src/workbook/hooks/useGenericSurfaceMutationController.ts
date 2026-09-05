import { useCallback, useState } from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookProtocolPatchRecordRequest } from "../adapters/workbookProtocolTypes";
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
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";

type GenericPatchMutationRequest = {
  readonly baseRowVersion: number;
  readonly changes: readonly WorkbookProtocolPatchRecordRequest["changes"][number][];
  readonly purpose: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
};
export type GenericSurfaceMutationController = {
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
}: {
  readonly mutationCommands: GenericMutationCommandPort;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onRefresh: () => Promise<void> | void;
  readonly refreshReferenceOptions: () => Promise<void> | void;
  readonly surfaceLabel: string;
  readonly sheetRef: SheetRef;
}): GenericSurfaceMutationController {
  const [mutationError, setMutationError] =
    useState<WorkbookInspectorErrorPresentation | null>(null);
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
  }, [beginMutationReport]);
  const clearMutationError = useCallback(() => setMutationError(null), []);
  const rejectMutationFailure = useCallback(
    (failure: WorkbookOperationFailure) =>
      setMutationError(workbookInspectorErrorPresentation(failure)),
    [],
  );
  const setValidationError = useCallback(
    (message: string) =>
      setMutationError(workbookInspectorLocalErrorPresentation(message)),
    [],
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
  }, [onRefresh, refreshReferenceOptions]);
  const submitPatchMutation = useCallback(
    async (request: GenericPatchMutationRequest) => {
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
    ],
  );
  return {
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
