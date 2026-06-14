import { useLayoutEffect, useState } from "react";
import type { WorkbookRow } from "../models/workbookTimelineModel";

type TimelineMutableRef<T> = {
  current: T;
};

export function useTimelineRows({
  draftCounterRef,
  rowsRef,
}: {
  readonly draftCounterRef: TimelineMutableRef<number>;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
}) {
  const [rows, setRows] = useState<WorkbookRow[]>(() => rowsRef.current);

  useLayoutEffect(() => {
    rowsRef.current = rows;
  }, [rows, rowsRef]);

  return {
    commands: {
      setRows,
    },
    refs: {
      draftCounterRef,
      rowsRef,
    },
    snapshot: {
      rows,
    },
  };
}
