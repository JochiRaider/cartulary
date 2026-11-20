import type { HTTPOperationID } from "@cartulary/protocol-ts/http";
import { expect, vi } from "vitest";
import type {
  HTTPOperationContractFailure,
  HTTPOperationError,
  HTTPOperationPublicError,
} from "../services/browserApi";

type FetchMock = {
  mock: {
    calls: ReadonlyArray<ReadonlyArray<unknown>>;
  };
};

type FetchCall = readonly [
  input: RequestInfo | URL,
  init?: RequestInit | undefined,
];

type JSONBody = Record<string, unknown>;

function asFetchCall(
  call: ReadonlyArray<unknown>,
  description: string,
): FetchCall {
  if (call.length === 0) {
    throw new Error(
      `expected ${description} to include a fetch input argument`,
    );
  }
  return [call[0] as RequestInfo | URL, call[1] as RequestInit | undefined];
}

function parseJSONBody<TBody extends JSONBody>(
  body: BodyInit | null | undefined,
  description: string,
): TBody {
  try {
    return JSON.parse(String(body ?? "{}")) as TBody;
  } catch (error) {
    throw new Error(
      `failed to parse JSON body for ${description}: ${
        error instanceof Error ? error.message : String(error)
      }`,
    );
  }
}

export function findFetchCalls(
  fetchMock: FetchMock,
  url: string,
  method: string,
): FetchCall[] {
  const expectedMethod = method.toUpperCase();
  return fetchMock.mock.calls.flatMap((call, index) => {
    const nextCall = asFetchCall(call, `fetch call #${index}`);
    const [input, init] = nextCall;
    const requestMethod = (init?.method ?? "GET").toUpperCase();
    if (String(input) !== url || requestMethod !== expectedMethod) {
      return [];
    }
    return [nextCall];
  });
}

export function findFetchCallsByPath(
  fetchMock: FetchMock,
  path: string,
  method: string,
): FetchCall[] {
  const expectedMethod = method.toUpperCase();
  return fetchMock.mock.calls.flatMap((call, index) => {
    const nextCall = asFetchCall(call, `fetch call #${index}`);
    const [input, init] = nextCall;
    const requestMethod = (init?.method ?? "GET").toUpperCase();
    const requestPath = new URL(String(input), "http://cartulary.test")
      .pathname;
    if (requestPath !== path || requestMethod !== expectedMethod) {
      return [];
    }
    return [nextCall];
  });
}

export function requireFetchCall(
  fetchMock: FetchMock,
  index: number,
  description = `fetch call #${index}`,
): FetchCall {
  const call = fetchMock.mock.calls[index];
  if (!call) {
    throw new Error(`missing ${description}`);
  }
  return asFetchCall(call, description);
}

export function requireJSONRequest<TBody extends JSONBody = JSONBody>(
  fetchMock: FetchMock,
  url: string,
  method: string,
  index = 0,
): { body: TBody; init: RequestInit | undefined } {
  const calls = findFetchCalls(fetchMock, url, method);
  const call = calls[index];
  if (!call) {
    throw new Error(
      `missing ${method.toUpperCase()} ${url} request at index ${index}; found ${calls.length} matching call(s)`,
    );
  }
  const [, init] = call;
  return {
    body: parseJSONBody<TBody>(
      init?.body,
      `${method.toUpperCase()} ${url} request at index ${index}`,
    ),
    init,
  };
}

export function requireJSONBodyAt<TBody extends JSONBody = JSONBody>(
  fetchMock: FetchMock,
  index: number,
  description = `fetch call #${index}`,
): TBody {
  const [, init] = requireFetchCall(fetchMock, index, description);
  return parseJSONBody<TBody>(init?.body, description);
}

export function readHeader(init: RequestInit | undefined, name: string) {
  const headers = init?.headers;
  if (headers instanceof Headers) {
    return headers.get(name) ?? "";
  }
  if (Array.isArray(headers)) {
    const match = headers.find(
      ([candidate]) => candidate.toLowerCase() === name.toLowerCase(),
    );
    return match?.[1] ?? "";
  }
  if (headers && typeof headers === "object") {
    const record = headers as Record<string, string | undefined>;
    for (const [key, value] of Object.entries(record)) {
      if (
        key.toLowerCase() === name.toLowerCase() &&
        typeof value === "string"
      ) {
        return value;
      }
    }
  }
  return "";
}

export function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}

export function errorResponse(
  code: string,
  status: number,
  details?: Record<string, unknown>,
) {
  return jsonResponse(
    {
      error: {
        code,
        details: details ?? {},
        message: code,
        request_id: "request-test",
        retryable: false,
        status,
      },
    },
    status,
  );
}

type OperationErrorIdentity<OperationID extends HTTPOperationID> =
  HTTPOperationPublicError<OperationID> extends infer Error
    ? Error extends HTTPOperationPublicError<OperationID>
      ? Pick<Error, "code" | "status">
      : never
    : never;

/**
 * A failed operation result as `fetchHTTPOperation` surfaces a decoded public
 * error. The operation is inferred from the receiving result type, so the
 * code must be one that operation declares for the status.
 */
export function operationRejection<
  OperationID extends HTTPOperationID = HTTPOperationID,
>(
  identity: OperationErrorIdentity<NoInfer<OperationID>>,
  details: Record<string, unknown> = {},
): {
  readonly ok: false;
  readonly status: number;
  readonly payload: { readonly error: HTTPOperationError<OperationID> };
} {
  const error = {
    ...identity,
    details,
    message: identity.code,
    request_id: "request-test",
    retryable: false,
  } as HTTPOperationPublicError<OperationID>;
  return { ok: false as const, status: identity.status, payload: { error } };
}

/**
 * A well-formed error response for one operation. The code must be one the
 * operation declares for the status. `internal_error` carries no details and
 * is not retryable; `service_unavailable` is retryable (Core 01 §3.3.6.1).
 */
export function operationErrorResponse<OperationID extends HTTPOperationID>(
  _operationID: OperationID,
  identity: OperationErrorIdentity<NoInfer<OperationID>>,
  fixture: {
    readonly details?: Record<string, unknown>;
    readonly message?: string;
    readonly retryable?: boolean;
  } = {},
): Response {
  const internal = identity.code === "internal_error";
  return jsonResponse(
    {
      error: {
        code: identity.code,
        status: identity.status,
        message: internal
          ? "internal_error"
          : (fixture.message ?? identity.code),
        request_id: "request-test",
        retryable: internal
          ? false
          : (fixture.retryable ?? identity.code === "service_unavailable"),
        details: internal ? {} : (fixture.details ?? {}),
      },
    },
    identity.status,
  );
}

/** A failed operation result as `fetchHTTPOperation` surfaces a client contract failure. */
export function operationContractFailure(
  operationID: HTTPOperationID,
  responseStatus = 200,
) {
  const error: HTTPOperationContractFailure = {
    code: "invalid_public_contract_response",
    details: {
      instance_path: "",
      operation_id: operationID,
      response_status: responseStatus,
      schema_id: `test.${operationID}`,
    },
    message: "The server returned an invalid public contract response.",
    retryable: true,
    status: 502,
  };
  return { ok: false as const, status: 502 as const, payload: { error } };
}

export function deferred<T>() {
  let resolve: (value: T) => void = () => {};
  let reject: (reason?: unknown) => void = () => {};
  const promise = new Promise<T>((nextResolve, nextReject) => {
    resolve = nextResolve;
    reject = nextReject;
  });
  return { promise, reject, resolve };
}

export function abortablePendingResponse(
  signal: AbortSignal | undefined,
  onAbort?: (signal: AbortSignal) => void,
) {
  return new Promise<Response>((_, reject) => {
    const abort = () => {
      if (signal) {
        onAbort?.(signal);
      }
      reject(new DOMException("Aborted", "AbortError"));
    };
    if (signal?.aborted) {
      abort();
      return;
    }
    signal?.addEventListener("abort", abort, { once: true });
  });
}

export async function expectStableFetchCount(
  fetchMock: FetchMock,
  expectedCount: number,
) {
  await vi.waitFor(() => {
    expect(fetchMock).toHaveBeenCalledTimes(expectedCount);
  });
  await flushMicrotasks();
  expect(fetchMock).toHaveBeenCalledTimes(expectedCount);
}

async function flushMicrotasks() {
  await Promise.resolve();
  await Promise.resolve();
}
