import { useCallback, useLayoutEffect, useRef, useState } from "react";
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

  useLayoutEffect(() => {
    rowsRef.current = rows;
  }, [rows]);

  return { nextDraftIndex, rows, rowsRef, setRows };
}
