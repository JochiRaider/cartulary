import { extensionRouteReadiness } from "../extensions/extensionAvailability";
import {
  importProfileId,
  importRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import { WorkbookImportController } from "../imports/WorkbookImportController";
import type { WorkbookImportSurfaceBinding } from "../imports/workbookImportBinding";
import type {
  AppSessionController,
  AppSessionSnapshot,
} from "./appSessionController";
import { createWorkflowPresentation } from "./createWorkflowPresentation";

function currentImportEligibility(session: AppSessionSnapshot) {
  return (
    session.extensions.kind === "ready" &&
    extensionRouteReadiness(
      session.extensions.value,
      importProfileId,
      importRouteFamily,
    ) === "available"
  );
}

export function createWorkbookImportWorkflow(
  options: () => {
    readonly accepting: () => boolean;
    readonly sessionController: AppSessionController;
    readonly currentIncidentId: () => string;
  },
) {
  const controller = new WorkbookImportController();
  let surface: WorkbookImportSurfaceBinding | null = null;
  let scope: { incidentId: string; actorId: string; lifetime: string } | null =
    null;
  let disposed = false;
  const retire = () => {
    if (disposed) return;
    surface = null;
    scope = null;
    controller.retire();
  };
  const synchronize = () => {
    if (disposed || !options().accepting()) return;
    const session = options().sessionController.getSnapshot();
    const incidentId = options().currentIncidentId();
    if (
      session.state === "anonymous" ||
      session.ended ||
      (scope &&
        (scope.incidentId !== incidentId ||
          (session.lifetime && scope.lifetime !== session.lifetime) ||
          (session.session && scope.actorId !== session.session.user_id)))
    ) {
      retire();
      return;
    }
    // Definitive session facts retire even while the presentation is detached.
    if (
      session.extensions.kind === "ready" &&
      !currentImportEligibility(session)
    ) {
      retire();
      return;
    }
    if (
      session.session &&
      !session.session.memberships.some((m) => m.incident_id === incidentId)
    ) {
      retire();
      return;
    }
    const binding = surface;
    if (binding && binding.incidentId !== incidentId) {
      retire();
      return;
    }
    if (
      binding?.readiness === "invalid" ||
      binding?.readiness === "unavailable"
    ) {
      retire();
      return;
    }
    if (
      !binding ||
      binding.readiness === "pending" ||
      !session.session ||
      !session.lifetime ||
      session.authenticationTransportPending ||
      session.extensions.kind !== "ready" ||
      session.state !== "authenticated"
    ) {
      controller.pause();
      return;
    }
    const acceptedScope = {
      incidentId,
      actorId: session.session.user_id,
      lifetime: session.lifetime,
    };
    scope = acceptedScope;
    const member = session.session.memberships.find(
      (m) => m.incident_id === incidentId,
    );
    controller.bind({
      ...binding,
      available: true,
      scope: acceptedScope,
      role: member ? (binding.role === null ? null : member.role) : "",
      current: () => {
        const now = options().sessionController.getSnapshot();
        return (
          !disposed &&
          options().accepting() &&
          options().currentIncidentId() === acceptedScope.incidentId &&
          now.state === "authenticated" &&
          !now.authenticationTransportPending &&
          currentImportEligibility(now) &&
          now.lifetime === acceptedScope.lifetime &&
          now.session?.user_id === acceptedScope.actorId
        );
      },
    });
  };
  const presentation = createWorkflowPresentation(
    () => ({
      ...options(),
      accepting: () => !disposed && options().accepting(),
    }),
    (binding: WorkbookImportSurfaceBinding | null) => {
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
      presentation.invalidate();
      retire();
    },
    dispose: () => {
      if (disposed) return;
      presentation.invalidate();
      disposed = true;
      surface = null;
      scope = null;
      controller.dispose();
    },
    attachWorkbook: presentation.attach,
  };
}
