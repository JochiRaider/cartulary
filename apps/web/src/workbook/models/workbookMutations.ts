import { apiPath } from "../../services/browserApi";
import { fetchWorkbookJSON } from "../../services/workbookApi";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import { parseMutationError } from "./genericWorkbookModel";

export type WorkbookMutationSaveState = "Syncing" | "Saved" | "Conflict";

export type ViewMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    row: EntityApiRow;
  };
};

async function submitViewRecordPatch({
  apiBase,
  baseRowVersion,
  changes,
  clientTxnId,
  recordId,
  viewSchemaId,
}: {
  readonly apiBase: string | undefined;
  readonly baseRowVersion: number;
  readonly changes: readonly Record<string, unknown>[];
  readonly clientTxnId: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
}) {
  return fetchWorkbookJSON<ViewMutationEnvelope>(
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
}

export async function submitWorkbookPatchMutation({
  apiBase,
  baseRowVersion,
  changes,
  clientTxnId,
  onConflict,
  recordId,
  setMutationError,
  setMutationState,
  viewSchemaId,
}: {
  readonly apiBase: string | undefined;
  readonly baseRowVersion: number;
  readonly changes: readonly Record<string, unknown>[];
  readonly clientTxnId: string;
  readonly onConflict?: ((payload: unknown) => void) | undefined;
  readonly recordId: string;
  readonly setMutationError: (message: string | null) => void;
  readonly setMutationState: (state: WorkbookMutationSaveState) => void;
  readonly viewSchemaId: string;
}) {
  setMutationState("Syncing");
  setMutationError(null);
  const result = await submitViewRecordPatch({
    apiBase,
    baseRowVersion,
    changes,
    clientTxnId,
    recordId,
    viewSchemaId,
  });
  if (!result.ok) {
    if (result.status === 409) onConflict?.(result.payload);
    setMutationState("Conflict");
    setMutationError(parseMutationError(result.payload));
    return null;
  }
  return result.payload;
}
