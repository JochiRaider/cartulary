import { useState } from "react";
import type {
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../components/TimelineHistoryPanel";

export function useTimelineHistoryState() {
  const [rowHistory, setRowHistory] = useState<RecordHistoryState>({
    recordId: null,
    status: "idle",
    data: null,
    message: null,
  });
  const [rowHistoryPendingAction, setRowHistoryPendingAction] =
    useState<RowHistoryPendingAction | null>(null);

  return {
    commands: {
      setRowHistory,
      setRowHistoryPendingAction,
    },
    snapshot: {
      rowHistory,
      rowHistoryPendingAction,
    },
  };
}
