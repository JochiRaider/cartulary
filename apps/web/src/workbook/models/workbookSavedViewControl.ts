import type { SheetRef } from "../../shared/sheetRef";
import type { SavedViewResource } from "./workbookSavedViews";

export type WorkbookSavedViewsResource =
  | { readonly kind: "loading" }
  | {
      readonly kind: "ready";
      readonly savedViews: readonly SavedViewResource[];
    }
  | { readonly kind: "unavailable"; readonly message: string }
  | {
      readonly kind: "invalid_selection";
      readonly savedViews: readonly SavedViewResource[];
      readonly selectedSavedViewId: string;
    };

export type SavedViewEditableScope = "private" | "shared";

export type SavedViewActionKind =
  | "create"
  | "update"
  | "duplicate"
  | "reset"
  | "delete"
  | "set_home"
  | "set_default";

export type SavedViewActionIdentity = {
  readonly surface: string;
  readonly savedViewId: string | null;
  readonly savedViewVersion: number | null;
  readonly actionKind: SavedViewActionKind;
  readonly generation: number;
};

export type SavedViewControlFeedback = {
  readonly kind: "success" | "error" | "notice";
  readonly message: string;
};

export type SavedViewSurfaceControlState = {
  readonly selectionKey: string;
  readonly displayName: string;
  readonly scope: SavedViewEditableScope;
  readonly panelOpen: boolean;
  readonly activeAction: SavedViewActionIdentity | null;
  readonly feedback: SavedViewControlFeedback | null;
};

export type SavedViewControlState = {
  readonly activeSurface: string;
  readonly surfaces: ReadonlyMap<string, SavedViewSurfaceControlState>;
};

export type SavedViewControlEvent =
  | {
      readonly type: "activate";
      readonly surface: string;
      readonly selectedSavedView: SavedViewResource | null;
    }
  | { readonly type: "toggle_panel"; readonly surface: string }
  | { readonly type: "close_panel"; readonly surface: string }
  | {
      readonly type: "change_name";
      readonly surface: string;
      readonly displayName: string;
    }
  | {
      readonly type: "change_scope";
      readonly surface: string;
      readonly scope: SavedViewEditableScope;
    }
  | {
      readonly type: "start_action";
      readonly identity: SavedViewActionIdentity;
    }
  | {
      readonly type: "complete_action";
      readonly identity: SavedViewActionIdentity;
      readonly feedback: SavedViewControlFeedback;
    }
  | {
      readonly type: "invalidate_action";
      readonly identity: SavedViewActionIdentity;
      readonly message: string;
    }
  | {
      readonly type: "publish_notice";
      readonly surface: string;
      readonly message: string;
    }
  | { readonly type: "clear_feedback"; readonly surface: string };

export type ActiveSurfaceSavedViewProjection = {
  readonly resourceKind: WorkbookSavedViewsResource["kind"];
  readonly resourceMessage: string | null;
  readonly savedViews: readonly SavedViewResource[];
  readonly privateSavedViews: readonly SavedViewResource[];
  readonly sharedSavedViews: readonly SavedViewResource[];
  readonly systemSavedViews: readonly SavedViewResource[];
  readonly selectedSavedView: SavedViewResource | null;
  readonly selectedSavedViewId: string;
};

export function workbookSavedViewsResource(
  savedViews: readonly SavedViewResource[],
  selectedSheetRef: SheetRef,
): WorkbookSavedViewsResource {
  if (
    selectedSheetRef.kind === "saved_view" &&
    !savedViews.some(
      (savedView) => savedView.saved_view_id === selectedSheetRef.id,
    )
  ) {
    return {
      kind: "invalid_selection",
      savedViews,
      selectedSavedViewId: selectedSheetRef.id,
    };
  }
  return { kind: "ready", savedViews };
}

export function projectActiveSurfaceSavedViews(
  resource: WorkbookSavedViewsResource,
  activeViewSchemaId: string,
  selectedSheetRef: SheetRef,
): ActiveSurfaceSavedViewProjection {
  const savedViews =
    resource.kind === "ready" || resource.kind === "invalid_selection"
      ? resource.savedViews.filter(
          (savedView) => savedView.view_schema_id === activeViewSchemaId,
        )
      : [];
  const selectedSavedView =
    selectedSheetRef.kind === "saved_view"
      ? (savedViews.find(
          (savedView) => savedView.saved_view_id === selectedSheetRef.id,
        ) ?? null)
      : null;
  return {
    resourceKind: resource.kind,
    resourceMessage: savedViewResourceMessage(resource),
    savedViews,
    privateSavedViews: savedViews.filter(
      (savedView) => savedView.scope === "private",
    ),
    sharedSavedViews: savedViews.filter(
      (savedView) => savedView.scope === "shared",
    ),
    systemSavedViews: savedViews.filter(
      (savedView) => savedView.scope === "system",
    ),
    selectedSavedView,
    selectedSavedViewId: selectedSavedView?.saved_view_id ?? "",
  };
}

export function parseSavedViewEditableScope(
  value: string,
): SavedViewEditableScope | null {
  return value === "private" || value === "shared" ? value : null;
}

export function createSavedViewControlState(
  surface: string,
  selectedSavedView: SavedViewResource | null,
): SavedViewControlState {
  return {
    activeSurface: surface,
    surfaces: new Map([[surface, defaultSurfaceState(selectedSavedView)]]),
  };
}

export function savedViewSurfaceControlState(
  state: SavedViewControlState,
  surface: string,
  selectedSavedView: SavedViewResource | null,
): SavedViewSurfaceControlState {
  const current = state.surfaces.get(surface);
  return current?.selectionKey === savedViewSelectionKey(selectedSavedView)
    ? current
    : defaultSurfaceState(selectedSavedView);
}

export function reduceSavedViewControl(
  state: SavedViewControlState,
  event: SavedViewControlEvent,
): SavedViewControlState {
  switch (event.type) {
    case "activate":
      return activateSavedViewSurface(state, event);
    case "toggle_panel":
      return updateExistingSurface(state, event.surface, (current) => ({
        ...current,
        panelOpen: !current.panelOpen,
      }));
    case "close_panel":
      return updateExistingSurface(state, event.surface, (current) =>
        current.panelOpen ? { ...current, panelOpen: false } : current,
      );
    case "change_name":
      return updateExistingSurface(state, event.surface, (current) => ({
        ...current,
        displayName: event.displayName,
      }));
    case "change_scope":
      return updateExistingSurface(state, event.surface, (current) => ({
        ...current,
        scope: event.scope,
      }));
    case "start_action":
      return startSavedViewAction(state, event.identity);
    case "complete_action":
      return completeSavedViewAction(state, event.identity, event.feedback);
    case "invalidate_action":
      return completeSavedViewAction(state, event.identity, {
        kind: "error",
        message: event.message,
      });
    case "publish_notice":
      return updateExistingSurface(state, event.surface, (current) => ({
        ...current,
        feedback: { kind: "notice", message: event.message },
      }));
    case "clear_feedback":
      return updateExistingSurface(state, event.surface, (current) =>
        current.feedback === null ? current : { ...current, feedback: null },
      );
  }
}

function activateSavedViewSurface(
  state: SavedViewControlState,
  event: Extract<SavedViewControlEvent, { readonly type: "activate" }>,
): SavedViewControlState {
  const current = state.surfaces.get(event.surface);
  const selectionKey = savedViewSelectionKey(event.selectedSavedView);
  const nextSurfaceState =
    current?.selectionKey === selectionKey
      ? current
      : {
          ...defaultSurfaceState(event.selectedSavedView),
          feedback:
            current?.feedback?.kind === "success" ? current.feedback : null,
        };
  if (state.activeSurface === event.surface && current === nextSurfaceState) {
    return state;
  }
  const surfaces = new Map(state.surfaces);
  closeInactiveSurface(surfaces, state.activeSurface, event.surface);
  surfaces.set(event.surface, nextSurfaceState);
  return { activeSurface: event.surface, surfaces };
}

function closeInactiveSurface(
  surfaces: Map<string, SavedViewSurfaceControlState>,
  previousSurface: string,
  nextSurface: string,
) {
  if (previousSurface === nextSurface) return;
  const previous = surfaces.get(previousSurface);
  if (previous === undefined) return;
  surfaces.set(previousSurface, {
    ...previous,
    panelOpen: false,
    activeAction: null,
  });
}

function startSavedViewAction(
  state: SavedViewControlState,
  identity: SavedViewActionIdentity,
): SavedViewControlState {
  return updateExistingSurface(state, identity.surface, (current) =>
    current.activeAction === null
      ? { ...current, activeAction: identity, feedback: null }
      : current,
  );
}

function completeSavedViewAction(
  state: SavedViewControlState,
  identity: SavedViewActionIdentity,
  feedback: SavedViewControlFeedback,
): SavedViewControlState {
  return updateExistingSurface(state, identity.surface, (current) =>
    sameSavedViewActionIdentity(current.activeAction, identity)
      ? {
          ...current,
          activeAction: null,
          feedback,
          panelOpen: feedback.kind === "error",
        }
      : current,
  );
}

function updateExistingSurface(
  state: SavedViewControlState,
  surface: string,
  update: (
    current: SavedViewSurfaceControlState,
  ) => SavedViewSurfaceControlState,
): SavedViewControlState {
  const current = state.surfaces.get(surface);
  if (current === undefined) return state;
  const next = update(current);
  return next === current ? state : withSurfaceState(state, surface, next);
}

export function sameSavedViewActionIdentity(
  left: SavedViewActionIdentity | null,
  right: SavedViewActionIdentity | null,
): boolean {
  return (
    left !== null &&
    right !== null &&
    left.surface === right.surface &&
    left.savedViewId === right.savedViewId &&
    left.savedViewVersion === right.savedViewVersion &&
    left.actionKind === right.actionKind &&
    left.generation === right.generation
  );
}

function savedViewResourceMessage(
  resource: WorkbookSavedViewsResource,
): string | null {
  switch (resource.kind) {
    case "loading":
      return "Loading saved views…";
    case "ready":
      return null;
    case "unavailable":
      return resource.message;
    case "invalid_selection":
      return "The selected saved view is no longer available. Showing the base surface.";
  }
}

function defaultSurfaceState(
  selectedSavedView: SavedViewResource | null,
): SavedViewSurfaceControlState {
  return {
    selectionKey: savedViewSelectionKey(selectedSavedView),
    displayName: selectedSavedView?.display_name ?? "Saved view",
    scope: selectedSavedView?.scope === "shared" ? "shared" : "private",
    panelOpen: false,
    activeAction: null,
    feedback: null,
  };
}

function savedViewSelectionKey(savedView: SavedViewResource | null): string {
  return savedView === null
    ? "base"
    : `${savedView.saved_view_id}:${savedView.saved_view_version}`;
}

function withSurfaceState(
  state: SavedViewControlState,
  surface: string,
  nextSurfaceState: SavedViewSurfaceControlState,
  activate = false,
): SavedViewControlState {
  const surfaces = new Map(state.surfaces);
  surfaces.set(surface, nextSurfaceState);
  return {
    activeSurface: activate ? surface : state.activeSurface,
    surfaces,
  };
}
