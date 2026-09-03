import type {
  CreateViewRowResponse,
  EvidenceCreateRequest,
} from "@cartulary/protocol-ts/http";
import {
  createUploadedEvidenceObjectBlob,
  evidenceAttachPublicErrorMessage,
} from "../../../services/workbookEvidence";
import type { WorkbookOperationExecutor } from "../../adapters/workbookOperationContract";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import { normalizeTimelineFullRow } from "../models/timelineRowModel";
import type {
  TimelineEvidenceAttachmentAccepted,
  TimelineEvidenceAttachmentPort,
  TimelineEvidenceCreated,
} from "../ports/TimelineEvidenceAttachmentPort";
import {
  buildAttachedEvidenceCreateRequest,
  buildAttachedEvidencePatchRequest,
} from "./timelineEvidenceRequestBuilders";

type TimelineEvidenceAdapterOptions = {
  readonly apiBase: string | undefined;
  readonly createClientTxnId: () => string;
  readonly incidentId: string;
};

type TimelineEvidenceCommand =
  | {
      readonly incidentId: string;
      readonly kind: "create";
      readonly request: ReturnType<typeof buildAttachedEvidenceCreateRequest>;
    }
  | {
      readonly kind: "patch";
      readonly recordId: string;
      readonly request: NonNullable<
        ReturnType<typeof buildAttachedEvidencePatchRequest>
      >;
    };

function invalidContract<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "The Timeline evidence response was invalid.",
    },
  };
}

function terminal<T>(message: string): WorkbookOperationOutcome<T> {
  return { kind: "rejected", failure: { kind: "terminal", message } };
}

function evidenceTitleFromFile(file: File): string {
  return file.name.trim() || "Workbook attachment";
}

export function createTimelineEvidenceAttachmentAdapter(
  options: TimelineEvidenceAdapterOptions,
): TimelineEvidenceAttachmentPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    attachEvidence: (input) =>
      attachEvidenceToTimeline(options, operations, input),
    createEvidence: ({ file }) =>
      createEvidenceFromFile(options, operations, file),
  };
}

async function createEvidenceFromFile(
  options: TimelineEvidenceAdapterOptions,
  operations: WorkbookOperationExecutor,
  file: File,
): Promise<WorkbookOperationOutcome<TimelineEvidenceCreated>> {
  try {
    const objectBlobId = await createUploadedEvidenceObjectBlob({
      apiBase: options.apiBase,
      createClientTxnId: options.createClientTxnId,
      file,
      incidentId: options.incidentId,
    });
    const request: EvidenceCreateRequest = {
      client_txn_id: options.createClientTxnId(),
      "evidence.collector_party_text": "Workbook upload",
      "evidence.initial_object_blob_id": objectBlobId,
      "evidence.lifecycle_state": "available",
      "evidence.title": evidenceTitleFromFile(file),
    };
    const outcome = await createEvidenceRowWithUncertainResponseRetry(
      operations,
      options.incidentId,
      request,
    );
    if (outcome.kind === "rejected") return outcome;
    const data = outcome.value.data;
    return data.view_schema_id === evidenceViewSchemaId
      ? {
          kind: "accepted",
          value: { evidenceRecordId: data.row.record_id },
        }
      : invalidContract();
  } catch (error) {
    return terminal(evidenceAttachPublicErrorMessage(error));
  }
}

async function createEvidenceRowWithUncertainResponseRetry(
  operations: WorkbookOperationExecutor,
  incidentId: string,
  request: EvidenceCreateRequest,
): Promise<WorkbookOperationOutcome<CreateViewRowResponse>> {
  try {
    return await operations.execute({
      operationID: "createViewRow",
      pathParameters: {
        incident_id: incidentId,
        view_schema_id: evidenceViewSchemaId,
      },
      request,
    });
  } catch {
    return operations.execute({
      operationID: "createViewRow",
      pathParameters: {
        incident_id: incidentId,
        view_schema_id: evidenceViewSchemaId,
      },
      request,
    });
  }
}

async function attachEvidenceToTimeline(
  options: TimelineEvidenceAdapterOptions,
  operations: WorkbookOperationExecutor,
  input: Parameters<TimelineEvidenceAttachmentPort["attachEvidence"]>[0],
): ReturnType<TimelineEvidenceAttachmentPort["attachEvidence"]> {
  let clientTxnId: string;
  try {
    clientTxnId = options.createClientTxnId();
  } catch {
    return {
      clientTxnId: null,
      outcome: terminal("A secure request identifier could not be created."),
    };
  }
  const command = timelineEvidenceCommand(
    options.incidentId,
    input.target,
    input.evidenceRecordId,
    clientTxnId,
  );
  if (command === null) {
    return {
      clientTxnId,
      outcome: {
        kind: "rejected",
        failure: {
          kind: "stale_target",
          message: "The selected Timeline row is stale.",
        },
      },
    };
  }
  input.onTimelineClientTxnId(clientTxnId);
  try {
    const outcome = await executeTimelineEvidenceCommand(operations, command);
    return {
      clientTxnId,
      outcome: normalizeTimelineEvidenceOutcome(
        outcome,
        input.evidenceRecordId,
        input.target.recordId,
      ),
    };
  } catch {
    return {
      clientTxnId,
      outcome: {
        kind: "rejected",
        failure: {
          kind: "retryable",
          message: "The Timeline evidence link could not be sent.",
        },
      },
    };
  }
}

function timelineEvidenceCommand(
  incidentId: string,
  target: Parameters<
    TimelineEvidenceAttachmentPort["attachEvidence"]
  >[0]["target"],
  evidenceRecordId: string,
  clientTxnId: string,
): TimelineEvidenceCommand | null {
  if (target.recordId === null) {
    return {
      incidentId,
      kind: "create",
      request: buildAttachedEvidenceCreateRequest(
        evidenceRecordId,
        clientTxnId,
      ),
    };
  }
  const request = buildAttachedEvidencePatchRequest(
    target,
    evidenceRecordId,
    clientTxnId,
  );
  return request === null
    ? null
    : {
        kind: "patch",
        recordId: target.recordId,
        request,
      };
}

function executeTimelineEvidenceCommand(
  operations: WorkbookOperationExecutor,
  command: TimelineEvidenceCommand,
): Promise<WorkbookOperationOutcome<CreateViewRowResponse>> {
  if (command.kind === "create") {
    return operations.execute({
      operationID: "createViewRow",
      pathParameters: {
        incident_id: command.incidentId,
        view_schema_id: timelineViewSchemaId,
      },
      request: command.request,
    });
  }
  return operations.execute({
    operationID: "patchRecord",
    pathParameters: { record_id: command.recordId },
    request: command.request,
  });
}

function normalizeTimelineEvidenceOutcome(
  outcome: WorkbookOperationOutcome<CreateViewRowResponse>,
  evidenceRecordId: string,
  expectedRecordId: string | null,
): WorkbookOperationOutcome<TimelineEvidenceAttachmentAccepted> {
  if (outcome.kind === "rejected") return outcome;
  const data = outcome.value.data;
  if (
    data.view_schema_id !== timelineViewSchemaId ||
    (expectedRecordId !== null && data.row.record_id !== expectedRecordId)
  ) {
    return invalidContract();
  }
  try {
    return {
      kind: "accepted",
      value: {
        evidenceRecordId,
        row: normalizeTimelineFullRow(
          data.row,
          "evidence attachment response row",
        ),
        viewSchemaId: data.view_schema_id,
      },
    };
  } catch {
    return invalidContract();
  }
}
