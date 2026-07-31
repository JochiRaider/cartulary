import {
  buildHTTPOperationPath,
  encodeHTTPOperationQuery,
  type HTTPOperationID,
  type HTTPQueryValue,
  validateHTTPOperationResponse,
} from "@cartulary/protocol-ts";
import { type PublicAPIError, publicErrorView } from "../shared/publicError";
import { createClientTransactionId } from "./clientTransactionId";
import {
  csrfCookieName,
  csrfHeaderName,
  readCookie,
  requestJSON,
  requestMultipartJSON,
} from "./httpTransport";

export type APIError = {
  code: string;
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

export type HTTPOperationResult<T> =
  | { readonly ok: true; readonly status: number; readonly payload: T }
  | {
      readonly ok: false;
      readonly status: number;
      readonly payload: { readonly error?: APIError };
    };

export { csrfCookieName, csrfHeaderName, publicErrorView, readCookie };

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

export async function fetchHTTPOperation<T>(options: {
  apiBase?: string | undefined;
  init?: RequestInit | undefined;
  onJSONParsed?: (() => void) | undefined;
  onResponse?: ((response: Response) => void) | undefined;
  operationID: HTTPOperationID;
  pathParameters?: Readonly<Record<string, string | number>> | undefined;
  query?: Readonly<Record<string, HTTPQueryValue>> | undefined;
}): Promise<HTTPOperationResult<T>> {
  const path =
    buildHTTPOperationPath(options.operationID, options.pathParameters) +
    encodeHTTPOperationQuery(options.operationID, options.query);
  const result = (await requestJSON<T>(
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
  )) as APIResult<T>;
  return validateHTTPOperationResult(options.operationID, result);
}

export async function fetchMultipartHTTPOperation<T>(options: {
  apiBase?: string | undefined;
  body: FormData;
  init?: Omit<RequestInit, "body" | "credentials" | "headers"> | undefined;
  operationID: HTTPOperationID;
  pathParameters?: Readonly<Record<string, string | number>> | undefined;
  query?: Readonly<Record<string, HTTPQueryValue>> | undefined;
}): Promise<HTTPOperationResult<T>> {
  const path =
    buildHTTPOperationPath(options.operationID, options.pathParameters) +
    encodeHTTPOperationQuery(options.operationID, options.query);
  const result = (await requestMultipartJSON<T>(
    apiPath(options.apiBase, path),
    options.body,
    options.init,
  )) as APIResult<T>;
  return validateHTTPOperationResult(options.operationID, result);
}

function validateHTTPOperationResult<T>(
  operationID: HTTPOperationID,
  result: APIResult<T>,
): HTTPOperationResult<T> {
  if (!result.ok) {
    return {
      ok: false,
      status: result.status,
      payload: result.payload as { error?: APIError },
    };
  }
  const validation = validateHTTPOperationResponse(operationID, result.payload);
  if (validation.ok) {
    return {
      ok: true,
      status: result.status,
      payload: result.payload as T,
    };
  }
  return {
    ok: false,
    status: 502,
    payload: {
      error: {
        code: "invalid_public_contract_response",
        details: {
          instance_path: validation.instancePath,
          operation_id: operationID,
          schema_id: validation.schemaId,
        },
        message: "The server returned an invalid public contract response.",
        retryable: true,
        status: 502,
      },
    },
  };
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
