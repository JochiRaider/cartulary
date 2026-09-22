import type { GridHandle } from "@cartulary/grid-adapter";
import type { WorkbookContinuityPort } from "../../continuity/workbookContinuityPort";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type {
  TimelineMutableRef,
  TimelineRowMutationEditorPort,
} from "../models/timelineControllerPorts";

/** Translates semantic mutation-editor commands at the grid boundary. */
export function createTimelineRowMutationEditorAdapter({
  continuityPort,
  gridHandleRef,
}: {
  readonly continuityPort: Pick<WorkbookContinuityPort, "focus">;
  readonly gridHandleRef: TimelineMutableRef<GridHandle | null>;
}): TimelineRowMutationEditorPort {
  return {
    activateEdit: ({ fieldKey, recordId, value, selectionRange }) => {
      gridHandleRef.current?.activateEdit(
        {
          fieldKey,
          rowIdentity: { kind: "core_record", recordId },
          surface: {
            kind: "view_schema",
            viewSchemaId: timelineViewSchemaId,
          },
        },
        { value, selectionRange },
      );
    },
    cancelEdit: ({ fieldKey, recordId }) => {
      gridHandleRef.current?.cancelEdit({
        fieldKey,
        rowIdentity: { kind: "core_record", recordId },
        surface: {
          kind: "view_schema",
          viewSchemaId: timelineViewSchemaId,
        },
      });
    },
    focus: ({ fieldKey, recordId }) => {
      continuityPort.focus({
        fieldKey,
        recordId,
        viewSchemaId: timelineViewSchemaId,
      });
    },
  };
}
