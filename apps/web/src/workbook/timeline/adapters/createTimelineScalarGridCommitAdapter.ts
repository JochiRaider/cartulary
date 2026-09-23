import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import type { TimelineQueueScalarSave } from "../models/timelineControllerPorts";
import type {
  RowValues,
  TimelineScalarEditorSurface,
} from "../models/timelineFieldRegistry";

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
