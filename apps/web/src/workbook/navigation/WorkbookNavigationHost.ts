import type { ViewContract } from "@cartulary/view-contracts";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookGridEntryFocusRequest } from "../models/workbookGridEntryFocus";
import type {
  WorkbookLayoutState,
  WorkbookQueryState,
} from "../models/workbookQuery";
import type { SavedViewResource } from "../models/workbookSavedViews";

/** Navigation consumes owner capabilities; shell composition adapts its runtime. */
export type WorkbookNavigationHost = {
  readonly snapshot: {
    readonly surface: string;
    readonly startupSheetRef: SheetRef;
    readonly activeContract: ViewContract;
    readonly selectedSavedView: SavedViewResource | null;
    readonly gridEntryFocusRequest: WorkbookGridEntryFocusRequest;
  };
  readonly commands: {
    readonly currentQueryStateForSurface: (id: string) => WorkbookQueryState;
    readonly currentLayoutStateForSurface: (id: string) => WorkbookLayoutState;
    readonly applyQueryStateForSurface: (
      id: string,
      state: WorkbookQueryState,
    ) => void;
    readonly applyLayoutStateForSurface: (
      id: string,
      state: WorkbookLayoutState,
    ) => void;
    readonly applyWorkbookIdentity: (
      identity: { readonly sheetRef: SheetRef; readonly viewSchemaId: string },
      options: { readonly focusFirstGridTarget: boolean },
    ) => void;
    readonly selectWorkbookSurface: (
      id: string,
      options: { readonly focusFirstGridTarget: boolean },
    ) => void;
    readonly selectExtensionWorkspace: (
      ref: Extract<SheetRef, { kind: "extension_workspace" }>,
    ) => void;
    readonly cancelGridEntryFocus: () => void;
  };
  readonly savedViews: {
    readonly retainNavigation: (id: string | null) => void;
    readonly read: (id: string) => Promise<SavedViewResource | null>;
    readonly isUnavailable: (id: string) => boolean;
    readonly acceptResource: (resource: SavedViewResource) => void;
  };
};
