import type { ReactNode } from "react";
import {
  IncidentCollaborationBoundary,
  useIncidentCollaborationSession,
} from "../../../collaboration/IncidentCollaborationSession";
import type { SheetRef } from "../../../shared/sheetRef";
import { useWorkbookCollaborationCoordinatorSession } from "../../collaboration/useWorkbookCollaborationCoordinator";
import type { WorkbookCollaborationCoordinator } from "../../collaboration/WorkbookCollaborationCoordinator";

export function TimelineCollaborationBoundary({
  apiBase,
  attachSession,
  children,
  incidentId,
  projection,
  sheetRef,
}: {
  readonly apiBase?: string | undefined;
  readonly attachSession: boolean;
  readonly children: ReactNode;
  readonly incidentId: string;
  readonly projection: WorkbookCollaborationCoordinator;
  readonly sheetRef: SheetRef;
}) {
  return (
    <IncidentCollaborationBoundary
      apiBase={apiBase}
      incidentId={incidentId}
      initialPresence={{
        sheet_ref: sheetRef,
        mode: "viewing",
      }}
    >
      {attachSession ? (
        <TimelineCollaborationSessionAttachment
          projection={projection}
          sheetRef={sheetRef}
        >
          {children}
        </TimelineCollaborationSessionAttachment>
      ) : (
        children
      )}
    </IncidentCollaborationBoundary>
  );
}

function TimelineCollaborationSessionAttachment({
  children,
  projection,
  sheetRef,
}: {
  readonly children: ReactNode;
  readonly projection: WorkbookCollaborationCoordinator;
  readonly sheetRef: SheetRef;
}) {
  const session = useIncidentCollaborationSession();
  useWorkbookCollaborationCoordinatorSession({
    projection,
    session,
    sheetRef,
  });
  return children;
}
