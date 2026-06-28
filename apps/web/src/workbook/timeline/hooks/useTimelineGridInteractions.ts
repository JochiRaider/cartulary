import type { GridColumn, GridRow } from "@cartulary/grid-adapter";
import type { WorkbookSurface } from "@cartulary/ui-contracts";
import { useCallback, useState } from "react";
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

  const updateWorkbookFocusAnchor = useCallback(
    (anchor: WorkbookFocusAnchor | null) => {
      workbookFocusAnchorRef.current = anchor;
      setWorkbookFocusAnchor(anchor);
    },
    [workbookFocusAnchorRef],
  );

  const updateTimelineFocusAnchor = useCallback(
    (recordId: string | null, fieldKey: string, surface: WorkbookSurface) => {
      if (
        recordId === null ||
        recordId.trim() === "" ||
        !timelineAnchorColumnsRef.current.some(
          (column) => column.fieldKey === fieldKey,
        )
      ) {
        updateWorkbookFocusAnchor(null);
        return;
      }
      updateWorkbookFocusAnchor({
        fieldKey,
        recordId,
        surface,
      });
    },
    [timelineAnchorColumnsRef, updateWorkbookFocusAnchor],
  );

  return {
    commands: {
      setViewportContinuityRequest,
      updateTimelineFocusAnchor,
      updateWorkbookFocusAnchor,
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
