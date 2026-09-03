import { useCallback, useEffect, useMemo, useState } from "react";
import type { useIncidentCollaborationSession } from "../../collaboration/IncidentCollaborationSession";
import type { AuthorizationRecoveryPort } from "../../shared/authorizationRecovery";
import { type SheetRef, sheetRefKey } from "../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import { useWorkbookCollaborationCoordinatorSession } from "../collaboration/useWorkbookCollaborationCoordinator";
import { WorkbookCollaborationCoordinator } from "../collaboration/WorkbookCollaborationCoordinator";
import type { WorkbookActiveSurfacePort } from "../collaboration/workbookSurfacePort";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";

type WorkbookCollaborationLifecycleOptions = {
  readonly activeSurfacePort: WorkbookActiveSurfacePort | null;
  readonly authorizationRecovery: AuthorizationRecoveryPort;
  readonly cancelGridEntryFocus: () => void;
  readonly collaborationSession: ReturnType<
    typeof useIncidentCollaborationSession
  >;
  readonly extensionInvalidation: () => void;
  readonly incidentId: string;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onAuthorizationRecovered: (result: {
    readonly role: WorkbookIncidentRole;
    readonly userId: string;
  }) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly queryInvalidation: (reason: WorkbookQueryInvalidationReason) => void;
  readonly sheetRef: SheetRef;
  readonly sheetReloadToken: number;
  readonly surface: string;
};

/** Owns collaboration invalidation, session projection, and surface registration. */
export function useWorkbookCollaborationLifecycle({
  activeSurfacePort,
  authorizationRecovery,
  cancelGridEntryFocus,
  collaborationSession,
  extensionInvalidation,
  incidentId,
  mutationRuntime,
  onAuthorizationRecovered,
  onIncidentAccessLost,
  queryInvalidation,
  sheetRef,
  sheetReloadToken,
  surface,
}: WorkbookCollaborationLifecycleOptions) {
  const [evidenceGeneration, setEvidenceGeneration] = useState(0);
  const [inspectorGeneration, setInspectorGeneration] = useState(0);
  const [continuityGeneration, setContinuityGeneration] = useState(0);
  const continuityInvalidation = useCallback(() => {
    cancelGridEntryFocus();
    setContinuityGeneration((current) => current + 1);
  }, [cancelGridEntryFocus]);
  const evidenceInvalidation = useCallback(() => {
    setEvidenceGeneration((current) => current + 1);
  }, []);
  const inspectorInvalidation = useCallback(() => {
    setInspectorGeneration((current) => current + 1);
  }, []);
  const projection = useMemo(
    () =>
      new WorkbookCollaborationCoordinator({
        authorizationRecovery,
        continuityInvalidation,
        evidenceInvalidation,
        extensionInvalidation,
        incidentId,
        initialSheetRef: {
          kind: "view_schema",
          id: timelineViewSchemaId,
        },
        inspectorInvalidation,
        mutationRuntime,
        onAuthorizationRecovered,
        onIncidentAccessLost,
        queryInvalidation,
      }),
    [
      authorizationRecovery,
      continuityInvalidation,
      evidenceInvalidation,
      extensionInvalidation,
      incidentId,
      inspectorInvalidation,
      mutationRuntime,
      onAuthorizationRecovered,
      onIncidentAccessLost,
      queryInvalidation,
    ],
  );
  const snapshot = useWorkbookCollaborationCoordinatorSession({
    projection,
    session: collaborationSession,
    sheetRef,
  });

  useEffect(() => {
    if (activeSurfacePort === null) return;
    return projection.registerActiveSurface(activeSurfacePort);
  }, [activeSurfacePort, projection]);

  const subjectKey = `${surface}:${sheetRefKey(sheetRef)}:${sheetReloadToken}`;
  return {
    continuityResetKey: `${subjectKey}:${continuityGeneration}`,
    inspectorResetKey: `${subjectKey}:${inspectorGeneration}:${evidenceGeneration}`,
    projection,
    snapshot,
  };
}
