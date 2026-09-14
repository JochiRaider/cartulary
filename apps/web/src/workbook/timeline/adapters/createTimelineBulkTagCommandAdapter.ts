import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookBatchOperationOwner } from "../../runtime/WorkbookBatchOperationOwner";
import type { TimelineBulkTagCommandPort } from "../ports/TimelineBulkTagCommandPort";
export function createTimelineBulkTagCommandAdapter(
  owner: Pick<WorkbookBatchOperationOwner, "admit">,
): TimelineBulkTagCommandPort {
  return {
    assignTag(input, admission) {
      const [first, ...rest] = input.targets;
      if (!first || !input.tagName.trim()) return null;
      return owner.admit(
        {
          operation: "applyWorkbookBulkMutation",
          request: {
            view_schema_id: timelineViewSchemaId,
            kind: "multi_row_tag_assignment_v1",
            tag_name: input.tagName,
            targets: [first, ...rest].map((target) => ({
              record_id: target.recordId,
              base_row_version: target.baseRowVersion,
            })) as [
              { record_id: string; base_row_version: number },
              ...{ record_id: string; base_row_version: number }[],
            ],
          },
          recordIds: input.targets.map((target) => target.recordId),
        },
        admission,
      );
    },
  };
}
