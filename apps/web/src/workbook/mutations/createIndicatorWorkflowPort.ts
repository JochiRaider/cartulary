import type {
  WorkbookOperationExecutor,
  WorkbookOperationResponse,
} from "../adapters/workbookOperationContract";
import type {
  IndicatorMutationAccepted,
  IndicatorObservation,
  IndicatorPage,
  IndicatorStateInterval,
  IndicatorWorkflowPort,
} from "./workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "./workbookOperationOutcome";

type ObservationMutationResponse =
  WorkbookOperationResponse<"createManualIndicatorObservation">;
type LifecycleMutationResponse =
  WorkbookOperationResponse<"appendIndicatorStateInterval">;

function invalidIdentity<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "terminal",
      message: "A secure transaction ID could not be created.",
    },
  };
}

function listQuery(input: {
  readonly cursorToken?: string | undefined;
  readonly limit?: number | undefined;
}) {
  return {
    ...(input.cursorToken === undefined
      ? {}
      : { cursor_token: input.cursorToken }),
    ...(input.limit === undefined ? {} : { limit: input.limit }),
  };
}

function observationPage(
  outcome: WorkbookOperationOutcome<
    WorkbookOperationResponse<"listIndicatorObservations">
  >,
): WorkbookOperationOutcome<IndicatorPage<IndicatorObservation>> {
  if (outcome.kind === "rejected") return outcome;
  return {
    kind: "accepted",
    value: {
      items: outcome.value.data.observations,
      paging: outcome.value.meta.paging ?? null,
    },
  };
}

function lifecyclePage(
  outcome: WorkbookOperationOutcome<
    WorkbookOperationResponse<"listIndicatorStateIntervals">
  >,
): WorkbookOperationOutcome<IndicatorPage<IndicatorStateInterval>> {
  if (outcome.kind === "rejected") return outcome;
  return {
    kind: "accepted",
    value: {
      items: outcome.value.data.intervals,
      paging: outcome.value.meta.paging ?? null,
    },
  };
}

function observationMutation(
  outcome: WorkbookOperationOutcome<ObservationMutationResponse>,
): WorkbookOperationOutcome<IndicatorMutationAccepted<IndicatorObservation>> {
  if (outcome.kind === "rejected") return outcome;
  return {
    kind: "accepted",
    value: {
      affectedRecords: outcome.value.data.affected_records,
      changeSetId: outcome.value.data.change_set_id,
      replayed: outcome.value.data.replayed,
      resource: outcome.value.data.observation,
    },
  };
}

function lifecycleMutation(
  outcome: WorkbookOperationOutcome<LifecycleMutationResponse>,
): WorkbookOperationOutcome<IndicatorMutationAccepted<IndicatorStateInterval>> {
  if (outcome.kind === "rejected") return outcome;
  return {
    kind: "accepted",
    value: {
      affectedRecords: outcome.value.data.affected_records,
      changeSetId: outcome.value.data.change_set_id,
      replayed: outcome.value.data.replayed,
      resource: outcome.value.data.interval,
    },
  };
}

export function createIndicatorWorkflowPort(options: {
  readonly createMutationID: (prefix: string) => string | null;
  readonly operations: WorkbookOperationExecutor;
}): IndicatorWorkflowPort {
  return {
    async listSourceObservations(input) {
      const outcome = await options.operations.execute({
        operationID: "listSourceRecordIndicatorObservations",
        pathParameters: { source_record_id: input.sourceRecordId },
        query: listQuery(input),
      });
      return observationPage(outcome);
    },
    async listObservations(input) {
      const outcome = await options.operations.execute({
        operationID: "listIndicatorObservations",
        pathParameters: { indicator_id: input.indicatorRecordId },
        query: listQuery(input),
      });
      return observationPage(outcome);
    },
    async listStateIntervals(input) {
      const outcome = await options.operations.execute({
        operationID: "listIndicatorStateIntervals",
        pathParameters: { indicator_id: input.indicatorRecordId },
        query: listQuery(input),
      });
      return lifecyclePage(outcome);
    },
    async createManualObservation(input) {
      const clientTxnId = options.createMutationID("indicator-observation");
      if (clientTxnId === null) return invalidIdentity();
      const outcome = await options.operations.execute({
        operationID: "createManualIndicatorObservation",
        pathParameters: { source_record_id: input.sourceRecordId },
        request: {
          client_txn_id: clientTxnId,
          base_row_version: input.baseRowVersion,
          source_field_key: input.sourceFieldKey,
          span_start_byte: input.spanStartByte,
          span_end_byte: input.spanEndByte,
          ...(input.parsedIndicatorType === undefined
            ? {}
            : { parsed_indicator_type: input.parsedIndicatorType }),
          ...(input.resolvedIndicatorRecordId === undefined
            ? {}
            : {
                resolved_indicator_record_id: input.resolvedIndicatorRecordId,
              }),
        },
      });
      return observationMutation(outcome);
    },
    async transitionObservation(input) {
      const clientTxnId = options.createMutationID(
        `indicator-observation-${input.action}`,
      );
      if (clientTxnId === null) return invalidIdentity();
      if (input.action === "resolve") {
        if (input.resolvedIndicatorRecordId === undefined) {
          return {
            kind: "rejected",
            failure: {
              kind: "validation",
              message: "A resolved Indicator record is required.",
            },
          };
        }
        return observationMutation(
          await options.operations.execute({
            operationID: "resolveIndicatorObservation",
            pathParameters: { observation_id: input.observationId },
            request: {
              client_txn_id: clientTxnId,
              base_row_version: input.baseRowVersion,
              resolved_indicator_record_id: input.resolvedIndicatorRecordId,
            },
          }),
        );
      }
      return observationMutation(
        await options.operations.execute({
          operationID:
            input.action === "dismiss"
              ? "dismissIndicatorObservation"
              : "restoreIndicatorObservation",
          pathParameters: { observation_id: input.observationId },
          request: {
            client_txn_id: clientTxnId,
            base_row_version: input.baseRowVersion,
          },
        }),
      );
    },
    async appendStateInterval(input) {
      const clientTxnId = options.createMutationID("indicator-lifecycle");
      if (clientTxnId === null) return invalidIdentity();
      return lifecycleMutation(
        await options.operations.execute({
          operationID: "appendIndicatorStateInterval",
          pathParameters: { indicator_id: input.indicatorRecordId },
          request: {
            client_txn_id: clientTxnId,
            base_row_version: input.baseRowVersion,
            lifecycle_state: input.lifecycleState,
            valid_from: input.validFrom,
            valid_to: input.validTo,
            confidence: input.confidence,
            rationale: input.rationale,
            support_refs: [...input.supportRefs],
            assessor: input.assessor,
          },
        }),
      );
    },
  };
}
