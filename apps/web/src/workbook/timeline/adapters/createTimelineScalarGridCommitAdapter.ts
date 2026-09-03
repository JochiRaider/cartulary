import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import type { TimelineScalarSaveOptions } from "../models/timelineControllerPorts";
import type {
  RowValues,
  TimelineScalarEditorSurface,
} from "../models/timelineFieldRegistry";

export type TimelineQueueScalarSave = (
  rowKey: string,
  focusField: keyof RowValues,
  options: TimelineScalarSaveOptions,
  currentValue?: string,
  onSettled?: (outcome: GridEditCommitOutcome) => void,
) => void;

export function createTimelineScalarGridCommitAdapter(
  queueScalarSave: TimelineQueueScalarSave,
) {
  return (
    rowKey: string,
    focusField: keyof RowValues,
    currentValue: string,
  ): Promise<GridEditCommitOutcome> =>
    new Promise((resolve) => {
      const surface: TimelineScalarEditorSurface = "grid";
      queueScalarSave(
        rowKey,
        focusField,
        {
          continueOnFreshDraft: false,
          preserveInputFocus: false,
          surface,
        },
        currentValue,
        resolve,
      );
    });
}
