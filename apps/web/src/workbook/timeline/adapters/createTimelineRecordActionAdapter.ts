import { createWorkbookOperationExecutor } from "../../mutations/workbookOperationExecutor";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type {
  TimelineRecordActionAccepted,
  TimelineRecordActionPort,
} from "../ports/TimelineRecordActionPort";

function invalidContract(): WorkbookOperationOutcome<TimelineRecordActionAccepted> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "The Timeline action response was invalid.",
    },
  };
}

export function createTimelineRecordActionAdapter(options: {
  readonly apiBase: string | undefined;
}): TimelineRecordActionPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async execute(input) {
      try {
        const outcome =
          input.action === "mark-reviewed"
            ? await operations.execute({
                operationID: "markTimelineRecordReviewed",
                pathParameters: { record_id: input.recordId },
                request: {
                  base_row_version: input.baseRowVersion,
                  client_txn_id: input.clientTxnId,
                  reason: "Reviewed from workbook",
                },
              })
            : await operations.execute({
                operationID: "supersedeRecord",
                pathParameters: { record_id: input.recordId },
                request: {
                  base_row_version: input.baseRowVersion,
                  client_txn_id: input.clientTxnId,
                  reason: "Superseded from workbook",
                  replacement_record_id: input.replacementRecordId ?? "",
                },
              });
        if (outcome.kind === "rejected") return outcome;
        const data = outcome.value.data;
        if (
          !("capture_state" in data) ||
          data.record_id !== input.recordId ||
          (input.action === "supersede" &&
            data.replacement_record_id !== input.replacementRecordId)
        ) {
          return invalidContract();
        }
        return {
          kind: "accepted",
          value: {
            captureState: data.capture_state,
            changeSetId: data.change_set_id,
            incidentId: data.incident_id,
            reason: data.reason,
            recordId: data.record_id,
            replacementRecordId: data.replacement_record_id,
            rowVersion: data.row_version,
          },
        };
      } catch {
        return {
          kind: "rejected",
          failure: {
            kind: "retryable",
            message: "The Timeline action could not be sent.",
          },
        };
      }
    },
  };
}
