import type { WorkbookOperationExecutor } from "../../adapters/workbookOperationContract";
import {
  captureRecordPatch,
  createRecordPatchTransport,
} from "../../adapters/workbookRecordPatchTransport";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type {
  GenericMutationCommandPort,
  GenericMutationOutcome,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import { decisionViewId } from "../coordination/decisionSupersessionModel";
import type { DecisionRecordWriteBoundary } from "../coordination/decisionSupersessionOperation";

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

export function createGenericMutationCommandPort(options: {
  readonly decisionWrites?: DecisionRecordWriteBoundary | undefined;
  readonly incidentId: string;
  readonly operations: WorkbookOperationExecutor;
  readonly transactionIds: SecureTransactionIdPort;
}): GenericMutationCommandPort {
  return {
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
