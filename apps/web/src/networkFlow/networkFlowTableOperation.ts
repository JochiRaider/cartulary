import type { ExtensionAvailabilityTag } from "../extensions/extensionAvailability";
import { createClientTransactionId } from "../services/clientTransactionId";
import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";
import type { NetworkFlowRequestError } from "./networkFlowErrors";

export type TableAction = "rename" | "delete";
export type TableAuthority = {
  readonly incidentId: string;
  readonly actorId: string | null;
  readonly sessionIdentity: string | null;
  readonly role: string | null;
  readonly open: boolean;
  readonly available: boolean;
  readonly profileAvailable: boolean;
  readonly availabilityTag: ExtensionAvailabilityTag | null;
};
export function tableAuthorityKey(value: TableAuthority): string {
  return JSON.stringify([
    value.incidentId,
    value.actorId,
    value.sessionIdentity,
    value.role,
    value.open,
    value.available,
    value.profileAvailable,
    value.availabilityTag?.epochId,
    value.availabilityTag?.generation.toString(),
  ]);
}
export function canReadTables(value: TableAuthority): boolean {
  return (
    value.actorId !== null &&
    value.sessionIdentity !== null &&
    value.open &&
    value.available &&
    value.profileAvailable &&
    value.availabilityTag !== null &&
    ["viewer", "editor", "reviewer", "admin"].includes(value.role ?? "")
  );
}
export function canMutateTable(
  value: TableAuthority,
  action: TableAction,
): boolean {
  return (
    canReadTables(value) &&
    (action === "rename"
      ? value.role === "editor" || value.role === "admin"
      : value.role === "reviewer" || value.role === "admin")
  );
}
export type TableNameResult =
  | { readonly ok: true; readonly name: string; readonly scalarCount: number }
  | {
      readonly ok: false;
      readonly reason:
        | "forbidden_control"
        | "empty_display_name"
        | "display_name_too_long";
      readonly message: string;
    };
/** Table names have a scalar limit, independently of saved-graph byte limits. */
export function normalizeTableDisplayName(value: string): TableNameResult {
  const name = value.normalize("NFC");
  if (
    Array.from(name).some((character) => {
      const scalar = character.codePointAt(0) as number;
      return (
        scalar <= 31 ||
        (scalar >= 127 && scalar <= 159) ||
        (scalar >= 0xd800 && scalar <= 0xdfff)
      );
    })
  )
    return {
      ok: false,
      reason: "forbidden_control",
      message: "Remove control characters from the table name.",
    };
  // Exact trim_unicode_whitespace_v1 set; controls have already been rejected.
  const trimmed = name.replace(
    /^[\u0020\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+|[\u0020\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+$/gu,
    "",
  );
  const scalarCount = Array.from(trimmed).length;
  if (scalarCount === 0)
    return {
      ok: false,
      reason: "empty_display_name",
      message: "Enter a table name.",
    };
  if (scalarCount > 64)
    return {
      ok: false,
      reason: "display_name_too_long",
      message: "Use at most 64 Unicode characters after normalization.",
    };
  return { ok: true, name: trimmed, scalarCount };
}
export type TableAttempt = {
  readonly action: TableAction;
  readonly authority: TableAuthority;
  readonly target: NetworkFlowTable;
  readonly transactionId: string;
  readonly normalizedName: string | null;
  readonly body: string;
};
export function captureTableAttempt(
  action: TableAction,
  authority: TableAuthority,
  target: NetworkFlowTable,
  name: string,
): TableAttempt {
  const normalized =
    action === "rename" ? normalizeTableDisplayName(name) : null;
  if (normalized?.ok === false) throw new Error(normalized.message);
  const transactionId = createClientTransactionId(`nf-table-${action}`);
  const normalizedName = normalized?.ok === true ? normalized.name : null;
  return Object.freeze({
    action,
    authority: Object.freeze({ ...authority }),
    target: Object.freeze({ ...target }),
    transactionId,
    normalizedName,
    body: JSON.stringify({
      client_txn_id: transactionId,
      base_table_version: target.table_version,
      ...(action === "rename" ? { display_name: normalizedName } : {}),
    }),
  });
}
export class TableWriteError extends Error {
  constructor(
    readonly certainty: "rejected" | "uncertain",
    message: string,
    readonly detail: NetworkFlowRequestError | null = null,
  ) {
    super(message);
    this.name = "TableWriteError";
  }
}
export function tableWriteFailure(error: unknown): TableWriteError {
  if (error instanceof TableWriteError) return error;
  return new TableWriteError(
    "uncertain",
    "The acknowledgement was not received. The request may have committed. Replay the exact request to recover its receipt.",
  );
}
export function validateTableReceipt(
  table: NetworkFlowTable,
  status: number,
  attempt: TableAttempt,
): NetworkFlowTable {
  const target = attempt.target;
  const mutable = new Set([
    "display_name",
    "table_version",
    "table_status",
    "updated_at",
    "deleted_at",
  ]);
  const normalized = normalizeTableDisplayName(table.display_name);
  const noop =
    attempt.action === "rename" &&
    attempt.normalizedName === target.display_name;
  if (
    status !== 200 ||
    table.incident_id !== attempt.authority.incidentId ||
    table.network_flow_table_id !== target.network_flow_table_id ||
    !normalized.ok ||
    normalized.name !== table.display_name ||
    Object.keys(target).some(
      (key) =>
        !mutable.has(key) &&
        table[key as keyof NetworkFlowTable] !==
          target[key as keyof NetworkFlowTable],
    ) ||
    table.table_version !== target.table_version + (noop ? 0 : 1) ||
    (attempt.action === "rename"
      ? table.table_status !== "active" ||
        table.deleted_at !== null ||
        table.display_name !== attempt.normalizedName ||
        (noop && table.updated_at !== target.updated_at)
      : table.table_status !== "soft_deleted" ||
        table.deleted_at === null ||
        table.display_name !== target.display_name)
  )
    throw new TableWriteError(
      "uncertain",
      "The acknowledgement does not match this request. Replay the exact request to recover its receipt.",
    );
  return Object.freeze({ ...table });
}
