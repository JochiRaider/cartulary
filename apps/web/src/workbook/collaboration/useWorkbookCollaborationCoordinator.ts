import { useEffect, useSyncExternalStore } from "react";
import type { IncidentCollaborationSessionValue } from "../../collaboration/IncidentCollaborationSession";
import type { WorkbookSheetRef } from "../../shared/workbookSheetRef";
import type { WorkbookCollaborationCoordinator } from "./WorkbookCollaborationCoordinator";

export function useWorkbookCollaborationCoordinator(
  projection: WorkbookCollaborationCoordinator,
) {
  return useSyncExternalStore(
    projection.subscribe,
    projection.getSnapshot,
    projection.getSnapshot,
  );
}

export function useWorkbookCollaborationCoordinatorSession({
  projection,
  session,
  sheetRef,
}: {
  readonly projection: WorkbookCollaborationCoordinator;
  readonly session: IncidentCollaborationSessionValue;
  readonly sheetRef: WorkbookSheetRef;
}) {
  useEffect(() => projection.retain(), [projection]);
  useEffect(() => projection.attachSession(session), [projection, session]);
  useEffect(() => {
    projection.setActiveSheet(sheetRef);
  }, [projection, sheetRef]);
  return useWorkbookCollaborationCoordinator(projection);
}
