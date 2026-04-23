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
