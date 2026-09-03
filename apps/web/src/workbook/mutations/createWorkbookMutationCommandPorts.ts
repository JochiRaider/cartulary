import type {
  ApplyWorkbookBulkMutationRequest,
  IssueEvidenceDownloadHandleRequest,
  IssueEvidencePreviewHandleRequest,
  MergeEntityRecordRequest,
  PatchRecordRequest,
  RollbackRecordRequest,
  SupersedeRecordRequest,
  SupersedeRecordResponse,
} from "@cartulary/protocol-ts/http";
import { resolvePublicEvidenceHandleHref } from "../../services/workbookEvidence";
import type { WorkbookOperationExecutor } from "../adapters/workbookOperationContract";
import { createWorkbookOperationExecutor } from "../adapters/workbookOperationExecutor";
import { createEvidenceAttachmentPort } from "../features/evidence/createEvidenceAttachmentPort";
import { createGenericMutationCommandPort } from "../features/generic/createGenericMutationCommandPort";
import { buildGenericCreateRequest } from "../features/generic/genericCreateRequestBuilder";
import { normalizeRecordHistoryData } from "../inspector/workbookRecordHistoryModel";
import { buildAssessmentCreatePayload } from "../models/assessmentWorkbookModel";
import {
  buildPatchRecordRequest,
  decodeCreateViewRowRequest,
} from "../models/workbookRequestDecoders";
import {
  assessmentsViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { createTimelineRelatedRecordCommandAdapter } from "../timeline/adapters/createTimelineRelatedRecordCommandAdapter";
import { normalizeTimelineFullRow } from "../timeline/models/timelineRowModel";
import { createIndicatorWorkflowPort } from "./createIndicatorWorkflowPort";
import type { SecureTransactionIdPort } from "./secureTransactionId";
import type {
  AssessmentCreateOutcome,
  DecisionSupersedeOutcome,
  EntityCreateOutcome,
  EntityMergeOutcome,
  EntityPatchOutcome,
  GenericViewMutationAccepted,
  RecordLifecycleAccepted,
  TaskLifecycleOutcome,
  TimelineFillOutcome,
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
}): Promise<TimelineFillOutcome> {
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

function timelineBulkTargets(
  targets: readonly {
    readonly baseRowVersion: number;
    readonly recordId: string;
  }[],
): ApplyWorkbookBulkMutationRequest["targets"] | null {
  const [firstTarget, ...remainingTargets] = targets;
  if (firstTarget === undefined) return null;
  return [
    {
      base_row_version: firstTarget.baseRowVersion,
      record_id: firstTarget.recordId,
    },
    ...remainingTargets.map((target) => ({
      base_row_version: target.baseRowVersion,
      record_id: target.recordId,
    })),
  ];
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
            } satisfies RollbackRecordRequest,
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
      fill: {
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
          const targets = timelineBulkTargets(input.targets);
          if (targets === null) {
            return { clientTxnId, outcome: invalidOperationPayload() };
          }
          const outcome = await executeTimelineBulkMutation({
            input: {
              client_txn_id: clientTxnId,
              field_key: input.fieldKey,
              kind: "fill_down_v1",
              targets,
              value: input.value,
              view_schema_id: timelineViewSchemaId,
            },
            incidentId: context.incidentId,
            operations,
          });
          return { clientTxnId, outcome };
        },
      },
      related: createTimelineRelatedRecordCommandAdapter({
        createClientTxnId: (prefix) => createId(context.transactionIds, prefix),
        incidentId: context.incidentId,
        operations,
      }),
    },
    generic: createGenericMutationCommandPort({
      incidentId: context.incidentId,
      operations,
      transactionIds: context.transactionIds,
    }),
    entity: {
      canCreateRecord(input) {
        return (
          buildGenericCreateRequest(
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
        const payload = buildGenericCreateRequest(
          input.contract,
          { ...input.draft },
          clientTxnId,
        );
        const request = decodeCreateViewRowRequest(input.contract, payload);
        if (request === null) return Promise.resolve(invalidOperationPayload());
        return operations
          .execute({
            operationID: "createViewRow",
            pathParameters: {
              incident_id: context.incidentId,
              view_schema_id: input.contract.viewSchemaId,
            },
            request,
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
        const request = buildPatchRecordRequest({
          baseRowVersion: input.baseRowVersion,
          changes: input.changes,
          clientTxnId,
          viewSchemaId: input.viewSchemaId,
        });
        if (request === null) return Promise.resolve(invalidOperationPayload());
        return operations
          .execute({
            operationID: "patchRecord",
            pathParameters: { record_id: input.recordId },
            request,
          })
          .then((outcome) =>
            normalizeEntityPatchOutcome(
              outcome,
              input.recordId,
              input.viewSchemaId,
            ),
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
            } satisfies MergeEntityRecordRequest,
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
            request: payload,
          })
          .then(normalizeAssessmentCreateOutcome);
      },
    },
    evidence: {
      ...createEvidenceAttachmentPort({
        apiBase: context.apiBase,
        incidentId: context.incidentId,
        operations,
        transactionIds: context.transactionIds,
      }),
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
