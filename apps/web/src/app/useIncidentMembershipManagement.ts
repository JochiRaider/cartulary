import { useLayoutEffect, useRef, useState } from "react";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import {
  listIncidentMembershipPage,
  mutateIncidentMembership,
} from "./api/incidentMembershipManagementClient";
import type { AppSessionController } from "./appSessionController";
import { IncidentMembershipManagementController } from "./incidentMembershipManagementController";

/** App owns the sole session lifetime; this binding retains just one incident workflow. */
export function useIncidentMembershipManagement(options: {
  readonly sessionController: AppSessionController;
  readonly recovery: AuthorizationRecoveryPort;
  readonly currentIncidentId: () => string;
  readonly onIncidentAccessLost: () => void;
  readonly onSessionLost: () => void;
}) {
  const current = useRef(options);
  current.current = options;
  const surface = useRef<WorkbookIncidentControlsRendererProps | null>(null);
  const workbook = useRef<Pick<
    WorkbookIncidentControlsRendererProps,
    "incidentId" | "onAuthorizationRecovered"
  > | null>(null);
  const [controller] = useState(
    () =>
      new IncidentMembershipManagementController({
        list: listIncidentMembershipPage,
        mutate: mutateIncidentMembership,
        isCurrent: (authority) => {
          const state = current.current.sessionController.getSnapshot();
          return (
            current.current.currentIncidentId() === authority.incidentId &&
            state.lifetime === authority.lifetime &&
            state.session?.user_id === authority.actorId &&
            state.session.memberships.some(
              (member) => member.incident_id === authority.incidentId,
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
      }),
  );
  const synchronize = () => {
    const state = current.current.sessionController.getSnapshot();
    const incidentId = current.current.currentIncidentId();
    const previous = controller.getSnapshot().authority;
    const membership = state.session?.memberships.find(
      (member) => member.incident_id === incidentId,
    );
    if (workbook.current && workbook.current.incidentId !== incidentId)
      workbook.current = null;
    if (
      workbook.current &&
      state.session &&
      membership &&
      previous?.role !== membership.role
    )
      workbook.current.onAuthorizationRecovered?.({
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
      workbook.current = {
        incidentId: props.incidentId,
        onAuthorizationRecovered: props.onAuthorizationRecovered,
      };
      synchronizeRef.current();
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
  };
}
