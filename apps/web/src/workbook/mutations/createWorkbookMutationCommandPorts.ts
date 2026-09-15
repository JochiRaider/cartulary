import type {
  ApplyWorkbookBulkMutationRequest,
  IssueEvidenceDownloadHandleRequest,
  IssueEvidencePreviewHandleRequest,
} from "@cartulary/protocol-ts/http";
import { resolvePublicEvidenceHandleHref } from "../../services/workbookEvidence";
import { createAssessmentAppendTransport } from "../adapters/createAssessmentAppendTransport";
import { createWorkbookRecordHistoryAdapter } from "../adapters/createWorkbookRecordHistoryAdapter";
import { createWorkbookOperationExecutor } from "../adapters/workbookOperationExecutor";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookBatchOperationOwner } from "../runtime/WorkbookBatchOperationOwner";
import { createTimelineRelatedRecordCommandAdapter } from "../timeline/adapters/createTimelineRelatedRecordCommandAdapter";
import type { SecureTransactionIdPort } from "./secureTransactionId";
import type { WorkbookMutationCommandPorts } from "./workbookMutationCommandPorts";

type CommandContext = {
  readonly batches: Pick<WorkbookBatchOperationOwner, "admit">;
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly transactionIds: SecureTransactionIdPort;
};

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
    assessment: createAssessmentAppendTransport(context.apiBase),
    evidence: {
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
