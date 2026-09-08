import type { HTTPOperationResponse } from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookPreferencePort } from "../ports/WorkbookPreferencePort";
import {
  type PreferenceKind,
  type PreferenceProblem,
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
const knownCodes: readonly PreferenceProblem["code"][] = [
  "invalid_mutation_payload",
  "invalid_query_request",
  "invalid_path_parameter",
  "authentication_required",
  "authorization_denied",
  "csrf_failed",
  "incident_not_found",
  "internal_error",
  "invalid_public_contract_response",
];
function failure<T>(
  status: number,
  code: unknown,
  write: boolean,
): PreferenceResult<T> {
  const safeCode =
    knownCodes.find((candidate) => candidate === code) ??
    "unknown_public_error";
  const definite =
    (status === 400 &&
      [
        "invalid_mutation_payload",
        "invalid_query_request",
        "invalid_path_parameter",
      ].includes(safeCode)) ||
    (status === 401 && safeCode === "authentication_required") ||
    (status === 403 &&
      ["csrf_failed", "authorization_denied"].includes(safeCode)) ||
    (status === 404 && safeCode === "incident_not_found");
  return {
    kind: write && !definite ? "uncertain" : "rejected",
    status,
    problem: { code: safeCode },
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
      const result = await fetchHTTPOperation<
        HTTPOperationResponse<typeof operationID>
      >({
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
        return failure(
          result.status,
          result.payload.error?.code,
          request !== undefined,
        );
      const resource = result.payload.data;
      if (
        result.status !== 200 ||
        !validPreferenceResource(resource, kind, options) ||
        (request &&
          !preferencePointersEqual(preferencePointer(resource), request.target))
      )
        return failure(
          502,
          "invalid_public_contract_response",
          request !== undefined,
        );
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
