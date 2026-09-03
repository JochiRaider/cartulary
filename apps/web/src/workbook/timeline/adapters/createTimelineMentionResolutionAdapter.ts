import type { ResolveEntityMentionRequest } from "@cartulary/protocol-ts/http";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import { buildMentionActionPayload } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type { TimelineMentionResolutionPort } from "../ports/TimelineMentionPort";

type TimelineMentionResolutionAccepted = Extract<
  Awaited<ReturnType<TimelineMentionResolutionPort["resolve"]>>,
  { readonly kind: "accepted" }
>["value"];

type OptionalString =
  | { readonly valid: true; readonly value: string | null }
  | { readonly valid: false };

export function createTimelineMentionResolutionAdapter(options: {
  readonly apiBase: string | undefined;
}): TimelineMentionResolutionPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async resolve(input) {
      const request = mentionResolutionRequest(input);
      if (request === null) return missingMentionVersion();
      try {
        const outcome = await operations.execute({
          operationID: "resolveEntityMention",
          pathParameters: { entity_mention_id: input.mentionId },
          request,
        });
        if (outcome.kind === "rejected") return outcome;
        return decodeMentionResolution(
          outcome.value.data,
          input.expectedSourceRecordId,
          input.mentionId,
        );
      } catch {
        return retryable();
      }
    },
  };
}

function mentionResolutionRequest(
  input: Parameters<TimelineMentionResolutionPort["resolve"]>[0],
): ResolveEntityMentionRequest | null {
  return buildMentionActionPayload(
    { mentionRowVersion: input.baseMentionRowVersion },
    input.action,
    input.clientTxnId,
    input.resolvedRecordId,
  );
}

function decodeMentionResolution(
  data: {
    readonly entity_mention: {
      readonly entity_mention_id: string;
      readonly entity_type?: unknown;
      readonly raw_text?: unknown;
      readonly resolution_method?: unknown;
      readonly row_version: number;
      readonly source_field_key?: unknown;
    };
    readonly source_record: {
      readonly record_id: string;
      readonly row_version: number;
    };
  },
  expectedSourceRecordId: string,
  expectedMentionId: string,
): WorkbookOperationOutcome<TimelineMentionResolutionAccepted> {
  if (
    data.source_record.record_id !== expectedSourceRecordId ||
    data.entity_mention.entity_mention_id !== expectedMentionId ||
    !validVersion(data.source_record.row_version) ||
    !validVersion(data.entity_mention.row_version)
  ) {
    return invalidContract();
  }
  const entityType = optionalEntityType(data.entity_mention.entity_type);
  const rawText = optionalString(data.entity_mention.raw_text);
  const resolutionMethod = optionalString(
    data.entity_mention.resolution_method,
  );
  const sourceFieldKey = optionalString(data.entity_mention.source_field_key);
  if (
    !entityType.valid ||
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
        entityType: entityType.value,
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
}

function validVersion(value: number): boolean {
  return Number.isSafeInteger(value) && value > 0;
}

function optionalEntityType(
  value: unknown,
):
  | { readonly valid: true; readonly value: "host" | "identity" | null }
  | { readonly valid: false } {
  if (value === undefined || value === null) {
    return { valid: true, value: null };
  }
  return value === "host" || value === "identity"
    ? { valid: true, value }
    : { valid: false };
}

function optionalString(value: unknown): OptionalString {
  if (value === undefined || value === null) {
    return { valid: true, value: null };
  }
  if (typeof value !== "string") return { valid: false };
  return { valid: true, value: value.trim() === "" ? null : value };
}

function missingMentionVersion(): WorkbookOperationOutcome<TimelineMentionResolutionAccepted> {
  return {
    kind: "rejected",
    failure: { kind: "validation", message: "Missing mention row version." },
  };
}

function invalidContract(): WorkbookOperationOutcome<TimelineMentionResolutionAccepted> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "Mention action source record was invalid.",
    },
  };
}

function retryable(): WorkbookOperationOutcome<TimelineMentionResolutionAccepted> {
  return {
    kind: "rejected",
    failure: {
      kind: "retryable",
      message: "The mention operation could not be sent.",
    },
  };
}
