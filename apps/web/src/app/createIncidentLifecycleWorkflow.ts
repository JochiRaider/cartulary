import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { IncidentResource } from "../shared/incidentResource";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import {
  mutateIncidentLifecycle,
  readLifecycleIncident,
} from "./api/incidentLifecycleClient";
import type { AppSessionController } from "./appSessionController";
import { createWorkflowPresentation } from "./createWorkflowPresentation";
import { IncidentLifecycleController } from "./incidentLifecycleController";
import type { LifecycleAuthority } from "./incidentLifecycleModel";

/** App owns lifetime and publication; the drawer borrows the retained workflow. */
export function createIncidentLifecycleWorkflow(
  options: () => {
    readonly accepting: () => boolean;
    sessionController: AppSessionController;
    recovery: AuthorizationRecoveryPort;
    currentIncidentId: () => string;
    onIncidentAccessLost: () => void;
    onSessionLost: () => void;
    onResourceAccepted: (
      resource: IncidentResource,
      authority: LifecycleAuthority,
    ) => void;
  },
) {
  let disposed = false;
  let surface: WorkbookIncidentControlsRendererProps | null = null;
  let workbook: WorkbookIncidentControlsRendererProps | null = null;
  const controller = new IncidentLifecycleController({
    read: readLifecycleIncident,
    mutate: mutateIncidentLifecycle,
    isCurrent: (authority) => {
      const s = options().sessionController.getSnapshot();
      return (
        !disposed &&
        options().accepting() &&
        options().currentIncidentId() === authority.incidentId &&
        s.lifetime === authority.lifetime &&
        s.session?.user_id === authority.actorId &&
        s.session.memberships.some(
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
      const state = options().sessionController.getSnapshot();
      if (
        state.lifetime !== authority.lifetime ||
        options().currentIncidentId() !== authority.incidentId
      )
        return;
      if (reason === "session") options().onSessionLost();
      else options().onIncidentAccessLost();
    },
    publishResource: (resource, authority) =>
      options().onResourceAccepted(resource, authority),
  });
  const synchronize = () => {
    if (disposed || !options().accepting()) return;
    const state = options().sessionController.getSnapshot();
    const incidentId = options().currentIncidentId();
    const previous = controller.getSnapshot().authority;
    const member = state.session?.memberships.find(
      (m) => m.incident_id === incidentId,
    );
    if (workbook?.incidentId !== incidentId) workbook = null;
    if (workbook && state.session && member && previous?.role !== member.role)
      workbook.onAuthorizationRecovered?.({
        role: member.role,
        userId: state.session.user_id,
      });
    controller.setAuthority(
      state.lifetime && state.session && member
        ? {
            incidentId,
            lifetime: state.lifetime,
            actorId: state.session.user_id,
            role: member.role,
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
      !member
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
        props.activeSection === "summary" &&
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
