import { type APIError, publicErrorView } from "../services/browserApi";

/** An observation deadline does not establish whether the server executed a write. */
export function observeAccountOperation<T>(
  request: (signal: AbortSignal) => Promise<T>,
) {
  const controller = new AbortController();
  type Outcome =
    | { kind: "completed"; value: T }
    | { kind: "timeout" | "transport" | "cancelled" };
  let finish!: (outcome: Outcome) => void;
  const result = new Promise<Outcome>((resolve) => {
    let completed = false;
    const timer = setTimeout(() => finish({ kind: "timeout" }), 30_000);
    finish = (outcome) => {
      if (completed) return;
      completed = true;
      clearTimeout(timer);
      controller.abort();
      resolve(outcome);
    };
  });
  let transport: Promise<T>;
  try {
    transport = request(controller.signal);
  } catch {
    transport = Promise.reject();
  }
  const settled = transport.then(
    (value) => {
      finish({ kind: "completed", value });
    },
    () => {
      finish({ kind: "transport" });
    },
  );
  return { result, settled, cancel: () => finish({ kind: "cancelled" }) };
}

/** Retain only the public diagnostic projection, never raw messages, tokens or seeds. */
export function accountOperationError(
  error: APIError | null | undefined,
): APIError | null {
  const view = publicErrorView(error);
  if (!view) return null;
  return {
    code: /^[a-z][a-z0-9_]{0,95}$/.test(view.code)
      ? view.code
      : "unknown_public_error",
    ...(view.status === null ? {} : { status: view.status }),
    details: Object.fromEntries(
      view.details.map(({ key, value }) => [key, value]),
    ),
  };
}
