import type { GridCoreRecordBulkSelection } from "@cartulary/grid-adapter";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { useWorkbookQueryPresentation } from "../../query/WorkbookQueryBrowsingContext";
import {
  planTimelineBulkTag,
  type TimelineBulkTagContext,
  timelineBulkTagMember,
} from "../models/timelineBulkTagPlan";
import type { WorkbookRow } from "../models/timelineRowModel";
import type {
  TimelineBulkTagAdmission,
  TimelineBulkTagCommandPort,
} from "../ports/TimelineBulkTagCommandPort";

export type TimelineBulkTagReadiness = {
  readonly subscribe: (listener: () => void) => () => void;
  readonly blockingReason: (recordIds: ReadonlySet<string>) => string | null;
};

type TimelineBulkTagControllerInput = {
  readonly context: TimelineBulkTagContext;
  readonly port: TimelineBulkTagCommandPort;
  readonly readiness: TimelineBulkTagReadiness;
  readonly precedingSaves: () => Promise<void>;
  readonly rows: readonly WorkbookRow[];
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
};

/** Selection is presentation identity. The batch owner captures execution identity. */
export function useTimelineBulkTagController(
  input: TimelineBulkTagControllerInput,
) {
  const browsing = useWorkbookQueryPresentation();
  const acceptedRows = browsing?.find(timelineViewSchemaId)?.getSnapshot()
    .accepted?.rows;
  const queryMembers = useMemo(
    () =>
      browsing === null
        ? null
        : new Set((acceptedRows ?? []).map((row) => row.record_id)),
    [acceptedRows, browsing],
  );
  const [selectedRecordIds, setSelectedRecordIds] = useState<
    ReadonlySet<string>
  >(() => new Set());
  const current = useRef({ input, browsing });
  const selected = useRef(selectedRecordIds);
  current.current = { input, browsing };
  const canAssign =
    input.context.authorized && input.context.capabilityAvailable;

  useEffect(() => {
    const members = new Set(
      canAssign
        ? input.rows
            .filter((row) => timelineBulkTagMember(row, queryMembers))
            .map((row) => row.recordId)
        : [],
    );
    setSelectedRecordIds((previous) => {
      const next = new Set([...previous].filter((id) => members.has(id)));
      selected.current = next.size === previous.size ? previous : next;
      return selected.current;
    });
  }, [canAssign, input.rows, queryMembers]);

  const changeSelectedRecordIds = useCallback((ids: ReadonlySet<string>) => {
    selected.current = new Set(ids);
    setSelectedRecordIds(selected.current);
  }, []);
  const gridSelection = useMemo<GridCoreRecordBulkSelection<WorkbookRow>>(
    () => ({
      isRecordSelectable: (row) =>
        canAssign && timelineBulkTagMember(row.data, queryMembers),
      onSelectedRecordIdsChange: changeSelectedRecordIds,
      selectedRecordIds,
    }),
    [canAssign, changeSelectedRecordIds, selectedRecordIds, queryMembers],
  );
  const getBlockingReason = useCallback(() => {
    const { input } = current.current;
    if (!input.context.authorized || !input.context.capabilityAvailable)
      return "Tag assignment is no longer available.";
    if (selected.current.size === 0)
      return "Select records to assign this tag.";
    return input.readiness.blockingReason(selected.current);
  }, []);
  const assignTag = useCallback(
    (tagName: string, delivery: object): TimelineBulkTagAdmission => {
      const { input, browsing } = current.current;
      const members =
        browsing === null
          ? null
          : new Set(
              (
                browsing.find(timelineViewSchemaId)?.getSnapshot().accepted
                  ?.rows ?? []
              ).map((row) => row.record_id),
            );
      const plan = planTimelineBulkTag({
        context: input.context,
        rows: input.rowsRef.current,
        queryMembers: members,
        selectedRecordIds: selected.current,
        tagName,
      });
      if (plan.kind === "reject")
        return {
          kind: "rejected",
          message:
            plan.reason === "empty_tag"
              ? "Enter a tag to assign."
              : plan.reason === "empty_selection"
                ? "Select records to assign this tag."
                : plan.reason === "partial_selection" ||
                    plan.reason === "invalid_target"
                  ? "Selection changed before assignment. Review the selected records and try again."
                  : "Tag assignment is no longer available.",
        };
      const reason = getBlockingReason();
      if (reason) return { kind: "rejected", message: reason };
      return input.port.assignTag(
        { tagName: plan.normalizedTagName, targets: plan.targets },
        { delivery, ready: input.precedingSaves() },
      );
    },
    [getBlockingReason],
  );
  return {
    snapshot: { gridSelection },
    controls: {
      assignTag,
      canAssign,
      selectedRecordIds,
      getBlockingReason,
      subscribeReadiness: input.readiness.subscribe,
      operations: input.port,
    },
  };
}
