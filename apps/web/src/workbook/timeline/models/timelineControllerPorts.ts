import type {
  GridCellAnchor,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import type {
  createTimelineCollaborationState,
  TimelineCollaborationAction,
  TimelineCollaborationEffect,
} from "../services/timelineCollaborationEffects";
import type { TimelinePresenceDraft } from "../services/workbookCollaborationMessages";
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

type DispatchCollaborationAction = (
  action: TimelineCollaborationAction,
) => readonly TimelineCollaborationEffect[];

export type TimelineLiveUpdateRefs = {
  readonly currentPresenceRef: TimelineMutableRef<TimelinePresenceDraft>;
  readonly dispatchCollaborationRef: TimelineMutableRef<DispatchCollaborationAction>;
  readonly presenceUpdateTimerRef: TimelineMutableRef<number | null>;
  readonly collaborationStateRef: TimelineMutableRef<
    ReturnType<typeof createTimelineCollaborationState>
  >;
  readonly socketReconnectAfterAuthRef: TimelineMutableRef<(() => void) | null>;
};
