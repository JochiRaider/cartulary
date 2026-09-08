import { useLayoutEffect, useRef, useState } from "react";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import { createWorkbookPreferenceAdapter } from "../workbook/adapters/createWorkbookPreferenceAdapter";
import { WorkbookPreferenceController } from "../workbook/preferences/WorkbookPreferenceController";
import type { PreferenceWorkbookBinding } from "../workbook/preferences/workbookPreferenceModel";
import { observeAccountOperation } from "./accountOperation";
import type { AppSessionController } from "./appSessionController";
export function useWorkbookPreferences(options: {
  sessionController: AppSessionController;
  recovery: AuthorizationRecoveryPort;
  currentIncidentId: () => string;
  onIncidentAccessLost: () => void;
  onSessionLost: () => void;
}) {
  const current = useRef(options);
  current.current = options;
  const workbook = useRef<PreferenceWorkbookBinding | null>(null);
  const [controller] = useState(
    () =>
      new WorkbookPreferenceController({
        observe: observeAccountOperation,
        port: (a) =>
          createWorkbookPreferenceAdapter({
            apiBase: a.apiBase,
            incidentId: a.incidentId,
            actorId: a.actorId,
          }),
        isCurrent: (a) => {
          const session = current.current.sessionController.getSnapshot();
          return (
            current.current.currentIncidentId() === a.incidentId &&
            session.lifetime === a.lifetime &&
            session.session?.user_id === a.actorId &&
            session.session.memberships.some(
              (m) => m.incident_id === a.incidentId,
            )
          );
        },
        recover: async (a, signal, admitted) => {
          const result = await current.current.recovery.recover({
            incidentId: a.incidentId,
            signal,
          });
          if (!admitted()) return { kind: "cancelled" };
          if (
            result.kind === "authorized" &&
            workbook.current?.incidentId === a.incidentId
          )
            workbook.current.onAuthorizationRecovered(result);
          return result;
        },
        lost: (reason, a) => {
          const session = current.current.sessionController.getSnapshot();
          if (
            session.lifetime !== a.lifetime ||
            current.current.currentIncidentId() !== a.incidentId
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
      (m) => m.incident_id === incidentId,
    );
    const previous = controller.getSnapshot().authority;
    if (workbook.current?.incidentId !== incidentId) workbook.current = null;
    controller.setAuthority(
      session.lifetime && session.session && membership
        ? {
            incidentId,
            actorId: session.session.user_id,
            lifetime: session.lifetime,
            role: membership.role,
            apiBase: workbook.current?.apiBase ?? previous?.apiBase,
          }
        : null,
    );
    if (
      workbook.current &&
      session.session &&
      membership &&
      previous?.role !== membership.role
    )
      workbook.current.onAuthorizationRecovered({
        role: membership.role,
        userId: session.session.user_id,
      });
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
  const mounted = useRef(false);
  useLayoutEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      controller.retire();
      queueMicrotask(() => {
        if (!mounted.current) controller.dispose();
      });
    };
  }, [controller]);
  return {
    controller,
    bindWorkbook: (binding: PreferenceWorkbookBinding | null) => {
      const session = current.current.sessionController.getSnapshot();
      if (
        binding &&
        (binding.incidentId !== current.current.currentIncidentId() ||
          (binding.actorId !== null &&
            binding.actorId !== session.session?.user_id))
      )
        return;
      workbook.current = binding;
      if (!binding) {
        controller.setSurface(null);
        return;
      }
      synchronize.current();
      controller.setSurface(binding.actorId === null ? null : binding.surface);
    },
  };
}
