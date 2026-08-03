import { apiPath, extractError, fetchJSON } from "../../services/browserApi";
import type {
  IndicatorObservation,
  IndicatorStateInterval,
  IndicatorWorkflowPort,
} from "./workbookMutationCommandPorts";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "./workbookOperationOutcome";

type ObservationListEnvelope = {
  readonly data: { readonly observations: readonly IndicatorObservation[] };
};

type LifecycleListEnvelope = {
  readonly data: { readonly intervals: readonly IndicatorStateInterval[] };
};

type ObservationMutationEnvelope = {
  readonly data: { readonly observation: IndicatorObservation };
};

type LifecycleMutationEnvelope = {
  readonly data: { readonly interval: IndicatorStateInterval };
};

function rejected(status: number, payload: unknown): WorkbookOperationFailure {
  const error = extractError(payload);
  const message = error?.message?.trim() || error?.code || "Request failed.";
  if (status === 401) {
    return { kind: "authentication_required", message };
  }
  if (status === 403) {
    return { kind: "authorization_lost", message };
  }
  if (status === 404) {
    return { kind: "stale_target", message };
  }
  if (status === 400 || status === 422) {
    return { kind: "validation", message };
  }
  if (error?.code === "client_txn_conflict") {
    return { kind: "client_txn_conflict", message };
  }
  if (
    status === 409 &&
    (error?.code === "row_version_conflict" ||
      error?.code === "illegal_transition")
  ) {
    return { kind: "stale_target", message };
  }
  if (error?.retryable || status === 429 || status >= 500) {
    return { kind: "retryable", message };
  }
  return { kind: "terminal", message };
}

async function request<Envelope, Value>(input: {
  readonly url: string;
  readonly init?: RequestInit | undefined;
  readonly select: (envelope: Envelope) => Value | null;
}): Promise<WorkbookOperationOutcome<Value>> {
  try {
    const result = await fetchJSON<Envelope>(input.url, input.init);
    if (!result.ok) {
      return {
        kind: "rejected",
        failure: rejected(result.status, result.payload),
      };
    }
    const value = input.select(result.payload as Envelope);
    return value === null
      ? {
          kind: "rejected",
          failure: {
            kind: "invalid_contract",
            message:
              "The server returned an invalid Indicator workflow result.",
          },
        }
      : { kind: "accepted", value };
  } catch {
    return {
      kind: "rejected",
      failure: {
        kind: "retryable",
        message: "Indicator workflow unavailable.",
      },
    };
  }
}

function invalidIdentity<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "terminal",
      message: "A secure transaction ID could not be created.",
    },
  };
}

function isArrayEnvelope(value: unknown, member: string): boolean {
  if (!value || typeof value !== "object" || !("data" in value)) return false;
  const data = value.data;
  return Boolean(
    data &&
      typeof data === "object" &&
      member in data &&
      Array.isArray((data as Record<string, unknown>)[member]),
  );
}

function isResourceEnvelope(value: unknown, member: string): boolean {
  if (!value || typeof value !== "object" || !("data" in value)) return false;
  const data = value.data;
  return Boolean(
    data &&
      typeof data === "object" &&
      member in data &&
      (data as Record<string, unknown>)[member] !== null &&
      typeof (data as Record<string, unknown>)[member] === "object",
  );
}

export function createIndicatorWorkflowPort(options: {
  readonly apiBase: string | undefined;
  readonly createMutationID: (prefix: string) => string | null;
}): IndicatorWorkflowPort {
  const jsonMutation = (body: unknown): RequestInit => ({
    method: "POST",
    body: JSON.stringify(body),
  });
  return {
    listSourceObservations({ sourceRecordId }) {
      return request<ObservationListEnvelope, readonly IndicatorObservation[]>({
        url: apiPath(
          options.apiBase,
          `/api/v1/records/${encodeURIComponent(sourceRecordId)}/indicator-observations`,
        ),
        select: (envelope) =>
          isArrayEnvelope(envelope, "observations")
            ? envelope.data.observations
            : null,
      });
    },
    listObservations({ indicatorRecordId }) {
      return request<ObservationListEnvelope, readonly IndicatorObservation[]>({
        url: apiPath(
          options.apiBase,
          `/api/v1/indicators/${encodeURIComponent(indicatorRecordId)}/observations`,
        ),
        select: (envelope) =>
          isArrayEnvelope(envelope, "observations")
            ? envelope.data.observations
            : null,
      });
    },
    listStateIntervals({ indicatorRecordId }) {
      return request<LifecycleListEnvelope, readonly IndicatorStateInterval[]>({
        url: apiPath(
          options.apiBase,
          `/api/v1/indicators/${encodeURIComponent(indicatorRecordId)}/state-intervals`,
        ),
        select: (envelope) =>
          isArrayEnvelope(envelope, "intervals")
            ? envelope.data.intervals
            : null,
      });
    },
    createManualObservation(input) {
      const clientTxnId = options.createMutationID("indicator-observation");
      if (clientTxnId === null) return Promise.resolve(invalidIdentity());
      return request<ObservationMutationEnvelope, IndicatorObservation>({
        url: apiPath(
          options.apiBase,
          `/api/v1/records/${encodeURIComponent(input.sourceRecordId)}/indicator-observations`,
        ),
        init: jsonMutation({
          client_txn_id: clientTxnId,
          base_row_version: input.baseRowVersion,
          source_field_key: input.sourceFieldKey,
          span_start_byte: input.spanStartByte,
          span_end_byte: input.spanEndByte,
          ...(input.parsedIndicatorType
            ? { parsed_indicator_type: input.parsedIndicatorType }
            : {}),
          ...(input.resolvedIndicatorRecordId
            ? { resolved_indicator_record_id: input.resolvedIndicatorRecordId }
            : {}),
        }),
        select: (envelope) =>
          isResourceEnvelope(envelope, "observation")
            ? envelope.data.observation
            : null,
      });
    },
    transitionObservation(input) {
      const clientTxnId = options.createMutationID(
        `indicator-observation-${input.action}`,
      );
      if (clientTxnId === null) return Promise.resolve(invalidIdentity());
      return request<ObservationMutationEnvelope, IndicatorObservation>({
        url: apiPath(
          options.apiBase,
          `/api/v1/indicator-observations/${encodeURIComponent(input.observationId)}/${input.action}`,
        ),
        init: jsonMutation({
          client_txn_id: clientTxnId,
          base_row_version: input.baseRowVersion,
          ...(input.action === "resolve" && input.resolvedIndicatorRecordId
            ? {
                resolved_indicator_record_id: input.resolvedIndicatorRecordId,
              }
            : {}),
        }),
        select: (envelope) =>
          isResourceEnvelope(envelope, "observation")
            ? envelope.data.observation
            : null,
      });
    },
    appendStateInterval(input) {
      const clientTxnId = options.createMutationID("indicator-lifecycle");
      if (clientTxnId === null) return Promise.resolve(invalidIdentity());
      return request<LifecycleMutationEnvelope, IndicatorStateInterval>({
        url: apiPath(
          options.apiBase,
          `/api/v1/indicators/${encodeURIComponent(input.indicatorRecordId)}/state-intervals`,
        ),
        init: jsonMutation({
          client_txn_id: clientTxnId,
          base_row_version: input.baseRowVersion,
          lifecycle_state: input.lifecycleState,
          valid_from: input.validFrom,
          valid_to: input.validTo,
          confidence: input.confidence,
          rationale: input.rationale,
          support_refs: input.supportRefs,
          assessor: input.assessor,
        }),
        select: (envelope) =>
          isResourceEnvelope(envelope, "interval")
            ? envelope.data.interval
            : null,
      });
    },
  };
}
