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
  useEffect(() => {
    const refresh = () => projection.refreshPresenceTime();
    const visible = () => {
      if (document.visibilityState === "visible") refresh();
    };
    document.addEventListener("visibilitychange", visible);
    window.addEventListener("focus", refresh);
    return () => {
      document.removeEventListener("visibilitychange", visible);
      window.removeEventListener("focus", refresh);
    };
  }, [projection]);
  return useWorkbookCollaborationCoordinator(projection);
}
