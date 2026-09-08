import { useLayoutEffect, useRef, useState } from "react";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import { createWorkbookSavedViewAdapter } from "../workbook/adapters/createWorkbookSavedViewAdapter";
import type { SavedViewBinding } from "../workbook/savedviews/savedViewOperationModel";
import { WorkbookSavedViewController } from "../workbook/savedviews/WorkbookSavedViewController";
import { observeAccountOperation } from "./accountOperation";
import type { AppSessionController } from "./appSessionController";

export function useWorkbookSavedViews(options: {
  sessionController: AppSessionController;
  recovery: AuthorizationRecoveryPort;
  currentIncidentId: () => string;
  onIncidentAccessLost: () => void;
  onSessionLost: () => void;
}) {
  const current = useRef(options);
  current.current = options;
  const binding = useRef<SavedViewBinding | null>(null);
  const [controller] = useState(
    () =>
      new WorkbookSavedViewController({
        observe: observeAccountOperation,
        port: (authority) =>
          createWorkbookSavedViewAdapter({
            apiBase: authority.apiBase,
            incidentId: authority.incidentId,
          }),
        isCurrent: (authority) => {
          const session = current.current.sessionController.getSnapshot();
          return (
            current.current.currentIncidentId() === authority.incidentId &&
            session.lifetime === authority.lifetime &&
            session.session?.user_id === authority.actorId &&
            session.session.memberships.some(
              (member) => member.incident_id === authority.incidentId,
            )
          );
        },
        recover: async (authority, signal, admitted) => {
          const result = await current.current.recovery.recover({
            incidentId: authority.incidentId,
            signal,
          });
          return admitted() ? result : { kind: "cancelled" };
        },
        lost: (reason, authority) => {
          const session = current.current.sessionController.getSnapshot();
          if (
            session.lifetime !== authority.lifetime ||
            current.current.currentIncidentId() !== authority.incidentId
          )
            return;
          if (reason === "session") current.current.onSessionLost();
          else current.current.onIncidentAccessLost();
        },
      }),
  );
  const sync = () => {
    const session = current.current.sessionController.getSnapshot();
    const incidentId = current.current.currentIncidentId();
    const membership = session.session?.memberships.find(
      (member) => member.incident_id === incidentId,
    );
    if (binding.current?.incidentId !== incidentId) binding.current = null;
    controller.setAuthority(
      session.session && session.lifetime && membership
        ? {
            incidentId,
            actorId: session.session.user_id,
            lifetime: session.lifetime,
            role: membership.role,
            apiBase:
              binding.current?.apiBase ??
              controller.getSnapshot().authority?.apiBase,
          }
        : null,
    );
  };
  const synchronize = useRef(sync);
  synchronize.current = sync;
  useLayoutEffect(() => synchronize.current());
  useLayoutEffect(() => {
    const unsubscribe = options.sessionController.subscribe(() =>
      synchronize.current(),
    );
    return () => {
      unsubscribe();
    };
  }, [options.sessionController]);
  useLayoutEffect(() => () => controller.retire(), [controller]);
  return {
    controller,
    bindWorkbook: (next: SavedViewBinding | null) => {
      if (next && next.incidentId !== current.current.currentIncidentId())
        return;
      binding.current = next;
      synchronize.current();
      controller.setBinding(next);
    },
  };
}
