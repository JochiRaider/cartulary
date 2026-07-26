import type { GridDensity, GridInteractionMode } from "@cartulary/grid-adapter";
import type { Dispatch, ReactNode, SetStateAction } from "react";
import type { WorkbookResolvedLayoutState } from "../../models/workbookLayout";
import type {
  FilterDraft,
  WorkbookQueryState,
} from "../../models/workbookQuery";
import type { WorkbookSheetRef } from "../../models/workbookStartup";
import type { EntityApiRow } from "./workbookTimelineModel";

export type TimelineWorkbookIncidentRole =
  | "viewer"
  | "editor"
  | "reviewer"
  | "admin"
  | "";

export type TimelineWorkbookEntityRow = {
  readonly entityType: "host" | "identity";
  readonly recordId: string;
  readonly rowVersion: number;
  readonly label: string;
  readonly secondaryText: string;
  readonly state: string;
  readonly aliasTexts: string[];
  readonly linkedEventCount: number;
  readonly rawRow: EntityApiRow;
  readonly identifiers: readonly {
    readonly key: string;
    readonly label: string;
    readonly value: string;
  }[];
};

export type TimelineWorkbookSurfaceRuntime = {
  readonly incident: {
    readonly id: string;
    readonly apiBase: string | undefined;
    readonly currentUserId: string | null;
    readonly currentRole: TimelineWorkbookIncidentRole | null;
    readonly sheetRef: WorkbookSheetRef;
    readonly inspectorResetKey: string;
    readonly reloadToken: number;
  };
  readonly query: {
    readonly state: WorkbookQueryState;
    readonly setState: Dispatch<SetStateAction<WorkbookQueryState>>;
    readonly filterDraft: FilterDraft;
    readonly setFilterDraft: Dispatch<SetStateAction<FilterDraft>>;
    readonly renderInlineControls: boolean;
    readonly savedViewSelector: ReactNode | undefined;
  };
  readonly entities: {
    readonly hosts: readonly TimelineWorkbookEntityRow[];
    readonly identities: readonly TimelineWorkbookEntityRow[];
    readonly index: Record<string, TimelineWorkbookEntityRow>;
    readonly refresh: (() => Promise<void> | void) | undefined;
  };
  readonly layout: {
    readonly density: GridDensity;
    readonly interactionMode: GridInteractionMode;
    readonly state: WorkbookResolvedLayoutState;
    readonly setColumnHidden: (fieldKey: string, hidden: boolean) => void;
    readonly moveColumn: (
      fieldKey: string,
      direction: "earlier" | "later",
    ) => void;
    readonly reorderColumn: (
      sourceFieldKey: string,
      targetFieldKey: string,
    ) => void;
    readonly setColumnWidth: (fieldKey: string, width: number) => void;
    readonly resetColumns: () => void;
  };
  readonly onIncidentAccessLost: (() => void) | undefined;
};
