import type {
  HTTPOperationID,
  HTTPOperationRequest,
  HTTPOperationResponse,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import { validatedPublicErrorReason } from "../../services/publicErrorIdentity";
import { validateDisplayName } from "../../shared/displayName";
import { normalizeWorkbookSavedViewPage } from "../models/workbookSavedViewPaginationMachine";
import {
  normalizeSavedViewResource,
  type SavedViewResource,
  savedViewJSONEqual,
} from "../models/workbookSavedViews";
import type {
  SavedViewProblem,
  SavedViewResult,
  WorkbookSavedViewChanges,
  WorkbookSavedViewPort,
} from "../ports/WorkbookSavedViewPort";
import { decodeWorkbookPublicError } from "./workbookPublicErrorDecoder";

function problem<T>(
  kind: SavedViewProblem["kind"],
  message: string,
  uncertain = false,
): SavedViewResult<T> {
  return {
    kind: uncertain ? "uncertain" : "rejected",
    failure: { kind, message },
  };
}

function persistenceQuery(
  query: NonNullable<WorkbookSavedViewChanges["queryJson"]>,
) {
  return {
    sort: query.sort.map((entry) => ({ ...entry })),
    filters: query.filters.map((entry) => ({
      ...entry,
      arg: structuredClone(entry.arg),
    })),
    ...(query.group_by === undefined ? {} : { group_by: query.group_by }),
  };
}

function persistenceFields(changes: WorkbookSavedViewChanges) {
  return {
    ...(changes.displayName === undefined
      ? {}
      : { display_name: changes.displayName }),
    ...(changes.scope === undefined ? {} : { scope: changes.scope }),
    ...(changes.layoutJson === undefined
      ? {}
      : {
          layout_json: {
            layout_schema_id: changes.layoutJson.layout_schema_id,
            column_order: [...changes.layoutJson.column_order],
            hidden_field_keys: [...changes.layoutJson.hidden_field_keys],
            column_widths: changes.layoutJson.column_widths.map((entry) => ({
              ...entry,
            })),
          },
        }),
    ...(changes.queryJson === undefined
      ? {}
      : {
          query_json: persistenceQuery(changes.queryJson),
        }),
  };
}

function failure<T>(
  status: number,
  payload: unknown,
  write: boolean,
  base?: SavedViewResource,
): SavedViewResult<T> {
  const decoded = decodeWorkbookPublicError(payload);
  if (decoded.kind !== "decoded" || decoded.envelope.error.status !== status) {
    return problem(
      "invalid_contract",
      "The server returned an invalid saved-view response.",
      write,
    );
  }
  const error = decoded.envelope.error;
  const common = { publicCode: error.code };
  if (status === 409 && error.code === "saved_view_version_conflict") {
    const details = error.details;
    if (
      !base ||
      details.saved_view_id !== base.saved_view_id ||
      details.base_saved_view_version !== base.saved_view_version ||
      typeof details.current_saved_view_version !== "number" ||
      !Number.isSafeInteger(details.current_saved_view_version) ||
      details.current_saved_view_version <= base.saved_view_version
    ) {
      return problem(
        "invalid_contract",
        "The server returned an invalid saved-view conflict response.",
        write,
      );
    }
    return {
      kind: "rejected",
      failure: {
        ...common,
        kind: "conflict",
        message:
          "This saved view changed. Review the current saved configuration before writing again.",
        conflict: {
          savedViewId: base.saved_view_id,
          baseVersion: base.saved_view_version,
          currentVersion: details.current_saved_view_version,
        },
      },
    };
  }
  if (
    status === 400 &&
    [
      "invalid_mutation_payload",
      "invalid_pagination_request",
      "invalid_path_parameter",
    ].includes(error.code)
  ) {
    const field = error.details.field;
    const reason = validatedPublicErrorReason(
      error.code,
      error.details.reason_code,
    );
    return {
      kind: "rejected",
      failure: {
        ...common,
        kind: "validation",
        message: "Check the saved-view values and try again.",
        ...(typeof field === "string" && /^[a-zA-Z0-9_.[\]]{1,160}$/.test(field)
          ? { field }
          : {}),
        ...(reason === undefined ? {} : { reason }),
      },
    };
  }
  if (status === 401 && error.code === "authentication_required") {
    return {
      kind: "rejected",
      failure: {
        ...common,
        kind: "authentication_required",
        message: "Sign in again to continue this saved-view action.",
      },
    };
  }
  if (
    status === 403 &&
    ["authorization_denied", "csrf_failed"].includes(error.code)
  ) {
    return {
      kind: "rejected",
      failure: {
        ...common,
        kind: "authorization_denied",
        message:
          "This saved-view action was not permitted. Check current access.",
      },
    };
  }
  if (
    status === 404 &&
    ["saved_view_not_found", "incident_not_found"].includes(error.code)
  ) {
    return {
      kind: "rejected",
      failure: {
        ...common,
        kind: "unavailable_target",
        message:
          "This saved view is unavailable. Check current access and refresh the list.",
      },
    };
  }
  return {
    kind: write ? "uncertain" : "rejected",
    failure: {
      ...common,
      kind: "terminal",
      message: "The saved-view request could not be confirmed.",
    },
  };
}

function patchCorrelates(
  resource: SavedViewResource,
  base: SavedViewResource,
  changes: WorkbookSavedViewChanges,
): boolean {
  if (
    resource.saved_view_id !== base.saved_view_id ||
    resource.incident_id !== base.incident_id ||
    resource.view_schema_id !== base.view_schema_id ||
    resource.owner_user_id !== base.owner_user_id ||
    resource.created_at !== base.created_at
  )
    return false;
  if (
    resource.display_name !==
      (changes.displayName === undefined
        ? base.display_name
        : validateDisplayName(changes.displayName).value) ||
    resource.scope !== (changes.scope ?? base.scope) ||
    (changes.queryJson === undefined &&
      !savedViewJSONEqual(resource.query_json, base.query_json)) ||
    (changes.layoutJson === undefined &&
      !savedViewJSONEqual(resource.layout_json, base.layout_json))
  )
    return false;
  if (resource.saved_view_version === base.saved_view_version)
    return savedViewJSONEqual(resource, base);
  return (
    resource.saved_view_version === base.saved_view_version + 1 &&
    resource.updated_at !== base.updated_at &&
    !savedViewJSONEqual(
      {
        ...resource,
        saved_view_version: base.saved_view_version,
        updated_at: base.updated_at,
      },
      base,
    )
  );
}

export function createWorkbookSavedViewAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): WorkbookSavedViewPort {
  async function execute<I extends HTTPOperationID>(
    operationID: I,
    signal: AbortSignal,
    write: boolean,
    request?: HTTPOperationRequest<I>,
    base?: SavedViewResource,
    savedViewId?: string,
  ) {
    try {
      const result = await fetchHTTPOperation<HTTPOperationResponse<I>>({
        apiBase: options.apiBase,
        operationID,
        pathParameters: {
          incident_id: options.incidentId,
          ...(savedViewId === undefined ? {} : { saved_view_id: savedViewId }),
        },
        init: {
          signal,
          method:
            operationID === "createIncidentSavedView"
              ? "POST"
              : operationID === "patchIncidentSavedView"
                ? "PATCH"
                : "DELETE",
          ...(request === undefined ? {} : { body: JSON.stringify(request) }),
        },
      });
      if (!result.ok)
        return failure<HTTPOperationResponse<I>>(
          result.status,
          result.payload,
          write,
          base,
        );
      if (
        result.status !==
        (operationID === "createIncidentSavedView" ? 201 : 200)
      )
        return problem<HTTPOperationResponse<I>>(
          "invalid_contract",
          "The server returned an invalid saved-view response.",
          write,
        );
      return { kind: "accepted" as const, value: result.payload };
    } catch (error) {
      if (error instanceof SyntaxError)
        return problem<HTTPOperationResponse<I>>(
          "invalid_contract",
          "The server returned malformed saved-view JSON.",
          write,
        );
      return problem<HTTPOperationResponse<I>>(
        "transport",
        "The saved-view request lost its connection. Its outcome is unknown.",
        write,
      );
    }
  }
  return {
    async listPage(input) {
      try {
        const result = await fetchHTTPOperation<
          HTTPOperationResponse<"listIncidentSavedViews">
        >({
          apiBase: options.apiBase,
          operationID: "listIncidentSavedViews",
          pathParameters: { incident_id: options.incidentId },
          query: {
            limit: input.limit,
            ...(input.cursorToken === null
              ? {}
              : { cursor_token: input.cursorToken }),
          },
          init: { method: "GET", signal: input.signal },
        });
        if (!result.ok) return failure(result.status, result.payload, false);
        const page =
          result.status === 200
            ? normalizeWorkbookSavedViewPage({
                incidentId: options.incidentId,
                limit: input.limit,
                paging: result.payload.meta.paging,
                savedViews: result.payload.data.saved_views,
              })
            : null;
        return page === null
          ? problem(
              "invalid_contract",
              "Saved views load failed: invalid response.",
            )
          : { kind: "accepted", value: page };
      } catch (error) {
        if (error instanceof SyntaxError)
          return problem(
            "invalid_contract",
            "The server returned malformed saved-view JSON.",
          );
        return problem("transport", "Saved views could not be loaded.");
      }
    },
    async create(input) {
      const fields = persistenceFields(input.definition);
      const request = {
        ...fields,
        display_name: input.definition.displayName,
        query_json: persistenceQuery(input.definition.queryJson),
        view_schema_id: input.definition.viewSchemaId,
      } satisfies HTTPOperationRequest<"createIncidentSavedView">;
      const result = await execute(
        "createIncidentSavedView",
        input.signal,
        true,
        request,
      );
      if (result.kind !== "accepted") return result;
      const resource = normalizeSavedViewResource(result.value.data);
      return resource !== null &&
        resource.incident_id === options.incidentId &&
        resource.view_schema_id === input.definition.viewSchemaId &&
        resource.scope === input.definition.scope &&
        resource.display_name ===
          validateDisplayName(input.definition.displayName).value
        ? { kind: "accepted", value: resource }
        : problem(
            "invalid_contract",
            "The server returned an invalid saved-view acknowledgement.",
            true,
          );
    },
    async patch(input) {
      if (input.base.scope === "system")
        return problem("validation", "System saved views cannot be updated.");
      if (input.base.incident_id !== options.incidentId)
        return problem(
          "validation",
          "The saved-view target is no longer current.",
        );
      const request = {
        base_saved_view_version: input.base.saved_view_version,
        ...persistenceFields(input.changes),
      } satisfies HTTPOperationRequest<"patchIncidentSavedView">;
      const result = await execute(
        "patchIncidentSavedView",
        input.signal,
        true,
        request,
        input.base,
        input.base.saved_view_id,
      );
      if (result.kind !== "accepted") return result;
      const resource = normalizeSavedViewResource(result.value.data);
      return resource !== null &&
        patchCorrelates(resource, input.base, input.changes)
        ? { kind: "accepted", value: resource }
        : problem(
            "invalid_contract",
            "The server returned an invalid saved-view acknowledgement.",
            true,
          );
    },
    async delete(input) {
      if (input.scope === "system")
        return problem("validation", "System saved views cannot be deleted.");
      const result = await execute(
        "deleteIncidentSavedView",
        input.signal,
        true,
        undefined,
        undefined,
        input.savedViewId,
      );
      if (result.kind !== "accepted") return result;
      return result.value.data.deleted === true &&
        result.value.data.saved_view_id === input.savedViewId
        ? { kind: "accepted", value: undefined }
        : problem(
            "invalid_contract",
            "The server returned an invalid saved-view acknowledgement.",
            true,
          );
    },
  };
}
