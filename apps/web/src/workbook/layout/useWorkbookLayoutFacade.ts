import type { GridDensity, GridInteractionMode } from "@cartulary/grid-adapter";
import { useWorkbookResponsiveLayout } from "./useWorkbookResponsiveLayout";
import type { WorkbookResolvedLayoutState } from "./workbookColumnLayout";
import {
  type AccountDensityMode,
  resolveEffectiveWorkbookDensity,
} from "./workbookDensity";
import type {
  WorkbookBlockMode,
  WorkbookChromeMode,
} from "./workbookResponsiveLayout";

export type WorkbookLayoutSnapshot = {
  readonly blockMode: WorkbookBlockMode;
  readonly chromeMode: WorkbookChromeMode;
  readonly density: GridDensity;
  readonly showStatusPresence: boolean;
};

export type WorkbookLayoutFacade = {
  readonly shell: WorkbookLayoutSnapshot;
  readonly surface: WorkbookSurfaceLayoutOwner;
};

export type WorkbookSurfaceLayoutOwner = {
  readonly commands: {
    readonly onColumnHiddenChange: (fieldKey: string, hidden: boolean) => void;
    readonly onColumnMove: (
      fieldKey: string,
      direction: "earlier" | "later",
    ) => void;
    readonly onColumnReorder: (
      sourceFieldKey: string,
      targetFieldKey: string,
    ) => void;
    readonly onColumnWidthChange: (fieldKey: string, width: number) => void;
    readonly onResetColumns: () => void;
  };
  readonly snapshot: {
    readonly chromeMode: WorkbookChromeMode;
    readonly density: GridDensity;
    readonly interactionMode: GridInteractionMode;
    readonly showStatusPresence: boolean;
    readonly state: WorkbookResolvedLayoutState;
  };
};

export function useWorkbookLayoutFacade({
  accountDensityMode,
  columnCommands,
  columnState,
  interactionMode,
  viewSchemaId,
}: {
  readonly accountDensityMode: AccountDensityMode | undefined;
  readonly columnCommands: WorkbookSurfaceLayoutOwner["commands"];
  readonly columnState: WorkbookResolvedLayoutState;
  readonly interactionMode: GridInteractionMode;
  readonly viewSchemaId: string;
}): WorkbookLayoutFacade {
  const responsive = useWorkbookResponsiveLayout();
  const density = resolveEffectiveWorkbookDensity(
    viewSchemaId,
    accountDensityMode,
  );
  return {
    shell: {
      blockMode: responsive.blockMode,
      chromeMode: responsive.chromeMode,
      density,
      showStatusPresence:
        responsive.chromeMode === "compact_desktop" ||
        responsive.chromeMode === "below_supported_minimum",
    },
    surface: {
      commands: columnCommands,
      snapshot: {
        chromeMode: responsive.chromeMode,
        density,
        interactionMode,
        showStatusPresence:
          responsive.chromeMode === "compact_desktop" ||
          responsive.chromeMode === "below_supported_minimum",
        state: columnState,
      },
    },
  };
}
