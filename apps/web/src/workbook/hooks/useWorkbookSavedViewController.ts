import type { ViewContract } from "@cartulary/view-contracts";
import {
  useCallback,
  useLayoutEffect,
  useMemo,
  useRef,
  useSyncExternalStore,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import {
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  type WorkbookLayoutState,
  type WorkbookQueryState,
  workbookLayoutStateFromSavedViewLayoutJson,
  workbookQueryStateFromSavedViewQueryJson,
} from "../models/workbookQuery";
import {
  type WorkbookSavedViewsResource,
  workbookSavedViewsResource,
} from "../models/workbookSavedViewControl";
import {
  fallbackIdentityAfterSavedViewDelete,
  savedViewConfigurationIsModified,
  savedViewIdentityForSelection,
} from "../models/workbookSavedViewRuntime";
import {
  type SavedViewResource,
  savedViewJSONEqual,
} from "../models/workbookSavedViews";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import type { SavedViewBinding } from "../savedviews/savedViewOperationModel";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";

type WorkbookIdentity = {
  readonly sheetRef: SheetRef;
  readonly viewSchemaId: string | null;
};

/** Binds portable working configuration and selection effects to the session-owned workflow. */
export function useWorkbookSavedViewController({
  activeContract,
  applyLayoutStateForSurface,
  applyQueryStateForSurface,
  applyWorkbookIdentity,
  currentLayoutStateForSurface,
  currentQueryStateForSurface,
  controller,
  bindWorkbook,
  incidentId,
  apiBase,
  selectionGeneration,
  authorizationRecovered,
  startupSheetRef,
}: {
  readonly activeContract: ViewContract;
  readonly applyLayoutStateForSurface: (
    id: string,
    state: WorkbookLayoutState,
  ) => void;
  readonly applyQueryStateForSurface: (
    id: string,
    state: WorkbookQueryState,
  ) => void;
  readonly applyWorkbookIdentity: (
    identity: WorkbookIdentity,
    options?: { readonly reloadSheet?: boolean },
  ) => void;
  readonly currentLayoutStateForSurface: (id: string) => WorkbookLayoutState;
  readonly currentQueryStateForSurface: (id: string) => WorkbookQueryState;
  readonly controller: WorkbookSavedViewController;
  readonly bindWorkbook: (binding: SavedViewBinding | null) => void;
  readonly incidentId: string;
  readonly apiBase?: string | undefined;
  readonly selectionGeneration: number;
  readonly authorizationRecovered: SavedViewBinding["authorizationRecovered"];
  readonly startupSheetRef: SheetRef;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const currentLayout = currentLayoutStateForSurface(
    activeContract.viewSchemaId,
  );
  const currentQuery = currentQueryStateForSurface(activeContract.viewSchemaId);
  const queryJson = buildSavedViewQueryJson(activeContract, currentQuery);
  const layoutJson = buildSavedViewLayoutJson(activeContract, currentLayout);
  const configuration = {
    incidentId,
    surface: activeContract.viewSchemaId,
    queryJson,
    layoutJson,
  };
  const working = useRef({ configuration, generation: 0 });
  if (!savedViewJSONEqual(configuration, working.current.configuration))
    working.current = {
      configuration,
      generation: working.current.generation + 1,
    };
  const selected =
    startupSheetRef.kind === "saved_view"
      ? (state.resources.find(
          (r) =>
            r.saved_view_id === startupSheetRef.id &&
            r.view_schema_id === activeContract.viewSchemaId,
        ) ?? null)
      : null;
  const applyConfiguration = useCallback<
    SavedViewBinding["applyConfiguration"]
  >(
    (id, query, layout) => {
      const contract = workbookContractForViewSchemaId(id);
      applyQueryStateForSurface(
        id,
        workbookQueryStateFromSavedViewQueryJson(contract, query),
      );
      applyLayoutStateForSurface(
        id,
        workbookLayoutStateFromSavedViewLayoutJson(contract, layout),
      );
    },
    [applyQueryStateForSurface, applyLayoutStateForSurface],
  );
  const selectSavedView = useCallback(
    (resource: SavedViewResource) => {
      const current = controller
        .getSnapshot()
        .resources.find(
          (r) =>
            r.saved_view_id === resource.saved_view_id &&
            r.saved_view_version === resource.saved_view_version,
        );
      if (!current) return;
      const contract = workbookContractForViewSchemaId(current.view_schema_id);
      applyQueryStateForSurface(
        current.view_schema_id,
        workbookQueryStateFromSavedViewQueryJson(contract, current.query_json),
      );
      applyLayoutStateForSurface(
        current.view_schema_id,
        workbookLayoutStateFromSavedViewLayoutJson(
          contract,
          current.layout_json,
        ),
      );
      applyWorkbookIdentity(savedViewIdentityForSelection(current), {
        reloadSheet: true,
      });
    },
    [
      controller,
      applyLayoutStateForSurface,
      applyQueryStateForSurface,
      applyWorkbookIdentity,
    ],
  );
  const bindingRef = useRef(bindWorkbook);
  bindingRef.current = bindWorkbook;
  useLayoutEffect(() => {
    bindingRef.current({
      apiBase,
      incidentId,
      subject: {
        viewSchemaId: activeContract.viewSchemaId,
        savedViewId:
          startupSheetRef.kind === "saved_view" ? startupSheetRef.id : null,
        savedViewVersion: selected?.saved_view_version ?? null,
      },
      sheetRef: startupSheetRef,
      selectionGeneration,
      workingGeneration: working.current.generation,
      queryJson,
      layoutJson,
      applyConfiguration,
      select: selectSavedView,
      deleted: (resource) => {
        const fallback = fallbackIdentityAfterSavedViewDelete(
          startupSheetRef,
          resource,
        );
        if (fallback) applyWorkbookIdentity(fallback);
      },
      authorizationRecovered,
    });
  });
  useLayoutEffect(() => () => bindingRef.current(null), []);
  const savedViewsResource = useMemo<WorkbookSavedViewsResource>(() => {
    if (state.list === "loading") return { kind: "loading" };
    if (state.list === "unavailable")
      return {
        kind: "unavailable",
        message: state.listProblem?.message ?? "Saved views are unavailable.",
      };
    return workbookSavedViewsResource(state.resources, startupSheetRef);
  }, [state.list, state.listProblem, state.resources, startupSheetRef]);
  return {
    commands: { selectSavedView, upsertSavedView: controller.acceptResource },
    snapshot: {
      savedViewsResource,
      activeSavedViewModified: savedViewConfigurationIsModified({
        contract: activeContract,
        currentLayoutState: currentLayout,
        currentQueryState: currentQuery,
        savedView: selected,
      }),
    },
  };
}
