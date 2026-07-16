import type { ViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useState } from "react";
import { apiPath } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  handleWorkbookLoadFailure,
  parseErrorMessage,
  readEnvelope,
} from "../../services/workbookApi";
import {
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  type WorkbookLayoutState,
  type WorkbookQueryState,
  workbookLayoutStateFromSavedViewLayoutJson,
} from "../models/workbookQuery";
import {
  fallbackIdentityAfterSavedViewDelete,
  removeSavedViewList,
  savedViewConfigurationIsModified,
  savedViewIdentityForSelection,
  savedViewQueryStateForRuntime,
  upsertSavedViewList,
} from "../models/workbookSavedViewRuntime";
import {
  normalizeSavedViewResource,
  type SavedViewEnvelope,
  type SavedViewListEnvelope,
  type SavedViewResource,
  savedViewLayoutJsonForPersistence,
  savedViewQueryJsonForPersistence,
} from "../models/workbookSavedViews";
import type { WorkbookSheetRef } from "../models/workbookStartup";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import { knownWorkbookViewSchemaId } from "../models/workbookSurfaceRegistry";

type WorkbookIdentity = {
  readonly sheetRef: WorkbookSheetRef;
  readonly viewSchemaId: string | null;
};

export function useWorkbookSavedViewController({
  activeContract,
  apiBase,
  applyLayoutStateForSurface,
  applyQueryStateForSurface,
  applyWorkbookIdentity,
  currentLayoutStateForSurface,
  currentQueryStateForSurface,
  incidentId,
  onIncidentAccessLost,
  startupSheetRef,
}: {
  readonly activeContract: ViewContract;
  readonly apiBase?: string | undefined;
  readonly applyLayoutStateForSurface: (
    viewSchemaId: string,
    layoutState: WorkbookLayoutState,
  ) => void;
  readonly applyQueryStateForSurface: (
    viewSchemaId: string,
    queryState: WorkbookQueryState,
  ) => void;
  readonly applyWorkbookIdentity: (
    identity: WorkbookIdentity,
    options?: { readonly reloadSheet?: boolean },
  ) => void;
  readonly currentLayoutStateForSurface: (
    viewSchemaId: string,
  ) => WorkbookLayoutState;
  readonly currentQueryStateForSurface: (
    viewSchemaId: string,
  ) => WorkbookQueryState;
  readonly incidentId: string;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly startupSheetRef: WorkbookSheetRef;
}) {
  const [savedViews, setSavedViews] = useState<SavedViewResource[]>([]);

  const upsertSavedView = useCallback((savedView: SavedViewResource) => {
    setSavedViews((current) => upsertSavedViewList(current, savedView));
  }, []);

  const selectSavedView = useCallback(
    (savedView: SavedViewResource) => {
      const nextSurface = knownWorkbookViewSchemaId(savedView.view_schema_id);
      const contract = workbookContractForViewSchemaId(nextSurface);
      applyQueryStateForSurface(
        nextSurface,
        savedViewQueryStateForRuntime(contract, savedView),
      );
      applyLayoutStateForSurface(
        nextSurface,
        workbookLayoutStateFromSavedViewLayoutJson(
          contract,
          savedView.layout_json,
        ),
      );
      applyWorkbookIdentity(savedViewIdentityForSelection(savedView), {
        reloadSheet: true,
      });
    },
    [
      applyLayoutStateForSurface,
      applyQueryStateForSurface,
      applyWorkbookIdentity,
    ],
  );

  const createSavedView = useCallback(
    async (input: {
      readonly displayName: string;
      readonly scope: "private" | "shared";
    }) => {
      const queryState = currentQueryStateForSurface(
        activeContract.viewSchemaId,
      );
      const layoutState = currentLayoutStateForSurface(
        activeContract.viewSchemaId,
      );
      const result = await fetchWorkbookJSON<SavedViewEnvelope>(
        apiPath(apiBase, `/api/v1/incidents/${incidentId}/saved-views`),
        {
          method: "POST",
          body: JSON.stringify({
            view_schema_id: activeContract.viewSchemaId,
            display_name: input.displayName,
            scope: input.scope,
            query_json: buildSavedViewQueryJson(activeContract, queryState),
            layout_json: buildSavedViewLayoutJson(activeContract, layoutState),
          }),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const savedView = normalizeSavedViewResource(
        readEnvelope<SavedViewEnvelope>(result.payload).data,
      );
      if (savedView === null) {
        throw new Error("Saved-view create returned an invalid resource.");
      }
      upsertSavedView(savedView);
      selectSavedView(savedView);
      return savedView;
    },
    [
      activeContract,
      apiBase,
      currentQueryStateForSurface,
      currentLayoutStateForSurface,
      incidentId,
      selectSavedView,
      upsertSavedView,
    ],
  );

  const duplicateSavedView = useCallback(
    async (source: SavedViewResource) => {
      const contract = workbookContractForViewSchemaId(source.view_schema_id);
      const result = await fetchWorkbookJSON<SavedViewEnvelope>(
        apiPath(apiBase, `/api/v1/incidents/${incidentId}/saved-views`),
        {
          method: "POST",
          body: JSON.stringify({
            view_schema_id: source.view_schema_id,
            display_name: `${source.display_name} Copy`,
            scope: "private",
            query_json: savedViewQueryJsonForPersistence(
              contract,
              source.query_json,
            ),
            layout_json: savedViewLayoutJsonForPersistence(
              contract,
              source.layout_json,
            ),
          }),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const savedView = normalizeSavedViewResource(
        readEnvelope<SavedViewEnvelope>(result.payload).data,
      );
      if (savedView === null) {
        throw new Error("Saved-view duplicate returned an invalid resource.");
      }
      upsertSavedView(savedView);
      selectSavedView(savedView);
      return savedView;
    },
    [apiBase, incidentId, selectSavedView, upsertSavedView],
  );

  const updateSavedView = useCallback(
    async (
      savedView: SavedViewResource,
      input: {
        readonly displayName: string;
        readonly scope: "private" | "shared";
      },
    ) => {
      const contract = workbookContractForViewSchemaId(
        savedView.view_schema_id,
      );
      const queryState = currentQueryStateForSurface(savedView.view_schema_id);
      const layoutState = currentLayoutStateForSurface(
        savedView.view_schema_id,
      );
      const result = await fetchWorkbookJSON<SavedViewEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/saved-views/${savedView.saved_view_id}`,
        ),
        {
          method: "PATCH",
          body: JSON.stringify({
            base_saved_view_version: savedView.saved_view_version,
            display_name: input.displayName,
            scope: input.scope,
            query_json: buildSavedViewQueryJson(contract, queryState),
            layout_json: buildSavedViewLayoutJson(contract, layoutState),
          }),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const updated = normalizeSavedViewResource(
        readEnvelope<SavedViewEnvelope>(result.payload).data,
      );
      if (updated === null) {
        throw new Error("Saved-view update returned an invalid resource.");
      }
      upsertSavedView(updated);
      return updated;
    },
    [
      apiBase,
      currentLayoutStateForSurface,
      currentQueryStateForSurface,
      incidentId,
      upsertSavedView,
    ],
  );

  const deleteSavedView = useCallback(
    async (savedView: SavedViewResource) => {
      const result = await fetchWorkbookJSON<Record<string, unknown>>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/saved-views/${savedView.saved_view_id}`,
        ),
        { method: "DELETE" },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      setSavedViews((current) =>
        removeSavedViewList(current, savedView.saved_view_id),
      );
      const fallback = fallbackIdentityAfterSavedViewDelete(
        startupSheetRef,
        savedView,
      );
      if (fallback !== null) {
        applyWorkbookIdentity(fallback, { reloadSheet: true });
      }
    },
    [apiBase, applyWorkbookIdentity, incidentId, startupSheetRef],
  );

  useEffect(() => {
    let cancelled = false;
    const nextSavedViews: SavedViewResource[] = [];
    const loadSavedViews = async () => {
      let cursorToken: string | null = null;
      do {
        const query = new URLSearchParams({ limit: "100" });
        if (cursorToken !== null) {
          query.set("cursor_token", cursorToken);
        }
        const result = await fetchWorkbookJSON<SavedViewListEnvelope>(
          apiPath(
            apiBase,
            `/api/v1/incidents/${incidentId}/saved-views?${query.toString()}`,
          ),
        );
        if (cancelled) {
          return;
        }
        if (!result.ok) {
          handleWorkbookLoadFailure(
            parseErrorMessage(result.payload),
            "Saved views load failed.",
            onIncidentAccessLost,
          );
          setSavedViews([]);
          return;
        }
        const envelope = readEnvelope<SavedViewListEnvelope>(result.payload);
        for (const savedView of envelope.data.saved_views) {
          const normalized = normalizeSavedViewResource(savedView);
          if (normalized !== null) {
            nextSavedViews.push(normalized);
          }
        }
        const paging = envelope.meta?.paging;
        cursorToken =
          paging?.has_more === true && paging.next_cursor
            ? paging.next_cursor
            : null;
      } while (cursorToken !== null);

      if (!cancelled) {
        setSavedViews(nextSavedViews);
      }
    };
    void loadSavedViews();
    return () => {
      cancelled = true;
    };
  }, [apiBase, incidentId, onIncidentAccessLost]);

  const activeSavedView = useMemo(
    () =>
      startupSheetRef.kind === "saved_view"
        ? (savedViews.find(
            (savedView) => savedView.saved_view_id === startupSheetRef.id,
          ) ?? null)
        : null,
    [savedViews, startupSheetRef],
  );
  const activeSavedViewModified = savedViewConfigurationIsModified({
    contract: activeContract,
    currentLayoutState: currentLayoutStateForSurface(
      activeContract.viewSchemaId,
    ),
    currentQueryState: currentQueryStateForSurface(activeContract.viewSchemaId),
    savedView:
      activeSavedView?.view_schema_id === activeContract.viewSchemaId
        ? activeSavedView
        : null,
  });

  return {
    commands: {
      createSavedView,
      deleteSavedView,
      duplicateSavedView,
      selectSavedView,
      updateSavedView,
      upsertSavedView,
    },
    snapshot: { activeSavedViewModified, savedViews },
  };
}
