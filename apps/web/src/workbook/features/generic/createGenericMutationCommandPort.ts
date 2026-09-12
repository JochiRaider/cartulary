import type { WorkbookOperationExecutor } from "../../adapters/workbookOperationContract";
import {
  captureRecordPatch,
  createRecordPatchTransport,
} from "../../adapters/workbookRecordPatchTransport";
import {
  decodeCreateRecordLinkedNoteRequest,
  decodeCreateViewRowRequest,
} from "../../models/workbookRequestDecoders";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type {
  GenericMutationCommandPort,
  GenericMutationOutcome,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import { decisionViewId } from "../coordination/decisionSupersessionModel";
import type { DecisionRecordWriteBoundary } from "../coordination/decisionSupersessionOperation";
import { buildGenericCreateRequest } from "./genericCreateRequestBuilder";

function operationIdentityFailure<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "terminal",
      message: "A secure transaction ID could not be created.",
    },
  };
}

function invalidOperationPayload<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: { kind: "validation", message: "invalid_mutation_payload" },
  };
}

function createId(
  transactionIds: SecureTransactionIdPort,
  prefix: string,
): string | null {
  try {
    return transactionIds.create(prefix);
  } catch {
    return null;
  }
}

function normalizeGenericMutationOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id: string;
      readonly row: Extract<
        GenericMutationOutcome,
        { kind: "accepted" }
      >["value"]["row"];
      readonly view_schema_id: string;
    };
  }>,
): GenericMutationOutcome {
  return outcome.kind === "rejected"
    ? outcome
    : {
        kind: "accepted",
        value: {
          changeSetId: outcome.value.data.change_set_id,
          row: outcome.value.data.row,
          viewSchemaId: outcome.value.data.view_schema_id,
        },
      };
}

export function createGenericMutationCommandPort(options: {
  readonly decisionWrites?: DecisionRecordWriteBoundary | undefined;
  readonly incidentId: string;
  readonly operations: WorkbookOperationExecutor;
  readonly transactionIds: SecureTransactionIdPort;
}): GenericMutationCommandPort {
  return {
    canCreateRecord(input) {
      return (
        buildGenericCreateRequest(
          input.contract,
          input.draft,
          "validation-only",
        ) !== null
      );
    },
    createRecord(input) {
      const clientTxnId = createId(
        options.transactionIds,
        `generic-create-${input.contract.viewSchemaId}`,
      );
      if (clientTxnId === null) {
        return Promise.resolve(operationIdentityFailure());
      }
      const payload = buildGenericCreateRequest(
        input.contract,
        input.draft,
        clientTxnId,
      );
      const request = decodeCreateViewRowRequest(input.contract, payload);
      if (request === null) return Promise.resolve(invalidOperationPayload());
      if (input.linkedNoteSourceRecordId === "") {
        return options.operations
          .execute({
            operationID: "createViewRow",
            pathParameters: {
              incident_id: options.incidentId,
              view_schema_id: input.contract.viewSchemaId,
            },
            request,
          })
          .then(normalizeGenericMutationOutcome);
      }
      const linkedNoteRequest = decodeCreateRecordLinkedNoteRequest(request);
      return linkedNoteRequest === null
        ? Promise.resolve(invalidOperationPayload())
        : options.operations
            .execute({
              operationID: "createRecordLinkedNote",
              pathParameters: { record_id: input.linkedNoteSourceRecordId },
              request: linkedNoteRequest,
            })
            .then(normalizeGenericMutationOutcome);
    },
    async patchRecord(input) {
      const boundary =
        input.viewSchemaId === decisionViewId
          ? options.decisionWrites
          : undefined;
      const release = boundary ? boundary.begin([input.recordId]) : () => {};
      if (!release)
        return {
          kind: "rejected",
          failure: {
            kind: "stale_target",
            message:
              "Recover this Decision in Decision actions before editing.",
          },
        };
      try {
        const clientTxnId = createId(
          options.transactionIds,
          `${input.purpose}-${input.viewSchemaId}`,
        );
        if (clientTxnId === null) {
          return Promise.resolve(operationIdentityFailure());
        }
        const request = captureRecordPatch({
          recordId: input.recordId,
          baseRowVersion: input.baseRowVersion,
          changes: input.changes,
          clientTxnId,
          viewSchemaId: input.viewSchemaId,
        });
        if (request === null) return invalidOperationPayload();
        const outcome = await createRecordPatchTransport(
          options.operations,
        ).send(request, new AbortController().signal);
        const result: GenericMutationOutcome =
          outcome.kind === "acknowledged"
            ? { kind: "accepted", value: outcome.receipt }
            : outcome.kind === "rejected"
              ? outcome
              : {
                  kind: "rejected",
                  failure: {
                    kind: "retryable",
                    message: "The patch outcome could not be confirmed.",
                  },
                };
        if (result.kind === "accepted") boundary?.acceptRow(result.value.row);
        return result;
      } finally {
        release();
      }
    },
  };
}
