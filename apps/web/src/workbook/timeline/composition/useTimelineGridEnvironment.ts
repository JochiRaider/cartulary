import type { GridColumn, GridHandle } from "@cartulary/grid-adapter";
import { dataTestIdSelector, draftCellTestId } from "@cartulary/ui-contracts";
import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createTimelineRowMutationEditorAdapter } from "../adapters/createTimelineRowMutationEditorAdapter";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineGridAnchorController } from "../hooks/useTimelineGridAnchorController";
import {
  type TimelineGridInteractionRefs,
  useTimelineGridInteractions,
} from "../hooks/useTimelineGridInteractions";
import {
  type TimelineViewportContinuityRequest,
  useTimelineViewportContinuityController,
} from "../hooks/useTimelineViewportContinuityController";
import type { WorkbookRow } from "../models/workbookTimelineModel";

type TimelineRowsRef = {
  readonly current: readonly WorkbookRow[];
};

export function useTimelineGridEnvironment({
  continuityResetKey,
  editorDraftRegistry,
  rowsRef,
}: {
  readonly continuityResetKey: string;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly rowsRef: TimelineRowsRef;
}) {
  const workbookFocusAnchorRef = useRef(null);
  const timelineAnchorColumnsRef = useRef<readonly GridColumn<WorkbookRow>[]>(
    [],
  );
  const timelineGridHandleRef = useRef<GridHandle | null>(null);
  const gridShellRef = useRef<HTMLDivElement | null>(null);
  const viewportContinuityTokenRef = useRef(1);
  const [gridShellWidth, setGridShellWidth] = useState(0);
  const interactionRefs: TimelineGridInteractionRefs = {
    gridHandleRef: timelineGridHandleRef,
    gridShellRef,
    timelineAnchorColumnsRef,
    viewportContinuityTokenRef,
    workbookFocusAnchorRef,
  };
  const interactions =
    useTimelineGridInteractions<TimelineViewportContinuityRequest>({
      continuityResetKey,
      refs: interactionRefs,
    });
  const viewportContinuity = useTimelineViewportContinuityController({
    editorDraftRegistry,
    gridHandleRef: timelineGridHandleRef,
    gridShellRef,
    setViewportContinuityRequest:
      interactions.commands.setViewportContinuityRequest,
    viewportContinuityRequest: interactions.snapshot.viewportContinuityRequest,
    viewportContinuityTokenRef,
  });
  const updateTimelineSurfaceFocusAnchor = useCallback(
    (recordId: string | null, fieldKey: string) => {
      interactions.commands.updateTimelineFocusAnchor(
        recordId,
        fieldKey,
        timelineViewSchemaId,
      );
    },
    [interactions.commands],
  );
  const anchors = useTimelineGridAnchorController({
    continuityPort: interactions.continuityPort,
    gridHandleRef: timelineGridHandleRef,
    rowsRef,
    timelineAnchorColumnsRef,
    updateTimelineSurfaceFocusAnchor,
    updateWorkbookFocusAnchor: interactions.commands.updateWorkbookFocusAnchor,
  });
  const mutationEditor = useMemo(
    () =>
      createTimelineRowMutationEditorAdapter({
        continuityPort: interactions.continuityPort,
        focusInput: (focusKey) => {
          viewportContinuity.commands
            .resolveInputElement(focusKey)
            ?.focus({ preventScroll: true });
        },
        gridHandleRef: timelineGridHandleRef,
      }),
    [interactions.continuityPort, viewportContinuity.commands],
  );

  useLayoutEffect(() => {
    const gridShell = gridShellRef.current;
    if (gridShell === null) {
      return;
    }
    const updateGridShellWidth = (width: number) => {
      const measuredWidth = Math.max(0, Math.floor(width));
      setGridShellWidth((current) =>
        current === measuredWidth ? current : measuredWidth,
      );
    };
    const updateFromElement = () => {
      updateGridShellWidth(gridShell.clientWidth);
    };
    const scheduleUpdateFromElement = () => {
      updateFromElement();
      window.requestAnimationFrame(updateFromElement);
    };

    updateFromElement();
    window.addEventListener("resize", scheduleUpdateFromElement);
    window.visualViewport?.addEventListener(
      "resize",
      scheduleUpdateFromElement,
    );

    if (typeof ResizeObserver === "undefined") {
      return () => {
        window.removeEventListener("resize", scheduleUpdateFromElement);
        window.visualViewport?.removeEventListener(
          "resize",
          scheduleUpdateFromElement,
        );
      };
    }

    const observer = new ResizeObserver(scheduleUpdateFromElement);
    observer.observe(gridShell);
    observer.observe(document.documentElement);
    return () => {
      window.removeEventListener("resize", scheduleUpdateFromElement);
      window.visualViewport?.removeEventListener(
        "resize",
        scheduleUpdateFromElement,
      );
      observer.disconnect();
    };
  }, []);

  const registerVisibleColumns = useCallback(
    (columns: readonly GridColumn<WorkbookRow>[]) => {
      timelineAnchorColumnsRef.current = columns;
    },
    [],
  );
  const focusDraftRow = useCallback(() => {
    document
      .querySelector<HTMLInputElement>(
        dataTestIdSelector(draftCellTestId("timeline.activity_synopsis_text")),
      )
      ?.focus({ preventScroll: false });
  }, []);
  return {
    commands: {
      anchors: anchors,
      focusDraftRow,
      registerVisibleColumns,
      updateTimelineSurfaceFocusAnchor,
      updateWorkbookFocusAnchor:
        interactions.commands.updateWorkbookFocusAnchor,
      viewportContinuity: viewportContinuity.commands,
    },
    ports: {
      continuity: interactions.continuityPort,
      mutationEditor,
    },
    refs: {
      gridHandle: timelineGridHandleRef,
      gridShell: gridShellRef,
      timelineAnchorColumns: timelineAnchorColumnsRef,
      viewportContinuityToken: viewportContinuityTokenRef,
      workbookFocusAnchor: workbookFocusAnchorRef,
    },
    snapshot: {
      gridShellWidth,
      viewportContinuityRequest:
        interactions.snapshot.viewportContinuityRequest,
      workbookFocusAnchor: interactions.snapshot.workbookFocusAnchor,
    },
  };
}
