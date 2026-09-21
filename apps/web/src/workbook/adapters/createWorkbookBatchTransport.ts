import {
  buildHTTPOperationPath,
  httpOperationBindings,
} from "@cartulary/protocol-ts/http";
import { requireViewContract } from "@cartulary/view-contracts";
import { apiPath, fetchHTTPOperation } from "../../services/browserApi";
import {
  workbookPasteColumns,
  workbookPasteTargets,
  workbookPasteViewSchemaId,
} from "../models/workbookClipboardPaste";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import type {
  WorkbookBatchAttempt,
  WorkbookBatchReceipt,
  WorkbookBatchTransport,
} from "../runtime/workbookBatchOperation";
import {
  parseSameFieldConflictPayload,
  type WorkbookSameFieldConflictPayload,
} from "../runtime/workbookConflictModel";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";

function pathFor(
  attempt: Pick<WorkbookBatchAttempt, "apiBase" | "authority" | "plan">,
): string {
  return apiPath(
    attempt.apiBase,
    buildHTTPOperationPath(attempt.plan.operation, {
      incident_id: attempt.authority.incidentId,
      view_schema_id: attempt.plan.request.view_schema_id,
    }),
  );
}

/** Captures bytes once. Source owners supply semantic plans and interpret rows. */
export function createWorkbookBatchTransport(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readScope?: WorkbookReadScopeSource;
}): WorkbookBatchTransport {
  return {
    capture(plan, authority, id) {
      if (
        authority.incidentId !== options.incidentId ||
        !id ||
        !plan.request.targets.length
      )
        throw new Error("Invalid batch scope");
      if (plan.operation === "pasteWorkbookClipboard") {
        if (
          !plan.request.clipboard_text ||
          !workbookPasteViewSchemaId(plan.request.view_schema_id) ||
          !workbookPasteColumns(plan.request.columns) ||
          !workbookPasteTargets(plan.request.targets)
        )
          throw new Error("Invalid paste plan");
      } else {
        const contract = requireViewContract(timelineViewSchemaId);
        if (
          plan.request.view_schema_id !== timelineViewSchemaId ||
          (plan.request.kind === "clear_cells_v1"
            ? !validClearRequest(plan.request)
            : plan.request.kind === "fill_down_v1"
              ? contract.fieldMap[plan.request.field_key ?? ""]
                  ?.gridEditable !== true
              : !plan.request.tag_name?.trim())
        )
          throw new Error("Invalid bulk plan");
      }
      const captured = { id, plan, authority, apiBase: options.apiBase };
      return {
        ...captured,
        path: pathFor(captured),
        body: JSON.stringify({ ...plan.request, client_txn_id: id }),
      };
    },
    async send(attempt, signal) {
      const scope = options.readScope?.() ?? null;
      if (
        attempt.authority.incidentId !== options.incidentId ||
        attempt.path !== pathFor(attempt)
      )
        return { kind: "uncertain" };
      let status: number | null = null;
      try {
        const result = await fetchHTTPOperation<{ data: unknown }>({
          apiBase: attempt.apiBase,
          operationID: attempt.plan.operation,
          pathParameters: {
            incident_id: attempt.authority.incidentId,
            view_schema_id: attempt.plan.request.view_schema_id,
          },
          init: {
            method: httpOperationBindings[attempt.plan.operation].method,
            body: attempt.body,
            signal,
          },
          onResponse: (response) => {
            status = response.status;
          },
        });
        if (!result.ok) {
          if (
            status === null ||
            status < 400 ||
            status >= 500 ||
            !result.payload.error?.code
          )
            return { kind: "uncertain" };
          const failure = classifyWorkbookOperationFailure(
            result.status,
            result.payload,
            attempt.plan.operation,
          );
          return failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : { kind: "rejected", failure };
        }
        const receipt = validateWorkbookBatchReceipt(
          attempt,
          result.payload.data,
        );
        return receipt
          ? {
              kind: "acknowledged",
              receipt: {
                ...receipt,
                rows: receipt.rows.map((row) =>
                  acceptWorkbookRowObservation(row, scope),
                ),
              },
            }
          : { kind: "uncertain" };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}

export function validateWorkbookBatchReceipt(
  attempt: WorkbookBatchAttempt,
  value: unknown,
): WorkbookBatchReceipt | null {
  if (
    !value ||
    typeof value !== "object" ||
    !("view_schema_id" in value) ||
    !("rows" in value)
  )
    return null;
  const data = value as Record<string, unknown>;
  const plan = attempt.plan;
  const view = plan.request.view_schema_id;
  if (data.view_schema_id !== view || !Array.isArray(data.rows)) return null;
  // Historical Entity receipts predate the required empty conflicts member.
  const conflictValues =
    data.conflicts === undefined && plan.entityType ? [] : data.conflicts;
  if (!Array.isArray(conflictValues)) return null;
  const changeSetId = data.change_set_id;
  if (
    changeSetId !== undefined &&
    (typeof changeSetId !== "string" || !changeSetId)
  )
    return null;
  if (data.rows.length > 0 !== (changeSetId !== undefined)) return null;
  try {
    const contract = requireViewContract(view);
    const rows = normalizeWorkbookViewRows(
      contract,
      data.rows,
      "batch receipt",
    );
    const targets = plan.request.targets;
    const records = new Map(
      targets.flatMap((target, index) =>
        "record_id" in target
          ? [[target.record_id, { target, index }] as const]
          : [],
      ),
    );
    const columns =
      plan.operation === "pasteWorkbookClipboard"
        ? plan.request.columns
        : plan.request.kind === "clear_cells_v1"
          ? (plan.request.field_keys ?? [])
          : [
              plan.request.kind === "fill_down_v1"
                ? plan.request.field_key
                : "timeline.tags",
            ];
    const conflicts: WorkbookSameFieldConflictPayload[] = [];
    let lastOrder = -1;
    for (const raw of conflictValues) {
      const conflict = parseSameFieldConflictPayload(raw);
      if (
        !conflict ||
        plan.entityType ||
        !Number.isSafeInteger(conflict.current_row_version) ||
        conflict.current_row_version < conflict.base_row_version ||
        typeof conflict.server_updated_by !== "string" ||
        !conflict.server_updated_by ||
        typeof conflict.server_updated_at !== "string" ||
        !Number.isFinite(Date.parse(conflict.server_updated_at))
      )
        return null;
      if (
        conflict.conflict_resolution_class === "text_compare_merge" &&
        (!raw || typeof raw !== "object" || !("base_value" in raw))
      )
        return null;
      const target = records.get(conflict.record_id);
      const column = columns.indexOf(conflict.field_key);
      if (
        !target ||
        column < 0 ||
        target.target.base_row_version !== conflict.base_row_version ||
        contract.fieldMap[conflict.field_key]?.conflictResolutionClass !==
          conflict.conflict_resolution_class ||
        (plan.operation === "applyWorkbookBulkMutation" &&
          plan.request.kind === "clear_cells_v1" &&
          conflict.client_value !== null)
      )
        return null;
      const order = target.index * columns.length + column;
      if (order <= lastOrder) return null;
      lastOrder = order;
      conflicts.push(conflict);
    }
    if (plan.entityType) {
      // Exact reuse can return the same record for several source rows.
      if (rows.length !== targets.length || conflicts.length) return null;
    } else {
      const seen = new Set<string>();
      let created = 0;
      let previousTargetIndex = -1;
      for (const row of rows) {
        if (seen.has(row.record_id)) return null;
        seen.add(row.record_id);
        const target = records.get(row.record_id);
        if (target) {
          if (row.row_version <= target.target.base_row_version) return null;
          if (
            plan.operation === "applyWorkbookBulkMutation" &&
            plan.request.kind === "clear_cells_v1"
          ) {
            if (target.index <= previousTargetIndex) return null;
            previousTargetIndex = target.index;
            for (const field of columns) {
              if (
                field &&
                !conflicts.some(
                  (conflict) =>
                    conflict.record_id === row.record_id &&
                    conflict.field_key === field,
                ) &&
                row.cells[field]?.value !== null
              )
                return null;
            }
          }
        } else {
          if (
            plan.operation !== "pasteWorkbookClipboard" ||
            row.row_version !== 1
          )
            return null;
          created++;
        }
      }
      if (
        created !==
        targets.filter((target) => "kind" in target && target.kind === "create")
          .length
      )
        return null;
      // Source owners omit unchanged record targets. Their absence is a legal
      // no-op; creates and every supplied conflict still require exact coverage.
    }
    return {
      viewSchemaId: view,
      changeSetId: typeof changeSetId === "string" ? changeSetId : null,
      rows,
      conflicts,
    };
  } catch {
    return null;
  }
}

function validClearRequest(
  request: Omit<
    import("./workbookProtocolTypes").WorkbookProtocolBulkRequest,
    "client_txn_id"
  >,
): boolean {
  const fields = request.field_keys;
  const contract = requireViewContract(timelineViewSchemaId);
  return (
    fields !== undefined &&
    fields.length >= 1 &&
    fields.length <= 10 &&
    new Set(fields).size === fields.length &&
    fields.every((key) => {
      const field = contract.fieldMap[key];
      return (
        field?.patchWritable &&
        field.gridEditable &&
        field.clearable &&
        field.writeKind === "direct_value"
      );
    }) &&
    request.field_key === undefined &&
    request.value === undefined &&
    request.tag_name === undefined &&
    request.targets.length <= 500 &&
    new Set(request.targets.map((target) => target.record_id)).size ===
      request.targets.length &&
    request.targets.every(
      (target) =>
        target.record_id !== "" &&
        Number.isSafeInteger(target.base_row_version) &&
        target.base_row_version > 0,
    )
  );
}
