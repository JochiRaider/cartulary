import type { CreateViewRowResponse } from "@cartulary/protocol-ts/http";
import type { ViewContract } from "@cartulary/view-contracts";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { decodeCreateViewRowRequest } from "../models/workbookRequestDecoders";
import { requireWorkbookSurfaceRegistration } from "../models/workbookSurfaceRegistration";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import type {
  WorkbookPendingMutationAccepted,
  WorkbookPendingMutationPort,
} from "../ports/WorkbookPendingMutationPort";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import type { PendingReplayUnitState } from "../utils/workbookPendingQueue";
import { invalidWorkbookAdapterResult } from "./workbookAdapterResult";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";
import {
  acceptedRecordMutation,
  type CapturedRecordPatch,
  captureRecordPatch,
} from "./workbookRecordPatchTransport";

const invalidMessage = "The Workbook mutation response was invalid.";

type PendingMutationAdapterOptions = {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly readScope?: WorkbookReadScopeSource;
  readonly recordTiming?:
    | ((name: string, details?: Readonly<Record<string, unknown>>) => void)
    | undefined;
};

type PendingMutationOperationExecutor = ReturnType<
  typeof createWorkbookOperationExecutor
>;

function invalidMutationResult<
  Accepted = WorkbookPendingMutationAccepted,
>(): WorkbookOperationOutcome<Accepted> {
  return invalidWorkbookAdapterResult(invalidMessage);
}

function observedTransport(
  unit: PendingReplayUnitState,
  recordTiming:
    | ((name: string, details?: Readonly<Record<string, unknown>>) => void)
    | undefined,
) {
  const details = { kind: unit.kind };
  return {
    onJSONParsed: () => {
      recordTiming?.("pending_fetch_json_parsed", details);
    },
    onResponseStatus: (status: number) => {
      recordTiming?.("pending_fetch_response", { ...details, status });
    },
  };
}

function contractForUnit(unit: PendingReplayUnitState): ViewContract | null {
  try {
    return requireWorkbookSurfaceRegistration(unit.viewSchemaId).contract;
  } catch {
    return null;
  }
}

function staleTargetResult(): WorkbookOperationOutcome<CreateViewRowResponse> {
  return {
    kind: "rejected",
    failure: {
      kind: "stale_target",
      message: "The Workbook row is no longer available.",
    },
  };
}

function executeCreate(
  operations: PendingMutationOperationExecutor,
  options: PendingMutationAdapterOptions,
  unit: PendingReplayUnitState,
) {
  if (unit.recordId !== null) {
    return Promise.resolve(invalidMutationResult<CreateViewRowResponse>());
  }
  const contract = contractForUnit(unit);
  const request =
    contract === null
      ? null
      : decodeCreateViewRowRequest(contract, {
          ...unit.payloadIntent,
          client_txn_id: unit.clientTxnId,
        });
  if (request === null) {
    return Promise.resolve(invalidMutationResult<CreateViewRowResponse>());
  }
  options.recordTiming?.("pending_fetch_request", { kind: unit.kind });
  return operations.execute({
    observeTransport: observedTransport(unit, options.recordTiming),
    operationID: "createViewRow",
    pathParameters: {
      incident_id: options.incidentId,
      view_schema_id: unit.viewSchemaId,
    },
    request,
  });
}

function executePatch(
  operations: PendingMutationOperationExecutor,
  options: PendingMutationAdapterOptions,
  committedRowVersion: number | null,
  unit: PendingReplayUnitState,
  attempts: Map<string, CapturedRecordPatch>,
) {
  if (
    committedRowVersion === null ||
    unit.recordId === null ||
    unit.identity.kind !== "patch" ||
    unit.identity.changes.length === 0
  ) {
    return Promise.resolve(staleTargetResult());
  }
  const captured =
    attempts.get(unit.clientTxnId) ??
    captureRecordPatch({
      recordId: unit.recordId,
      baseRowVersion: committedRowVersion,
      changes: unit.identity.changes,
      clientTxnId: unit.clientTxnId,
      viewSchemaId: unit.viewSchemaId,
    });
  if (captured === null) {
    return Promise.resolve(invalidMutationResult<CreateViewRowResponse>());
  }
  attempts.set(unit.clientTxnId, captured);
  const request = JSON.parse(captured.body);
  return operations.execute({
    observeTransport: observedTransport(unit, options.recordTiming),
    operationID: "patchRecord",
    pathParameters: { record_id: captured.recordId },
    request,
  });
}

function executeOperation(
  operations: PendingMutationOperationExecutor,
  options: PendingMutationAdapterOptions,
  committedRowVersion: number | null,
  unit: PendingReplayUnitState,
  attempts: Map<string, CapturedRecordPatch>,
): Promise<WorkbookOperationOutcome<CreateViewRowResponse>> {
  return unit.kind === "create"
    ? executeCreate(operations, options, unit)
    : executePatch(operations, options, committedRowVersion, unit, attempts);
}

function responseCorrelatesToUnit(
  response: CreateViewRowResponse["data"],
  unit: PendingReplayUnitState,
): boolean {
  return (
    response.view_schema_id === unit.viewSchemaId &&
    response.change_set_id.trim() !== ""
  );
}

function normalizedMutationRow(
  contract: ViewContract,
  response: CreateViewRowResponse["data"],
  unit: PendingReplayUnitState,
) {
  try {
    return (
      normalizeWorkbookViewRows(
        contract,
        [response.row],
        `${unit.viewSchemaId} mutation response`,
      )[0] ?? null
    );
  } catch {
    return null;
  }
}

function rowCorrelatesToUnit(
  committedRowVersion: number | null,
  row: NonNullable<ReturnType<typeof normalizedMutationRow>>,
  unit: PendingReplayUnitState,
): boolean {
  if (row.view_schema_id !== unit.viewSchemaId || row.row_version < 1) {
    return false;
  }
  return (
    unit.kind === "create" ||
    (row.record_id === unit.recordId &&
      committedRowVersion !== null &&
      row.row_version > committedRowVersion)
  );
}

function normalizedAcceptedMutation(
  contract: ViewContract,
  committedRowVersion: number | null,
  response: CreateViewRowResponse["data"],
  unit: PendingReplayUnitState,
): WorkbookOperationOutcome<WorkbookPendingMutationAccepted> {
  if (!responseCorrelatesToUnit(response, unit)) {
    return invalidMutationResult();
  }
  const row = normalizedMutationRow(contract, response, unit);
  if (row === null || !rowCorrelatesToUnit(committedRowVersion, row, unit)) {
    return invalidMutationResult();
  }
  const accepted = acceptedRecordMutation(
    response,
    unit.viewSchemaId,
    unit.recordId ?? undefined,
  );
  return accepted
    ? { kind: "accepted", value: accepted }
    : invalidMutationResult();
}

export function createWorkbookPendingMutationAdapter(
  options: PendingMutationAdapterOptions,
): WorkbookPendingMutationPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  const attempts = new Map<string, CapturedRecordPatch>();
  let retired = false;
  return {
    retire() {
      retired = true;
      attempts.clear();
    },
    async execute({ committedRowVersion, unit }) {
      const scope = options.readScope?.() ?? null;
      if (retired) return invalidMutationResult();
      if (unit.incidentId !== options.incidentId) {
        return invalidMutationResult();
      }
      const contract = contractForUnit(unit);
      if (contract === null) return invalidMutationResult();
      try {
        const outcome = await executeOperation(
          operations,
          options,
          committedRowVersion,
          unit,
          attempts,
        );
        const result =
          outcome.kind === "rejected"
            ? outcome
            : normalizedAcceptedMutation(
                contract,
                attempts.get(unit.clientTxnId)?.baseRowVersion ??
                  committedRowVersion,
                outcome.value.data,
                unit,
              );
        // A malformed response cannot prove that the server rejected a write.
        if (
          result.kind === "rejected" &&
          result.failure.kind === "invalid_contract"
        )
          return {
            kind: "rejected",
            failure: {
              kind: "retryable",
              message:
                "The save acknowledgement could not be verified. Retrying the captured request.",
            },
          };
        if (
          result.kind === "accepted" ||
          (result.failure.kind !== "retryable" &&
            result.failure.kind !== "authentication_required" &&
            result.failure.kind !== "authorization_lost")
        )
          attempts.delete(unit.clientTxnId);
        return result.kind === "accepted"
          ? {
              ...result,
              value: {
                ...result.value,
                row: acceptWorkbookRowObservation(result.value.row, scope),
              },
            }
          : result;
      } catch {
        return {
          kind: "rejected",
          failure: {
            kind: "retryable",
            message: "The Workbook mutation could not be sent.",
          },
        };
      }
    },
  };
}
