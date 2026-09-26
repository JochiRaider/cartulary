import { createWorkflowAttachment } from "../shared/workflowAttachment";
import type { AppSessionController } from "./appSessionController";

/** Capture the authorizing lifetime once; updates cannot transfer an old handle. */
export function createWorkflowPresentation<
  Binding extends { incidentId: string; actorId?: string | null },
>(
  environment: () => {
    sessionController: AppSessionController;
    currentIncidentId: () => string;
    accepting: () => boolean;
  },
  apply: (binding: Binding | null) => void,
) {
  return createWorkflowAttachment<Binding>((binding) => {
    const owner = environment();
    const source = owner.sessionController;
    const state = source.getSnapshot();
    if (
      !owner.accepting() ||
      !state.session ||
      !state.lifetime ||
      binding.incidentId !== owner.currentIncidentId() ||
      (binding.actorId != null && binding.actorId !== state.session.user_id) ||
      !state.session.memberships.some(
        (m) => m.incident_id === binding.incidentId,
      )
    )
      return null;
    return () => {
      const now = environment();
      const accepted = source.getSnapshot();
      return (
        now.accepting() &&
        now.sessionController === source &&
        now.currentIncidentId() === binding.incidentId &&
        accepted.lifetime === state.lifetime &&
        accepted.session?.user_id === state.session?.user_id
      );
    };
  }, apply);
}
