import { useEffect, useSyncExternalStore } from "react";
import type { IncidentCollaborationSessionValue } from "../../collaboration/IncidentCollaborationSession";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookCollaborationCoordinator } from "./WorkbookCollaborationCoordinator";

export type WorkbookCollaborationStore = Pick<
  WorkbookCollaborationCoordinator,
  "getSnapshot" | "subscribe"
>;

export function useWorkbookCollaborationCoordinator(
  projection: WorkbookCollaborationStore,
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
  readonly sheetRef: SheetRef;
}) {
  useEffect(() => projection.retain(), [projection]);
  useEffect(() => projection.attachSession(session), [projection, session]);
  useEffect(() => {
    projection.setActiveSheet(sheetRef);
  }, [projection, sheetRef]);
  return useWorkbookCollaborationCoordinator(projection);
}
