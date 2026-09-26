import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import { createWorkbookSavedViewAdapter } from "../workbook/adapters/createWorkbookSavedViewAdapter";
import type { SavedViewBinding } from "../workbook/savedviews/savedViewOperationModel";
import { WorkbookSavedViewController } from "../workbook/savedviews/WorkbookSavedViewController";
import { observeAccountOperation } from "./accountOperation";
import type { AppSessionController } from "./appSessionController";
import { createWorkflowPresentation } from "./createWorkflowPresentation";

export function createWorkbookSavedViewsWorkflow(
  options: () => {
    readonly accepting: () => boolean;
    sessionController: AppSessionController;
    recovery: AuthorizationRecoveryPort;
    currentIncidentId: () => string;
    onIncidentAccessLost: () => void;
    onSessionLost: () => void;
  },
) {
  let disposed = false;
  let binding: SavedViewBinding | null = null;
  const controller = new WorkbookSavedViewController({
    observe: observeAccountOperation,
    port: (authority) =>
      createWorkbookSavedViewAdapter({
        apiBase: authority.apiBase,
        incidentId: authority.incidentId,
      }),
    isCurrent: (authority) => {
      const session = options().sessionController.getSnapshot();
      return (
        !disposed &&
        options().accepting() &&
        options().currentIncidentId() === authority.incidentId &&
        session.lifetime === authority.lifetime &&
        session.session?.user_id === authority.actorId &&
        session.session.memberships.some(
          (member) => member.incident_id === authority.incidentId,
        )
      );
    },
    recover: async (authority, signal, admitted) => {
      const result = await options().recovery.recover({
        incidentId: authority.incidentId,
        signal,
      });
      return admitted() ? result : { kind: "cancelled" };
    },
    lost: (reason, authority) => {
      const session = options().sessionController.getSnapshot();
      if (
        session.lifetime !== authority.lifetime ||
        options().currentIncidentId() !== authority.incidentId
      )
        return;
      if (reason === "session") options().onSessionLost();
      else options().onIncidentAccessLost();
    },
  });
  const sync = () => {
    if (disposed || !options().accepting()) return;
    const session = options().sessionController.getSnapshot();
    const incidentId = options().currentIncidentId();
    const membership = session.session?.memberships.find(
      (member) => member.incident_id === incidentId,
    );
    if (binding?.incidentId !== incidentId) binding = null;
    controller.setAuthority(
      session.session && session.lifetime && membership
        ? {
            incidentId,
            actorId: session.session.user_id,
            lifetime: session.lifetime,
            role: membership.role,
            apiBase:
              binding?.apiBase ?? controller.getSnapshot().authority?.apiBase,
          }
        : null,
    );
  };
  const presentation = createWorkflowPresentation(
    () => ({
      ...options(),
      accepting: () => !disposed && options().accepting(),
    }),
    (next: SavedViewBinding | null) => {
      if (disposed || !options().accepting()) return;
      if (next && next.incidentId !== options().currentIncidentId()) return;
      binding = next;
      sync();
      controller.setBinding(next);
    },
  );
  return {
    controller,
    synchronize: sync,
    retire: () => {
      if (disposed) return;
      presentation.invalidate();
      binding = null;
      controller.retire();
    },
    dispose: () => {
      if (disposed) return;
      presentation.invalidate();
      disposed = true;
      binding = null;
      controller.dispose();
    },
    attachWorkbook: presentation.attach,
  };
}
