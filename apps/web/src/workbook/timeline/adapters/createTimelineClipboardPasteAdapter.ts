import type { PasteWorkbookClipboardRequest } from "@cartulary/protocol-ts/http";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import {
  parseSameFieldConflictPayload,
  type WorkbookSameFieldConflictPayload,
} from "../../runtime/workbookConflictModel";
import { normalizeTimelineFullRow } from "../models/workbookTimelineModel";
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

function nonEmptyTuple<Value>(
  values: readonly Value[],
): [Value, ...Value[]] | null {
  const [first, ...rest] = values;
  return first === undefined ? null : [first, ...rest];
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
      const columns = nonEmptyTuple(input.columns);
      const targets = nonEmptyTuple(
        input.targets.map((target) =>
          target.kind === "create"
            ? { kind: "create" as const }
            : {
                base_row_version: target.baseRowVersion,
                kind: "record" as const,
                record_id: target.recordId,
              },
        ),
      );
      if (targets === null || columns === null) {
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
        columns,
        format: input.format,
        start_field_key: input.startFieldKey,
        targets,
        view_schema_id: timelineViewSchemaId,
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
          const conflicts: WorkbookSameFieldConflictPayload[] = [];
          for (const value of data.conflicts ?? []) {
            const conflict = parseSameFieldConflictPayload(value);
            if (conflict === null) return invalidContract();
            conflicts.push(conflict);
          }
          return {
            kind: "accepted",
            value: {
              conflicts,
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
