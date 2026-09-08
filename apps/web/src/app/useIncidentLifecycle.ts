import { useLayoutEffect, useRef, useState } from "react";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { IncidentResource } from "../shared/incidentResource";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import {
  mutateIncidentLifecycle,
  readLifecycleIncident,
} from "./api/incidentLifecycleClient";
import type { AppSessionController } from "./appSessionController";
import { IncidentLifecycleController } from "./incidentLifecycleController";
import type { LifecycleAuthority } from "./incidentLifecycleModel";

/** App owns lifetime and publication; the drawer borrows the retained workflow. */
export function useIncidentLifecycle(options: {
  sessionController: AppSessionController;
  recovery: AuthorizationRecoveryPort;
  currentIncidentId: () => string;
  onIncidentAccessLost: () => void;
  onSessionLost: () => void;
  onResourceAccepted: (
    resource: IncidentResource,
    authority: LifecycleAuthority,
  ) => void;
}) {
  const current = useRef(options);
  current.current = options;
  const surface = useRef<WorkbookIncidentControlsRendererProps | null>(null);
  const workbook = useRef<WorkbookIncidentControlsRendererProps | null>(null);
  const [controller] = useState(
    () =>
      new IncidentLifecycleController({
        read: readLifecycleIncident,
        mutate: mutateIncidentLifecycle,
        isCurrent: (authority) => {
          const s = current.current.sessionController.getSnapshot();
          return (
            current.current.currentIncidentId() === authority.incidentId &&
            s.lifetime === authority.lifetime &&
            s.session?.user_id === authority.actorId &&
            s.session.memberships.some(
              (m) => m.incident_id === authority.incidentId,
            )
          );
        },
        recover: async (authority, signal, admitted) => {
          const result = await current.current.recovery.recover({
            incidentId: authority.incidentId,
            signal,
          });
          if (!admitted()) return { kind: "cancelled" };
          if (
            result.kind === "authorized" &&
            workbook.current?.incidentId === authority.incidentId
          )
            workbook.current.onAuthorizationRecovered?.(result);
          return result;
        },
        lost: (reason, authority) => {
          const state = current.current.sessionController.getSnapshot();
          if (
            state.lifetime !== authority.lifetime ||
            current.current.currentIncidentId() !== authority.incidentId
          )
            return;
          if (reason === "session") current.current.onSessionLost();
          else current.current.onIncidentAccessLost();
        },
        publishResource: (resource, authority) =>
          current.current.onResourceAccepted(resource, authority),
      }),
  );
  const synchronize = () => {
    const state = current.current.sessionController.getSnapshot();
    const incidentId = current.current.currentIncidentId();
    const previous = controller.getSnapshot().authority;
    const member = state.session?.memberships.find(
      (m) => m.incident_id === incidentId,
    );
    if (workbook.current?.incidentId !== incidentId) workbook.current = null;
    if (
      workbook.current &&
      state.session &&
      member &&
      previous?.role !== member.role
    )
      workbook.current.onAuthorizationRecovered?.({
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
              surface.current?.incidentId === incidentId
                ? surface.current.apiBase
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
      current.current.onIncidentAccessLost();
  };
  const sync = useRef(synchronize);
  sync.current = synchronize;
  useLayoutEffect(() => sync.current());
  useLayoutEffect(() => {
    const unsubscribe = options.sessionController.subscribe(() =>
      sync.current(),
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
    bindSurface: (props: WorkbookIncidentControlsRendererProps | null) => {
      surface.current = props;
      if (!props) {
        controller.setActive(false);
        return;
      }
      workbook.current = props;
      sync.current();
      controller.setActive(
        props.activeSection === "summary" &&
          document.visibilityState !== "hidden",
      );
    },
  };
}
