import type {
  GridCellAnchor,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import type { TimelinePresenceDraft } from "../services/workbookCollaborationMessages";
import type {
  createWorkbookSocketLifecycleState,
  WorkbookSocketLifecycleAction,
  WorkbookSocketLifecycleEffect,
} from "../services/workbookSocketLifecycle";
import type {
  FocusFieldKey,
  TimelineScalarEditorSurface,
  WorkbookRow,
} from "./workbookTimelineModel";

export type TimelineMutableRef<T> = {
  current: T;
};

export type TimelinePasteTargetResolution = {
  readonly anchor: GridCellAnchor | null;
  readonly targetResolution: GridPasteTargetResolution;
};

export type TimelineScalarSaveOptions = {
  readonly allowZeroFieldCreate?: boolean | undefined;
  readonly continueOnFreshDraft: boolean;
  readonly preserveInputFocus: boolean;
  readonly surface: TimelineScalarEditorSurface;
};

export type PendingReplayRuntimeMeta = {
  focusField: FocusFieldKey;
  focusKey: string;
  surface: TimelineScalarEditorSurface;
  rowSnapshot: WorkbookRow;
  continueOnFreshDraft: boolean;
  detectAutoResolution: boolean;
  promoteToCommittedRowInspect: boolean;
  viewportContinuityToken: number;
};

type DispatchSocketLifecycle = (
  action: WorkbookSocketLifecycleAction,
) => WorkbookSocketLifecycleEffect[];

export type TimelineLiveUpdateRefs = {
  readonly currentPresenceRef: TimelineMutableRef<TimelinePresenceDraft>;
  readonly dispatchSocketLifecycleRef: TimelineMutableRef<DispatchSocketLifecycle>;
  readonly presenceUpdateTimerRef: TimelineMutableRef<number | null>;
  readonly socketConnectionIDRef: TimelineMutableRef<string | null>;
  readonly socketLifecycleRef: TimelineMutableRef<
    ReturnType<typeof createWorkbookSocketLifecycleState>
  >;
  readonly socketReconnectAfterAuthRef: TimelineMutableRef<(() => void) | null>;
};
