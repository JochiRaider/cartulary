import type { ViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { SheetRef } from "../../shared/sheetRef";
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
  type SavedViewResource,
  savedViewLayoutJsonForPersistence,
  savedViewQueryJsonForPersistence,
} from "../models/workbookSavedViews";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import { knownWorkbookViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";
import { workbookOperationFailureIsAccessLoss } from "../ports/WorkbookPortResult";
import type { WorkbookSavedViewPort } from "../ports/WorkbookSavedViewPort";

type WorkbookIdentity = {
  readonly sheetRef: SheetRef;
  readonly viewSchemaId: string | null;
};

function acceptedSavedViewResult<Accepted>(
  result: WorkbookPortResult<Accepted>,
  onIncidentAccessLost: (() => void) | undefined,
): Accepted {
  if (result.kind === "aborted") {
    throw new Error("Saved-view request was aborted.");
  }
  if (result.kind === "rejected") {
    if (workbookOperationFailureIsAccessLoss(result.failure)) {
      onIncidentAccessLost?.();
    }
    throw new Error(result.failure.message);
  }
  return result.value;
}

export function useWorkbookSavedViewController({
  activeContract,
  applyLayoutStateForSurface,
  applyQueryStateForSurface,
  applyWorkbookIdentity,
  currentLayoutStateForSurface,
  currentQueryStateForSurface,
  onIncidentAccessLost,
  savedViewPort,
  startupSheetRef,
}: {
  readonly activeContract: ViewContract;
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
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly savedViewPort: WorkbookSavedViewPort;
  readonly startupSheetRef: SheetRef;
}) {
  const [savedViews, setSavedViews] = useState<SavedViewResource[]>([]);
  const contextVersionRef = useRef(0);
  const savedViewPortRef = useRef(savedViewPort);
  if (savedViewPortRef.current !== savedViewPort) {
    savedViewPortRef.current = savedViewPort;
    contextVersionRef.current += 1;
  }

  useEffect(
    () => () => {
      contextVersionRef.current += 1;
    },
    [],
  );

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
      const contextVersion = contextVersionRef.current;
      const result = await savedViewPort.create({
        definition: {
          displayName: input.displayName,
          layoutJson: buildSavedViewLayoutJson(activeContract, layoutState),
          queryJson: buildSavedViewQueryJson(activeContract, queryState),
          scope: input.scope,
          viewSchemaId: activeContract.viewSchemaId,
        },
        signal: new AbortController().signal,
      });
      const savedView = acceptedSavedViewResult(result, onIncidentAccessLost);
      if (contextVersionRef.current !== contextVersion) {
        throw new Error("Saved-view create was superseded.");
      }
      upsertSavedView(savedView);
      selectSavedView(savedView);
      return savedView;
    },
    [
      activeContract,
      currentQueryStateForSurface,
      currentLayoutStateForSurface,
      onIncidentAccessLost,
      savedViewPort,
      selectSavedView,
      upsertSavedView,
    ],
  );

  const duplicateSavedView = useCallback(
    async (source: SavedViewResource) => {
      const contract = workbookContractForViewSchemaId(source.view_schema_id);
      const contextVersion = contextVersionRef.current;
      const result = await savedViewPort.create({
        definition: {
          displayName: `${source.display_name} Copy`,
          layoutJson: savedViewLayoutJsonForPersistence(
            contract,
            source.layout_json,
          ),
          queryJson: savedViewQueryJsonForPersistence(
            contract,
            source.query_json,
          ),
          scope: "private",
          viewSchemaId: source.view_schema_id,
        },
        signal: new AbortController().signal,
      });
      const savedView = acceptedSavedViewResult(result, onIncidentAccessLost);
      if (contextVersionRef.current !== contextVersion) {
        throw new Error("Saved-view duplicate was superseded.");
      }
      upsertSavedView(savedView);
      selectSavedView(savedView);
      return savedView;
    },
    [onIncidentAccessLost, savedViewPort, selectSavedView, upsertSavedView],
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
      const contextVersion = contextVersionRef.current;
      const result = await savedViewPort.patch({
        baseVersion: savedView.saved_view_version,
        definition: {
          displayName: input.displayName,
          layoutJson: buildSavedViewLayoutJson(contract, layoutState),
          queryJson: buildSavedViewQueryJson(contract, queryState),
          scope: input.scope,
        },
        savedViewId: savedView.saved_view_id,
        scope: savedView.scope,
        signal: new AbortController().signal,
        viewSchemaId: savedView.view_schema_id,
      });
      const updated = acceptedSavedViewResult(result, onIncidentAccessLost);
      if (contextVersionRef.current !== contextVersion) {
        throw new Error("Saved-view update was superseded.");
      }
      upsertSavedView(updated);
      return updated;
    },
    [
      currentLayoutStateForSurface,
      currentQueryStateForSurface,
      onIncidentAccessLost,
      savedViewPort,
      upsertSavedView,
    ],
  );

  const deleteSavedView = useCallback(
    async (savedView: SavedViewResource) => {
      const contextVersion = contextVersionRef.current;
      const result = await savedViewPort.delete({
        savedViewId: savedView.saved_view_id,
        scope: savedView.scope,
        signal: new AbortController().signal,
      });
      acceptedSavedViewResult(result, onIncidentAccessLost);
      if (contextVersionRef.current !== contextVersion) {
        throw new Error("Saved-view delete was superseded.");
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
    [
      applyWorkbookIdentity,
      onIncidentAccessLost,
      savedViewPort,
      startupSheetRef,
    ],
  );

  useEffect(() => {
    const controller = new AbortController();
    const nextSavedViews: SavedViewResource[] = [];
    const seenCursors = new Set<string>();
    const seenSavedViewIds = new Set<string>();
    const loadSavedViews = async () => {
      let cursorToken: string | null = null;
      do {
        const result = await savedViewPort.listPage({
          cursorToken,
          limit: 100,
          signal: controller.signal,
        });
        if (controller.signal.aborted || result.kind === "aborted") {
          return;
        }
        if (result.kind === "rejected") {
          if (workbookOperationFailureIsAccessLoss(result.failure)) {
            onIncidentAccessLost?.();
          }
          setSavedViews([]);
          return;
        }
        for (const savedView of result.value.savedViews) {
          if (seenSavedViewIds.has(savedView.saved_view_id)) {
            setSavedViews([]);
            return;
          }
          seenSavedViewIds.add(savedView.saved_view_id);
          nextSavedViews.push(savedView);
        }
        cursorToken = result.value.nextCursor;
        if (cursorToken !== null) {
          if (seenCursors.has(cursorToken)) {
            setSavedViews([]);
            return;
          }
          seenCursors.add(cursorToken);
        }
      } while (cursorToken !== null);

      if (!controller.signal.aborted) {
        setSavedViews(nextSavedViews);
      }
    };
    void loadSavedViews();
    return () => {
      controller.abort();
    };
  }, [onIncidentAccessLost, savedViewPort]);

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
