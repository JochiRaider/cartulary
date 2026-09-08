import { useLayoutEffect, useRef, useState } from "react";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import { listIncidentMembershipAuditPage } from "./api/incidentMembershipAuditClient";
import type { AppSessionController } from "./appSessionController";
import { IncidentMembershipAuditController } from "./incidentMembershipAuditController";

/** App retains exactly one incident's filters; existing session/recovery owners accept authority. */
export function useIncidentMembershipAudit(options: {
  readonly sessionController: AppSessionController;
  readonly recovery: AuthorizationRecoveryPort;
  readonly currentIncidentId: () => string;
  readonly onIncidentAccessLost: () => void;
  readonly onSessionLost: () => void;
}) {
  const current = useRef(options);
  current.current = options;
  const surface = useRef<WorkbookIncidentControlsRendererProps | null>(null);
  const [controller] = useState(
    () =>
      new IncidentMembershipAuditController({
        list: listIncidentMembershipAuditPage,
        isCurrent: (authority) => {
          const state = current.current.sessionController.getSnapshot();
          return (
            current.current.currentIncidentId() === authority.incidentId &&
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
          const result = await current.current.recovery.recover({
            incidentId: authority.incidentId,
            signal,
          });
          if (!admitted()) return { kind: "cancelled" };
          if (result.kind === "authorized")
            surface.current?.onAuthorizationRecovered?.(result);
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
          else if (reason === "incident")
            current.current.onIncidentAccessLost();
          else void surface.current?.onSessionRoleChange?.();
        },
      }),
  );
  const synchronize = () => {
    const state = current.current.sessionController.getSnapshot();
    const incidentId = current.current.currentIncidentId();
    const previous = controller.getSnapshot().authority;
    const membership = state.session?.memberships.find(
      (member) => member.incident_id === incidentId,
    );
    if (
      surface.current?.incidentId === incidentId &&
      state.session &&
      membership &&
      surface.current.currentIncidentRole !== membership.role
    ) {
      surface.current.onAuthorizationRecovered?.({
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
            apiBase: surface.current
              ? surface.current.apiBase
              : previous?.apiBase,
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
      current.current.onIncidentAccessLost();
  };
  const synchronizeRef = useRef(synchronize);
  synchronizeRef.current = synchronize;
  useLayoutEffect(() => {
    synchronizeRef.current();
  });
  useLayoutEffect(() => {
    const unsubscribe = options.sessionController.subscribe(() =>
      synchronizeRef.current(),
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
      synchronizeRef.current();
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
  };
}
