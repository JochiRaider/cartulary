import {
  buildHTTPOperationPath,
  type CreateViewRowResponse,
  httpOperationBindings,
} from "@cartulary/protocol-ts/http";
import { apiPath, fetchHTTPOperation } from "../../services/browserApi";
import {
  indicatorCreateConstraints,
  indicatorCreateRequest,
  indicatorCreateTypes,
} from "../features/indicators/indicatorCreateModel";
import type {
  IndicatorCreateAttempt,
  IndicatorCreateReceipt,
  IndicatorCreateTransport,
} from "../features/indicators/indicatorCreateOperation";
import {
  observationIndicatorView,
  observationUUID,
  observationVersion,
} from "../features/indicators/observationModel";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { decodeCreateViewRowRequest } from "../models/workbookRequestDecoders";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";

const parameters = (attempt: Pick<IndicatorCreateAttempt, "authority">) => ({
  incident_id: attempt.authority.incidentId,
  view_schema_id: observationIndicatorView,
});
export function createIndicatorCreateTransport(options: {
  apiBase: string | undefined;
  incidentId: string;
}): IndicatorCreateTransport {
  return {
    capture(authority, generation, observation, contract, values, id) {
      if (
        authority.incidentId !== options.incidentId ||
        observation.incident_id !== options.incidentId
      )
        throw new Error("Indicator creation context changed");
      const request = indicatorCreateRequest(contract, values, id);
      if (!request) throw new Error("Invalid canonical proposal");
      const admitted = decodeCreateViewRowRequest(contract, request);
      if (!admitted)
        throw new Error("Canonical create request is not admitted");
      return {
        id,
        authority: structuredClone(authority),
        generation,
        observation: structuredClone(observation),
        contract,
        operation: "createViewRow",
        path: apiPath(
          options.apiBase,
          buildHTTPOperationPath("createViewRow", parameters({ authority })),
        ),
        apiBase: options.apiBase,
        body: JSON.stringify(admitted),
      };
    },
    async send(attempt, signal) {
      if (
        attempt.operation !== "createViewRow" ||
        attempt.authority.incidentId !== options.incidentId ||
        attempt.path !==
          apiPath(
            attempt.apiBase,
            buildHTTPOperationPath("createViewRow", parameters(attempt)),
          )
      )
        return { kind: "uncertain" };
      let status: number | null = null;
      try {
        const response = await fetchHTTPOperation<CreateViewRowResponse>({
          apiBase: attempt.apiBase,
          operationID: "createViewRow",
          pathParameters: parameters(attempt),
          init: {
            method: httpOperationBindings.createViewRow.method,
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
            "createViewRow",
          );
          return failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : { kind: "rejected", failure };
        }
        const receipt = validateIndicatorCreateReceipt(
          attempt,
          response.payload,
          response.status,
        );
        return receipt ? { kind: "accepted", receipt } : { kind: "uncertain" };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}
/** Retain the complete public result; status classifies the operation, not insertion. */
export function validateIndicatorCreateReceipt(
  attempt: Pick<IndicatorCreateAttempt, "contract" | "body">,
  response: CreateViewRowResponse,
  status: number,
): IndicatorCreateReceipt | null {
  try {
    const data = response.data;
    if (
      (status !== 200 && status !== 201) ||
      data.view_schema_id !== observationIndicatorView ||
      data.view_schema_id !== attempt.contract.viewSchemaId ||
      !observationUUID(data.change_set_id) ||
      !observationUUID(data.row.record_id) ||
      !observationVersion(data.row.row_version)
    )
      return null;
    const row = normalizeWorkbookViewRows(
      attempt.contract,
      [data.row],
      "Canonical Indicator create",
    )[0];
    if (!row) return null;
    const type = row.cells["indicator.indicator_type"]?.value,
      kind = row.cells["indicator.value_kind"]?.value,
      value = row.cells["indicator.display_value"]?.value;
    const submitted = JSON.parse(attempt.body) as Record<string, unknown>;
    if (
      type !== submitted["indicator.indicator_type"] ||
      kind !== submitted["indicator.value_kind"]
    )
      return null;
    if (
      typeof type !== "string" ||
      !(indicatorCreateTypes as readonly string[]).includes(type) ||
      typeof kind !== "string" ||
      !(indicatorCreateConstraints.valueKinds as readonly string[]).includes(
        kind,
      ) ||
      typeof value !== "string" ||
      !value.trim()
    )
      return null;
    return { status, response, row };
  } catch {
    return null;
  }
}
