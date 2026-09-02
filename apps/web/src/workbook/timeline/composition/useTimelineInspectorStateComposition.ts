import { requireViewContract } from "@cartulary/view-contracts";
import { type SetStateAction, useCallback, useRef } from "react";
import type {
  WorkbookContinuityAnchor,
  WorkbookContinuityPort,
  WorkbookContinuityToken,
} from "../../continuity/workbookContinuityPort";
import { useWorkbookInspectorCoordinator } from "../../inspector/useWorkbookInspectorCoordinator";
import { workbookInspectorStateIsOpen } from "../../models/workbookInspectorModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { useTimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import { useTimelineHistoryState } from "../hooks/useTimelineHistoryState";
import { useTimelineInspectorSelection } from "../hooks/useTimelineInspectorSelection";
import { selectTimelineInspectorHistorySubject } from "../models/timelineHistoryModel";
import type { DismissedMention } from "../models/workbookMentionChips";
import type { WorkbookRow } from "../models/workbookTimelineModel";

const timelineInspectorConfig =
  requireViewContract(timelineViewSchemaId).inspectorConfig;

export function useTimelineInspectorStateComposition({
  continuity,
  currentIncidentRole,
  dismissedMentionsByRow,
  inspectorResetKey,
  rows,
  selectedMentionRef,
  workbookFocusAnchorRef,
}: {
  readonly continuity: WorkbookContinuityPort;
  readonly currentIncidentRole: string | null | undefined;
  readonly dismissedMentionsByRow: Record<string, DismissedMention[]>;
  readonly inspectorResetKey: string;
  readonly rows: readonly WorkbookRow[];
  readonly selectedMentionRef: string | null;
  readonly workbookFocusAnchorRef: {
    readonly current: WorkbookContinuityAnchor | null;
  };
}) {
  const inspectorContinuityTokenRef = useRef<WorkbookContinuityToken | null>(
    null,
  );
  const selection = useTimelineInspectorSelection({
    currentIncidentRole,
    dismissedMentionsByRow,
    rows,
    selectedMentionRef,
  });
  const selectedRow = selection.snapshot.selectedRow;
  const history = useTimelineHistoryState({
    draftRow: selection.snapshot.draftRow,
    selectedRow,
  });
  const inspectorSubject = selectTimelineInspectorHistorySubject({
    draftRow: selection.snapshot.draftRow,
    rowHistory: history.snapshot.rowHistory,
    selectedRow,
  });
  const coordinator = useWorkbookInspectorCoordinator({
    actionPorts: {
      resetOwnerState: ({ scope }) => {
        selection.commands.setInspectorMessage(null);
        if (scope === "surface") {
          continuity.clear();
          selection.commands.setSelectedRowId(null);
        }
      },
      restoreFocus: () => {
        const token = inspectorContinuityTokenRef.current;
        inspectorContinuityTokenRef.current = null;
        if (token !== null) {
          continuity.restore(token);
        }
      },
    },
    config: timelineInspectorConfig,
    lifecycleKey: inspectorResetKey,
    subject: inspectorSubject,
  });
  const isOpen = workbookInspectorStateIsOpen(coordinator.snapshot);
  const elementRegistry = useTimelineInspectorElementRegistry(
    coordinator.snapshot,
  );
  const setOpen = useCallback(
    (next: SetStateAction<boolean>) => {
      const nextOpen = typeof next === "function" ? next(isOpen) : next;
      if (nextOpen && !isOpen) {
        inspectorContinuityTokenRef.current = continuity.capture();
      } else if (!nextOpen && isOpen) {
        inspectorContinuityTokenRef.current = continuity.capture(
          workbookFocusAnchorRef.current,
        );
      }
      coordinator.commands.setOpen(next);
    },
    [continuity, coordinator.commands, isOpen, workbookFocusAnchorRef],
  );
  return {
    commands: {
      history: history.commands,
      publishFeedback: selection.commands.setInspectorMessage,
      selectRow: selection.commands.setSelectedRowId,
      setOpen,
    },
    ports: { elements: elementRegistry },
    refs: {
      continuityToken: inspectorContinuityTokenRef,
    },
    snapshot: {
      history: history.snapshot,
      lifecycle: coordinator.snapshot,
      selection: {
        ...selection.snapshot,
        selectedRowWorkflowSubject:
          inspectorSubject?.kind === "live" ? inspectorSubject : null,
      },
    },
  };
}
