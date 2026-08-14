import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import type { TimelineRowStoreCommands } from "../models/timelineControllerPorts";
import {
  createDraftRow,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

export function useTimelineRows() {
  const [rows, setRows] = useState<WorkbookRow[]>(() => [createDraftRow(1)]);
  const rowsRef = useRef(rows);
  const draftCounterRef = useRef(2);

  const nextDraftIndex = useCallback(() => {
    const value = draftCounterRef.current;
    draftCounterRef.current += 1;
    return value;
  }, []);

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
