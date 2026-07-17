import type {
  JsonRequestContextLike,
  JsonRequestMethod,
  PublicJsonResult,
} from "./publicJsonClient";

const testControlRoutePrefix = "/api/v1/test/";

function canonicalOrigin(value: string, label: string): string {
  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    throw new Error(`${label} must be an absolute HTTP origin`);
  }
  if (
    (parsed.protocol !== "http:" && parsed.protocol !== "https:") ||
    parsed.username !== "" ||
    parsed.password !== "" ||
    parsed.pathname !== "/" ||
    parsed.search !== "" ||
    parsed.hash !== ""
  ) {
    throw new Error(`${label} must be an absolute HTTP origin`);
  }
  return parsed.origin;
}

function requireControlPath(path: string): string {
  if (!path.startsWith(testControlRoutePrefix) || path.startsWith("//")) {
    throw new Error(
      `test-control path must start with ${testControlRoutePrefix}`,
    );
  }
  const parsed = new URL(path, "https://cartulary.invalid");
  if (parsed.origin !== "https://cartulary.invalid") {
    throw new Error(
      "test-control path must be relative to the approved origin",
    );
  }
  return `${parsed.pathname}${parsed.search}`;
}

export class TestControlClient {
  readonly #endpointOrigin: string;
  readonly #request: JsonRequestContextLike;
  readonly #requestOrigin: string;
  readonly #token: string;

  constructor(options: {
    readonly approvedEndpointOrigins: readonly string[];
    readonly approvedRequestOrigins: readonly string[];
    readonly endpointOrigin: string;
    readonly request: JsonRequestContextLike;
    readonly requestOrigin: string;
    readonly token: string;
  }) {
    const token = options.token.trim();
    if (token === "") {
      throw new Error("test-control token must be non-empty");
    }
    const endpointOrigin = canonicalOrigin(
      options.endpointOrigin,
      "test-control endpoint origin",
    );
    const approvedEndpointOrigins = new Set(
      options.approvedEndpointOrigins.map((approved) =>
        canonicalOrigin(approved, "approved test-control endpoint origin"),
      ),
    );
    if (!approvedEndpointOrigins.has(endpointOrigin)) {
      throw new Error("test-control endpoint host is not approved");
    }
    const requestOrigin = canonicalOrigin(
      options.requestOrigin,
      "test-control request origin",
    );
    const approvedRequestOrigins = new Set(
      options.approvedRequestOrigins.map((approved) =>
        canonicalOrigin(approved, "approved test-control request origin"),
      ),
    );
    if (!approvedRequestOrigins.has(requestOrigin)) {
      throw new Error("test-control request origin is not approved");
    }
    this.#endpointOrigin = endpointOrigin;
    this.#request = options.request;
    this.#requestOrigin = requestOrigin;
    this.#token = token;
  }

  async request(options: {
    readonly body?: unknown;
    readonly method: JsonRequestMethod;
    readonly path: string;
  }): Promise<PublicJsonResult> {
    const path = requireControlPath(options.path);
    const requestOptions: {
      data?: unknown;
      headers: Record<string, string>;
      method: JsonRequestMethod;
    } = {
      headers: {
        Origin: this.#requestOrigin,
        "X-Cartulary-Test-Route-Token": this.#token,
      },
      method: options.method,
    };
    if ("body" in options) {
      requestOptions.data = options.body;
    }
    const response = await this.#request.fetch(
      `${this.#endpointOrigin}${path}`,
      requestOptions,
    );
    return Object.freeze({
      body: await response.json(),
      headers: Object.freeze({ ...response.headers() }),
      ok: response.ok(),
      status: response.status(),
    });
  }
}
