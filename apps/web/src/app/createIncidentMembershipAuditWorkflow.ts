import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import { listIncidentMembershipAuditPage } from "./api/incidentMembershipAuditClient";
import type { AppSessionController } from "./appSessionController";
import { createWorkflowPresentation } from "./createWorkflowPresentation";
import { IncidentMembershipAuditController } from "./incidentMembershipAuditController";

/** App retains exactly one incident's filters; existing session/recovery owners accept authority. */
export function createIncidentMembershipAuditWorkflow(
  options: () => {
    readonly accepting: () => boolean;
    readonly sessionController: AppSessionController;
    readonly recovery: AuthorizationRecoveryPort;
    readonly currentIncidentId: () => string;
    readonly onIncidentAccessLost: () => void;
    readonly onSessionLost: () => void;
  },
) {
  let disposed = false;
  let surface: WorkbookIncidentControlsRendererProps | null = null;
  const controller = new IncidentMembershipAuditController({
    list: listIncidentMembershipAuditPage,
    isCurrent: (authority) => {
      const state = options().sessionController.getSnapshot();
      return (
        !disposed &&
        options().accepting() &&
        options().currentIncidentId() === authority.incidentId &&
        state.lifetime === authority.lifetime &&
        state.session?.user_id === authority.actorId &&
        state.session.memberships.some(
          (member) =>
            member.incident_id === authority.incidentId &&
            member.role === "admin",
        )
      );
    },
    recover: async (authority, signal, admitted) => {
      const result = await options().recovery.recover({
        incidentId: authority.incidentId,
        signal,
      });
      if (!admitted()) return { kind: "cancelled" };
      if (result.kind === "authorized")
        surface?.onAuthorizationRecovered?.(result);
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
      else if (reason === "incident") options().onIncidentAccessLost();
      else void surface?.onSessionRoleChange?.();
    },
  });
  const synchronize = () => {
    if (disposed || !options().accepting()) return;
    const state = options().sessionController.getSnapshot();
    const incidentId = options().currentIncidentId();
    const previous = controller.getSnapshot().authority;
    const membership = state.session?.memberships.find(
      (member) => member.incident_id === incidentId,
    );
    if (
      surface?.incidentId === incidentId &&
      state.session &&
      membership &&
      surface.currentIncidentRole !== membership.role
    ) {
      surface.onAuthorizationRecovered?.({
        role: membership.role,
        userId: state.session.user_id,
      });
    }
    const authority =
      state.lifetime &&
      state.session &&
      incidentId &&
      membership?.role === "admin"
        ? {
            lifetime: state.lifetime,
            actorId: state.session.user_id,
            incidentId,
            apiBase: surface ? surface.apiBase : previous?.apiBase,
          }
        : null;
    controller.setAuthority(authority);
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
      synchronize();
      if (
        props.currentIncidentRole !== null &&
        props.currentIncidentRole !== "admin"
      )
        controller.setAuthority(null);
      controller.setActive(
        props.currentIncidentRole === "admin" &&
          props.activeSection === "membership-audit" &&
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
      controller.retire();
    },
    dispose: () => {
      if (disposed) return;
      presentation.invalidate();
      disposed = true;
      surface = null;
      controller.dispose();
    },
    attachSurface: presentation.attach,
  };
}
