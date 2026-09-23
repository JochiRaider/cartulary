import {
  buildHTTPOperationPath,
  type CreateManualIndicatorObservationResponse,
  httpOperationBindings,
} from "@cartulary/protocol-ts/http";
import { apiPath, fetchHTTPOperation } from "../../services/browserApi";
import {
  observationTimeKey,
  observationUUID,
  observationVersion,
} from "../features/indicators/observationModel";
import type {
  ObservationAttempt,
  ObservationReceipt,
  ObservationTransportPort,
} from "../features/indicators/observationOperation";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";

const operations = {
  create: "createManualIndicatorObservation",
  resolve: "resolveIndicatorObservation",
  dismiss: "dismissIndicatorObservation",
  restore: "restoreIndicatorObservation",
} as const;
function parameters(attempt: Pick<ObservationAttempt, "intent">) {
  return attempt.intent.action === "create"
    ? { source_record_id: attempt.intent.source.recordId }
    : { observation_id: attempt.intent.observation.observation_id };
}
export function createObservationTransport(options: {
  apiBase: string | undefined;
  incidentId: string;
}): ObservationTransportPort {
  return {
    capture(authority, generation, intent, id) {
      if (authority.incidentId !== options.incidentId)
        throw new Error("Observation incident changed");
      const operation = operations[intent.action];
      const body =
        intent.action === "create"
          ? {
              client_txn_id: id,
              base_row_version: intent.source.rowVersion,
              source_field_key: intent.source.fieldKey,
              span_start_byte: intent.selection.startByte,
              span_end_byte: intent.selection.endByte,
              ...(intent.parsedType === undefined
                ? {}
                : { parsed_indicator_type: intent.parsedType }),
              ...(intent.targetId === undefined
                ? {}
                : { resolved_indicator_record_id: intent.targetId }),
            }
          : {
              client_txn_id: id,
              base_row_version: intent.observation.row_version,
              ...(intent.action === "resolve"
                ? { resolved_indicator_record_id: intent.targetId }
                : {}),
            };
      return {
        id,
        authority: structuredClone(authority),
        generation,
        intent: structuredClone(intent),
        operation,
        apiBase: options.apiBase,
        path: apiPath(
          options.apiBase,
          buildHTTPOperationPath(operation, parameters({ intent })),
        ),
        body: JSON.stringify(body),
      };
    },
    async send(attempt, signal) {
      const operationID = operations[attempt.intent.action];
      if (
        attempt.operation !== operationID ||
        attempt.authority.incidentId !== options.incidentId ||
        attempt.path !==
          apiPath(
            attempt.apiBase,
            buildHTTPOperationPath(operationID, parameters(attempt)),
          )
      )
        return { kind: "uncertain" };
      let status: number | null = null;
      try {
        const response =
          await fetchHTTPOperation<CreateManualIndicatorObservationResponse>({
            apiBase: attempt.apiBase,
            operationID,
            pathParameters: parameters(attempt),
            init: {
              method: httpOperationBindings[operationID].method,
              body: attempt.body,
              signal,
            },
            onResponse: (value) => {
              status = value.status;
            },
          });
        if (!response.ok) {
          if (
            status === null ||
            status < 400 ||
            status >= 500 ||
            !response.payload.error?.code
          )
            return { kind: "uncertain" };
          const failure = classifyWorkbookOperationFailure(
            response.status,
            response.payload,
            operationID,
          );
          return failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : { kind: "rejected", failure };
        }
        const receipt = validateObservationReceipt(
          attempt,
          response.payload.data,
          response.status,
        );
        return receipt
          ? { kind: "acknowledged", receipt }
          : { kind: "uncertain" };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}

/** The generated decoder checks shape; this checks operation meaning and binding. */
function validateObservationReceipt(
  attempt: ObservationAttempt,
  data: ObservationReceipt,
  status: number,
): ObservationReceipt | null {
  const item = data.observation,
    intent = attempt.intent;
  if (
    !item ||
    !observationUUID(data.change_set_id) ||
    typeof data.replayed !== "boolean" ||
    !Array.isArray(data.affected_records) ||
    !observationUUID(item.observation_id) ||
    !observationVersion(item.row_version) ||
    item.incident_id !== attempt.authority.incidentId ||
    !observationUUID(item.created_by_user_id) ||
    !observationTimeKey(item.created_at) ||
    !item.origin_locator ||
    !item.source_field_key ||
    typeof item.observed_text !== "string"
  )
    return null;
  if (
    intent.action === "create"
      ? data.replayed
        ? status !== 200
        : status !== 201
      : status !== 200
  )
    return null;
  const sourceId =
    intent.action === "create"
      ? intent.source.recordId
      : intent.observation.source_record_id;
  const target =
    intent.action === "create"
      ? (intent.targetId ?? null)
      : intent.action === "resolve"
        ? intent.targetId
        : null;
  const state =
    intent.action === "dismiss"
      ? "dismissed"
      : target
        ? "resolved"
        : "unresolved";
  if (
    item.source_record_id !== sourceId ||
    item.resolution_status !== state ||
    item.resolved_indicator_record_id !== target
  )
    return null;
  if (state === "unresolved") {
    if (
      item.resolved_by_user_id !== null ||
      item.resolved_at !== null ||
      item.resolution_method !== null
    )
      return null;
  } else if (
    item.resolved_by_user_id !== attempt.authority.actorId ||
    !item.resolved_at ||
    !observationTimeKey(item.resolved_at) ||
    !item.resolution_method
  )
    return null;
  if (intent.action === "create") {
    if (
      item.row_version !== 1 ||
      item.source_field_key !== intent.source.fieldKey ||
      item.observed_text !== intent.selection.text ||
      item.origin_kind !== "manual_entry" ||
      item.created_by_user_id !== attempt.authority.actorId ||
      (intent.parsedType !== undefined &&
        item.parsed_indicator_type !== intent.parsedType)
    )
      return null;
  } else {
    const previous = intent.observation;
    if (
      item.observation_id !== previous.observation_id ||
      item.row_version !== previous.row_version + 1
    )
      return null;
    for (const key of [
      "source_field_key",
      "origin_kind",
      "origin_locator",
      "observed_text",
      "parsed_indicator_type",
      "normalized_candidate",
      "created_by_user_id",
    ] as const)
      if (item[key] !== previous[key]) return null;
    if (
      observationTimeKey(item.created_at) !==
      observationTimeKey(previous.created_at)
    )
      return null;
  }
  const expected = new Set([sourceId]);
  if (target) expected.add(target);
  if (
    intent.action !== "create" &&
    intent.observation.resolved_indicator_record_id
  )
    expected.add(intent.observation.resolved_indicator_record_id);
  let previousId = "";
  for (const record of data.affected_records) {
    if (
      !observationUUID(record.record_id) ||
      record.record_id <= previousId ||
      !observationVersion(record.row_version) ||
      !expected.delete(record.record_id)
    )
      return null;
    if (
      intent.action === "create" &&
      record.record_id === sourceId &&
      record.row_version !== intent.source.rowVersion + 1
    )
      return null;
    previousId = record.record_id;
  }
  return expected.size === 0 ? data : null;
}
