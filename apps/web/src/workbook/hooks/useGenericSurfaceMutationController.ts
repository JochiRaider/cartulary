import { useCallback, useRef, useState } from "react";
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

type GenericMutationSaveState = "Syncing" | "Saved" | "Conflict";

type GenericPatchMutationRequest = {
  readonly baseRowVersion: number;
  readonly changes: readonly Record<string, unknown>[];
  readonly purpose: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
};

export type GenericSurfaceMutationController = {
  readonly beginMutation: () => void;
  readonly clearMutationError: () => void;
  readonly completeGenericMutation: () => Promise<void>;
  readonly markMutationConflict: () => void;
  readonly markMutationSaved: () => void;
  readonly mutationError: WorkbookInspectorErrorPresentation | null;
  readonly mutationState: GenericMutationSaveState;
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
}: {
  readonly mutationCommands: GenericMutationCommandPort;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onRefresh: () => Promise<void> | void;
  readonly refreshReferenceOptions: () => Promise<void> | void;
  readonly surfaceLabel: string;
}): GenericSurfaceMutationController {
  const [mutationError, setMutationError] =
    useState<WorkbookInspectorErrorPresentation | null>(null);
  const [mutationState, setMutationState] =
    useState<GenericMutationSaveState>("Saved");
  const finishExplicitMutationRef = useRef<(() => void) | null>(null);

  const beginMutation = useCallback(() => {
    finishExplicitMutationRef.current?.();
    finishExplicitMutationRef.current = mutationRuntime.beginExplicitMutation();
    setMutationState("Syncing");
    setMutationError(null);
  }, [mutationRuntime]);

  const clearMutationError = useCallback(() => {
    setMutationError(null);
  }, []);

  const markMutationSaved = useCallback(() => {
    finishExplicitMutationRef.current?.();
    finishExplicitMutationRef.current = null;
    setMutationState("Saved");
  }, []);

  const markMutationConflict = useCallback(() => {
    finishExplicitMutationRef.current?.();
    finishExplicitMutationRef.current = null;
    setMutationState("Conflict");
  }, []);

  const rejectMutationFailure = useCallback(
    (failure: WorkbookOperationFailure) => {
      finishExplicitMutationRef.current?.();
      finishExplicitMutationRef.current = null;
      setMutationState("Conflict");
      setMutationError(workbookInspectorErrorPresentation(failure));
    },
    [],
  );

  const setValidationError = useCallback((message: string) => {
    setMutationError(workbookInspectorLocalErrorPresentation(message));
  }, []);

  const completeGenericMutation = useCallback(async () => {
    try {
      await onRefresh();
      await refreshReferenceOptions();
    } catch (error) {
      finishExplicitMutationRef.current?.();
      finishExplicitMutationRef.current = null;
      setMutationState("Conflict");
      setMutationError(
        workbookInspectorLocalErrorPresentation(
          error instanceof Error ? error.message : "Workbook refresh failed.",
        ),
      );
      return;
    }
    finishExplicitMutationRef.current?.();
    finishExplicitMutationRef.current = null;
    setMutationState("Saved");
  }, [onRefresh, refreshReferenceOptions]);

  const submitPatchMutation = useCallback(
    async ({
      baseRowVersion,
      changes,
      purpose,
      recordId,
      viewSchemaId,
    }: GenericPatchMutationRequest) => {
      beginMutation();
      const result = await mutationCommands.patchRecord({
        baseRowVersion,
        changes,
        purpose,
        recordId,
        viewSchemaId,
      });
      if (result.kind === "rejected") {
        if (result.failure.kind === "same_field_conflict") {
          mutationRuntime.registerConflict({
            conflict: result.failure.conflict,
            focusKey: `${recordId}:${result.failure.conflict.field_key}`,
            rowLabel: recordId,
            surfaceLabel,
            viewSchemaId,
          });
        }
        rejectMutationFailure(result.failure);
        return null;
      }
      return result.value;
    },
    [
      beginMutation,
      mutationCommands,
      mutationRuntime,
      rejectMutationFailure,
      surfaceLabel,
    ],
  );

  return {
    beginMutation,
    clearMutationError,
    completeGenericMutation,
    markMutationConflict,
    markMutationSaved,
    mutationError,
    mutationState,
    rejectMutationFailure,
    setValidationError,
    submitPatchMutation,
  };
}
