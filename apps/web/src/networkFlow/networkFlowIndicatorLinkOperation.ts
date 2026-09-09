import type { ExtensionAvailabilityTag } from "../extensions/extensionAvailability";
import { extractError } from "../services/browserApi";
import { createClientTransactionId } from "../services/clientTransactionId";
import type {
  NetworkFlowGraphResult,
  NetworkFlowIndicatorLinkRequest,
  NetworkFlowIndicatorLinkResult,
  NetworkFlowIndicatorSelector,
  NetworkFlowIndicatorTarget,
  NetworkFlowRowRef,
} from "../services/networkFlowContractAdapter";
import {
  networkFlowErrorMetadata,
  validNetworkFlowErrorEnvelope,
} from "../services/networkFlowContractAdapter";
import {
  compatibleAtomicIPTarget,
  coreAtomicIPType,
} from "../services/networkFlowIndicatorAdapter";

export type IndicatorLinkAuthority = {
  readonly incidentId: string;
  readonly actorId: string | null;
  readonly session: string;
  readonly sessionResolved: boolean;
  readonly role: string | null;
  readonly open: boolean;
  readonly available: boolean;
  readonly profileAvailable: boolean;
  readonly availabilityTag: ExtensionAvailabilityTag | null;
};

export type NetworkFlowIndicatorLinkCandidate = {
  readonly candidateValue: string;
  readonly key: string;
  readonly label: string;
  readonly selector: NetworkFlowIndicatorSelector;
  readonly sourceRefs: readonly NetworkFlowRowRef[];
  readonly sourceTableIds: readonly string[];
  readonly sourceTableRefs: NetworkFlowGraphResult["source_table_refs"];
};

export type IndicatorLinkAttempt = {
  readonly authority: IndicatorLinkAuthority;
  readonly candidate: NetworkFlowIndicatorLinkCandidate;
  readonly selectionRevision: number;
  readonly draftRevision: number;
  readonly sourceLimit: number;
  readonly request: NetworkFlowIndicatorLinkRequest;
  readonly body: string;
};

export type IndicatorLinkFeedback = {
  readonly kind: "validation" | "stale" | "denied" | "recovery" | "read_failed";
  readonly message: string;
  readonly field: "confirmation" | "target" | null;
};

export class IndicatorLinkWriteError extends Error {
  constructor(
    readonly certainty: "not_dispatched" | "rejected" | "uncertain",
    readonly feedback: IndicatorLinkFeedback,
  ) {
    super(feedback.message);
    this.name = "IndicatorLinkWriteError";
  }
}

/** An invalid rejection is a read failure, not evidence that the write rolled back. */
export function validIndicatorLinkRejection(
  status: number,
  payload: unknown,
): boolean {
  if (!validNetworkFlowErrorEnvelope(status, payload)) return false;
  const error = extractError(payload);
  if (!error || status < 400 || status >= 500) return false;
  if (!error.code.startsWith("network_flow_")) return true;
  const contract = networkFlowErrorMetadata.errors.find(
    (item) => item.code === error.code,
  );
  const details = error.details;
  if (
    !contract ||
    contract.http_status !== status ||
    !details ||
    details.retry_action !== contract.retry_action
  )
    return false;
  const linkError = [
    "network_flow_indicator_link_ambiguous",
    "network_flow_invalid_indicator_selector",
    "network_flow_invalid_indicator_target",
    "network_flow_indicator_link_forbidden",
  ].includes(error.code);
  const reasons = networkFlowErrorMetadata.reason_registries.find(
    (item) =>
      item.error_code.split(",").includes(error.code) ||
      (linkError && item.error_code === "indicator-link errors"),
  );
  if (!reasons?.reason_codes.some((reason) => reason === details.reason_code))
    return false;
  const required: Readonly<Record<string, readonly string[]>> = {
    network_flow_invalid_request: ["field", "expected_contract", "actual_kind"],
    network_flow_source_changed: [
      "approved_source_content_sha256",
      "observed_source_content_sha256",
      "approved_mapping_fingerprint",
      "observed_mapping_fingerprint",
    ],
    network_flow_table_not_found: [
      "network_flow_table_id",
      "table_status",
      "allowed_states",
    ],
    network_flow_table_not_active: [
      "network_flow_table_id",
      "table_status",
      "allowed_states",
    ],
    network_flow_invalid_table_scope: ["mode", "table_ids", "limit_key"],
    network_flow_invalid_filter: ["field_key", "op", "filter_index"],
    network_flow_invalid_time_range: ["field", "start_utc", "end_utc"],
    network_flow_graph_query_stale: ["graph_query_digest"],
    network_flow_resource_limit_exceeded: [
      "limit_key",
      "limit",
      "actual",
      "phase",
    ],
    network_flow_graph_limit_exceeded: [
      "limit_key",
      "limit",
      "actual",
      "phase",
    ],
    network_flow_counter_sum_limit_exceeded: [
      "limit_key",
      "limit",
      "actual",
      "phase",
    ],
  };
  const members = linkError
    ? ["selector_kind", "field_key", "target_mode", "resolved_candidate_value"]
    : (required[error.code] ?? []);
  return members.every((name) => Object.hasOwn(details, name));
}

export function canLinkIndicator(authority: IndicatorLinkAuthority): boolean {
  return (
    authority.sessionResolved &&
    authority.actorId !== null &&
    authority.available &&
    authority.profileAvailable &&
    authority.open &&
    (authority.role === "editor" || authority.role === "admin")
  );
}

export function sameIndicatorLinkAuthority(
  a: IndicatorLinkAuthority,
  b: IndicatorLinkAuthority,
): boolean {
  return (
    a.incidentId === b.incidentId &&
    a.actorId === b.actorId &&
    a.session === b.session &&
    a.sessionResolved === b.sessionResolved &&
    a.role === b.role &&
    a.open === b.open &&
    a.available === b.available &&
    a.profileAvailable === b.profileAvailable &&
    a.availabilityTag?.epochId === b.availabilityTag?.epochId &&
    a.availabilityTag?.generation === b.availabilityTag?.generation
  );
}

export function captureIndicatorLinkAttempt(input: {
  readonly authority: IndicatorLinkAuthority;
  readonly candidate: NetworkFlowIndicatorLinkCandidate;
  readonly selectionRevision: number;
  readonly draftRevision: number;
  readonly sourceLimit: number;
  readonly target: NetworkFlowIndicatorTarget;
  readonly confirmation: string;
}): IndicatorLinkAttempt {
  const request: NetworkFlowIndicatorLinkRequest = {
    schema_id: "cartulary.network_flow.indicator_link_request.v1",
    client_txn_id: createClientTransactionId("nf-indicator-link"),
    selector: structuredClone(input.candidate.selector),
    target: structuredClone(input.target),
    observation_mode: "binding_only",
    confirm_exact_value: input.confirmation,
  };
  return freezeLinkValue({
    authority: structuredClone(input.authority),
    candidate: structuredClone(input.candidate),
    selectionRevision: input.selectionRevision,
    draftRevision: input.draftRevision,
    sourceLimit: input.sourceLimit,
    request,
    body: JSON.stringify(request),
  });
}

export function validateIndicatorLinkReceipt(
  result: NetworkFlowIndicatorLinkResult,
  status: number,
  attempt: IndicatorLinkAttempt,
): NetworkFlowIndicatorLinkResult {
  const binding = result.binding;
  const candidate = attempt.candidate;
  const refs = binding.source_row_refs;
  if (
    status !== (result.duplicate ? 200 : 201) ||
    binding.incident_id !== attempt.authority.incidentId ||
    binding.candidate_value !== candidate.candidateValue ||
    !compatibleAtomicIPTarget(
      binding.target_indicator_ref,
      candidate.candidateValue,
    ) ||
    refs.length === 0 ||
    refs.length > attempt.sourceLimit ||
    new Set(refs.map((ref) => ref.network_flow_row_id)).size !== refs.length ||
    binding.source_row_refs_total_count < refs.length ||
    binding.source_row_refs_truncated !==
      binding.source_row_refs_total_count > refs.length ||
    (binding.selector_kind === "row_field_value" &&
      (refs.length !== 1 || binding.source_row_refs_truncated)) ||
    (binding.selector_kind === "row_refs" && binding.source_row_refs_truncated)
  )
    throw invalidLinkReceipt();
  if (
    attempt.request.target.mode === "existing_indicator" &&
    binding.target_indicator_ref.indicator_id.toLowerCase() !==
      attempt.request.target.indicator_id.toLowerCase()
  )
    throw invalidLinkReceipt();
  // Reuse identifies the binding tuple. Its original selector, creator, time,
  // and graph truncation metadata need not describe this invocation.
  if (!result.duplicate && binding.selector_kind !== candidate.selector.kind)
    throw invalidLinkReceipt();
  if (
    candidate.selector.kind === "row_refs" ||
    candidate.selector.kind === "row_field_value"
  ) {
    const expected = new Map(
      candidate.sourceRefs.map((ref) => [ref.network_flow_row_id, ref]),
    );
    if (
      refs.length !== expected.size ||
      refs.some((ref) => {
        const source = expected.get(ref.network_flow_row_id);
        return (
          source === undefined ||
          source.network_flow_table_id !== ref.network_flow_table_id ||
          source.source_row_number !== ref.source_row_number ||
          source.mapping_fingerprint !== ref.mapping_fingerprint
        );
      })
    )
      throw invalidLinkReceipt();
  } else if (
    refs.some(
      (ref) =>
        !candidate.sourceTableIds.includes(ref.network_flow_table_id) ||
        candidate.sourceTableRefs.find(
          (table) => table.network_flow_table_id === ref.network_flow_table_id,
        )?.mapping_fingerprint !== ref.mapping_fingerprint,
    )
  )
    throw invalidLinkReceipt();
  if (coreAtomicIPType(candidate.candidateValue) === null)
    throw invalidLinkReceipt();
  return freezeLinkValue(structuredClone(result));
}

function invalidLinkReceipt(): IndicatorLinkWriteError {
  return new IndicatorLinkWriteError("uncertain", {
    kind: "read_failed",
    field: null,
    message:
      "The link response could not be verified. The binding may exist. Recover this exact request before starting another link.",
  });
}

function freezeLinkValue<T>(value: T): T {
  if (value !== null && typeof value === "object") {
    for (const child of Object.values(value)) freezeLinkValue(child);
    Object.freeze(value);
  }
  return value;
}
