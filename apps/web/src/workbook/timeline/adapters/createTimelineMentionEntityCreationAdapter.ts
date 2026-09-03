import type { CreateViewRowRequest } from "@cartulary/protocol-ts/http";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type {
  TimelineMentionEntityCreated,
  TimelineMentionEntityCreationPort,
} from "../ports/TimelineMentionPort";

type EntityCreateCommand = {
  readonly request: CreateViewRowRequest;
  readonly viewSchemaId: string;
};

export function createTimelineMentionEntityCreationAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): TimelineMentionEntityCreationPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async createEntity(input) {
      const command = entityCreateCommand(input);
      if (command === null) return invalidEntityCreate();
      try {
        const outcome = await operations.execute({
          operationID: "createViewRow",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: command.viewSchemaId,
          },
          request: command.request,
        });
        if (outcome.kind === "rejected") return outcome;
        return outcome.value.data.view_schema_id === command.viewSchemaId
          ? {
              kind: "accepted",
              value: { recordId: outcome.value.data.row.record_id },
            }
          : invalidContract();
      } catch {
        return retryable();
      }
    },
  };
}

function entityCreateCommand(
  input: Parameters<TimelineMentionEntityCreationPort["createEntity"]>[0],
): EntityCreateCommand | null {
  const rawText = input.rawText.trim();
  if (rawText === "") return null;
  if (input.entityType === "host") {
    return {
      request: {
        client_txn_id: input.clientTxnId,
        "host.display_name": rawText,
        ...(rawText.includes(".")
          ? { "host.fqdn": rawText }
          : { "host.hostname": rawText }),
      },
      viewSchemaId: hostsViewSchemaId,
    };
  }
  return {
    request: {
      client_txn_id: input.clientTxnId,
      "identity.display_name": rawText,
      ...(rawText.includes("@")
        ? { "identity.email": rawText, "identity.upn": rawText }
        : { "identity.sam_account_name": rawText }),
    },
    viewSchemaId: identitiesViewSchemaId,
  };
}

function invalidEntityCreate(): WorkbookOperationOutcome<TimelineMentionEntityCreated> {
  return {
    kind: "rejected",
    failure: {
      kind: "validation",
      message: "Cannot create an entity from this mention.",
    },
  };
}

function invalidContract(): WorkbookOperationOutcome<TimelineMentionEntityCreated> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "Mention action source record was invalid.",
    },
  };
}

function retryable(): WorkbookOperationOutcome<TimelineMentionEntityCreated> {
  return {
    kind: "rejected",
    failure: {
      kind: "retryable",
      message: "The mention operation could not be sent.",
    },
  };
}
