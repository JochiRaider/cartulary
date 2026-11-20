import {
  buildHTTPOperationPath,
  type ErrorEnvelope,
  encodeHTTPOperationQuery,
  errorEnvelopeDecoder,
  type HTTPOperationErrorCode,
  type HTTPOperationErrorStatus,
  type HTTPOperationID,
  type HTTPOperationResponse,
  type HTTPQueryValue,
  httpOperationBindings,
  validateHTTPOperationResponse,
} from "@cartulary/protocol-ts/http";
import {
  type ClientErrorCode,
  type PublicAPIError,
  publicErrorView,
} from "../shared/publicError";
import { createClientTransactionId } from "./clientTransactionId";
import {
  csrfHeaderName,
  readCookie,
  requestJSON,
  requestMultipartJSON,
} from "./httpTransport";

/**
 * A retained public error. `Code` narrows the codes feature code may branch
 * on; decoded operation failures supply `HTTPOperationFailureCode<Op>` and
 * client-originated failures supply `ClientErrorCode` members.
 */
export type APIError<Code extends string = string> = {
  code: Code;
  details?: Record<string, unknown>;
  message?: string;
  request_id?: string;
  retryable?: boolean;
  status?: number;
} & PublicAPIError;

export type APIResult<T> = {
  ok: boolean;
  status: number;
  payload: T | { error?: APIError };
};

type PublicErrorObject = ErrorEnvelope["error"];

/**
 * A public error decoded against its operation's contract (Core 03
 * REQ-03-086a): the envelope is valid, `status` equals the HTTP status, and
 * `code` is a code the operation declares for that status. Every operation
 * declares `500` `internal_error` and `503` `service_unavailable` (Core 01
 * REQ-01-702), so neither needs a rule of its own.
 */
export type HTTPOperationPublicError<
  OperationID extends HTTPOperationID = HTTPOperationID,
> = Readonly<Omit<PublicErrorObject, "code" | "status">> &
  {
    readonly [Status in HTTPOperationErrorStatus<OperationID>]: {
      readonly code: HTTPOperationErrorCode<OperationID, Status>;
      readonly status: Status;
    };
  }[HTTPOperationErrorStatus<OperationID>];

/**
 * The client contract failure for a response that fails its operation
 * contract. Feature code must not derive behaviour from the rejected payload.
 */
export type HTTPOperationContractFailure = {
  readonly code: Extract<ClientErrorCode, "invalid_public_contract_response">;
  readonly details: Readonly<Record<string, unknown>> & {
    readonly instance_path: string;
    readonly operation_id: HTTPOperationID;
    readonly response_status: number;
    readonly schema_id: string;
  };
  readonly message: string;
  /** A client-originated failure carries no server correlation. */
  readonly request_id?: never;
  readonly retryable: true;
  readonly status: 502;
};

export type HTTPOperationError<
  OperationID extends HTTPOperationID = HTTPOperationID,
> = HTTPOperationPublicError<OperationID> | HTTPOperationContractFailure;

/**
 * Every failure code feature code may branch on for the selected operations:
 * their declared codes and the client contract failure (Core 03 REQ-03-086a).
 */
export type HTTPOperationFailureCode<
  OperationID extends HTTPOperationID = HTTPOperationID,
> = HTTPOperationError<OperationID>["code"];

export type HTTPOperationResult<
  T,
  OperationID extends HTTPOperationID = HTTPOperationID,
> =
  | { readonly ok: true; readonly status: number; readonly payload: T }
  | {
      readonly ok: false;
      readonly status: number;
      readonly payload: { readonly error: HTTPOperationError<OperationID> };
    };

type HTTPOperationRequestOptions<OperationID extends HTTPOperationID> = {
  apiBase?: string | undefined;
  operationID: OperationID;
  pathParameters?: Readonly<Record<string, string | number>> | undefined;
  query?: Readonly<Record<string, HTTPQueryValue>> | undefined;
  onResponse?: ((response: Response) => void) | undefined;
};

type HTTPOperationFetchOptions<OperationID extends HTTPOperationID> =
  HTTPOperationRequestOptions<OperationID> & {
    init?: RequestInit | undefined;
    onJSONParsed?: (() => void) | undefined;
  };

type HTTPOperationMultipartOptions<OperationID extends HTTPOperationID> =
  HTTPOperationRequestOptions<OperationID> & {
    body: FormData;
    init?: Omit<RequestInit, "body" | "credentials" | "headers"> | undefined;
  };

/** Operation identities and declared codes for feature layers that may not import the protocol package. */
export type { HTTPOperationErrorCode, HTTPOperationID };
export { csrfHeaderName, publicErrorView, readCookie };

export function apiPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    return path;
  }
  return `${trimmedBase.replace(/\/$/, "")}${path}`;
}

export async function fetchJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<APIResult<T>> {
  return requestJSON<T>(input, init, {
    contentType: "mutations",
    responseParsing: "content-aware",
  }) as Promise<APIResult<T>>;
}

/**
 * Sends one generated operation and decodes its response against the
 * operation contract (Core 03 REQ-03-086a). Without type arguments the
 * operation ID is inferred, typing the success payload and the declared error
 * codes; an explicit payload type alone types error codes over every bound
 * operation.
 */
export function fetchHTTPOperation<OperationID extends HTTPOperationID>(
  options: HTTPOperationFetchOptions<OperationID>,
): Promise<
  HTTPOperationResult<HTTPOperationResponse<OperationID>, OperationID>
>;
export function fetchHTTPOperation<
  T,
  OperationID extends HTTPOperationID = HTTPOperationID,
>(
  options: HTTPOperationFetchOptions<OperationID>,
): Promise<HTTPOperationResult<T, OperationID>>;
export async function fetchHTTPOperation(
  options: HTTPOperationFetchOptions<HTTPOperationID>,
): Promise<HTTPOperationResult<unknown>> {
  const path =
    buildHTTPOperationPath(options.operationID, options.pathParameters) +
    encodeHTTPOperationQuery(options.operationID, options.query);
  const result = (await requestJSON<unknown>(
    apiPath(options.apiBase, path),
    options.init,
    {
      contentType: "mutations",
      responseParsing: "content-aware",
      ...(options.onJSONParsed === undefined
        ? {}
        : { onJSONParsed: options.onJSONParsed }),
      ...(options.onResponse === undefined
        ? {}
        : { onResponse: options.onResponse }),
    },
  )) as APIResult<unknown>;
  return validateHTTPOperationResult(options.operationID, result);
}

export function fetchMultipartHTTPOperation<
  OperationID extends HTTPOperationID,
>(
  options: HTTPOperationMultipartOptions<OperationID>,
): Promise<
  HTTPOperationResult<HTTPOperationResponse<OperationID>, OperationID>
>;
export function fetchMultipartHTTPOperation<
  T,
  OperationID extends HTTPOperationID = HTTPOperationID,
>(
  options: HTTPOperationMultipartOptions<OperationID>,
): Promise<HTTPOperationResult<T, OperationID>>;
export async function fetchMultipartHTTPOperation(
  options: HTTPOperationMultipartOptions<HTTPOperationID>,
): Promise<HTTPOperationResult<unknown>> {
  const path =
    buildHTTPOperationPath(options.operationID, options.pathParameters) +
    encodeHTTPOperationQuery(options.operationID, options.query);
  const result = (await requestMultipartJSON<unknown>(
    apiPath(options.apiBase, path),
    options.body,
    options.init,
    options.onResponse === undefined ? {} : { onResponse: options.onResponse },
  )) as APIResult<unknown>;
  return validateHTTPOperationResult(options.operationID, result);
}

type HTTPOperationErrorDecoding<OperationID extends HTTPOperationID> =
  | { readonly ok: true; readonly error: HTTPOperationPublicError<OperationID> }
  | { readonly ok: false; readonly instancePath: string };

/** Core 03 REQ-03-086a: decode one public error against its operation. */
function decodeHTTPOperationError<OperationID extends HTTPOperationID>(
  operationID: OperationID,
  status: number,
  payload: unknown,
): HTTPOperationErrorDecoding<OperationID> {
  const decoded = errorEnvelopeDecoder.decode(payload);
  if (!decoded.ok) {
    return { ok: false, instancePath: decoded.error.instancePath };
  }
  const error = decoded.value.error;
  if (status < 400 || error.status !== status) {
    return { ok: false, instancePath: "/error/status" };
  }
  const declaredCodes: Readonly<Record<string, readonly string[]>> =
    httpOperationBindings[operationID as HTTPOperationID].error_codes_by_status;
  if (!(declaredCodes[String(status)] ?? []).includes(error.code)) {
    return { ok: false, instancePath: "/error/code" };
  }
  return { ok: true, error: error as HTTPOperationPublicError<OperationID> };
}

function contractFailure(
  operationID: HTTPOperationID,
  responseStatus: number,
  failure: { readonly instancePath: string; readonly schemaId: string },
): {
  readonly ok: false;
  readonly status: 502;
  readonly payload: { readonly error: HTTPOperationContractFailure };
} {
  return {
    ok: false,
    status: 502,
    payload: {
      error: {
        code: "invalid_public_contract_response",
        details: {
          instance_path: failure.instancePath,
          operation_id: operationID,
          response_status: responseStatus,
          schema_id: failure.schemaId,
        },
        message: "The server returned an invalid public contract response.",
        retryable: true,
        status: 502,
      },
    },
  };
}

function validateHTTPOperationResult<T, OperationID extends HTTPOperationID>(
  operationID: OperationID,
  result: APIResult<T>,
): HTTPOperationResult<T, OperationID> {
  if (!result.ok) {
    const decoded = decodeHTTPOperationError(
      operationID,
      result.status,
      result.payload,
    );
    if (decoded.ok) {
      return {
        ok: false,
        status: result.status,
        payload: { error: decoded.error },
      };
    }
    return contractFailure(operationID, result.status, {
      instancePath: decoded.instancePath,
      schemaId: errorEnvelopeDecoder.schemaId,
    });
  }
  const validation = validateHTTPOperationResponse(
    operationID,
    result.payload,
    result.status,
  );
  if (validation.ok) {
    return {
      ok: true,
      status: result.status,
      payload: result.payload as T,
    };
  }
  return contractFailure(operationID, result.status, validation);
}

export function extractError(payload: unknown): APIError | null {
  if (!payload || typeof payload !== "object") {
    return null;
  }
  return ((payload as { error?: APIError }).error ?? null) as APIError | null;
}

export function clientTxnID(prefix: string): string {
  return createClientTransactionId(prefix);
}
