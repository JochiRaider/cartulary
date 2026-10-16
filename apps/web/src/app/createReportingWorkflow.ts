import { extensionRouteReadiness } from "../extensions/extensionAvailability";
import {
  reportingProfileId,
  reportingReleasesRoute,
  reportingSnapshotsRoute,
} from "../extensions/extensionWorkspaceIdentities";
import { ReportingController } from "../reporting/ReportingController";
import { createReportingClient } from "../services/reportingClient";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import type {
  AppSessionController,
  AppSessionSnapshot,
} from "./appSessionController";
import { createWorkflowPresentation } from "./createWorkflowPresentation";

function eligible(state: AppSessionSnapshot) {
  return (
    state.extensions.kind === "ready" &&
    [reportingSnapshotsRoute, reportingReleasesRoute].every(
      (route) =>
        extensionRouteReadiness(
          state.extensions.kind === "ready" ? state.extensions.value : [],
          reportingProfileId,
          route,
        ) === "available",
    )
  );
}
/** Session authority outlives presentation; no observation restarts on attachment. */
export function createReportingWorkflow(
  options: () => {
    readonly accepting: () => boolean;
    readonly sessionController: AppSessionController;
    readonly currentIncidentId: () => string;
  },
) {
  const controller = new ReportingController();
  let surface: WorkbookIncidentControlsRendererProps | null = null;
  let scope: { incidentId: string; actorId: string; lifetime: string } | null =
    null;
  let disposed = false;
  let apiBase: string | undefined;
  let client = createReportingClient();
  const retire = () => {
    surface = null;
    scope = null;
    controller.retire();
  };
  const synchronize = () => {
    if (disposed || !options().accepting()) return;
    const state = options().sessionController.getSnapshot(),
      incidentId = options().currentIncidentId();
    const member = state.session?.memberships.find(
      (m) => m.incident_id === incidentId,
    );
    if (
      state.state === "anonymous" ||
      state.ended ||
      (scope &&
        (scope.incidentId !== incidentId ||
          (state.lifetime && scope.lifetime !== state.lifetime) ||
          (state.session && scope.actorId !== state.session.user_id))) ||
      (state.session && !member) ||
      (state.extensions.kind === "ready" && !eligible(state))
    ) {
      retire();
      return;
    }
    if (
      !surface ||
      surface.incidentId !== incidentId ||
      state.state !== "authenticated" ||
      state.authenticationTransportPending ||
      !state.session ||
      !state.lifetime ||
      !member ||
      !eligible(state) ||
      surface.currentIncidentRole === null
    ) {
      controller.pause();
      return;
    }
    if (apiBase !== surface.apiBase) {
      controller.retire();
      apiBase = surface.apiBase;
      client = createReportingClient(apiBase);
    }
    const accepted = {
      incidentId,
      actorId: state.session.user_id,
      lifetime: state.lifetime,
    };
    scope = accepted;
    controller.bind({
      ...accepted,
      client,
      role: member.role,
      accessFailure: () => {
        controller.pause();
        void surface?.onSessionRoleChange?.();
      },
      current: () => {
        const now = options().sessionController.getSnapshot();
        return (
          !disposed &&
          options().accepting() &&
          options().currentIncidentId() === accepted.incidentId &&
          now.state === "authenticated" &&
          !now.authenticationTransportPending &&
          eligible(now) &&
          now.lifetime === accepted.lifetime &&
          now.session?.user_id === accepted.actorId &&
          now.session.memberships.some(
            (m) =>
              m.incident_id === accepted.incidentId && m.role === member.role,
          )
        );
      },
    });
    controller.setPresented(
      surface.activeSection === "snapshot-reports" &&
        document.visibilityState !== "hidden",
    );
  };
  const presentation = createWorkflowPresentation(
    () => ({
      ...options(),
      accepting: () => !disposed && options().accepting(),
    }),
    (binding: WorkbookIncidentControlsRendererProps | null) => {
      if (disposed || !options().accepting()) return;
      surface = binding;
      if (!binding) controller.setPresented(false);
      synchronize();
    },
  );
  return {
    controller,
    synchronize,
    retire: () => {
      if (!disposed) {
        presentation.invalidate();
        retire();
      }
    },
    dispose: () => {
      if (!disposed) {
        presentation.invalidate();
        retire();
        disposed = true;
        controller.dispose();
      }
    },
    attachSurface: presentation.attach,
  };
}
