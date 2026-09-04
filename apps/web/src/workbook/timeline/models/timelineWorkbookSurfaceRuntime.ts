import type { Dispatch, SetStateAction } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookClipboardPastePort } from "../../adapters/WorkbookClipboardPastePort";
import type { WorkbookCollaborationCoordinator } from "../../collaboration/WorkbookCollaborationCoordinator";
import type { WorkbookViewBarWorkingSetBinding } from "../../components/WorkbookViewBar";
import type { WorkbookSurfaceLayoutOwner } from "../../layout/useWorkbookLayoutFacade";
import type { WorkbookGridEntryFocusOwner } from "../../models/workbookGridEntryFocus";
import type {
  FilterDraft,
  WorkbookQueryState,
} from "../../models/workbookQuery";
import type {
  IndicatorWorkflowPort,
  TimelineMutationCommandPorts,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookIncidentPort } from "../../ports/WorkbookIncidentPort";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";

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
  readonly rawRow: WorkbookQueryRow;
  readonly identifiers: readonly {
    readonly key: string;
    readonly label: string;
    readonly value: string;
  }[];
};

export type TimelineWorkbookSurfaceRuntime = {
  readonly attachCollaborationSession: boolean;
  readonly collaborationProjection: WorkbookCollaborationCoordinator;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly clipboardPaste: WorkbookClipboardPastePort;
  readonly mutationCommands: TimelineMutationCommandPorts;
  readonly indicatorWorkflow: IndicatorWorkflowPort;
  readonly gridEntryFocus: WorkbookGridEntryFocusOwner;
  readonly incident: {
    readonly id: string;
    readonly apiBase: string | undefined;
    readonly continuityResetKey: string;
    readonly currentUserId: string | null;
    readonly currentRole: TimelineWorkbookIncidentRole | null;
    readonly incidentPort: WorkbookIncidentPort;
    readonly sheetRef: SheetRef;
    readonly inspectorResetKey: string;
    readonly reloadToken: number;
  };
  readonly query: {
    readonly viewQuery: WorkbookViewQueryPort;
    readonly state: WorkbookQueryState;
    readonly setState: Dispatch<SetStateAction<WorkbookQueryState>>;
    readonly filterDraft: FilterDraft;
    readonly setFilterDraft: Dispatch<SetStateAction<FilterDraft>>;
    readonly renderInlineControls: boolean;
    readonly viewBarWorkingSet: WorkbookViewBarWorkingSetBinding | null;
  };
  readonly entities: {
    readonly hosts: readonly TimelineWorkbookEntityRow[];
    readonly identities: readonly TimelineWorkbookEntityRow[];
    readonly index: Record<string, TimelineWorkbookEntityRow>;
    readonly refresh: (() => Promise<void> | void) | undefined;
  };
  readonly layout: WorkbookSurfaceLayoutOwner;
  readonly onActivateConflict:
    | ((invoker: HTMLButtonElement) => void)
    | undefined;
  readonly onIncidentAccessLost: (() => void) | undefined;
};
