import type {
  CreateViewRowRequest,
  CreateViewRowResponse,
  PatchRecordRequest,
} from "@cartulary/protocol-ts";
import type { ViewContract } from "@cartulary/view-contracts";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { requireWorkbookSurfaceRegistration } from "../models/workbookSurfaceRegistration";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import type {
  WorkbookPendingMutationAccepted,
  WorkbookPendingMutationPort,
} from "../ports/WorkbookPendingMutationPort";
import type { PendingReplayUnitState } from "../utils/workbookPendingQueue";
import { invalidWorkbookAdapterResult } from "./workbookAdapterResult";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

const invalidMessage = "The Workbook mutation response was invalid.";

type PendingMutationAdapterOptions = {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
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
  const details = {
    clientTxnId: unit.clientTxnId,
    kind: unit.kind,
    rowKey: unit.rowKey,
  };
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
  return operations.execute({
    observeTransport: observedTransport(unit, options.recordTiming),
    operationID: "createViewRow",
    pathParameters: {
      incident_id: options.incidentId,
      view_schema_id: unit.viewSchemaId,
    },
    request: {
      ...unit.payloadIntent,
      client_txn_id: unit.clientTxnId,
    } as CreateViewRowRequest,
  });
}

function executePatch(
  operations: PendingMutationOperationExecutor,
  options: PendingMutationAdapterOptions,
  committedRowVersion: number | null,
  unit: PendingReplayUnitState,
) {
  if (
    committedRowVersion === null ||
    unit.recordId === null ||
    unit.identity.kind !== "patch" ||
    unit.identity.changes.length === 0
  ) {
    return Promise.resolve(staleTargetResult());
  }
  return operations.execute({
    observeTransport: observedTransport(unit, options.recordTiming),
    operationID: "patchRecord",
    pathParameters: { record_id: unit.recordId },
    request: {
      base_row_version: committedRowVersion,
      changes: unit.identity.changes,
      client_txn_id: unit.clientTxnId,
      view_schema_id: unit.viewSchemaId,
    } as PatchRecordRequest,
  });
}

function executeOperation(
  operations: PendingMutationOperationExecutor,
  options: PendingMutationAdapterOptions,
  committedRowVersion: number | null,
  unit: PendingReplayUnitState,
): Promise<WorkbookOperationOutcome<CreateViewRowResponse>> {
  return unit.kind === "create"
    ? executeCreate(operations, options, unit)
    : executePatch(operations, options, committedRowVersion, unit);
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
    (row.record_id === unit.recordId && committedRowVersion !== null)
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
  return {
    kind: "accepted",
    value: {
      changeSetId: response.change_set_id,
      row: { ...row, view_schema_id: unit.viewSchemaId },
      viewSchemaId: response.view_schema_id,
    },
  };
}

export function createWorkbookPendingMutationAdapter(
  options: PendingMutationAdapterOptions,
): WorkbookPendingMutationPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async execute({ committedRowVersion, unit }) {
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
        );
        return outcome.kind === "rejected"
          ? outcome
          : normalizedAcceptedMutation(
              contract,
              committedRowVersion,
              outcome.value.data,
              unit,
            );
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
