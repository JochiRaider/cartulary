import type { WorkbookOperationExecutor } from "../../adapters/workbookOperationContract";
import type { WorkbookProtocolCreateViewRowRequest } from "../../adapters/workbookProtocolTypes";
import { extractEmailFromPartyText } from "../../models/genericWorkbookModel";
import {
  buildPatchRecordRequest,
  decodeCreateRecordLinkedNoteRequest,
  decodeCreateViewRowRequest,
} from "../../models/workbookRequestDecoders";
import { partiesViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type {
  GenericMutationCommandPort,
  GenericMutationOutcome,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
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
    patchRecord(input) {
      const clientTxnId = createId(
        options.transactionIds,
        `${input.purpose}-${input.viewSchemaId}`,
      );
      if (clientTxnId === null) {
        return Promise.resolve(operationIdentityFailure());
      }
      const request = buildPatchRecordRequest({
        baseRowVersion: input.baseRowVersion,
        changes: input.changes,
        clientTxnId,
        viewSchemaId: input.viewSchemaId,
      });
      return request === null
        ? Promise.resolve(invalidOperationPayload())
        : options.operations
            .execute({
              operationID: "patchRecord",
              pathParameters: { record_id: input.recordId },
              request,
            })
            .then(normalizeGenericMutationOutcome);
    },
    createPartyFromText(input) {
      const clientTxnId = createId(
        options.transactionIds,
        `party-from-text-${input.originViewSchemaId}`,
      );
      if (clientTxnId === null) {
        return Promise.resolve(operationIdentityFailure());
      }
      const email = extractEmailFromPartyText(input.rawText);
      const request = {
        client_txn_id: clientTxnId,
        "party.display_name": input.rawText,
        "party.party_kind": "person",
        ...(email === null ? {} : { "party.primary_email": email }),
      } satisfies WorkbookProtocolCreateViewRowRequest;
      return options.operations
        .execute({
          operationID: "createViewRow",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: partiesViewSchemaId,
          },
          request,
        })
        .then(normalizeGenericMutationOutcome);
    },
  };
}
