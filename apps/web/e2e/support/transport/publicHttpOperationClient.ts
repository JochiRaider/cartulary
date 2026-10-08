import { errorRegistry } from "@cartulary/protocol-ts/errors";
import {
  buildHTTPOperationPath,
  encodeHTTPOperationQuery,
  type HTTPOperationID,
  type HTTPOperationRequest,
  type HTTPOperationResponse,
  type HTTPQueryValue,
  httpOperationBindings,
  validateHTTPOperationResponse,
} from "@cartulary/protocol-ts/http";

import {
  type JsonRequestContextLike,
  type RepeatedHeaderJsonResponse,
  requestPublicJsonObserved,
} from "./publicJsonClient";

type BodyOption<OperationID extends HTTPOperationID> = [
  HTTPOperationRequest<OperationID>,
] extends [undefined]
  ? { readonly body?: never }
  : { readonly body: HTTPOperationRequest<OperationID> };

type PublicHttpOperationOptions<OperationID extends HTTPOperationID> =
  BodyOption<OperationID> & {
    readonly headers?: Record<string, string>;
    readonly operationID: OperationID;
    readonly pathParameters?: Readonly<Record<string, string | number>>;
    readonly query?: Readonly<Record<string, HTTPQueryValue>>;
    readonly request: JsonRequestContextLike;
  };

type PublicHttpOperationResult<OperationID extends HTTPOperationID> =
  | {
      readonly ok: true;
      readonly payload: HTTPOperationResponse<OperationID>;
      readonly status: number;
    }
  | {
      readonly ok: false;
      readonly payload: unknown;
      readonly status: number;
    };

type PublicHttpOperationObservedResult<OperationID extends HTTPOperationID> =
  | (Extract<PublicHttpOperationResult<OperationID>, { readonly ok: true }> & {
      readonly response: RepeatedHeaderJsonResponse;
    })
  | Extract<PublicHttpOperationResult<OperationID>, { readonly ok: false }>;

type ObservedHttpOperationRequest = {
  readonly method: () => string;
};

type ObservedHttpOperationResponse = {
  readonly ok: () => boolean;
  readonly request: () => ObservedHttpOperationRequest;
  readonly status: () => number;
  readonly text: () => Promise<string>;
};

/** Diagnostics expose contract identity, never transport bodies or capabilities. */
export function publicHttpOperationFailure(
  operationID: HTTPOperationID,
  status: number,
  payload: unknown,
): string {
  const envelope =
    typeof payload === "object" && payload !== null
      ? (payload as { error?: unknown })
      : null;
  const error =
    typeof envelope?.error === "object" && envelope.error !== null
      ? (envelope.error as { code?: unknown })
      : null;
  const code =
    typeof error?.code === "string" &&
    errorRegistry.errors.some((entry) => entry.code === error.code)
      ? error.code
      : "unrecognized_error";
  return `public HTTP operation ${operationID} failed: HTTP ${status}; code=${code}`;
}

export async function publicHttpOperation<OperationID extends HTTPOperationID>(
  options: PublicHttpOperationOptions<OperationID>,
): Promise<PublicHttpOperationResult<OperationID>> {
  return (await executePublicHttpOperation(options)).result;
}

export async function publicHttpOperationObserved<
  OperationID extends HTTPOperationID,
>(
  options: PublicHttpOperationOptions<OperationID>,
): Promise<PublicHttpOperationObservedResult<OperationID>> {
  const execution = await executePublicHttpOperation(options);
  if (!execution.result.ok) {
    return execution.result;
  }
  if (typeof execution.response.headersArray !== "function") {
    return invalidPublicContractResult(
      options.operationID,
      execution.result.status,
      {
        observation_error: "repeated_headers_unavailable",
      },
    );
  }
  return {
    ...execution.result,
    response: execution.response as RepeatedHeaderJsonResponse,
  };
}

async function executePublicHttpOperation<OperationID extends HTTPOperationID>(
  options: PublicHttpOperationOptions<OperationID>,
) {
  const binding = httpOperationBindings[options.operationID];
  const path =
    buildHTTPOperationPath(options.operationID, options.pathParameters) +
    encodeHTTPOperationQuery(options.operationID, options.query);
  const observation = await requestPublicJsonObserved({
    ...("body" in options ? { body: options.body } : {}),
    ...(options.headers === undefined ? {} : { headers: options.headers }),
    method: binding.method,
    path,
    request: options.request,
  });
  const { response, result: publicJsonResult } = observation;
  if (!publicJsonResult.ok) {
    return {
      response,
      result: {
        ok: false,
        payload: publicJsonResult.body,
        status: publicJsonResult.status,
      } satisfies PublicHttpOperationResult<OperationID>,
    };
  }
  const validation = validateHTTPOperationResponse(
    options.operationID,
    publicJsonResult.body,
  );
  if (
    !binding.success_statuses.some(
      (expectedStatus) => expectedStatus === publicJsonResult.status,
    ) ||
    !validation.ok
  ) {
    return {
      response,
      result: invalidPublicContractResult(
        options.operationID,
        publicJsonResult.status,
        {
          ...(validation.ok
            ? {}
            : {
                instance_path: validation.instancePath,
                schema_id: validation.schemaId,
              }),
        },
      ),
    };
  }
  return {
    response,
    result: {
      ok: true,
      payload: publicJsonResult.body as HTTPOperationResponse<OperationID>,
      status: publicJsonResult.status,
    } satisfies PublicHttpOperationResult<OperationID>,
  };
}

function invalidPublicContractResult<OperationID extends HTTPOperationID>(
  operationID: OperationID,
  receivedStatus: number,
  details: Readonly<Record<string, unknown>>,
): Extract<PublicHttpOperationResult<OperationID>, { readonly ok: false }> {
  return {
    ok: false,
    payload: {
      error: {
        code: "invalid_public_contract_response",
        details: {
          ...details,
          operation_id: operationID,
          received_status: receivedStatus,
        },
        message: "The server returned an invalid public contract response.",
        retryable: true,
        status: 502,
      },
    },
    status: 502,
  };
}

export async function readHttpOperationResponse<
  OperationID extends HTTPOperationID,
>(
  response: ObservedHttpOperationResponse,
  operationID: OperationID,
): Promise<HTTPOperationResponse<OperationID>> {
  const binding = httpOperationBindings[operationID];
  const request = response.request();
  const responseBody = await response
    .text()
    .catch(() => "response body unavailable");
  if (!response.ok()) {
    let failurePayload: unknown;
    try {
      failurePayload = JSON.parse(responseBody);
    } catch {
      failurePayload = undefined;
    }
    throw new Error(
      publicHttpOperationFailure(
        operationID,
        response.status(),
        failurePayload,
      ),
    );
  }
  let payload: unknown;
  try {
    payload = JSON.parse(responseBody);
  } catch {
    throw new Error(
      observedOperationDiagnostic(
        operationID,
        response,
        "invalid_public_contract_response: response body is not JSON",
      ),
    );
  }
  const validation = validateHTTPOperationResponse(operationID, payload);
  const methodMatches = request.method() === binding.method;
  const statusMatches = binding.success_statuses.some(
    (status) => status === response.status(),
  );
  if (!methodMatches || !statusMatches || !validation.ok) {
    throw new Error(
      observedOperationDiagnostic(
        operationID,
        response,
        [
          "invalid_public_contract_response",
          ...(methodMatches ? [] : [`expected_method=${binding.method}`]),
          ...(statusMatches
            ? []
            : [`expected_status=${binding.success_statuses.join(",")}`]),
          ...(validation.ok ? [] : [`schema_id=${validation.schemaId}`]),
        ].join("; "),
      ),
    );
  }
  return payload as HTTPOperationResponse<OperationID>;
}

function observedOperationDiagnostic(
  operationID: HTTPOperationID,
  response: ObservedHttpOperationResponse,
  reason: string,
) {
  return `public HTTP operation ${operationID} failed: ${reason}; HTTP ${response.status()}`;
}
