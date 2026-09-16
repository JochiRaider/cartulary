import type { SheetRef } from "../../shared/sheetRef";
import type { SavedViewResource } from "./workbookSavedViews";

/** Selected-resource projection. Discovery pages never determine identity. */
export type WorkbookSavedViewsResource = {
  readonly kind: "loading" | "ready" | "unavailable";
  readonly selectedSavedView: SavedViewResource | null;
  readonly selectedSavedViewId: string;
  readonly message: string | null;
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
  readonly selectedSavedView: SavedViewResource | null;
  readonly selectedSavedViewId: string;
};

export function workbookSavedViewsResource(
  observation:
    | import("../savedviews/SavedViewResourceObserver").SavedViewObservation
    | undefined,
  selectedSheetRef: SheetRef,
): WorkbookSavedViewsResource {
  if (selectedSheetRef.kind !== "saved_view")
    return {
      kind: "ready",
      selectedSavedView: null,
      selectedSavedViewId: "",
      message: null,
    };
  return {
    kind: observation?.resource
      ? "ready"
      : observation?.status === "unavailable" || observation?.problem
        ? "unavailable"
        : "loading",
    selectedSavedView: observation?.resource ?? null,
    selectedSavedViewId: selectedSheetRef.id,
    message:
      observation?.problem?.message ??
      (observation?.pending && !observation.resource
        ? "Loading selected saved view…"
        : null),
  };
}

export function projectActiveSurfaceSavedViews(
  resource: WorkbookSavedViewsResource,
  activeViewSchemaId: string,
  selectedSheetRef: SheetRef,
): ActiveSurfaceSavedViewProjection {
  const selected = resource.selectedSavedView;
  return {
    resourceKind: resource.kind,
    resourceMessage: resource.message,
    selectedSavedView:
      selected?.view_schema_id === activeViewSchemaId &&
      selectedSheetRef.kind === "saved_view" &&
      selected.saved_view_id === selectedSheetRef.id
        ? selected
        : null,
    selectedSavedViewId:
      selectedSheetRef.kind === "saved_view" ? selectedSheetRef.id : "",
  };
}

export function parseSavedViewEditableScope(
  value: string,
): SavedViewEditableScope | null {
  return value === "private" || value === "shared" ? value : null;
}
