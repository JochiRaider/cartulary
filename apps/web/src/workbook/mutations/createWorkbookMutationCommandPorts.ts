import type {
  ApplyWorkbookBulkMutationRequest,
  IssueEvidenceDownloadHandleRequest,
  IssueEvidencePreviewHandleRequest,
} from "@cartulary/protocol-ts/http";
import { resolvePublicEvidenceHandleHref } from "../../services/workbookEvidence";
import { createAssessmentAppendTransport } from "../adapters/createAssessmentAppendTransport";
import { createWorkbookRecordHistoryAdapter } from "../adapters/createWorkbookRecordHistoryAdapter";
import { createWorkbookOperationExecutor } from "../adapters/workbookOperationExecutor";
import type { DecisionRecordWriteBoundary } from "../features/coordination/decisionSupersessionOperation";
import { createEvidenceAttachmentPort } from "../features/evidence/createEvidenceAttachmentPort";
import { createGenericMutationCommandPort } from "../features/generic/createGenericMutationCommandPort";
import { buildPatchRecordRequest } from "../models/workbookRequestDecoders";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookBatchOperationOwner } from "../runtime/WorkbookBatchOperationOwner";
import { createTimelineRelatedRecordCommandAdapter } from "../timeline/adapters/createTimelineRelatedRecordCommandAdapter";
import type {
  EntityRecordWriteBoundary,
  EntityRecordWriteTarget,
} from "./entityRecordWriteBoundary";
import type { SecureTransactionIdPort } from "./secureTransactionId";
import type {
  EntityPatchOutcome,
  GenericViewMutationAccepted,
  WorkbookMutationCommandPorts,
} from "./workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "./workbookOperationOutcome";

type CommandContext = {
  readonly batches: Pick<WorkbookBatchOperationOwner, "admit">;
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
          "Finish the earlier batch or recover the pending merge before changing these records.",
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
        fillDown(input, admission) {
          const targets = timelineBulkTargets(input.targets);
          if (!targets) return null;
          return context.batches.admit(
            {
              operation: "applyWorkbookBulkMutation",
              request: {
                field_key: input.fieldKey,
                kind: "fill_down_v1",
                targets,
                value: input.value,
                view_schema_id: timelineViewSchemaId,
              },
              recordIds: input.targets.map((target) => target.recordId),
            },
            admission,
          );
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
      patchRecord(input) {
        return executeEntityWrite(
          context,
          {
            recordIds: [input.recordId],
            entityType:
              input.viewSchemaId === "cartulary.view.hosts.v1"
                ? "host"
                : input.viewSchemaId === "cartulary.view.identities.v1"
                  ? "identity"
                  : undefined,
          },
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
    assessment: createAssessmentAppendTransport(context.apiBase),
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
