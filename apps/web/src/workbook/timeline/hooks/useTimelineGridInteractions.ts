import type { GridColumn, GridHandle } from "@cartulary/grid-adapter";
import { useCallback, useState } from "react";
import { useWorkbookGridContinuity } from "../../continuity/useWorkbookGridContinuity";
import type { WorkbookContinuityAnchor } from "../../continuity/workbookContinuityPort";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";

type TimelineMutableRef<T> = {
  current: T;
};

export type TimelineGridInteractionRefs = {
  readonly gridHandleRef: TimelineMutableRef<GridHandle | null>;
  readonly gridShellRef: TimelineMutableRef<HTMLDivElement | null>;
  readonly timelineAnchorColumnsRef: TimelineMutableRef<
    readonly GridColumn<WorkbookRow>[]
  >;
  readonly viewportContinuityTokenRef: TimelineMutableRef<number>;
  readonly workbookFocusAnchorRef: TimelineMutableRef<WorkbookContinuityAnchor | null>;
};

export function useTimelineGridInteractions<TViewportContinuityRequest>({
  continuityResetKey,
  refs,
}: {
  readonly continuityResetKey: string;
  readonly refs: TimelineGridInteractionRefs;
}) {
  const {
    gridShellRef,
    gridHandleRef,
    timelineAnchorColumnsRef,
    viewportContinuityTokenRef,
    workbookFocusAnchorRef,
  } = refs;
  const [viewportContinuityRequest, setViewportContinuityRequest] =
    useState<TViewportContinuityRequest | null>(null);
  const continuity = useWorkbookGridContinuity({
    columns: timelineAnchorColumnsRef.current,
    continuityResetKey,
    gridHandleRef,
    selectionRef: workbookFocusAnchorRef,
    viewSchemaId: timelineViewSchemaId,
  });

  const updateWorkbookFocusAnchor = useCallback(
    (anchor: WorkbookContinuityAnchor | null) => {
      if (anchor === null) {
        continuity.port.clear();
      } else {
        continuity.port.select(anchor);
      }
    },
    [continuity.port],
  );

  const updateTimelineFocusAnchor = useCallback(
    (recordId: string | null, fieldKey: string, surface: string) => {
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
        viewSchemaId: surface,
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
      gridHandleRef,
      timelineAnchorColumnsRef,
      viewportContinuityTokenRef,
      workbookFocusAnchorRef,
    },
    snapshot: {
      viewportContinuityRequest,
      workbookFocusAnchor: continuity.snapshot.anchor,
    },
    continuityPort: continuity.port,
  };
}
