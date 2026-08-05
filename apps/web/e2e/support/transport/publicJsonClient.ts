export type JsonRequestMethod = "DELETE" | "GET" | "PATCH" | "POST" | "PUT";

type JsonResponseLike = {
  headers: () => Record<string, string>;
  headersArray?: () => readonly JsonResponseHeader[];
  json: () => Promise<unknown>;
  ok: () => boolean;
  status: () => number;
};

type JsonResponseHeader = {
  readonly name: string;
  readonly value: string;
};

export type RepeatedHeaderJsonResponse = JsonResponseLike & {
  headersArray: () => readonly JsonResponseHeader[];
};

export type JsonRequestContextLike = {
  fetch: (
    path: string,
    options: {
      data?: unknown;
      headers?: Record<string, string>;
      method: JsonRequestMethod;
    },
  ) => Promise<JsonResponseLike>;
};

export type PublicJsonResult = {
  readonly body: unknown;
  readonly headers: Readonly<Record<string, string>>;
  readonly ok: boolean;
  readonly status: number;
};

type PublicJsonOptions = {
  readonly body?: unknown;
  readonly headers?: Record<string, string>;
  readonly method: JsonRequestMethod;
  readonly path: string;
  readonly request: JsonRequestContextLike;
};

export function atJsonOrigin(
  request: JsonRequestContextLike,
  origin: string,
): JsonRequestContextLike {
  const parsed = new URL(origin);
  if (
    (parsed.protocol !== "http:" && parsed.protocol !== "https:") ||
    parsed.pathname !== "/" ||
    parsed.search !== "" ||
    parsed.hash !== ""
  ) {
    throw new Error("JSON endpoint origin must be an absolute HTTP origin");
  }
  return {
    fetch: (path, options) =>
      request.fetch(`${parsed.origin}${requireSameOriginPath(path)}`, options),
  };
}

function requireSameOriginPath(path: string): string {
  if (!path.startsWith("/") || path.startsWith("//")) {
    throw new Error("public JSON path must be a relative same-origin path");
  }
  const parsed = new URL(path, "https://cartulary.invalid");
  if (parsed.origin !== "https://cartulary.invalid") {
    throw new Error("public JSON path must be a relative same-origin path");
  }
  return `${parsed.pathname}${parsed.search}`;
}

export async function requestPublicJson(
  options: PublicJsonOptions,
): Promise<PublicJsonResult> {
  return (await requestPublicJsonObserved(options)).result;
}

export async function requestPublicJsonObserved(options: PublicJsonOptions) {
  const path = requireSameOriginPath(options.path);
  const requestOptions: {
    data?: unknown;
    headers?: Record<string, string>;
    method: JsonRequestMethod;
  } = { method: options.method };
  if ("body" in options) {
    requestOptions.data = options.body;
  }
  if (options.headers !== undefined) {
    requestOptions.headers = options.headers;
  }
  const response = await options.request.fetch(path, requestOptions);
  const body = await response.json();
  const result = Object.freeze({
    body,
    headers: Object.freeze({ ...response.headers() }),
    ok: response.ok(),
    status: response.status(),
  });
  return Object.freeze({ response, result });
}
