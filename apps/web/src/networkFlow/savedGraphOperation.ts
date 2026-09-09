import type { ExtensionAvailabilityTag } from "../extensions/extensionAvailability";
import { createClientTransactionId } from "../services/clientTransactionId";
import {
  type NetworkFlowGraphSemanticQuery,
  type NetworkFlowSavedGraph,
  type NetworkFlowSavedGraphAccepted,
  normalizeSavedGraphDisplayName,
} from "../services/networkFlowContractAdapter";
import { NetworkFlowRequestError } from "./networkFlowErrors";

export type SavedGraphAction = "create" | "rename" | "refresh" | "retire";
export type SavedGraphAuthority = {
  readonly incidentId: string;
  readonly actorId: string | null;
  readonly session: string | object;
  readonly role: string | null;
  readonly open: boolean;
  readonly available: boolean;
  readonly availabilityTag: ExtensionAvailabilityTag | null;
};
export type SavedGraphIntent = {
  readonly kind: SavedGraphAction;
  readonly authority: SavedGraphAuthority;
  readonly target: NetworkFlowSavedGraph | null;
  readonly query: NetworkFlowGraphSemanticQuery | null;
};
export type SavedGraphAttempt = {
  readonly intent: SavedGraphIntent;
  readonly transactionId: string;
  readonly name: string | null;
  readonly body: string;
};
export type SavedGraphReceipt =
  | { readonly kind: "accepted"; readonly value: NetworkFlowSavedGraphAccepted }
  | { readonly kind: "renamed"; readonly value: NetworkFlowSavedGraph }
  | { readonly kind: "retired" };
export type SavedGraphFailureCategory =
  | "validation"
  | "version_conflict"
  | "transaction_conflict"
  | "quota"
  | "authorization"
  | "lifecycle"
  | "transport"
  | "invalid_response";
export class SavedGraphWriteError extends Error {
  constructor(
    readonly category: SavedGraphFailureCategory,
    readonly certainty: "rejected" | "uncertain",
    message: string,
    readonly detail: NetworkFlowRequestError | null = null,
  ) {
    super(message);
    this.name = "SavedGraphWriteError";
  }
}
export function savedGraphWriteFailure(caught: unknown): SavedGraphWriteError {
  if (caught instanceof SavedGraphWriteError) return caught;
  if (caught instanceof NetworkFlowRequestError) {
    const category: SavedGraphFailureCategory = caught.code.includes(
      "version_conflict",
    )
      ? "version_conflict"
      : caught.code.includes("idempotency") ||
          caught.code.includes("transaction")
        ? "transaction_conflict"
        : caught.code.includes("limit") ||
            caught.code.includes("quota") ||
            caught.status === 429
          ? "quota"
          : caught.status === 401 ||
              caught.status === 403 ||
              caught.code === "incident_not_found"
            ? "authorization"
            : caught.status === 404 ||
                caught.code === "incident_closed" ||
                caught.code.includes("source")
              ? "lifecycle"
              : caught.status >= 500 || caught.status === 0
                ? "transport"
                : "validation";
    return new SavedGraphWriteError(
      category,
      caught.status >= 400 && caught.status < 500 ? "rejected" : "uncertain",
      caught.message,
      caught,
    );
  }
  return new SavedGraphWriteError(
    "transport",
    "uncertain",
    "The acknowledgement was not received. The request may have committed. Replay this exact attempt to recover its receipt.",
  );
}
export function sameSavedGraphScope(
  a: SavedGraphAuthority,
  b: SavedGraphAuthority,
): boolean {
  return (
    a.incidentId === b.incidentId &&
    a.actorId === b.actorId &&
    a.session === b.session
  );
}
export function sameSavedGraphAuthority(
  a: SavedGraphAuthority,
  b: SavedGraphAuthority,
): boolean {
  return (
    sameSavedGraphScope(a, b) &&
    a.role === b.role &&
    a.open === b.open &&
    a.available === b.available &&
    a.availabilityTag?.epochId === b.availabilityTag?.epochId &&
    a.availabilityTag?.generation === b.availabilityTag?.generation
  );
}
export function canReadSavedGraphs(a: SavedGraphAuthority): boolean {
  return (
    a.actorId !== null &&
    a.open &&
    a.available &&
    a.availabilityTag !== null &&
    ["viewer", "editor", "reviewer", "admin"].includes(a.role ?? "")
  );
}
export function canMutateSavedGraph(
  a: SavedGraphAuthority,
  kind: SavedGraphAction,
): boolean {
  return (
    canReadSavedGraphs(a) &&
    (kind === "retire"
      ? a.role === "reviewer" || a.role === "admin"
      : a.role === "editor" || a.role === "admin")
  );
}
export function captureSavedGraphIntent(
  kind: SavedGraphAction,
  authority: SavedGraphAuthority,
  target: NetworkFlowSavedGraph | null,
  query: NetworkFlowGraphSemanticQuery | null,
): SavedGraphIntent {
  return Object.freeze({
    kind,
    authority: Object.freeze({ ...authority }),
    target: target === null ? null : freezeJSON(target),
    query: query === null ? null : freezeJSON(query),
  });
}
function freezeJSON<T>(value: T): T {
  const copy = structuredClone(value);
  const freeze = (item: unknown) => {
    if (item && typeof item === "object") {
      for (const child of Object.values(item)) freeze(child);
      Object.freeze(item);
    }
  };
  freeze(copy);
  return copy;
}
export function prepareSavedGraphAttempt(
  intent: SavedGraphIntent,
  draft: string,
  identify = createClientTransactionId,
): SavedGraphAttempt {
  let name: string | null = null;
  if (intent.kind === "create" || intent.kind === "rename") {
    const normalized = normalizeSavedGraphDisplayName(draft);
    if (!normalized.ok)
      throw new SavedGraphWriteError(
        "validation",
        "rejected",
        normalized.reason,
      );
    name = normalized.name;
  }
  if (
    (intent.kind === "create" && intent.query === null) ||
    (intent.kind !== "create" && intent.target === null)
  )
    throw new SavedGraphWriteError(
      "validation",
      "rejected",
      "Capture a current graph before submitting.",
    );
  const transactionId = identify(`nf-graph-view-${intent.kind}`);
  const schemas = {
    create: "cartulary.network_flow.graph_view_create_request.v3",
    rename: "cartulary.network_flow.graph_view_rename_request.v2",
    refresh: "cartulary.network_flow.graph_view_refresh_request.v1",
    retire: "cartulary.network_flow.graph_view_retire_request.v1",
  } as const;
  const body = JSON.stringify({
    schema_id: schemas[intent.kind],
    client_txn_id: transactionId,
    ...(intent.kind === "create"
      ? { semantic_query: intent.query }
      : { base_graph_view_version: intent.target?.graph_view_version }),
    ...(name === null ? {} : { display_name: name }),
  });
  return Object.freeze({ intent, transactionId, name, body });
}

export type SavedGraphLoadState =
  | "idle"
  | "loading"
  | "refreshing"
  | "ready"
  | "error";
