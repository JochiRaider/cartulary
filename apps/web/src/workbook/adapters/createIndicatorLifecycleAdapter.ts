import {
  type AppendIndicatorStateIntervalRequest,
  type AppendIndicatorStateIntervalResponse,
  buildHTTPOperationPath,
  httpOperationBindings,
} from "@cartulary/protocol-ts/http";
import { apiPath, fetchHTTPOperation } from "../../services/browserApi";
import {
  canonicalLifecycleUUID,
  lifecycleUTCTimestamp,
} from "../features/indicators/indicatorLifecycleModel";
import type {
  IndicatorLifecycleTransportPort,
  LifecycleAttempt,
} from "../features/indicators/indicatorLifecycleOperation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { createIndicatorLifecycleReader } from "./createIndicatorLifecycleReader";
import type { IndicatorLifecycleReceipt } from "./indicatorLifecycleProtocol";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";

const operationID = "appendIndicatorStateInterval";
const route = (base: string | undefined, recordId: string) =>
  apiPath(
    base,
    buildHTTPOperationPath(operationID, { indicator_id: recordId }),
  );
export function createIndicatorLifecycleAdapter(options: {
  apiBase: string | undefined;
  incidentId: string;
  readScope?: WorkbookReadScopeSource;
}): IndicatorLifecycleTransportPort {
  return {
    ...createIndicatorLifecycleReader(options),
    capture(authority, generation, draft, values, id) {
      if (authority.incidentId !== options.incidentId)
        throw new Error("Indicator incident changed");
      return {
        id,
        authority: structuredClone(authority),
        generation,
        draft: structuredClone(draft),
        values: structuredClone(values),
        apiBase: options.apiBase,
        path: route(options.apiBase, draft.recordId),
        body: JSON.stringify({
          client_txn_id: id,
          base_row_version: draft.baseRowVersion,
          ...values,
        } satisfies AppendIndicatorStateIntervalRequest),
      };
    },
    async send(attempt, signal) {
      if (
        attempt.authority.incidentId !== options.incidentId ||
        attempt.path !== route(attempt.apiBase, attempt.draft.recordId)
      )
        return { kind: "uncertain" };
      let status: number | null = null;
      try {
        const response =
          await fetchHTTPOperation<AppendIndicatorStateIntervalResponse>({
            apiBase: attempt.apiBase,
            operationID,
            pathParameters: { indicator_id: attempt.draft.recordId },
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
        const receipt = validatedLifecycleReceipt(
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

function storedTime(time: string | null): string | null {
  if (time === null) return null;
  return time.replace(/\.(\d{1,6})\d*Z$/u, (_match, fraction: string) => {
    const value = fraction.replace(/0+$/u, "");
    return value ? `.${value}Z` : "Z";
  });
}
function validatedLifecycleReceipt(
  attempt: LifecycleAttempt,
  data: IndicatorLifecycleReceipt,
  status: number,
): IndicatorLifecycleReceipt | null {
  const item = data.interval,
    values = attempt.values;
  if (
    !item ||
    !canonicalLifecycleUUID(data.change_set_id) ||
    !canonicalLifecycleUUID(item.interval_id) ||
    item.incident_id !== attempt.authority.incidentId ||
    item.indicator_record_id !== attempt.draft.recordId ||
    item.created_by_user_id !== attempt.authority.actorId ||
    !Number.isSafeInteger(item.row_version) ||
    item.row_version < 1 ||
    lifecycleUTCTimestamp(item.created_at) !== item.created_at ||
    lifecycleUTCTimestamp(item.assessed_at) !== item.assessed_at ||
    item.lifecycle_state !== values.lifecycle_state ||
    item.valid_from !== storedTime(values.valid_from) ||
    item.valid_to !== storedTime(values.valid_to) ||
    item.confidence !== values.confidence ||
    item.rationale !== values.rationale ||
    item.assessor !== values.assessor ||
    item.support_refs.length !== values.support_refs.length ||
    new Set(item.support_refs).size !== item.support_refs.length ||
    item.support_refs.some(
      (id) => !canonicalLifecycleUUID(id) || !values.support_refs.includes(id),
    ) ||
    (status === 201
      ? data.replayed !== false
      : status !== 200 || data.replayed !== true)
  )
    return null;
  let previous = "";
  let indicatorFound = false;
  for (const record of data.affected_records) {
    if (
      !canonicalLifecycleUUID(record.record_id) ||
      record.record_id <= previous ||
      !Number.isSafeInteger(record.row_version) ||
      record.row_version < 1
    )
      return null;
    previous = record.record_id;
    if (record.record_id === attempt.draft.recordId) {
      if (record.row_version <= attempt.draft.baseRowVersion) return null;
      indicatorFound = true;
    }
  }
  return indicatorFound ? data : null;
}
