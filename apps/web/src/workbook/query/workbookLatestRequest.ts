export type LatestQueryRuntime = {
  controller: AbortController | null;
  sequence: number;
};

export type LatestQueryRequest = {
  isCurrent: () => boolean;
  signal: AbortSignal;
};

export function beginLatestQuery(runtime: {
  current: LatestQueryRuntime;
}): LatestQueryRequest {
  const previousController = runtime.current.controller;
  const controller = new AbortController();
  const sequence = runtime.current.sequence + 1;
  runtime.current = { controller, sequence };
  previousController?.abort();

  return {
    signal: controller.signal,
    isCurrent: () =>
      runtime.current.sequence === sequence &&
      runtime.current.controller === controller &&
      !controller.signal.aborted,
  };
}

export function abortLatestQuery(runtime: { current: LatestQueryRuntime }) {
  runtime.current.controller?.abort();
  runtime.current = {
    controller: null,
    sequence: runtime.current.sequence + 1,
  };
}

export function isAbortError(error: unknown): boolean {
  return error instanceof DOMException
    ? error.name === "AbortError"
    : error instanceof Error && error.name === "AbortError";
}
