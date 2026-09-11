import type {
  PasteWorkbookClipboardRequest,
  PasteWorkbookClipboardResponse,
} from "@cartulary/protocol-ts/http";
import {
  workbookPasteColumns,
  workbookPasteTargets,
  workbookPasteViewSchemaId,
} from "../models/workbookClipboardPaste";
import type { EntityRecordWriteBoundary } from "../mutations/entityRecordWriteBoundary";
import type { SecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import type {
  WorkbookClipboardPasteAccepted,
  WorkbookClipboardPasteInput,
  WorkbookClipboardPastePort,
  WorkbookClipboardPasteResult,
} from "./WorkbookClipboardPastePort";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

const invalidMessage = "The Workbook paste request or response was invalid.";

function invalidPasteOutcome(): WorkbookOperationOutcome<WorkbookClipboardPasteAccepted> {
  return {
    kind: "rejected",
    failure: { kind: "invalid_contract", message: invalidMessage },
  };
}

function retryablePasteOutcome(): WorkbookOperationOutcome<WorkbookClipboardPasteAccepted> {
  return {
    kind: "rejected",
    failure: {
      kind: "retryable",
      message: "The Workbook paste could not be sent.",
    },
  };
}

function secureIdFailure(): WorkbookClipboardPasteResult {
  return {
    clientTxnId: null,
    outcome: {
      kind: "rejected",
      failure: {
        kind: "terminal",
        message: "A secure transaction ID could not be created.",
      },
    },
  };
}

function validPasteInput(input: WorkbookClipboardPasteInput): boolean {
  if (
    input.clipboard_text.length === 0 ||
    input.start_field_key.trim().length === 0 ||
    workbookPasteViewSchemaId(input.view_schema_id) === null ||
    workbookPasteColumns(input.columns) === null ||
    workbookPasteTargets(input.targets) === null
  ) {
    return false;
  }
  return true;
}

function acceptedPaste(
  request: PasteWorkbookClipboardRequest,
  data: PasteWorkbookClipboardResponse["data"],
): WorkbookOperationOutcome<WorkbookClipboardPasteAccepted> {
  if (data.view_schema_id !== request.view_schema_id) {
    return invalidPasteOutcome();
  }
  return {
    kind: "accepted",
    value: {
      changeSetId: data.change_set_id ?? null,
      conflicts: [...(data.conflicts ?? [])],
      rows: data.rows,
      viewSchemaId: request.view_schema_id,
    },
  };
}

export function createWorkbookClipboardPasteAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly transactionIds: SecureTransactionIdPort;
  readonly entityWrites?: EntityRecordWriteBoundary;
}): WorkbookClipboardPastePort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async paste(input) {
      if (!validPasteInput(input)) {
        return { clientTxnId: null, outcome: invalidPasteOutcome() };
      }
      const entityType =
        input.view_schema_id === "cartulary.view.hosts.v1"
          ? "host"
          : input.view_schema_id === "cartulary.view.identities.v1"
            ? "identity"
            : null;
      const release =
        entityType && options.entityWrites
          ? options.entityWrites.begin({
              recordIds: input.targets.flatMap((target) =>
                target.kind === "record" ? [target.record_id] : [],
              ),
              ...(input.targets.some((target) => target.kind === "create")
                ? { unknownEntityType: entityType }
                : {}),
            })
          : () => {};
      if (release === null)
        return {
          clientTxnId: null,
          outcome: {
            kind: "rejected",
            failure: {
              kind: "stale_target",
              message:
                "Recover the pending merge in Merge actions before pasting into these records.",
            },
          },
        };
      try {
        let clientTxnId: string;
        try {
          clientTxnId = options.transactionIds.create(
            `${input.view_schema_id}-clipboard-paste`,
          );
        } catch {
          return secureIdFailure();
        }
        input.onClientTxnId?.(clientTxnId);
        const { onClientTxnId: _onClientTxnId, ...requestInput } = input;
        const request: PasteWorkbookClipboardRequest = {
          ...requestInput,
          client_txn_id: clientTxnId,
        };
        try {
          const result = await operations.execute({
            operationID: "pasteWorkbookClipboard",
            pathParameters: {
              incident_id: options.incidentId,
              view_schema_id: request.view_schema_id,
            },
            request,
          });
          if (result.kind === "accepted" && entityType)
            for (const row of result.value.data.rows)
              options.entityWrites?.acceptVersion(
                row.record_id,
                row.row_version,
              );
          return {
            clientTxnId,
            outcome:
              result.kind === "rejected"
                ? result
                : acceptedPaste(request, result.value.data),
          };
        } catch {
          return { clientTxnId, outcome: retryablePasteOutcome() };
        }
      } finally {
        release();
      }
    },
  };
}
