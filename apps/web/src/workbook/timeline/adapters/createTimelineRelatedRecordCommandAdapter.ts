import type { WorkbookOperationExecutor } from "../../adapters/workbookOperationContract";
import { buildGenericCreateRequest } from "../../features/generic/genericCreateRequestBuilder";
import { decodeCreateViewRowRequest } from "../../models/workbookRequestDecoders";
import type {
  TimelineRelatedRecordCreated,
  TimelineRelatedRecordPort,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";

type TimelineRelatedCommandAdapterOptions = {
  readonly createClientTxnId: (prefix: string) => string | null;
  readonly incidentId: string;
  readonly operations: WorkbookOperationExecutor;
};

function identityFailure<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "terminal",
      message: "A secure transaction ID could not be created.",
    },
  };
}

function invalidPayload<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: { kind: "validation", message: "invalid_mutation_payload" },
  };
}

function invalidContract<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "The server returned an inconsistent Workbook operation result.",
    },
  };
}

function retryable<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "retryable",
      message: "The Workbook operation could not be sent.",
    },
  };
}

export function createTimelineRelatedRecordCommandAdapter(
  options: TimelineRelatedCommandAdapterOptions,
): TimelineRelatedRecordPort {
  return {
    createRelatedRecord: (input) => createRelatedRecord(options, input),
  };
}

async function createRelatedRecord(
  options: TimelineRelatedCommandAdapterOptions,
  input: Parameters<TimelineRelatedRecordPort["createRelatedRecord"]>[0],
): ReturnType<TimelineRelatedRecordPort["createRelatedRecord"]> {
  const clientTxnId = options.createClientTxnId(
    `timeline-create-related-${input.featureGroupKey}`,
  );
  if (clientTxnId === null) return identityFailure();
  const payload = buildGenericCreateRequest(
    input.contract,
    { ...input.draft },
    clientTxnId,
  );
  const request = decodeCreateViewRowRequest(input.contract, payload);
  if (request === null) return invalidPayload();
  try {
    const outcome = await options.operations.execute({
      operationID: "createViewRow",
      pathParameters: {
        incident_id: options.incidentId,
        view_schema_id: input.contract.viewSchemaId,
      },
      request,
    });
    if (outcome.kind === "rejected") return outcome;
    const data = outcome.value.data;
    return data.view_schema_id === input.contract.viewSchemaId
      ? {
          kind: "accepted",
          value: {
            changeSetId: data.change_set_id,
            recordId: data.row.record_id,
            viewSchemaId: data.view_schema_id,
          },
        }
      : invalidContract<TimelineRelatedRecordCreated>();
  } catch {
    return retryable();
  }
}
