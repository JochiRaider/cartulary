import type {
  PasteWorkbookClipboardRequest,
  PasteWorkbookClipboardResponse,
} from "@cartulary/protocol-ts/http";
import {
  workbookPasteColumns,
  workbookPasteTargets,
  workbookPasteViewSchemaId,
} from "../models/workbookClipboardPaste";
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
}): WorkbookClipboardPastePort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async paste(input) {
      if (!validPasteInput(input)) {
        return { clientTxnId: null, outcome: invalidPasteOutcome() };
      }
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
    },
  };
}
