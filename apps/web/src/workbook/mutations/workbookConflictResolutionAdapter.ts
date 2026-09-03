import { createWorkbookOperationExecutor } from "../adapters/workbookOperationExecutor";
import type { WorkbookProtocolResolveConflictRequest } from "../adapters/workbookProtocolTypes";
import type { WorkbookOperationOutcome } from "./workbookOperationOutcome";

export type WorkbookResolvedMutation = {
  readonly row: unknown;
  readonly viewSchemaId: string;
};

export async function executeWorkbookConflictResolution(input: {
  readonly apiBase: string | undefined;
  readonly conflictToken: string;
  readonly recordId: string;
  readonly request: WorkbookProtocolResolveConflictRequest;
}): Promise<WorkbookOperationOutcome<WorkbookResolvedMutation>> {
  try {
    const outcome = await createWorkbookOperationExecutor({
      apiBase: input.apiBase,
    }).execute({
      operationID: "resolveRecordSameFieldConflict",
      pathParameters: {
        conflict_token: input.conflictToken,
        record_id: input.recordId,
      },
      request: input.request,
    });
    if (outcome.kind === "rejected") return outcome;
    return {
      kind: "accepted",
      value: {
        row: outcome.value.data.row,
        viewSchemaId: outcome.value.data.view_schema_id,
      },
    };
  } catch {
    return {
      kind: "rejected",
      failure: {
        kind: "retryable",
        message: "The conflict resolution could not be sent.",
      },
    };
  }
}
