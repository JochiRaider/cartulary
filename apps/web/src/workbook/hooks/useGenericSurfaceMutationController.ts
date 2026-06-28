import { useCallback, useState } from "react";
import { apiPath } from "../../services/browserApi";
import { fetchJSON, readEnvelope } from "../../services/workbookApi";
import { parseMutationError } from "../models/genericWorkbookModel";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";

export type GenericMutationSaveState = "Syncing" | "Saved" | "Conflict";

export type GenericViewMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    row: EntityApiRow;
  };
};

export type GenericPatchMutationRequest = {
  readonly baseRowVersion: number;
  readonly changes: readonly Record<string, unknown>[];
  readonly clientTxnId: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
};

export type GenericSurfaceMutationController = {
  readonly beginMutation: () => void;
  readonly clearMutationError: () => void;
  readonly completeGenericMutation: <TEnvelope>(
    payload: unknown,
  ) => Promise<TEnvelope>;
  readonly markMutationConflict: () => void;
  readonly markMutationSaved: () => void;
  readonly mutationError: string | null;
  readonly mutationState: GenericMutationSaveState;
  readonly rejectMutationPayload: (payload: unknown) => void;
  readonly setValidationError: (message: string) => void;
  readonly submitPatchMutation: (
    request: GenericPatchMutationRequest,
  ) => Promise<unknown | null>;
};

export function useGenericSurfaceMutationController({
  apiBase,
  onRefresh,
  refreshReferenceOptions,
}: {
  readonly apiBase: string | undefined;
  readonly onRefresh: () => Promise<void> | void;
  readonly refreshReferenceOptions: () => Promise<void> | void;
}): GenericSurfaceMutationController {
  const [mutationError, setMutationError] = useState<string | null>(null);
  const [mutationState, setMutationState] =
    useState<GenericMutationSaveState>("Saved");

  const beginMutation = useCallback(() => {
    setMutationState("Syncing");
    setMutationError(null);
  }, []);

  const clearMutationError = useCallback(() => {
    setMutationError(null);
  }, []);

  const markMutationSaved = useCallback(() => {
    setMutationState("Saved");
  }, []);

  const markMutationConflict = useCallback(() => {
    setMutationState("Conflict");
  }, []);

  const rejectMutationPayload = useCallback((payload: unknown) => {
    setMutationState("Conflict");
    setMutationError(parseMutationError(payload));
  }, []);

  const setValidationError = useCallback((message: string) => {
    setMutationError(message);
  }, []);

  const completeGenericMutation = useCallback(
    async <TEnvelope>(payload: unknown) => {
      const envelope = readEnvelope<TEnvelope>(payload);
      try {
        await onRefresh();
        await refreshReferenceOptions();
      } catch (error) {
        setMutationState("Conflict");
        setMutationError(
          error instanceof Error ? error.message : "Workbook refresh failed.",
        );
        return envelope;
      }
      setMutationState("Saved");
      return envelope;
    },
    [onRefresh, refreshReferenceOptions],
  );

  const submitPatchMutation = useCallback(
    async ({
      baseRowVersion,
      changes,
      clientTxnId,
      recordId,
      viewSchemaId,
    }: GenericPatchMutationRequest) => {
      beginMutation();
      const result = await fetchJSON<GenericViewMutationEnvelope>(
        apiPath(apiBase, `/api/v1/records/${recordId}`),
        {
          method: "PATCH",
          body: JSON.stringify({
            view_schema_id: viewSchemaId,
            base_row_version: baseRowVersion,
            client_txn_id: clientTxnId,
            changes,
          }),
        },
      );
      if (!result.ok) {
        rejectMutationPayload(result.payload);
        return null;
      }
      return result.payload;
    },
    [apiBase, beginMutation, rejectMutationPayload],
  );

  return {
    beginMutation,
    clearMutationError,
    completeGenericMutation,
    markMutationConflict,
    markMutationSaved,
    mutationError,
    mutationState,
    rejectMutationPayload,
    setValidationError,
    submitPatchMutation,
  };
}
