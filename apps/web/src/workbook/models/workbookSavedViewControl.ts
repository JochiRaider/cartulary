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
  | "delete";

export type SavedViewControlFeedback = {
  readonly kind: "success" | "error" | "notice";
  readonly message: string;
};
export type SavedViewSurfaceControlState = {
  readonly selectionKey: string;
  readonly displayName: string;
  readonly scope: SavedViewEditableScope;
  readonly panelOpen: boolean;
  readonly busy: boolean;
  readonly feedback: SavedViewControlFeedback | null;
};
export type SavedViewControlEvent =
  | {
      readonly type: "toggle_panel" | "close_panel" | "clear_feedback";
      readonly surface: string;
    }
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
      readonly type: "publish_notice";
      readonly surface: string;
      readonly message: string;
    };

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
