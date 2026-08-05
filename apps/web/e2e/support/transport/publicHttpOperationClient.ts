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
  requestPublicJson,
} from "./publicJsonClient";

type BodyOption<OperationID extends HTTPOperationID> = [
  HTTPOperationRequest<OperationID>,
] extends [undefined]
  ? { readonly body?: never }
  : { readonly body: HTTPOperationRequest<OperationID> };

export type PublicHttpOperationOptions<OperationID extends HTTPOperationID> =
  BodyOption<OperationID> & {
    readonly headers?: Record<string, string>;
    readonly operationID: OperationID;
    readonly pathParameters?: Readonly<Record<string, string | number>>;
    readonly query?: Readonly<Record<string, HTTPQueryValue>>;
    readonly request: JsonRequestContextLike;
  };

export type PublicHttpOperationResult<OperationID extends HTTPOperationID> =
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

export type ObservedHttpOperationRequest = {
  readonly method: () => string;
  readonly postData: () => string | null;
};

export type ObservedHttpOperationResponse = {
  readonly ok: () => boolean;
  readonly request: () => ObservedHttpOperationRequest;
  readonly status: () => number;
  readonly text: () => Promise<string>;
  readonly url: () => string;
};

export async function publicHttpOperation<OperationID extends HTTPOperationID>(
  options: PublicHttpOperationOptions<OperationID>,
): Promise<PublicHttpOperationResult<OperationID>> {
  const binding = httpOperationBindings[options.operationID];
  const path =
    buildHTTPOperationPath(options.operationID, options.pathParameters) +
    encodeHTTPOperationQuery(options.operationID, options.query);
  const response = await requestPublicJson({
    ...("body" in options ? { body: options.body } : {}),
    ...(options.headers === undefined ? {} : { headers: options.headers }),
    method: binding.method,
    path,
    request: options.request,
  });
  if (!response.ok) {
    return {
      ok: false,
      payload: response.body,
      status: response.status,
    };
  }
  const validation = validateHTTPOperationResponse(
    options.operationID,
    response.body,
  );
  if (
    !binding.success_statuses.some(
      (expectedStatus) => expectedStatus === response.status,
    ) ||
    !validation.ok
  ) {
    return {
      ok: false,
      payload: {
        error: {
          code: "invalid_public_contract_response",
          details: {
            ...(validation.ok
              ? {}
              : {
                  instance_path: validation.instancePath,
                  schema_id: validation.schemaId,
                }),
            operation_id: options.operationID,
            received_status: response.status,
          },
          message: "The server returned an invalid public contract response.",
          retryable: true,
          status: 502,
        },
      },
      status: 502,
    };
  }
  return {
    ok: true,
    payload: response.body as HTTPOperationResponse<OperationID>,
    status: response.status,
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
  const responseBody = await response.text().catch((error: unknown) => {
    return `<<failed to read response body: ${String(error)}>>`;
  });
  if (!response.ok()) {
    throw new Error(
      observedOperationDiagnostic(
        operationID,
        response,
        request,
        responseBody,
        `HTTP ${response.status()}`,
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
        request,
        responseBody,
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
        request,
        responseBody,
        [
          "invalid_public_contract_response",
          ...(methodMatches ? [] : [`expected_method=${binding.method}`]),
          ...(statusMatches
            ? []
            : [`expected_status=${binding.success_statuses.join(",")}`]),
          ...(validation.ok
            ? []
            : [
                `schema_id=${validation.schemaId}`,
                `instance_path=${validation.instancePath}`,
              ]),
        ].join("; "),
      ),
    );
  }
  return payload as HTTPOperationResponse<OperationID>;
}

function observedOperationDiagnostic(
  operationID: HTTPOperationID,
  response: ObservedHttpOperationResponse,
  request: ObservedHttpOperationRequest,
  responseBody: string,
  reason: string,
) {
  return [
    `public HTTP operation ${operationID} failed: ${reason}`,
    `method=${request.method()}`,
    `url=${response.url()}`,
    `request_body=${truncateDiagnostic(request.postData() ?? "")}`,
    `response_body=${truncateDiagnostic(responseBody)}`,
  ].join("\n");
}

function truncateDiagnostic(value: string) {
  const limit = 4000;
  if (value.length <= limit) {
    return value;
  }
  return `${value.slice(0, limit)}...<truncated ${value.length - limit} chars>`;
}
