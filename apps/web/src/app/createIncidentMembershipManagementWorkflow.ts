import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import {
  listIncidentMembershipPage,
  mutateIncidentMembership,
} from "./api/incidentMembershipManagementClient";
import type { AppSessionController } from "./appSessionController";
import { createWorkflowPresentation } from "./createWorkflowPresentation";
import { IncidentMembershipManagementController } from "./incidentMembershipManagementController";

/** App owns the sole session lifetime; this binding retains just one incident workflow. */
export function createIncidentMembershipManagementWorkflow(
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
  let workbook: Pick<
    WorkbookIncidentControlsRendererProps,
    "incidentId" | "onAuthorizationRecovered"
  > | null = null;
  const controller = new IncidentMembershipManagementController({
    list: listIncidentMembershipPage,
    mutate: mutateIncidentMembership,
    isCurrent: (authority) => {
      const state = options().sessionController.getSnapshot();
      return (
        !disposed &&
        options().accepting() &&
        options().currentIncidentId() === authority.incidentId &&
        state.lifetime === authority.lifetime &&
        state.session?.user_id === authority.actorId &&
        state.session.memberships.some(
          (member) => member.incident_id === authority.incidentId,
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
  });
  const synchronize = () => {
    if (disposed || !options().accepting()) return;
    const state = options().sessionController.getSnapshot();
    const incidentId = options().currentIncidentId();
    const previous = controller.getSnapshot().authority;
    const membership = state.session?.memberships.find(
      (member) => member.incident_id === incidentId,
    );
    if (workbook && workbook.incidentId !== incidentId) workbook = null;
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
            lifetime: state.lifetime,
            actorId: state.session.user_id,
            incidentId,
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
      workbook = {
        incidentId: props.incidentId,
        onAuthorizationRecovered: props.onAuthorizationRecovered,
      };
      synchronize();
      const authority = controller.getSnapshot().authority;
      if (
        authority &&
        props.currentIncidentRole &&
        props.currentIncidentRole !== "admin"
      )
        controller.setAuthority({
          ...authority,
          role: props.currentIncidentRole,
        });
      controller.setActive(
        props.activeSection === "memberships" &&
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
