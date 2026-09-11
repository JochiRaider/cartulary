import type {
  ApplyWorkbookBulkMutationRequest,
  IssueEvidenceDownloadHandleRequest,
  IssueEvidencePreviewHandleRequest,
} from "@cartulary/protocol-ts/http";
import { resolvePublicEvidenceHandleHref } from "../../services/workbookEvidence";
import { createWorkbookRecordHistoryAdapter } from "../adapters/createWorkbookRecordHistoryAdapter";
import type { WorkbookOperationExecutor } from "../adapters/workbookOperationContract";
import { createWorkbookOperationExecutor } from "../adapters/workbookOperationExecutor";
import type { DecisionRecordWriteBoundary } from "../features/coordination/decisionSupersessionOperation";
import { createEvidenceAttachmentPort } from "../features/evidence/createEvidenceAttachmentPort";
import { createGenericMutationCommandPort } from "../features/generic/createGenericMutationCommandPort";
import { buildGenericCreateRequest } from "../features/generic/genericCreateRequestBuilder";
import { buildAssessmentCreatePayload } from "../models/assessmentWorkbookModel";
import {
  buildPatchRecordRequest,
  decodeCreateViewRowRequest,
} from "../models/workbookRequestDecoders";
import {
  assessmentsViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { createTimelineRelatedRecordCommandAdapter } from "../timeline/adapters/createTimelineRelatedRecordCommandAdapter";
import { normalizeTimelineFullRow } from "../timeline/models/timelineRowModel";
import type {
  EntityRecordWriteBoundary,
  EntityRecordWriteTarget,
} from "./entityRecordWriteBoundary";
import type { SecureTransactionIdPort } from "./secureTransactionId";
import type {
  AssessmentCreateOutcome,
  EntityCreateOutcome,
  EntityPatchOutcome,
  GenericViewMutationAccepted,
  TimelineFillOutcome,
  WorkbookMutationCommandPorts,
} from "./workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "./workbookOperationOutcome";

type CommandContext = {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly transactionIds: SecureTransactionIdPort;
  readonly entityWrites?: EntityRecordWriteBoundary;
  readonly decisionWrites?: DecisionRecordWriteBoundary;
};

async function executeEntityWrite(
  context: CommandContext,
  target: EntityRecordWriteTarget,
  run: () => Promise<EntityPatchOutcome>,
): Promise<EntityPatchOutcome> {
  const release = context.entityWrites
    ? context.entityWrites.begin(target)
    : () => {};
  if (release === null)
    return {
      kind: "rejected",
      failure: {
        kind: "stale_target",
        message:
          "Recover the pending merge in Merge actions before changing these records.",
      },
    };
  try {
    const outcome = await run();
    if (outcome.kind === "accepted")
      context.entityWrites?.acceptVersion(
        outcome.value.row.record_id,
        outcome.value.row.row_version,
      );
    return outcome;
  } finally {
    release();
  }
}

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
    records: createWorkbookRecordHistoryAdapter(context),
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
      decisionWrites: context.decisionWrites,
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
        return executeEntityWrite(
          context,
          {
            recordIds: [],
            unknownEntityType:
              input.contract.viewSchemaId === "cartulary.view.hosts.v1"
                ? "host"
                : "identity",
          },
          () => {
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
            if (request === null)
              return Promise.resolve(invalidOperationPayload());
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
                normalizeEntityCreateOutcome(
                  outcome,
                  input.contract.viewSchemaId,
                ),
              );
          },
        );
      },
      patchRecord(input) {
        return executeEntityWrite(
          context,
          { recordIds: [input.recordId] },
          () => {
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
            if (request === null)
              return Promise.resolve(invalidOperationPayload());
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
  };
}
