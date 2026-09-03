import type { WorkbookOperationExecutor } from "../../adapters/workbookOperationContract";
import { buildGenericCreateRequest } from "../../features/generic/genericCreateRequestBuilder";
import { decodeCreateViewRowRequest } from "../../models/workbookRequestDecoders";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type {
  TimelineRelatedEvidenceLinked,
  TimelineRelatedRecordCreated,
  TimelineRelatedRecordPort,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import { normalizeTimelineFullRow } from "../models/timelineRowModel";
import { buildAttachedEvidencePatchRequest } from "./timelineEvidenceRequestBuilders";

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
    linkCreatedEvidence: (input) => linkCreatedEvidence(options, input),
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

async function linkCreatedEvidence(
  options: TimelineRelatedCommandAdapterOptions,
  input: Parameters<TimelineRelatedRecordPort["linkCreatedEvidence"]>[0],
): ReturnType<TimelineRelatedRecordPort["linkCreatedEvidence"]> {
  const clientTxnId = options.createClientTxnId(
    "timeline-link-created-evidence",
  );
  if (clientTxnId === null) return identityFailure();
  const request = buildAttachedEvidencePatchRequest(
    input.sourceRow,
    input.createdRecordId,
    clientTxnId,
  );
  if (request === null || input.sourceRow.recordId === null) {
    return {
      kind: "rejected",
      failure: {
        kind: "stale_target",
        message: "Created evidence, but the selected row version is stale.",
      },
    };
  }
  try {
    const outcome = await options.operations.execute({
      operationID: "patchRecord",
      pathParameters: { record_id: input.sourceRow.recordId },
      request,
    });
    if (outcome.kind === "rejected") return outcome;
    const data = outcome.value.data;
    if (
      data.view_schema_id !== timelineViewSchemaId ||
      data.row.record_id !== input.sourceRow.recordId
    ) {
      return invalidContract();
    }
    try {
      return {
        kind: "accepted",
        value: {
          changeSetId: data.change_set_id,
          row: normalizeTimelineFullRow(
            data.row,
            "related evidence link response row",
          ),
          viewSchemaId: data.view_schema_id,
        },
      };
    } catch {
      return invalidContract<TimelineRelatedEvidenceLinked>();
    }
  } catch {
    return retryable();
  }
}
