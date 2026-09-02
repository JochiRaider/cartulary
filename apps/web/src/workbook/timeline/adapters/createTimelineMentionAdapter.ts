import type {
  CreateViewRowRequest,
  ResolveEntityMentionRequest,
} from "@cartulary/protocol-ts/http";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import { buildMentionActionPayload } from "../../collaboration/workbookCollaborationMessages";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type {
  TimelineMentionEntityCreated,
  TimelineMentionPort,
} from "../ports/TimelineMentionPort";

function invalidContract<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "Mention action source record was invalid.",
    },
  };
}

function retryable<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "retryable",
      message: "The mention operation could not be sent.",
    },
  };
}

type OptionalString =
  | { readonly valid: true; readonly value: string | null }
  | { readonly valid: false };

function optionalString(value: unknown): OptionalString {
  if (value === undefined || value === null) {
    return { valid: true, value: null };
  }
  if (typeof value !== "string") return { valid: false };
  return {
    valid: true,
    value: value.trim() === "" ? null : value,
  };
}

function entityCreateRequest(input: {
  readonly clientTxnId: string;
  readonly entityType: "host" | "identity";
  readonly rawText: string;
}): {
  readonly request: CreateViewRowRequest;
  readonly viewSchemaId: string;
} | null {
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

export function createTimelineMentionAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): TimelineMentionPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async createEntity(input) {
      const create = entityCreateRequest(input);
      if (create === null) {
        return {
          kind: "rejected",
          failure: {
            kind: "validation",
            message: "Cannot create an entity from this mention.",
          },
        };
      }
      try {
        const outcome = await operations.execute({
          operationID: "createViewRow",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: create.viewSchemaId,
          },
          request: create.request,
        });
        if (outcome.kind === "rejected") return outcome;
        return outcome.value.data.view_schema_id === create.viewSchemaId
          ? {
              kind: "accepted",
              value: { recordId: outcome.value.data.row.record_id },
            }
          : invalidContract<TimelineMentionEntityCreated>();
      } catch {
        return retryable();
      }
    },
    async resolve(input) {
      const request: ResolveEntityMentionRequest | null =
        buildMentionActionPayload(
          {
            mentionRowVersion: input.baseMentionRowVersion,
          },
          input.action,
          input.clientTxnId,
          input.resolvedRecordId,
        );
      if (request === null) {
        return {
          kind: "rejected",
          failure: {
            kind: "validation",
            message: "Missing mention row version.",
          },
        };
      }
      try {
        const outcome = await operations.execute({
          operationID: "resolveEntityMention",
          pathParameters: { entity_mention_id: input.mentionId },
          request,
        });
        if (outcome.kind === "rejected") return outcome;
        const data = outcome.value.data;
        if (
          data.source_record.record_id !== input.expectedSourceRecordId ||
          data.entity_mention.entity_mention_id !== input.mentionId ||
          !Number.isSafeInteger(data.source_record.row_version) ||
          data.source_record.row_version < 1 ||
          !Number.isSafeInteger(data.entity_mention.row_version) ||
          data.entity_mention.row_version < 1
        ) {
          return invalidContract();
        }
        const entityType = data.entity_mention.entity_type;
        const rawText = optionalString(data.entity_mention.raw_text);
        const resolutionMethod = optionalString(
          data.entity_mention.resolution_method,
        );
        const sourceFieldKey = optionalString(
          data.entity_mention.source_field_key,
        );
        if (
          (entityType !== undefined &&
            entityType !== null &&
            entityType !== "host" &&
            entityType !== "identity") ||
          !rawText.valid ||
          !resolutionMethod.valid ||
          !sourceFieldKey.valid
        ) {
          return invalidContract();
        }
        return {
          kind: "accepted",
          value: {
            entityMention: {
              entityType:
                entityType === "host" || entityType === "identity"
                  ? entityType
                  : null,
              rawText: rawText.value,
              resolutionMethod: resolutionMethod.value,
              rowVersion: data.entity_mention.row_version,
              sourceFieldKey: sourceFieldKey.value,
            },
            sourceRecord: {
              recordId: data.source_record.record_id,
              rowVersion: data.source_record.row_version,
            },
          },
        };
      } catch {
        return retryable();
      }
    },
  };
}
