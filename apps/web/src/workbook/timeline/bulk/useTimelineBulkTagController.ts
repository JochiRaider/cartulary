import type {
  GridCoreRecordBulkSelection,
  GridDataRow,
} from "@cartulary/grid-adapter";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { TimelineBulkMutationPort } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookRow } from "../models/workbookTimelineModel";

type TimelineBulkTagPort = Pick<TimelineBulkMutationPort, "assignTag">;

export type TimelineBulkTagMessage = {
  readonly kind: "error" | "success";
  readonly message: string;
};

export function useTimelineBulkTagController({
  canAssign,
  port,
  refreshRows,
  rows,
  rowsRef,
}: {
  readonly canAssign: boolean;
  readonly port: TimelineBulkTagPort;
  readonly refreshRows: () => Promise<void>;
  readonly rows: readonly WorkbookRow[];
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
}) {
  const [selectedRecordIds, setSelectedRecordIds] = useState<
    ReadonlySet<string>
  >(() => new Set());
  const [tagName, setTagName] = useState("");
  const [message, setMessage] = useState<TimelineBulkTagMessage | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const activeRef = useRef(false);
  const canAssignRef = useRef(canAssign);
  const generationRef = useRef(0);
  const submissionInFlightRef = useRef(false);

  useEffect(() => {
    if (canAssignRef.current !== canAssign) {
      canAssignRef.current = canAssign;
      generationRef.current += 1;
    }
  }, [canAssign]);
  useEffect(() => {
    activeRef.current = true;
    return () => {
      activeRef.current = false;
      generationRef.current += 1;
    };
  }, []);

  useEffect(() => {
    const selectableIds = new Set(
      canAssign
        ? rows.flatMap((row) =>
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
      return next.size === current.size ? current : next;
    });
  }, [canAssign, rows]);

  const changeSelectedRecordIds = useCallback(
    (recordIds: ReadonlySet<string>) => {
      setSelectedRecordIds(new Set(recordIds));
      setMessage(null);
    },
    [],
  );
  const changeTagName = useCallback((value: string) => {
    setTagName(value);
    setMessage(null);
  }, []);

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
    const normalizedTagName = tagName.trim();
    if (
      !canAssignRef.current ||
      normalizedTagName === "" ||
      selectedRecordIds.size === 0 ||
      submissionInFlightRef.current
    ) {
      return;
    }
    const selectedRows = rowsRef.current.filter(
      (row) =>
        row.recordId !== null &&
        selectedRecordIds.has(row.recordId) &&
        row.rowVersion !== null &&
        row.pendingSignature === null,
    );
    if (selectedRows.length !== selectedRecordIds.size) {
      setMessage({
        kind: "error",
        message:
          "Selection changed before the command could be submitted. Review the selected rows and try again.",
      });
      return;
    }

    const generation = generationRef.current;
    submissionInFlightRef.current = true;
    setSubmitting(true);
    setMessage(null);
    const result = await port.assignTag({
      tagName: normalizedTagName,
      targets: selectedRows.map((row) => ({
        recordId: row.recordId ?? "",
        baseRowVersion: row.rowVersion ?? 0,
      })),
    });
    submissionInFlightRef.current = false;
    if (
      !activeRef.current ||
      generation !== generationRef.current ||
      !canAssignRef.current
    ) {
      if (activeRef.current) setSubmitting(false);
      return;
    }
    if (result.kind === "rejected") {
      setSubmitting(false);
      setMessage({ kind: "error", message: result.failure.message });
      return;
    }
    await refreshRows();
    if (
      !activeRef.current ||
      generation !== generationRef.current ||
      !canAssignRef.current
    ) {
      if (activeRef.current) setSubmitting(false);
      return;
    }
    setSubmitting(false);
    if (result.value.conflictCount > 0) {
      setMessage({
        kind: "error",
        message: `Assigned tag to ${result.value.affectedRowCount} selected record${result.value.affectedRowCount === 1 ? "" : "s"}; ${result.value.conflictCount} ${result.value.conflictCount === 1 ? "record changed and needs" : "records changed and need"} review.`,
      });
      return;
    }
    setMessage({
      kind: "success",
      message: `Assigned tag to ${selectedRows.length} selected record${selectedRows.length === 1 ? "" : "s"}.`,
    });
  }, [port, refreshRows, rowsRef, selectedRecordIds, tagName]);

  return {
    commands: {
      assignTag,
      changeSelectedRecordIds,
      changeTagName,
    },
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
