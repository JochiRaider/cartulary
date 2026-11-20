import {
  type JsonRequestContextLike,
  type JsonRequestMethod,
  type MultipartRequestPart,
  requireSameOriginPath,
} from "./publicJsonClient";

type PublicJsonOptions = {
  readonly body?: unknown;
  readonly headers?: Record<string, string>;
  readonly method: JsonRequestMethod;
  readonly multipart?: Readonly<Record<string, MultipartRequestPart>>;
  readonly path: string;
  readonly request: JsonRequestContextLike;
};

export async function requestPublicJsonObserved(options: PublicJsonOptions) {
  const path = requireSameOriginPath(options.path);
  if ("body" in options && options.multipart !== undefined) {
    throw new Error("a public request sends either a JSON body or multipart");
  }
  const requestOptions: {
    data?: unknown;
    headers?: Record<string, string>;
    method: JsonRequestMethod;
    multipart?: Readonly<Record<string, MultipartRequestPart>>;
  } = { method: options.method };
  if ("body" in options) {
    requestOptions.data = options.body;
  }
  if (options.multipart !== undefined) {
    requestOptions.multipart = options.multipart;
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
