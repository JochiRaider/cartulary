import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import {
  patchIncidentMetadata,
  readIncidentMetadata,
} from "./api/incidentMetadataClient";
import type { AppSessionController } from "./appSessionController";
import { IncidentMetadataController } from "./incidentMetadataController";

/** Retains one workflow above drawer mounts; App remains the session authority. */
export function useIncidentMetadata(options: {
  sessionController: AppSessionController;
  recovery: AuthorizationRecoveryPort;
  currentIncidentId: () => string;
  onIncidentAccessLost: () => void;
  onSessionLost: () => void;
}) {
  const current = useRef(options);
  current.current = options;
  const surface = useRef<WorkbookIncidentControlsRendererProps | null>(null);
  const workbook = useRef<WorkbookIncidentControlsRendererProps | null>(null);
  const [controller] = useState(
    () =>
      new IncidentMetadataController({
        read: readIncidentMetadata,
        patch: patchIncidentMetadata,
        isCurrent: (authority) => {
          const state = current.current.sessionController.getSnapshot();
          return (
            current.current.currentIncidentId() === authority.incidentId &&
            state.lifetime === authority.lifetime &&
            state.session?.user_id === authority.actorId &&
            state.session.memberships.some(
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
          if (
            current.current.sessionController.getSnapshot().lifetime !==
              authority.lifetime ||
            current.current.currentIncidentId() !== authority.incidentId
          )
            return;
          if (reason === "session") current.current.onSessionLost();
          else current.current.onIncidentAccessLost();
        },
        publishResource: (resource, authority) => {
          if (workbook.current?.incidentId === authority.incidentId)
            workbook.current.onIncidentResourceAccepted?.(resource);
        },
      }),
  );
  const synchronize = () => {
    const state = current.current.sessionController.getSnapshot();
    const incidentId = current.current.currentIncidentId();
    const previous = controller.getSnapshot().authority;
    const membership = state.session?.memberships.find(
      (m) => m.incident_id === incidentId,
    );
    if (workbook.current?.incidentId !== incidentId) workbook.current = null;
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
            incidentId,
            lifetime: state.lifetime,
            actorId: state.session.user_id,
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
  const resource = useSyncExternalStore(
    controller.subscribe,
    () => controller.getSnapshot().resource,
  );
  return {
    controller,
    resource,
    bindSurface: (props: WorkbookIncidentControlsRendererProps | null) => {
      surface.current = props;
      if (!props) {
        controller.setActive(false);
        return;
      }
      workbook.current = props;
      sync.current();
      controller.setActive(
        props.activeSection === "incident-fields" &&
          document.visibilityState !== "hidden",
      );
    },
  };
}
