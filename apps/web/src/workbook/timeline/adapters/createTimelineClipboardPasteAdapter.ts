import type { PasteWorkbookClipboardRequest } from "@cartulary/protocol-ts";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createWorkbookOperationExecutor } from "../../mutations/workbookOperationExecutor";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import { parseSameFieldConflictFields } from "../../runtime/workbookConflictModel";
import {
  normalizeTimelineFullRow,
  type SameFieldConflictPayload,
} from "../models/workbookTimelineModel";
import type {
  TimelineClipboardPasteAccepted,
  TimelineClipboardPastePort,
} from "../ports/TimelineClipboardPastePort";

function invalidContract(): WorkbookOperationOutcome<TimelineClipboardPasteAccepted> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "The Timeline paste response was invalid.",
    },
  };
}

function normalizeConflict(value: unknown): SameFieldConflictPayload | null {
  const fields = parseSameFieldConflictFields(value);
  if (
    fields === null ||
    !("client_value" in fields) ||
    !("server_value" in fields)
  ) {
    return null;
  }
  return fields as SameFieldConflictPayload;
}

export function createTimelineClipboardPasteAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): TimelineClipboardPastePort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async paste(input) {
      if (input.targets.length === 0 || input.columns.length === 0) {
        return {
          kind: "rejected",
          failure: {
            kind: "validation",
            message: "The Timeline paste target was invalid.",
          },
        };
      }
      const request: PasteWorkbookClipboardRequest = {
        client_txn_id: input.clientTxnId,
        clipboard_text: input.clipboardText,
        columns: input.columns as PasteWorkbookClipboardRequest["columns"],
        format: input.format,
        start_field_key: input.startFieldKey,
        targets: input.targets.map((target) =>
          target.kind === "create"
            ? { kind: "create" as const }
            : {
                base_row_version: target.baseRowVersion,
                kind: "record" as const,
                record_id: target.recordId,
              },
        ) as PasteWorkbookClipboardRequest["targets"],
        view_schema_id:
          timelineViewSchemaId as PasteWorkbookClipboardRequest["view_schema_id"],
      };
      try {
        const outcome = await operations.execute({
          operationID: "pasteWorkbookClipboard",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: timelineViewSchemaId,
          },
          request,
        });
        if (outcome.kind === "rejected") return outcome;
        const data = outcome.value.data;
        if (data.view_schema_id !== timelineViewSchemaId) {
          return invalidContract();
        }
        try {
          const rows = data.rows.map((row) =>
            normalizeTimelineFullRow(row, "clipboard paste response row"),
          );
          const conflicts = (data.conflicts ?? []).map(normalizeConflict);
          if (conflicts.some((conflict) => conflict === null)) {
            return invalidContract();
          }
          return {
            kind: "accepted",
            value: {
              conflicts: conflicts as SameFieldConflictPayload[],
              rows,
            },
          };
        } catch {
          return invalidContract();
        }
      } catch {
        return {
          kind: "rejected",
          failure: {
            kind: "retryable",
            message: "The Timeline paste could not be sent.",
          },
        };
      }
    },
  };
}
