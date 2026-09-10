import { type APIError, publicErrorView } from "../services/browserApi";

export { observeAsyncOperation as observeAccountOperation } from "../services/asyncObservation";

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
