import type { GridColumn, GridRow } from "@cartulary/grid-adapter";
import { useState } from "react";
import type { WorkbookFocusAnchor } from "../../utils/workbookGridFocus";
import type { WorkbookRow } from "../models/workbookTimelineModel";

type TimelineMutableRef<T> = {
  current: T;
};

export type TimelineGridInteractionRefs = {
  readonly gridShellRef: TimelineMutableRef<HTMLDivElement | null>;
  readonly rowInputRefs: TimelineMutableRef<
    Map<string, HTMLInputElement | HTMLTextAreaElement>
  >;
  readonly rowInputTestIdsRef: TimelineMutableRef<Map<string, string>>;
  readonly timelineAnchorColumnsRef: TimelineMutableRef<
    readonly GridColumn<WorkbookRow>[]
  >;
  readonly timelineAnchorRowsRef: TimelineMutableRef<
    readonly GridRow<WorkbookRow>[]
  >;
  readonly viewportContinuityTokenRef: TimelineMutableRef<number>;
  readonly workbookFocusAnchorRef: TimelineMutableRef<WorkbookFocusAnchor | null>;
};

export function useTimelineGridInteractions<TViewportContinuityRequest>({
  refs,
}: {
  readonly refs: TimelineGridInteractionRefs;
}) {
  const [workbookFocusAnchor, setWorkbookFocusAnchor] =
    useState<WorkbookFocusAnchor | null>(null);
  const {
    gridShellRef,
    rowInputRefs,
    rowInputTestIdsRef,
    timelineAnchorColumnsRef,
    timelineAnchorRowsRef,
    viewportContinuityTokenRef,
    workbookFocusAnchorRef,
  } = refs;
  const [viewportContinuityRequest, setViewportContinuityRequest] =
    useState<TViewportContinuityRequest | null>(null);

  return {
    commands: {
      setViewportContinuityRequest,
      setWorkbookFocusAnchor,
    },
    refs: {
      gridShellRef,
      rowInputRefs,
      rowInputTestIdsRef,
      timelineAnchorColumnsRef,
      timelineAnchorRowsRef,
      viewportContinuityTokenRef,
      workbookFocusAnchorRef,
    },
    snapshot: {
      viewportContinuityRequest,
      workbookFocusAnchor,
    },
  };
}
