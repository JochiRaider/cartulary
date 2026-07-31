import type { ViewContract } from "@cartulary/view-contracts";
import { apiPath } from "../../services/browserApi";
import { fetchWorkbookJSON } from "../../services/workbookApi";
import { createAndAttachEvidenceBlob } from "../../services/workbookEvidence";
import { buildAssessmentCreatePayload } from "../models/assessmentWorkbookModel";
import {
  buildGenericCreatePayload,
  extractEmailFromPartyText,
} from "../models/genericWorkbookModel";
import {
  assessmentsViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  buildAttachedEvidencePatchPayload,
  type WorkbookRow,
} from "../timeline/models/workbookTimelineModel";
import type { SecureTransactionIdPort } from "./secureTransactionId";
import type {
  WorkbookMutationCommandPorts,
  WorkbookMutationCommandResult,
} from "./workbookMutationCommandPorts";

type CommandContext = {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly transactionIds: SecureTransactionIdPort;
};

function identityFailure(): WorkbookMutationCommandResult {
  return {
    ok: false,
    status: 0,
    payload: {
      error: {
        message: "A secure transaction ID could not be created.",
      },
    },
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

function executeJSON<T>(
  apiBase: string | undefined,
  path: string,
  method: string,
  body: unknown,
): Promise<WorkbookMutationCommandResult<T>> {
  return fetchWorkbookJSON<T>(apiPath(apiBase, path), {
    method,
    body: JSON.stringify(body),
  });
}

function createRecordCommand(
  context: CommandContext,
  contract: ViewContract,
  draft: Readonly<Record<string, string>>,
  prefix: string,
  path: string,
): Promise<WorkbookMutationCommandResult> {
  const clientTxnId = createId(context.transactionIds, prefix);
  if (clientTxnId === null) return Promise.resolve(identityFailure());
  const payload = buildGenericCreatePayload(
    contract,
    { ...draft },
    clientTxnId,
  );
  if (payload === null) {
    return Promise.resolve({
      ok: false,
      status: 422,
      payload: { error: { message: "invalid_mutation_payload" } },
    });
  }
  return executeJSON(context.apiBase, path, "POST", payload);
}

function patchRecordCommand(
  context: CommandContext,
  input: {
    readonly baseRowVersion: number;
    readonly changes: readonly Record<string, unknown>[];
    readonly purpose: string;
    readonly recordId: string;
    readonly viewSchemaId: string;
  },
): Promise<WorkbookMutationCommandResult> {
  const clientTxnId = createId(
    context.transactionIds,
    `${input.purpose}-${input.viewSchemaId}`,
  );
  if (clientTxnId === null) return Promise.resolve(identityFailure());
  return executeJSON(
    context.apiBase,
    `/api/v1/records/${input.recordId}`,
    "PATCH",
    {
      view_schema_id: input.viewSchemaId,
      base_row_version: input.baseRowVersion,
      client_txn_id: clientTxnId,
      changes: input.changes,
    },
  );
}

function relatedRecordPath(incidentId: string, contract: ViewContract): string {
  return `/api/v1/incidents/${incidentId}/views/${contract.viewSchemaId}/rows`;
}

function linkCreatedEvidenceCommand(
  context: CommandContext,
  sourceRow: WorkbookRow,
  createdRecordId: string,
): Promise<WorkbookMutationCommandResult> {
  const clientTxnId = createId(
    context.transactionIds,
    "timeline-link-created-evidence",
  );
  if (clientTxnId === null) return Promise.resolve(identityFailure());
  const payload = buildAttachedEvidencePatchPayload(
    sourceRow,
    createdRecordId,
    clientTxnId,
  );
  if (payload === null || sourceRow.recordId === null) {
    return Promise.resolve({
      ok: false,
      status: 409,
      payload: {
        error: {
          message: "Created evidence, but the selected row version is stale.",
        },
      },
    });
  }
  return executeJSON(
    context.apiBase,
    `/api/v1/records/${sourceRow.recordId}`,
    "PATCH",
    payload,
  );
}

export function createWorkbookMutationCommandPorts(
  context: CommandContext,
): WorkbookMutationCommandPorts {
  return {
    timeline: {
      createLogicalActionId() {
        const clientTxnId = createId(context.transactionIds, "timeline-client");
        if (clientTxnId === null) {
          throw new Error("A secure request identifier could not be created.");
        }
        return clientTxnId;
      },
      createConflictRecoveryId() {
        const clientTxnId = createId(context.transactionIds, "timeline-client");
        if (clientTxnId === null) {
          throw new Error("A secure request identifier could not be created.");
        }
        return clientTxnId;
      },
      assignTag(input) {
        const clientTxnId = createId(context.transactionIds, "timeline-client");
        if (clientTxnId === null) return Promise.resolve(identityFailure());
        return executeJSON(
          context.apiBase,
          `/api/v1/incidents/${context.incidentId}/views/${timelineViewSchemaId}/bulk-mutations`,
          "POST",
          {
            view_schema_id: timelineViewSchemaId,
            client_txn_id: clientTxnId,
            kind: "multi_row_tag_assignment_v1",
            tag_name: input.tagName,
            targets: input.targets.map((target) => ({
              record_id: target.recordId,
              base_row_version: target.baseRowVersion,
            })),
          },
        );
      },
      async fillDown(input) {
        const clientTxnId = createId(context.transactionIds, "timeline-client");
        if (clientTxnId === null) {
          return { ...identityFailure(), clientTxnId: null };
        }
        input.onClientTxnId(clientTxnId);
        const result = await executeJSON(
          context.apiBase,
          `/api/v1/incidents/${context.incidentId}/views/${timelineViewSchemaId}/bulk-mutations`,
          "POST",
          {
            view_schema_id: timelineViewSchemaId,
            client_txn_id: clientTxnId,
            kind: "fill_down_v1",
            field_key: input.fieldKey,
            value: input.value,
            targets: input.targets.map((target) => ({
              record_id: target.recordId,
              base_row_version: target.baseRowVersion,
            })),
          },
        );
        return { ...result, clientTxnId };
      },
      createRelatedRecord(input) {
        return createRecordCommand(
          context,
          input.contract,
          input.draft,
          `timeline-create-related-${input.featureGroupKey}`,
          relatedRecordPath(context.incidentId, input.contract),
        );
      },
      linkCreatedEvidence(input) {
        return linkCreatedEvidenceCommand(
          context,
          input.sourceRow,
          input.createdRecordId,
        );
      },
    },
    generic: {
      canCreateRecord(input) {
        return (
          buildGenericCreatePayload(
            input.contract,
            { ...input.draft },
            "validation-only",
          ) !== null
        );
      },
      createRecord(input) {
        const path =
          input.linkedNoteSourceRecordId === ""
            ? relatedRecordPath(context.incidentId, input.contract)
            : `/api/v1/records/${input.linkedNoteSourceRecordId}/linked-notes`;
        return createRecordCommand(
          context,
          input.contract,
          input.draft,
          `generic-create-${input.contract.viewSchemaId}`,
          path,
        );
      },
      patchRecord(input) {
        return patchRecordCommand(context, input);
      },
      createPartyFromText(input) {
        const clientTxnId = createId(
          context.transactionIds,
          `party-from-text-${input.originViewSchemaId}`,
        );
        if (clientTxnId === null) return Promise.resolve(identityFailure());
        const payload: Record<string, unknown> = {
          client_txn_id: clientTxnId,
          "party.display_name": input.rawText,
          "party.party_kind": "person",
        };
        const email = extractEmailFromPartyText(input.rawText);
        if (email !== null) payload["party.primary_email"] = email;
        return executeJSON(
          context.apiBase,
          `/api/v1/incidents/${context.incidentId}/views/${partiesViewSchemaId}/rows`,
          "POST",
          payload,
        );
      },
    },
    entity: {
      canCreateRecord(input) {
        return (
          buildGenericCreatePayload(
            input.contract,
            { ...input.draft },
            "validation-only",
          ) !== null
        );
      },
      createRecord(input) {
        return createRecordCommand(
          context,
          input.contract,
          input.draft,
          `entity-create-${input.contract.viewSchemaId}`,
          relatedRecordPath(context.incidentId, input.contract),
        );
      },
      patchRecord(input) {
        return patchRecordCommand(context, input);
      },
      pasteCreate(input) {
        const clientTxnId = createId(
          context.transactionIds,
          `${input.viewSchemaId}-paste`,
        );
        if (clientTxnId === null) return Promise.resolve(identityFailure());
        return executeJSON(
          context.apiBase,
          `/api/v1/incidents/${context.incidentId}/views/${input.viewSchemaId}/clipboard-paste`,
          "POST",
          {
            view_schema_id: input.viewSchemaId,
            client_txn_id: clientTxnId,
            clipboard_text: input.clipboardText,
            format: input.format,
            start_field_key: input.startFieldKey,
            columns: input.columns,
            targets: Array.from({ length: input.targetCount }, () => ({
              kind: "create",
            })),
          },
        );
      },
      merge(input) {
        const clientTxnId = createId(context.transactionIds, "merge");
        if (clientTxnId === null) return Promise.resolve(identityFailure());
        return executeJSON(
          context.apiBase,
          `/api/v1/records/${input.survivorRecordId}/merge`,
          "POST",
          {
            loser_record_id: input.loserRecordId,
            survivor_base_row_version: input.survivorBaseRowVersion,
            loser_base_row_version: input.loserBaseRowVersion,
            client_txn_id: clientTxnId,
            reason: input.reason,
          },
        );
      },
    },
    assessment: {
      create(input) {
        const clientTxnId = createId(context.transactionIds, "assessment");
        if (clientTxnId === null) return Promise.resolve(identityFailure());
        const payload = buildAssessmentCreatePayload(input.draft, clientTxnId);
        if (payload === null) {
          return Promise.resolve({
            ok: false,
            status: 422,
            payload: { error: { message: "invalid_mutation_payload" } },
          });
        }
        return executeJSON(
          context.apiBase,
          `/api/v1/incidents/${context.incidentId}/views/${assessmentsViewSchemaId}/rows`,
          "POST",
          payload,
        );
      },
    },
    evidence: {
      async attach(input) {
        const createClientTxnId = createId(
          context.transactionIds,
          "evidence-blob",
        );
        const attachClientTxnId = createId(
          context.transactionIds,
          "evidence-attach",
        );
        if (createClientTxnId === null || attachClientTxnId === null) {
          throw new Error("A secure transaction ID could not be created.");
        }
        await createAndAttachEvidenceBlob({
          apiBase: context.apiBase,
          attachClientTxnId: () => attachClientTxnId,
          baseRowVersion: input.baseRowVersion,
          createClientTxnId: () => createClientTxnId,
          evidenceRecordId: input.evidenceRecordId,
          file: input.file,
          incidentId: context.incidentId,
        });
      },
    },
    coordination: {
      updateTaskLifecycle(input) {
        const clientTxnId = createId(context.transactionIds, "task-lifecycle");
        if (clientTxnId === null) return Promise.resolve(identityFailure());
        const changes: Record<string, unknown>[] = [
          { field_key: "task.status", value: input.status },
        ];
        if (input.status === "blocked") {
          changes.push({
            field_key: "task.blocked_reason",
            value: input.blockedReason,
          });
        }
        return executeJSON(
          context.apiBase,
          `/api/v1/records/${input.recordId}`,
          "PATCH",
          {
            view_schema_id: taskRequestsViewSchemaId,
            base_row_version: input.baseRowVersion,
            client_txn_id: clientTxnId,
            changes,
          },
        );
      },
      supersedeDecision(input) {
        const clientTxnId = createId(
          context.transactionIds,
          "decision-supersede",
        );
        if (clientTxnId === null) return Promise.resolve(identityFailure());
        return executeJSON(
          context.apiBase,
          `/api/v1/records/${input.targetRecordId}/supersede`,
          "POST",
          {
            base_row_version: input.baseRowVersion,
            client_txn_id: clientTxnId,
            replacement_record_id: input.replacementRecordId,
            reason: input.reason,
          },
        );
      },
    },
  };
}
