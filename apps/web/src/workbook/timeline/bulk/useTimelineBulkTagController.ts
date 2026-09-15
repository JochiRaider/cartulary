import type {
  GridCoreRecordBulkSelection,
  GridDataRow,
} from "@cartulary/grid-adapter";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { useWorkbookQueryPresentation } from "../../query/WorkbookQueryBrowsingContext";
import {
  planTimelineBulkTag,
  type TimelineBulkTagContext,
  type TimelineBulkTagPlan,
} from "../models/timelineBulkTagPlan";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { TimelineBulkTagCommandPort } from "../ports/TimelineBulkTagCommandPort";

type TimelineBulkTagMessage = {
  readonly kind: "error" | "success";
  readonly message: string;
};

type TimelineBulkTagControllerInput = {
  readonly context: TimelineBulkTagContext;
  readonly port: TimelineBulkTagCommandPort;
  readonly precedingSaves: () => Promise<void>;
  readonly rows: readonly WorkbookRow[];
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
};

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
  const [tagName, setTagName] = useState("");
  const [message, setMessage] = useState<TimelineBulkTagMessage | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const inputRef = useRef(input);
  const selectedRecordIdsRef = useRef(selectedRecordIds);
  const tagNameRef = useRef(tagName);
  const submissionInFlightRef = useRef(false);
  inputRef.current = input;
  selectedRecordIdsRef.current = selectedRecordIds;
  tagNameRef.current = tagName;

  useEffect(() => {
    const selectableIds = new Set(
      input.context.authorized && input.context.capabilityAvailable
        ? input.rows.flatMap((row) =>
            row.recordId !== null &&
            row.rowVersion !== null &&
            row.pendingSignature === null &&
            (queryMembers === null || queryMembers.has(row.recordId))
              ? [row.recordId]
              : [],
          )
        : [],
    );
    setSelectedRecordIds((current) => {
      const next = new Set(
        [...current].filter((recordId) => selectableIds.has(recordId)),
      );
      selectedRecordIdsRef.current = next;
      return next.size === current.size ? current : next;
    });
  }, [
    input.context.authorized,
    input.context.capabilityAvailable,
    input.rows,
    queryMembers,
  ]);

  const changeSelectedRecordIds = useCallback(
    (recordIds: ReadonlySet<string>) => {
      const next = new Set(recordIds);
      selectedRecordIdsRef.current = next;
      setSelectedRecordIds(next);
      setMessage(null);
    },
    [],
  );
  const changeTagName = useCallback((value: string) => {
    tagNameRef.current = value;
    setTagName(value);
    setMessage(null);
  }, []);

  const canAssign =
    input.context.authorized && input.context.capabilityAvailable;
  const gridSelection = useMemo<GridCoreRecordBulkSelection<WorkbookRow>>(
    () => ({
      isRecordSelectable: (row: GridDataRow<WorkbookRow>) =>
        canAssign &&
        row.data.pendingSignature === null &&
        (queryMembers === null ||
          (row.data.recordId !== null && queryMembers.has(row.data.recordId))),
      onSelectedRecordIdsChange: changeSelectedRecordIds,
      selectedRecordIds,
    }),
    [canAssign, changeSelectedRecordIds, selectedRecordIds, queryMembers],
  );

  const assignTag = useCallback(async () => {
    if (submissionInFlightRef.current) return;
    const current = inputRef.current;
    const plan = planTimelineBulkTag({
      context: current.context,
      rows: current.rowsRef.current,
      selectedRecordIds: selectedRecordIdsRef.current,
      tagName: tagNameRef.current,
    });
    if (plan.kind === "reject") {
      publishBulkTagRejection(plan.reason, setMessage);
      return;
    }
    submissionInFlightRef.current = true;
    setSubmitting(true);
    setMessage(null);
    current.port.assignTag(
      { tagName: plan.normalizedTagName, targets: plan.targets },
      { delivery: {}, ready: current.precedingSaves() },
    );
    queueMicrotask(() => {
      submissionInFlightRef.current = false;
      setSubmitting(false);
    });
  }, []);

  return {
    commands: { assignTag, changeSelectedRecordIds, changeTagName },
    snapshot: {
      canAssign,
      canSubmit:
        canAssign &&
        !submitting &&
        selectedRecordIds.size > 0 &&
        tagName.trim() !== "",
      gridSelection,
      message,
      selectedRecordIds,
      submitting,
      tagName,
    },
  };
}

function publishBulkTagRejection(
  reason: Extract<TimelineBulkTagPlan, { kind: "reject" }>["reason"],
  setMessage: (message: TimelineBulkTagMessage | null) => void,
): void {
  if (reason === "empty_tag" || reason === "empty_selection") return;
  setMessage({
    kind: "error",
    message:
      reason === "partial_selection" || reason === "invalid_target"
        ? "Selection changed before the command could be submitted. Review the selected rows and try again."
        : "Tag assignment is no longer available.",
  });
}
