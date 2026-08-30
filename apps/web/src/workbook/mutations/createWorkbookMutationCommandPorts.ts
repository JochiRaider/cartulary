import type {
  ApplyWorkbookBulkMutationRequest,
  AttachBlobToEvidenceRecordRequest,
  CreateObjectBlobSlotRequest,
  CreateRecordLinkedNoteRequest,
  CreateViewRowRequest,
  IssueEvidenceDownloadHandleRequest,
  IssueEvidencePreviewHandleRequest,
  MergeEntityRecordRequest,
  PasteWorkbookClipboardRequest,
  PatchRecordRequest,
  RollbackRecordRequest,
  SupersedeRecordRequest,
  SupersedeRecordResponse,
} from "@cartulary/protocol-ts/http";
import {
  resolvePublicEvidenceHandleHref,
  uploadEvidenceObjectBlobTarget,
} from "../../services/workbookEvidence";
import {
  createWorkbookOperationExecutor,
  type WorkbookOperationExecutor,
} from "../adapters/workbookOperationExecutor";
import { buildAssessmentCreatePayload } from "../models/assessmentWorkbookModel";
import {
  buildGenericCreatePayload,
  extractEmailFromPartyText,
} from "../models/genericWorkbookModel";
import {
  assessmentsViewSchemaId,
  evidenceViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { normalizeRecordHistoryData } from "../timeline/models/timelineHistoryModel";
import {
  buildAttachedEvidencePatchPayload,
  normalizeTimelineFullRow,
} from "../timeline/models/workbookTimelineModel";
import { createIndicatorWorkflowPort } from "./createIndicatorWorkflowPort";
import type { SecureTransactionIdPort } from "./secureTransactionId";
import type {
  AssessmentCreateOutcome,
  DecisionSupersedeOutcome,
  EntityCreateOutcome,
  EntityMergeOutcome,
  EntityPasteOutcome,
  EntityPatchOutcome,
  GenericMutationOutcome,
  GenericViewMutationAccepted,
  RecordLifecycleAccepted,
  TaskLifecycleOutcome,
  TimelineBulkMutationOutcome,
  TimelineRelatedEvidenceLinked,
  TimelineRelatedRecordCreated,
  WorkbookMutationCommandPorts,
} from "./workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "./workbookOperationOutcome";

type CommandContext = {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly transactionIds: SecureTransactionIdPort;
};

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

function invalidOperationContract<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "The server returned an inconsistent Workbook operation result.",
    },
  };
}

function retryableOperationFailure<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "retryable",
      message: "The Workbook operation could not be sent.",
    },
  };
}

async function executeTimelineBulkMutation(options: {
  readonly incidentId: string;
  readonly input: ApplyWorkbookBulkMutationRequest;
  readonly operations: WorkbookOperationExecutor;
}): Promise<TimelineBulkMutationOutcome> {
  if (
    options.input.client_txn_id === "" ||
    options.input.targets.length === 0
  ) {
    return options.input.client_txn_id === ""
      ? operationIdentityFailure()
      : invalidOperationPayload();
  }
  try {
    const outcome = await options.operations.execute({
      operationID: "applyWorkbookBulkMutation",
      pathParameters: {
        incident_id: options.incidentId,
        view_schema_id: timelineViewSchemaId,
      },
      request: options.input,
    });
    if (outcome.kind === "rejected") return outcome;
    const data = outcome.value.data;
    if (data.view_schema_id !== timelineViewSchemaId) {
      return invalidOperationContract();
    }
    try {
      for (const row of data.rows) {
        normalizeTimelineFullRow(row, "bulk mutation response row");
      }
    } catch {
      return invalidOperationContract();
    }
    return {
      kind: "accepted",
      value: {
        affectedRowCount: data.rows.length,
        changeSetId: data.change_set_id ?? null,
        conflictCount: data.conflicts?.length ?? 0,
      },
    };
  } catch {
    return retryableOperationFailure();
  }
}

function normalizeTimelineRelatedCreateOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id: string;
      readonly row: GenericViewMutationAccepted["row"];
      readonly view_schema_id: string;
    };
  }>,
  expectedViewSchemaId: string,
): WorkbookOperationOutcome<TimelineRelatedRecordCreated> {
  if (outcome.kind === "rejected") return outcome;
  const data = outcome.value.data;
  if (data.view_schema_id !== expectedViewSchemaId) {
    return invalidOperationContract();
  }
  return {
    kind: "accepted",
    value: {
      changeSetId: data.change_set_id,
      recordId: data.row.record_id,
      viewSchemaId: data.view_schema_id,
    },
  };
}

function normalizeTimelineRelatedLinkOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id: string;
      readonly row: GenericViewMutationAccepted["row"];
      readonly view_schema_id: string;
    };
  }>,
  expectedRecordId: string,
): WorkbookOperationOutcome<TimelineRelatedEvidenceLinked> {
  if (outcome.kind === "rejected") return outcome;
  const data = outcome.value.data;
  if (
    data.view_schema_id !== timelineViewSchemaId ||
    data.row.record_id !== expectedRecordId
  ) {
    return invalidOperationContract();
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
    return invalidOperationContract();
  }
}

function normalizeEntityCreateOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id: string;
      readonly row: GenericViewMutationAccepted["row"];
      readonly view_schema_id: string;
    };
  }>,
  expectedViewSchemaId: string,
): EntityCreateOutcome {
  if (outcome.kind === "rejected") return outcome;
  if (outcome.value.data.view_schema_id !== expectedViewSchemaId) {
    return invalidOperationContract();
  }
  return {
    kind: "accepted",
    value: {
      changeSetId: outcome.value.data.change_set_id,
      row: outcome.value.data.row,
      viewSchemaId: outcome.value.data.view_schema_id,
    },
  };
}

function normalizeEntityPatchOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id: string;
      readonly row: GenericViewMutationAccepted["row"];
      readonly view_schema_id: string;
    };
  }>,
  expectedRecordId: string,
  expectedViewSchemaId: string,
): EntityPatchOutcome {
  if (outcome.kind === "rejected") return outcome;
  if (
    outcome.value.data.row.record_id !== expectedRecordId ||
    outcome.value.data.view_schema_id !== expectedViewSchemaId
  ) {
    return invalidOperationContract();
  }
  return {
    kind: "accepted",
    value: {
      changeSetId: outcome.value.data.change_set_id,
      row: outcome.value.data.row,
      viewSchemaId: outcome.value.data.view_schema_id,
    },
  };
}

function normalizeEntityPasteOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id?: string | undefined;
      readonly rows: readonly GenericViewMutationAccepted["row"][];
      readonly view_schema_id: string;
    };
  }>,
  expectedViewSchemaId: string,
): EntityPasteOutcome {
  if (outcome.kind === "rejected") return outcome;
  if (outcome.value.data.view_schema_id !== expectedViewSchemaId) {
    return invalidOperationContract();
  }
  return {
    kind: "accepted",
    value: {
      changeSetId: outcome.value.data.change_set_id ?? null,
      rows: outcome.value.data.rows,
      viewSchemaId: outcome.value.data.view_schema_id,
    },
  };
}

function normalizeEntityMergeOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id: string;
      readonly loser_record_id: string;
      readonly loser_row_version: number;
      readonly merged_into_record_id: string;
      readonly record_type: "host" | "identity";
      readonly survivor_record_id: string;
      readonly survivor_row_version: number;
    };
  }>,
  expectedLoserRecordId: string,
  expectedSurvivorRecordId: string,
): EntityMergeOutcome {
  if (outcome.kind === "rejected") return outcome;
  const data = outcome.value.data;
  if (
    data.loser_record_id !== expectedLoserRecordId ||
    data.survivor_record_id !== expectedSurvivorRecordId ||
    data.merged_into_record_id !== expectedSurvivorRecordId
  ) {
    return invalidOperationContract();
  }
  return {
    kind: "accepted",
    value: {
      changeSetId: data.change_set_id,
      loserRecordId: data.loser_record_id,
      loserRowVersion: data.loser_row_version,
      mergedIntoRecordId: data.merged_into_record_id,
      recordType: data.record_type,
      survivorRecordId: data.survivor_record_id,
      survivorRowVersion: data.survivor_row_version,
    },
  };
}

function normalizeAssessmentCreateOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id: string;
      readonly row: GenericViewMutationAccepted["row"];
      readonly view_schema_id: string;
    };
  }>,
): AssessmentCreateOutcome {
  if (outcome.kind === "rejected") return outcome;
  if (outcome.value.data.view_schema_id !== assessmentsViewSchemaId) {
    return invalidOperationContract();
  }
  return {
    kind: "accepted",
    value: {
      changeSetId: outcome.value.data.change_set_id,
      row: outcome.value.data.row,
      viewSchemaId: outcome.value.data.view_schema_id,
    },
  };
}

function normalizeTaskLifecycleOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id: string;
      readonly row: GenericViewMutationAccepted["row"];
      readonly view_schema_id: string;
    };
  }>,
  expectedRecordId: string,
  status: Parameters<
    WorkbookMutationCommandPorts["coordination"]["updateTaskLifecycle"]
  >[0]["status"],
): TaskLifecycleOutcome {
  if (outcome.kind === "rejected") return outcome;
  const data = outcome.value.data;
  if (
    data.view_schema_id !== taskRequestsViewSchemaId ||
    data.row.record_id !== expectedRecordId
  ) {
    return invalidOperationContract();
  }
  return {
    kind: "accepted",
    value: {
      changeSetId: data.change_set_id,
      row: data.row,
      status,
      viewSchemaId: data.view_schema_id,
    },
  };
}

function normalizeDecisionSupersedeOutcome(
  outcome: WorkbookOperationOutcome<SupersedeRecordResponse>,
  expectedTargetRecordId: string,
  expectedReplacementRecordId: string,
): DecisionSupersedeOutcome {
  if (outcome.kind === "rejected") return outcome;
  const data = outcome.value.data;
  if (
    !("target_record_id" in data) ||
    data.view_schema_id !== "cartulary.view.decisions.v1" ||
    data.target_record_id !== expectedTargetRecordId ||
    data.superseding_record_id !== expectedReplacementRecordId
  ) {
    return invalidOperationContract();
  }
  return {
    kind: "accepted",
    value: {
      changeSetId: data.change_set_id,
      replacementRecordId: data.superseding_record_id,
      replacementRowVersion: data.superseding_row_version,
      targetRecordId: data.target_record_id,
      targetRowVersion: data.target_row_version,
      targetStatus: data.target_status,
      viewSchemaId: data.view_schema_id,
    },
  };
}

function normalizeGenericMutationOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly change_set_id: string;
      readonly row: GenericViewMutationAccepted["row"];
      readonly view_schema_id: string;
    };
  }>,
): GenericMutationOutcome {
  if (outcome.kind === "rejected") return outcome;
  return {
    kind: "accepted",
    value: {
      changeSetId: outcome.value.data.change_set_id,
      row: outcome.value.data.row,
      viewSchemaId: outcome.value.data.view_schema_id,
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

export function createWorkbookMutationCommandPorts(
  context: CommandContext,
): WorkbookMutationCommandPorts {
  const operations = createWorkbookOperationExecutor({
    apiBase: context.apiBase,
  });
  return {
    records: {
      async execute(input) {
        const clientTxnId = createId(
          context.transactionIds,
          `record-${input.action}`,
        );
        if (clientTxnId === null) return operationIdentityFailure();
        try {
          const outcome = await operations.execute({
            operationID:
              input.action === "delete" ? "deleteRecord" : "restoreRecord",
            pathParameters: { record_id: input.recordId },
            request: {
              base_row_version: input.baseRowVersion,
              client_txn_id: clientTxnId,
              reason: input.reason,
            },
          });
          return normalizeRecordLifecycleOutcome(outcome, input.recordId);
        } catch {
          return retryableOperationFailure();
        }
      },
      async loadHistory(input) {
        try {
          const outcome = await operations.execute({
            operationID: "getRecordHistory",
            pathParameters: { record_id: input.recordId },
          });
          if (outcome.kind === "rejected") return outcome;
          const data = normalizeRecordHistoryData(outcome.value.data);
          return data !== null && data.record_id === input.recordId
            ? { kind: "accepted", value: data }
            : invalidOperationContract();
        } catch {
          return retryableOperationFailure();
        }
      },
      async rollback(input) {
        const clientTxnId = createId(context.transactionIds, "record-rollback");
        if (clientTxnId === null) return operationIdentityFailure();
        try {
          const outcome = await operations.execute({
            operationID: "rollbackRecord",
            pathParameters: { record_id: input.recordId },
            request: {
              base_row_version: input.baseRowVersion,
              client_txn_id: clientTxnId,
              reason: input.reason,
              target: input.target,
            } as unknown as RollbackRecordRequest,
          });
          return normalizeRecordLifecycleOutcome(outcome, input.recordId);
        } catch {
          return retryableOperationFailure();
        }
      },
    },
    indicators: createIndicatorWorkflowPort({
      createMutationID: (prefix) => createId(context.transactionIds, prefix),
      operations,
    }),
    timeline: {
      identity: {
        createLogicalActionId() {
          const clientTxnId = createId(
            context.transactionIds,
            "timeline-client",
          );
          if (clientTxnId === null) {
            throw new Error(
              "A secure request identifier could not be created.",
            );
          }
          return clientTxnId;
        },
        createConflictRecoveryId() {
          const clientTxnId = createId(
            context.transactionIds,
            "timeline-client",
          );
          if (clientTxnId === null) {
            throw new Error(
              "A secure request identifier could not be created.",
            );
          }
          return clientTxnId;
        },
      },
      bulk: {
        async assignTag(input) {
          const clientTxnId = createId(
            context.transactionIds,
            "timeline-client",
          );
          if (clientTxnId === null) return operationIdentityFailure();
          return executeTimelineBulkMutation({
            input: {
              client_txn_id: clientTxnId,
              kind: "multi_row_tag_assignment_v1",
              tag_name: input.tagName,
              targets: input.targets.map((target) => ({
                record_id: target.recordId,
                base_row_version: target.baseRowVersion,
              })) as ApplyWorkbookBulkMutationRequest["targets"],
              view_schema_id:
                timelineViewSchemaId as ApplyWorkbookBulkMutationRequest["view_schema_id"],
            },
            incidentId: context.incidentId,
            operations,
          });
        },
        async fillDown(input) {
          const clientTxnId = createId(
            context.transactionIds,
            "timeline-client",
          );
          if (clientTxnId === null) {
            return {
              clientTxnId: null,
              outcome: operationIdentityFailure(),
            };
          }
          input.onClientTxnId(clientTxnId);
          const outcome = await executeTimelineBulkMutation({
            input: {
              client_txn_id: clientTxnId,
              field_key: input.fieldKey,
              kind: "fill_down_v1",
              targets: input.targets.map((target) => ({
                record_id: target.recordId,
                base_row_version: target.baseRowVersion,
              })) as ApplyWorkbookBulkMutationRequest["targets"],
              value: input.value,
              view_schema_id:
                timelineViewSchemaId as ApplyWorkbookBulkMutationRequest["view_schema_id"],
            },
            incidentId: context.incidentId,
            operations,
          });
          return { clientTxnId, outcome };
        },
      },
      related: {
        async createRelatedRecord(input) {
          const clientTxnId = createId(
            context.transactionIds,
            `timeline-create-related-${input.featureGroupKey}`,
          );
          if (clientTxnId === null) return operationIdentityFailure();
          const request = buildGenericCreatePayload(
            input.contract,
            { ...input.draft },
            clientTxnId,
          );
          if (request === null) return invalidOperationPayload();
          try {
            const outcome = await operations.execute({
              operationID: "createViewRow",
              pathParameters: {
                incident_id: context.incidentId,
                view_schema_id: input.contract.viewSchemaId,
              },
              request: request as CreateViewRowRequest,
            });
            return normalizeTimelineRelatedCreateOutcome(
              outcome,
              input.contract.viewSchemaId,
            );
          } catch {
            return retryableOperationFailure();
          }
        },
        async linkCreatedEvidence(input) {
          const clientTxnId = createId(
            context.transactionIds,
            "timeline-link-created-evidence",
          );
          if (clientTxnId === null) return operationIdentityFailure();
          const request = buildAttachedEvidencePatchPayload(
            input.sourceRow,
            input.createdRecordId,
            clientTxnId,
          );
          if (request === null || input.sourceRow.recordId === null) {
            return {
              kind: "rejected",
              failure: {
                kind: "stale_target",
                message:
                  "Created evidence, but the selected row version is stale.",
              },
            };
          }
          try {
            const outcome = await operations.execute({
              operationID: "patchRecord",
              pathParameters: { record_id: input.sourceRow.recordId },
              request: request as PatchRecordRequest,
            });
            return normalizeTimelineRelatedLinkOutcome(
              outcome,
              input.sourceRow.recordId,
            );
          } catch {
            return retryableOperationFailure();
          }
        },
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
        const clientTxnId = createId(
          context.transactionIds,
          `generic-create-${input.contract.viewSchemaId}`,
        );
        if (clientTxnId === null)
          return Promise.resolve(operationIdentityFailure());
        const payload = buildGenericCreatePayload(
          input.contract,
          { ...input.draft },
          clientTxnId,
        );
        if (payload === null) return Promise.resolve(invalidOperationPayload());
        return input.linkedNoteSourceRecordId === ""
          ? operations
              .execute({
                operationID: "createViewRow",
                pathParameters: {
                  incident_id: context.incidentId,
                  view_schema_id: input.contract.viewSchemaId,
                },
                request: payload as CreateViewRowRequest,
              })
              .then(normalizeGenericMutationOutcome)
          : operations
              .execute({
                operationID: "createRecordLinkedNote",
                pathParameters: {
                  record_id: input.linkedNoteSourceRecordId,
                },
                request: payload as CreateRecordLinkedNoteRequest,
              })
              .then(normalizeGenericMutationOutcome);
      },
      patchRecord(input) {
        const clientTxnId = createId(
          context.transactionIds,
          `${input.purpose}-${input.viewSchemaId}`,
        );
        if (clientTxnId === null)
          return Promise.resolve(operationIdentityFailure());
        return operations
          .execute({
            operationID: "patchRecord",
            pathParameters: { record_id: input.recordId },
            request: {
              view_schema_id: input.viewSchemaId,
              base_row_version: input.baseRowVersion,
              client_txn_id: clientTxnId,
              changes: input.changes,
            } as PatchRecordRequest,
          })
          .then(normalizeGenericMutationOutcome);
      },
      createPartyFromText(input) {
        const clientTxnId = createId(
          context.transactionIds,
          `party-from-text-${input.originViewSchemaId}`,
        );
        if (clientTxnId === null)
          return Promise.resolve(operationIdentityFailure());
        const payload: Record<string, unknown> & {
          readonly client_txn_id: string;
        } = {
          client_txn_id: clientTxnId,
          "party.display_name": input.rawText,
          "party.party_kind": "person",
        };
        const email = extractEmailFromPartyText(input.rawText);
        if (email !== null) payload["party.primary_email"] = email;
        return operations
          .execute({
            operationID: "createViewRow",
            pathParameters: {
              incident_id: context.incidentId,
              view_schema_id: partiesViewSchemaId,
            },
            request: payload as CreateViewRowRequest,
          })
          .then(normalizeGenericMutationOutcome);
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
        const clientTxnId = createId(
          context.transactionIds,
          `entity-create-${input.contract.viewSchemaId}`,
        );
        if (clientTxnId === null)
          return Promise.resolve(operationIdentityFailure());
        const payload = buildGenericCreatePayload(
          input.contract,
          { ...input.draft },
          clientTxnId,
        );
        if (payload === null) return Promise.resolve(invalidOperationPayload());
        return operations
          .execute({
            operationID: "createViewRow",
            pathParameters: {
              incident_id: context.incidentId,
              view_schema_id: input.contract.viewSchemaId,
            },
            request: payload as CreateViewRowRequest,
          })
          .then((outcome) =>
            normalizeEntityCreateOutcome(outcome, input.contract.viewSchemaId),
          );
      },
      patchRecord(input) {
        const clientTxnId = createId(
          context.transactionIds,
          `${input.purpose}-${input.viewSchemaId}`,
        );
        if (clientTxnId === null)
          return Promise.resolve(operationIdentityFailure());
        return operations
          .execute({
            operationID: "patchRecord",
            pathParameters: { record_id: input.recordId },
            request: {
              view_schema_id: input.viewSchemaId,
              base_row_version: input.baseRowVersion,
              client_txn_id: clientTxnId,
              changes: input.changes,
            } as PatchRecordRequest,
          })
          .then((outcome) =>
            normalizeEntityPatchOutcome(
              outcome,
              input.recordId,
              input.viewSchemaId,
            ),
          );
      },
      pasteCreate(input) {
        const clientTxnId = createId(
          context.transactionIds,
          `${input.viewSchemaId}-paste`,
        );
        if (clientTxnId === null)
          return Promise.resolve(operationIdentityFailure());
        const [firstColumn, ...remainingColumns] = input.columns;
        const format =
          input.format === "auto" ||
          input.format === "csv" ||
          input.format === "tsv"
            ? input.format
            : null;
        if (
          firstColumn === undefined ||
          input.targetCount < 1 ||
          format === null
        ) {
          return Promise.resolve(invalidOperationPayload());
        }
        const request: PasteWorkbookClipboardRequest = {
          view_schema_id:
            input.viewSchemaId as PasteWorkbookClipboardRequest["view_schema_id"],
          client_txn_id: clientTxnId,
          clipboard_text: input.clipboardText,
          format,
          start_field_key: input.startFieldKey,
          columns: [firstColumn, ...remainingColumns],
          targets: [
            { kind: "create" },
            ...Array.from({ length: input.targetCount - 1 }, () => ({
              kind: "create" as const,
            })),
          ],
        };
        return operations
          .execute({
            operationID: "pasteWorkbookClipboard",
            pathParameters: {
              incident_id: context.incidentId,
              view_schema_id: input.viewSchemaId,
            },
            request,
          })
          .then((outcome) =>
            normalizeEntityPasteOutcome(outcome, input.viewSchemaId),
          );
      },
      merge(input) {
        const clientTxnId = createId(context.transactionIds, "merge");
        if (clientTxnId === null)
          return Promise.resolve(operationIdentityFailure());
        return operations
          .execute({
            operationID: "mergeEntityRecord",
            pathParameters: {
              survivor_record_id: input.survivorRecordId,
            },
            request: {
              loser_record_id: input.loserRecordId,
              survivor_base_row_version: input.survivorBaseRowVersion,
              loser_base_row_version: input.loserBaseRowVersion,
              client_txn_id: clientTxnId,
              reason: input.reason,
            } as MergeEntityRecordRequest,
          })
          .then((outcome) =>
            normalizeEntityMergeOutcome(
              outcome,
              input.loserRecordId,
              input.survivorRecordId,
            ),
          );
      },
    },
    assessment: {
      canCreate(input) {
        return (
          buildAssessmentCreatePayload(input.draft, "validation-only") !== null
        );
      },
      create(input) {
        const clientTxnId = createId(context.transactionIds, "assessment");
        if (clientTxnId === null)
          return Promise.resolve(operationIdentityFailure());
        const payload = buildAssessmentCreatePayload(input.draft, clientTxnId);
        if (payload === null) {
          return Promise.resolve(invalidOperationPayload());
        }
        return operations
          .execute({
            operationID: "createViewRow",
            pathParameters: {
              incident_id: context.incidentId,
              view_schema_id: assessmentsViewSchemaId,
            },
            request: payload as CreateViewRowRequest,
          })
          .then(normalizeAssessmentCreateOutcome);
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
        const availableClientTxnId = createId(
          context.transactionIds,
          "evidence-available",
        );
        if (
          createClientTxnId === null ||
          attachClientTxnId === null ||
          availableClientTxnId === null
        ) {
          return operationIdentityFailure();
        }
        if (input.file.size <= 0) return invalidOperationPayload();
        const createBlob = await operations.execute({
          operationID: "createObjectBlobSlot",
          request: {
            incident_id: context.incidentId,
            client_txn_id: createClientTxnId,
            byte_size: input.file.size,
            filename_hint: input.file.name || null,
            content_type_hint: input.file.type || null,
          } satisfies CreateObjectBlobSlotRequest,
        });
        if (createBlob.kind === "rejected") return createBlob;
        const upload = await uploadEvidenceObjectBlobTarget(
          context.apiBase,
          createBlob.value.data.upload_target,
          input.file,
        );
        if (upload.kind === "rejected") {
          return {
            kind: "rejected",
            failure: {
              kind: upload.retryable ? "retryable" : "terminal",
              message: upload.message,
            },
          };
        }
        const objectBlobId = createBlob.value.data.object_blob_id;
        const attach = await operations.execute({
          operationID: "attachBlobToEvidenceRecord",
          pathParameters: { record_id: input.evidenceRecordId },
          request: {
            object_blob_id: objectBlobId,
            base_row_version: input.baseRowVersion,
            client_txn_id: attachClientTxnId,
          } satisfies AttachBlobToEvidenceRecordRequest,
        });
        if (attach.kind === "rejected") return attach;
        if (
          attach.value.data.row.record_id !== input.evidenceRecordId ||
          attach.value.data.object_blob_id !== objectBlobId
        ) {
          return invalidOperationContract();
        }
        const available = await operations.execute({
          operationID: "patchRecord",
          pathParameters: { record_id: input.evidenceRecordId },
          request: {
            view_schema_id: evidenceViewSchemaId,
            base_row_version: attach.value.data.row.row_version,
            client_txn_id: availableClientTxnId,
            changes: [
              {
                field_key: "evidence.lifecycle_state",
                value: "available",
              },
            ],
          } satisfies PatchRecordRequest,
        });
        if (available.kind === "rejected") return available;
        if (
          available.value.data.view_schema_id !== evidenceViewSchemaId ||
          available.value.data.row.record_id !== input.evidenceRecordId ||
          available.value.data.row.cells["evidence.lifecycle_state"]?.value !==
            "available"
        ) {
          return invalidOperationContract();
        }
        return {
          kind: "accepted",
          value: { evidenceRecordId: input.evidenceRecordId },
        };
      },
      async issueHandle(input) {
        const operationID =
          input.kind === "preview"
            ? "issueEvidencePreviewHandle"
            : "issueEvidenceDownloadHandle";
        const outcome = await operations.execute({
          operationID,
          pathParameters: { record_id: input.evidenceRecordId },
          request: {} satisfies IssueEvidencePreviewHandleRequest &
            IssueEvidenceDownloadHandleRequest,
        });
        if (outcome.kind === "rejected") return outcome;
        const href = resolvePublicEvidenceHandleHref(outcome.value.data.href);
        if (
          href === null ||
          outcome.value.data.record_id !== input.evidenceRecordId
        ) {
          return {
            kind: "rejected",
            failure: {
              kind: "invalid_contract",
              message: "Evidence handle is unavailable.",
            },
          };
        }
        return {
          kind: "accepted",
          value: {
            filename: outcome.value.data.filename,
            href,
            previewKind: outcome.value.data.preview_kind ?? null,
          },
        };
      },
    },
    coordination: {
      async updateTaskLifecycle(input) {
        const clientTxnId = createId(context.transactionIds, "task-lifecycle");
        if (clientTxnId === null) return operationIdentityFailure();
        const changes: PatchRecordRequest["changes"] = [
          { field_key: "task.status", value: input.status },
        ];
        if (input.status === "blocked") {
          changes.push({
            field_key: "task.blocked_reason",
            value: input.blockedReason,
          });
        }
        return normalizeTaskLifecycleOutcome(
          await operations.execute({
            operationID: "patchRecord",
            pathParameters: { record_id: input.recordId },
            request: {
              view_schema_id: taskRequestsViewSchemaId,
              base_row_version: input.baseRowVersion,
              client_txn_id: clientTxnId,
              changes,
            } satisfies PatchRecordRequest,
          }),
          input.recordId,
          input.status,
        );
      },
      async supersedeDecision(input) {
        const clientTxnId = createId(
          context.transactionIds,
          "decision-supersede",
        );
        if (clientTxnId === null) return operationIdentityFailure();
        return normalizeDecisionSupersedeOutcome(
          await operations.execute({
            operationID: "supersedeRecord",
            pathParameters: { record_id: input.targetRecordId },
            request: {
              base_row_version: input.baseRowVersion,
              client_txn_id: clientTxnId,
              replacement_record_id: input.replacementRecordId,
              reason: input.reason,
            } satisfies SupersedeRecordRequest,
          }),
          input.targetRecordId,
          input.replacementRecordId,
        );
      },
    },
  };
}

function normalizeRecordLifecycleOutcome(
  outcome: WorkbookOperationOutcome<{
    readonly data: {
      readonly record_id: string;
      readonly row_version: number;
    };
  }>,
  recordId: string,
): WorkbookOperationOutcome<RecordLifecycleAccepted> {
  if (outcome.kind === "rejected") return outcome;
  return outcome.value.data.record_id === recordId
    ? {
        kind: "accepted",
        value: {
          recordId: outcome.value.data.record_id,
          rowVersion: outcome.value.data.row_version,
        },
      }
    : invalidOperationContract();
}
