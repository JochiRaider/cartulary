import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import type { TimelineRowStoreCommands } from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { WorkbookTimelineMutationOwner } from "../mutations/WorkbookTimelineMutationOwner";

export function useTimelineRows(owner: WorkbookTimelineMutationOwner) {
  const [rows, setRows] = useState<WorkbookRow[]>(owner.initialRows);
  const rowsRef = useRef(rows);
  const nextDraftIndex = owner.capture.allocateDraftIndex;

  const replaceRows = useCallback((nextRows: WorkbookRow[]) => {
    setRows(nextRows);
  }, []);
  const updateRows = useCallback<TimelineRowStoreCommands["updateRows"]>(
    (updater) => {
      setRows((current) => updater(current));
    },
    [],
  );
  const commands = useMemo<TimelineRowStoreCommands>(
    () => ({ replaceRows, updateRows }),
    [replaceRows, updateRows],
  );

  useLayoutEffect(() => {
    rowsRef.current = rows;
  }, [rows]);

  return { commands, nextDraftIndex, rows, rowsRef };
}
