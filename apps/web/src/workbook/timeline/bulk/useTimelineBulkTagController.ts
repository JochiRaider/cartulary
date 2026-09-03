import type {
  GridCoreRecordBulkSelection,
  GridDataRow,
} from "@cartulary/grid-adapter";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  planTimelineBulkTag,
  type TimelineBulkTagContext,
  type TimelineBulkTagPlan,
  timelineBulkTagSubmissionIsCurrent,
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
  readonly refreshRows: () => Promise<void>;
  readonly rows: readonly WorkbookRow[];
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
};

export function useTimelineBulkTagController(
  input: TimelineBulkTagControllerInput,
) {
  const [selectedRecordIds, setSelectedRecordIds] = useState<
    ReadonlySet<string>
  >(() => new Set());
  const [tagName, setTagName] = useState("");
  const [message, setMessage] = useState<TimelineBulkTagMessage | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const activeRef = useRef(false);
  const inputRef = useRef(input);
  const selectedRecordIdsRef = useRef(selectedRecordIds);
  const tagNameRef = useRef(tagName);
  const submissionInFlightRef = useRef(false);
  inputRef.current = input;
  selectedRecordIdsRef.current = selectedRecordIds;
  tagNameRef.current = tagName;

  useEffect(() => {
    activeRef.current = true;
    return () => {
      activeRef.current = false;
    };
  }, []);

  useEffect(() => {
    const selectableIds = new Set(
      input.context.authorized && input.context.capabilityAvailable
        ? input.rows.flatMap((row) =>
            row.recordId !== null &&
            row.rowVersion !== null &&
            row.pendingSignature === null
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
  }, [input.context.authorized, input.context.capabilityAvailable, input.rows]);

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
        canAssign && row.data.pendingSignature === null,
      onSelectedRecordIdsChange: changeSelectedRecordIds,
      selectedRecordIds,
    }),
    [canAssign, changeSelectedRecordIds, selectedRecordIds],
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
    await executeBulkTagSubmission({
      activeRef,
      inputRef,
      plan,
      selectedRecordIdsRef,
      setMessage,
      setSubmitting,
      submissionInFlightRef,
      tagNameRef,
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

async function executeBulkTagSubmission(options: {
  readonly activeRef: { readonly current: boolean };
  readonly inputRef: { readonly current: TimelineBulkTagControllerInput };
  readonly plan: Extract<TimelineBulkTagPlan, { kind: "dispatch" }>;
  readonly selectedRecordIdsRef: { readonly current: ReadonlySet<string> };
  readonly setMessage: (message: TimelineBulkTagMessage | null) => void;
  readonly setSubmitting: (submitting: boolean) => void;
  readonly submissionInFlightRef: { current: boolean };
  readonly tagNameRef: { readonly current: string };
}): Promise<void> {
  const result = await options.inputRef.current.port.assignTag({
    tagName: options.plan.normalizedTagName,
    targets: options.plan.targets,
  });
  options.submissionInFlightRef.current = false;
  if (!options.activeRef.current) return;
  const current = options.inputRef.current;
  const stillAuthorized =
    current.context.authorized &&
    current.context.capabilityAvailable &&
    current.context.surfaceKey === options.plan.surfaceKey;
  options.setSubmitting(false);
  if (!stillAuthorized) return;
  if (result.kind === "rejected") {
    if (submissionIsCurrent(options)) {
      options.setMessage({ kind: "error", message: result.failure.message });
    }
    return;
  }
  await current.refreshRows();
  if (!options.activeRef.current || !submissionIsCurrent(options)) return;
  options.setMessage(bulkTagAcceptedMessage(result.value, options.plan));
}

function submissionIsCurrent(options: {
  readonly inputRef: { readonly current: TimelineBulkTagControllerInput };
  readonly plan: Extract<TimelineBulkTagPlan, { kind: "dispatch" }>;
  readonly selectedRecordIdsRef: { readonly current: ReadonlySet<string> };
  readonly tagNameRef: { readonly current: string };
}): boolean {
  return timelineBulkTagSubmissionIsCurrent({
    context: options.inputRef.current.context,
    plan: options.plan,
    selectedRecordIds: options.selectedRecordIdsRef.current,
    tagName: options.tagNameRef.current,
  });
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

function bulkTagAcceptedMessage(
  accepted: {
    readonly affectedRowCount: number;
    readonly conflictCount: number;
  },
  plan: Extract<TimelineBulkTagPlan, { kind: "dispatch" }>,
): TimelineBulkTagMessage {
  if (accepted.conflictCount > 0) {
    return {
      kind: "error",
      message: `Assigned tag to ${accepted.affectedRowCount} selected record${accepted.affectedRowCount === 1 ? "" : "s"}; ${accepted.conflictCount} ${accepted.conflictCount === 1 ? "record changed and needs" : "records changed and need"} review.`,
    };
  }
  return {
    kind: "success",
    message: `Assigned tag to ${plan.selectedCount} selected record${plan.selectedCount === 1 ? "" : "s"}.`,
  };
}
