import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import { createWorkbookPreferenceAdapter } from "../workbook/adapters/createWorkbookPreferenceAdapter";
import { WorkbookPreferenceController } from "../workbook/preferences/WorkbookPreferenceController";
import type { PreferenceWorkbookBinding } from "../workbook/preferences/workbookPreferenceModel";
import { observeAccountOperation } from "./accountOperation";
import type { AppSessionController } from "./appSessionController";
import { createWorkflowPresentation } from "./createWorkflowPresentation";
export function createWorkbookPreferencesWorkflow(
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
  let workbook: PreferenceWorkbookBinding | null = null;
  const controller = new WorkbookPreferenceController({
    observe: observeAccountOperation,
    port: (a) =>
      createWorkbookPreferenceAdapter({
        apiBase: a.apiBase,
        incidentId: a.incidentId,
        actorId: a.actorId,
      }),
    isCurrent: (a) => {
      const session = options().sessionController.getSnapshot();
      return (
        !disposed &&
        options().accepting() &&
        options().currentIncidentId() === a.incidentId &&
        session.lifetime === a.lifetime &&
        session.session?.user_id === a.actorId &&
        session.session.memberships.some((m) => m.incident_id === a.incidentId)
      );
    },
    recover: async (a, signal, admitted) => {
      const result = await options().recovery.recover({
        incidentId: a.incidentId,
        signal,
      });
      if (!admitted()) return { kind: "cancelled" };
      if (result.kind === "authorized" && workbook?.incidentId === a.incidentId)
        workbook.onAuthorizationRecovered(result);
      return result;
    },
    lost: (reason, a) => {
      const session = options().sessionController.getSnapshot();
      if (
        session.lifetime !== a.lifetime ||
        options().currentIncidentId() !== a.incidentId
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
      (m) => m.incident_id === incidentId,
    );
    const previous = controller.getSnapshot().authority;
    if (workbook?.incidentId !== incidentId) workbook = null;
    controller.setAuthority(
      session.lifetime && session.session && membership
        ? {
            incidentId,
            actorId: session.session.user_id,
            lifetime: session.lifetime,
            role: membership.role,
            apiBase: workbook?.apiBase ?? previous?.apiBase,
          }
        : null,
    );
    if (
      workbook &&
      session.session &&
      membership &&
      previous?.role !== membership.role
    )
      workbook.onAuthorizationRecovered({
        role: membership.role,
        userId: session.session.user_id,
      });
  };
  const presentation = createWorkflowPresentation(
    () => ({
      ...options(),
      accepting: () => !disposed && options().accepting(),
    }),
    (binding: PreferenceWorkbookBinding | null) => {
      if (disposed || !options().accepting()) return;
      const session = options().sessionController.getSnapshot();
      if (
        binding &&
        (binding.incidentId !== options().currentIncidentId() ||
          (binding.actorId !== null &&
            binding.actorId !== session.session?.user_id))
      )
        return;
      workbook = binding;
      if (!binding) {
        controller.setSurface(null);
        return;
      }
      sync();
      controller.setSurface(binding.actorId === null ? null : binding.surface);
    },
  );
  return {
    controller,
    synchronize: sync,
    retire: () => {
      if (disposed) return;
      presentation.invalidate();
      workbook = null;
      controller.retire();
    },
    dispose: () => {
      if (disposed) return;
      presentation.invalidate();
      disposed = true;
      workbook = null;
      controller.dispose();
    },
    attachWorkbook: presentation.attach,
  };
}
