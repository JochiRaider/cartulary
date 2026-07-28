import { useEffect, useSyncExternalStore } from "react";
import type { IncidentCollaborationSessionValue } from "../../collaboration/IncidentCollaborationSession";
import type { WorkbookSheetRef } from "../models/workbookStartup";
import type { WorkbookCollaborationProjection } from "./WorkbookCollaborationProjection";

export function useWorkbookCollaborationProjection(
  projection: WorkbookCollaborationProjection,
) {
  return useSyncExternalStore(
    projection.subscribe,
    projection.getSnapshot,
    projection.getSnapshot,
  );
}

export function useWorkbookCollaborationProjectionSession({
  projection,
  session,
  sheetRef,
}: {
  readonly projection: WorkbookCollaborationProjection;
  readonly session: IncidentCollaborationSessionValue;
  readonly sheetRef: WorkbookSheetRef;
}) {
  useEffect(() => projection.attachSession(session), [projection, session]);
  useEffect(() => {
    projection.setActiveSheet(sheetRef);
  }, [projection, sheetRef]);
  useEffect(() => () => projection.dispose(), [projection]);
  return useWorkbookCollaborationProjection(projection);
}
