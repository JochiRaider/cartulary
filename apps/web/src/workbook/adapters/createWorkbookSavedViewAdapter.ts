import type { SavedViewResource } from "../models/workbookSavedViews";
import { normalizeSavedViewResource } from "../models/workbookSavedViews";
import type {
  WorkbookSavedViewDefinition,
  WorkbookSavedViewPort,
} from "../ports/WorkbookSavedViewPort";
import {
  invalidWorkbookAdapterResult,
  normalizeWorkbookAdapterFailure,
  workbookAdapterCaughtResult,
} from "./workbookAdapterResult";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function correlatedSavedView(
  value: unknown,
  incidentId: string,
  options: {
    readonly minimumVersion?: number;
    readonly savedViewId?: string;
    readonly viewSchemaId?: string;
  } = {},
): SavedViewResource | null {
  if (!isRecord(value)) {
    return null;
  }
  const resource = value;
  if (
    resource.incident_id !== incidentId ||
    (options.savedViewId !== undefined &&
      resource.saved_view_id !== options.savedViewId) ||
    (options.viewSchemaId !== undefined &&
      resource.view_schema_id !== options.viewSchemaId)
  ) {
    return null;
  }
  const savedView = normalizeSavedViewResource(value);
  if (
    savedView === null ||
    (options.minimumVersion !== undefined &&
      savedView.saved_view_version < options.minimumVersion)
  ) {
    return null;
  }
  return savedView;
}

function generatedPersistenceDefinition(
  definition: Omit<WorkbookSavedViewDefinition, "viewSchemaId">,
) {
  return {
    display_name: definition.displayName,
    layout_json: {
      column_order: [...definition.layoutJson.column_order],
      column_widths: definition.layoutJson.column_widths.map((entry) => ({
        field_key: entry.field_key,
        width_px: entry.width_px,
      })),
      hidden_field_keys: [...definition.layoutJson.hidden_field_keys],
      layout_schema_id: definition.layoutJson.layout_schema_id,
    },
    query_json: {
      filters: definition.queryJson.filters.map((filter) => ({
        arg: { ...filter.arg },
        field_key: filter.field_key,
        op: filter.op,
      })),
      sort: definition.queryJson.sort.map((entry) => ({
        direction: entry.direction,
        field_key: entry.field_key,
      })),
      ...(definition.queryJson.group_by === undefined
        ? {}
        : { group_by: definition.queryJson.group_by }),
    },
    scope: definition.scope,
  };
}

function generatedDefinition(definition: WorkbookSavedViewDefinition) {
  return {
    ...generatedPersistenceDefinition(definition),
    view_schema_id: definition.viewSchemaId,
  };
}

function immutableSystemViewResult(message: string) {
  return {
    kind: "rejected" as const,
    failure: { kind: "validation" as const, message },
  };
}

export function createWorkbookSavedViewAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): WorkbookSavedViewPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async listPage(input) {
      const message = "Saved views load failed.";
      try {
        const outcome = await operations.execute({
          operationID: "listIncidentSavedViews",
          pathParameters: { incident_id: options.incidentId },
          query: {
            limit: input.limit,
            ...(input.cursorToken === null
              ? {}
              : { cursor_token: input.cursorToken }),
          },
          signal: input.signal,
        });
        if (outcome.kind === "rejected") {
          return normalizeWorkbookAdapterFailure(outcome, message);
        }
        const paging = outcome.value.meta.paging;
        if (
          paging === undefined ||
          paging.limit !== input.limit ||
          (paging.has_more &&
            (paging.next_cursor === null ||
              paging.next_cursor.trim() === "")) ||
          (!paging.has_more && paging.next_cursor !== null)
        ) {
          return invalidWorkbookAdapterResult(message);
        }
        const savedViews: SavedViewResource[] = [];
        for (const value of outcome.value.data.saved_views) {
          const savedView = correlatedSavedView(value, options.incidentId, {
            minimumVersion: 1,
          });
          if (savedView === null) {
            return invalidWorkbookAdapterResult(message);
          }
          savedViews.push(savedView);
        }
        return {
          kind: "accepted",
          value: {
            nextCursor: paging.has_more ? paging.next_cursor : null,
            savedViews,
          },
        };
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
    async create(input) {
      const message = "Saved-view create failed.";
      try {
        const outcome = await operations.execute({
          operationID: "createIncidentSavedView",
          pathParameters: { incident_id: options.incidentId },
          request: generatedDefinition(input.definition),
          signal: input.signal,
        });
        if (outcome.kind === "rejected") {
          return normalizeWorkbookAdapterFailure(outcome, message);
        }
        const savedView = correlatedSavedView(
          outcome.value.data,
          options.incidentId,
          { minimumVersion: 1, viewSchemaId: input.definition.viewSchemaId },
        );
        return savedView === null
          ? invalidWorkbookAdapterResult(message)
          : { kind: "accepted", value: savedView };
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
    async patch(input) {
      const message = "Saved-view update failed.";
      if (input.scope === "system") {
        return immutableSystemViewResult(message);
      }
      try {
        const definition = generatedPersistenceDefinition(input.definition);
        const outcome = await operations.execute({
          operationID: "patchIncidentSavedView",
          pathParameters: {
            incident_id: options.incidentId,
            saved_view_id: input.savedViewId,
          },
          request: {
            base_saved_view_version: input.baseVersion,
            display_name: input.definition.displayName,
            layout_json: definition.layout_json,
            query_json: definition.query_json,
            scope: input.definition.scope,
          },
          signal: input.signal,
        });
        if (outcome.kind === "rejected") {
          return normalizeWorkbookAdapterFailure(outcome, message);
        }
        const savedView = correlatedSavedView(
          outcome.value.data,
          options.incidentId,
          {
            minimumVersion: input.baseVersion + 1,
            savedViewId: input.savedViewId,
            viewSchemaId: input.viewSchemaId,
          },
        );
        return savedView === null
          ? invalidWorkbookAdapterResult(message)
          : { kind: "accepted", value: savedView };
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
    async delete(input) {
      const message = "Saved-view delete failed.";
      if (input.scope === "system") {
        return immutableSystemViewResult(message);
      }
      try {
        const outcome = await operations.execute({
          operationID: "deleteIncidentSavedView",
          pathParameters: {
            incident_id: options.incidentId,
            saved_view_id: input.savedViewId,
          },
          signal: input.signal,
        });
        if (outcome.kind === "rejected") {
          return normalizeWorkbookAdapterFailure(outcome, message);
        }
        return outcome.value.data.deleted === true &&
          outcome.value.data.saved_view_id === input.savedViewId
          ? { kind: "accepted", value: undefined }
          : invalidWorkbookAdapterResult(message);
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
  };
}
