import {
  fetchHTTPOperation,
  type HTTPOperationError,
} from "../../services/browserApi";
import { isFailureCode, isUnappliedRefusal } from "../../shared/publicError";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookPreferencePort } from "../ports/WorkbookPreferencePort";
import {
  type PreferenceHTTPOperation,
  type PreferenceKind,
  type PreferenceResources,
  type PreferenceResult,
  preferencePointer,
  preferencePointersEqual,
  validPreferenceResource,
} from "../preferences/workbookPreferenceModel";

const operations = {
  home: {
    get: "getCurrentUserWorkbookPreferences",
    put: "putCurrentUserWorkbookPreferences",
  },
  default: {
    get: "getIncidentDefaultWorkbookPreferences",
    put: "putIncidentDefaultWorkbookPreferences",
  },
} as const;
/**
 * Only declared rejections, and a refusal before apply, are definite; other
 * write failures stay uncertain (Core 03 REQ-03-086a). Preference writes are
 * never replayed.
 */
function failure<T>(
  error: HTTPOperationError<PreferenceHTTPOperation>,
  write: boolean,
): PreferenceResult<T> {
  const definite =
    isFailureCode(error.code, [
      "invalid_mutation_payload",
      "session_required",
      "csrf_verification_failed",
      "authorization_denied",
      "incident_not_found",
    ]) || isUnappliedRefusal(error.status, error);
  return {
    kind: write && !definite ? "uncertain" : "rejected",
    status: error.status,
    problem: { code: error.code },
  };
}
export function createWorkbookPreferenceAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly actorId: string;
}): WorkbookPreferencePort {
  async function execute<K extends PreferenceKind>(
    kind: K,
    signal: AbortSignal,
    request?: { target: SheetRef | null },
  ): Promise<PreferenceResult<PreferenceResources[K]>> {
    const operationID = request ? operations[kind].put : operations[kind].get;
    try {
      const result = await fetchHTTPOperation({
        apiBase: options.apiBase,
        operationID,
        pathParameters: { incident_id: options.incidentId },
        init: {
          signal,
          method: request ? "PUT" : "GET",
          ...(request
            ? {
                body: JSON.stringify(
                  kind === "home"
                    ? { home_sheet_ref: request.target }
                    : { default_sheet_ref: request.target },
                ),
              }
            : {}),
        },
      });
      if (!result.ok)
        return failure(result.payload.error, request !== undefined);
      const resource = result.payload.data;
      if (
        result.status !== 200 ||
        !validPreferenceResource(resource, kind, options) ||
        (request &&
          !preferencePointersEqual(preferencePointer(resource), request.target))
      )
        return {
          kind: request ? "uncertain" : "rejected",
          status: 502,
          problem: { code: "invalid_public_contract_response" },
        };
      const value = structuredClone(resource);
      const pointer = preferencePointer(value);
      if (pointer !== null) Object.freeze(pointer);
      Object.freeze(value);
      return { kind: "accepted", value };
    } catch {
      return {
        kind: request ? "uncertain" : "rejected",
        status: 0,
        problem: { code: "transport" },
      };
    }
  }
  return {
    readHome: ({ signal }) => execute("home", signal),
    readDefault: ({ signal }) => execute("default", signal),
    setHomeSheet: ({ signal, sheetRef }) =>
      execute("home", signal, { target: sheetRef }),
    setDefaultSheet: ({ signal, sheetRef }) =>
      execute("default", signal, { target: sheetRef }),
  };
}
