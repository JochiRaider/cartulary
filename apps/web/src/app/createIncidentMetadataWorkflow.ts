import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { IncidentResource } from "../shared/incidentResource";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import {
  patchIncidentMetadata,
  readIncidentMetadata,
} from "./api/incidentMetadataClient";
import type { AppSessionController } from "./appSessionController";
import { createWorkflowPresentation } from "./createWorkflowPresentation";
import { IncidentMetadataController } from "./incidentMetadataController";
import type { MetadataAuthority } from "./incidentMetadataModel";

/** Retains one workflow above drawer mounts; App remains the session authority. */
export function createIncidentMetadataWorkflow(
  options: () => {
    readonly accepting: () => boolean;
    onResourceAccepted: (
      resource: IncidentResource,
      authority: MetadataAuthority,
    ) => void;
    sessionController: AppSessionController;
    recovery: AuthorizationRecoveryPort;
    currentIncidentId: () => string;
    onIncidentAccessLost: () => void;
    onSessionLost: () => void;
  },
) {
  let disposed = false;
  let surface: WorkbookIncidentControlsRendererProps | null = null;
  let workbook: WorkbookIncidentControlsRendererProps | null = null;
  const controller = new IncidentMetadataController({
    read: readIncidentMetadata,
    patch: patchIncidentMetadata,
    isCurrent: (authority) => {
      const state = options().sessionController.getSnapshot();
      return (
        !disposed &&
        options().accepting() &&
        options().currentIncidentId() === authority.incidentId &&
        state.lifetime === authority.lifetime &&
        state.session?.user_id === authority.actorId &&
        state.session.memberships.some(
          (m) => m.incident_id === authority.incidentId,
        )
      );
    },
    recover: async (authority, signal, admitted) => {
      const result = await options().recovery.recover({
        incidentId: authority.incidentId,
        signal,
      });
      if (!admitted()) return { kind: "cancelled" };
      if (
        result.kind === "authorized" &&
        workbook?.incidentId === authority.incidentId
      )
        workbook.onAuthorizationRecovered?.(result);
      return result;
    },
    lost: (reason, authority) => {
      if (
        options().sessionController.getSnapshot().lifetime !==
          authority.lifetime ||
        options().currentIncidentId() !== authority.incidentId
      )
        return;
      if (reason === "session") options().onSessionLost();
      else options().onIncidentAccessLost();
    },
    publishResource: (resource, authority) => {
      options().onResourceAccepted(resource, authority);
    },
  });
  const synchronize = () => {
    if (disposed || !options().accepting()) return;
    const state = options().sessionController.getSnapshot();
    const incidentId = options().currentIncidentId();
    const previous = controller.getSnapshot().authority;
    const membership = state.session?.memberships.find(
      (m) => m.incident_id === incidentId,
    );
    if (workbook?.incidentId !== incidentId) workbook = null;
    if (
      workbook &&
      state.session &&
      membership &&
      previous?.role !== membership.role
    )
      workbook.onAuthorizationRecovered?.({
        role: membership.role,
        userId: state.session.user_id,
      });
    controller.setAuthority(
      state.lifetime && state.session && membership
        ? {
            incidentId,
            lifetime: state.lifetime,
            actorId: state.session.user_id,
            role: membership.role,
            apiBase:
              surface?.incidentId === incidentId
                ? surface.apiBase
                : previous?.apiBase,
          }
        : null,
    );
    if (
      previous &&
      previous.incidentId === incidentId &&
      state.lifetime === previous.lifetime &&
      state.session &&
      !membership
    )
      options().onIncidentAccessLost();
  };
  const presentation = createWorkflowPresentation(
    () => ({
      ...options(),
      accepting: () => !disposed && options().accepting(),
    }),
    (props: WorkbookIncidentControlsRendererProps | null) => {
      if (disposed || !options().accepting()) return;
      surface = props;
      if (!props) {
        controller.setActive(false);
        return;
      }
      workbook = props;
      synchronize();
      controller.setActive(
        props.activeSection === "incident-fields" &&
          document.visibilityState !== "hidden",
      );
    },
  );
  return {
    controller,
    synchronize: synchronize,
    retire: () => {
      if (disposed) return;
      presentation.invalidate();
      surface = null;
      workbook = null;
      controller.retire();
    },
    dispose: () => {
      if (disposed) return;
      presentation.invalidate();
      disposed = true;
      surface = null;
      workbook = null;
      controller.dispose();
    },
    attachSurface: presentation.attach,
  };
}
