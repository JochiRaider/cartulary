import { useCallback, useReducer, useRef } from "react";
import type { WorkbookRecordSubject } from "../ports/WorkbookRecordSubject";
import {
  initialWorkbookRecordHistoryState,
  type WorkbookRecordHistoryEvent,
  workbookRecordHistoryReducer,
} from "./workbookRecordHistoryModel";

/** One state instance shared by its subject composition and History binding. */
export function useWorkbookRecordHistoryState(
  subject: WorkbookRecordSubject | null = null,
) {
  const [snapshot, reactDispatch] = useReducer(
    workbookRecordHistoryReducer,
    subject,
    initialWorkbookRecordHistoryState,
  );
  const current = useRef(snapshot);
  current.current = snapshot;
  const dispatch = useCallback((event: WorkbookRecordHistoryEvent) => {
    current.current = workbookRecordHistoryReducer(current.current, event);
    reactDispatch(event);
    return current.current;
  }, []);
  return { snapshot, dispatch };
}
