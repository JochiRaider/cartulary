import { useLayoutEffect, useRef, useState } from "react";
import type { AppWorkflowOptions } from "../app/AppWorkflows";
import { IncidentResourceController } from "../app/incidentResourceController";
import { useAppWorkflows } from "../app/useAppWorkflows";

type Options = Pick<
  AppWorkflowOptions,
  "sessionController" | "currentIncidentId"
> &
  Partial<Omit<AppWorkflowOptions, "sessionController" | "currentIncidentId">>;

/** Real eight-owner composition with a production Incident resource owner. */
export function useAppWorkflowsForTest(options: Options) {
  const committed = useRef(options);
  const workflowsRef = useRef<ReturnType<typeof useAppWorkflows> | null>(null);
  const [resources] = useState(
    () =>
      new IncidentResourceController({
        isCurrent: (authority) => {
          const current = committed.current;
          const state = current.sessionController.getSnapshot();
          return (
            current.currentIncidentId() === authority.incidentId &&
            state.lifetime === authority.lifetime &&
            state.session?.user_id === authority.actorId &&
            state.session.memberships.some(
              (m) => m.incident_id === authority.incidentId,
            )
          );
        },
        accepted: (resource) => {
          workflowsRef.current?.metadata.controller.acceptResource(
            resource,
            false,
          );
          workflowsRef.current?.lifecycle.controller.acceptResource(
            resource,
            false,
          );
        },
      }),
  );
  const workflows = useAppWorkflows({
    ...options,
    recovery:
      options.recovery ??
      options.sessionController.recoveryPort(
        (id) => id === committed.current.currentIncidentId(),
      ),
    onIncidentAccessLost: options.onIncidentAccessLost ?? resources.retire,
    onSessionLost:
      options.onSessionLost ?? (() => options.sessionController.sessionLost()),
    onResourceAccepted: (resource, authority) => {
      if (resources.accept(resource, authority))
        committed.current.onResourceAccepted?.(resource, authority);
    },
  });
  useLayoutEffect(() => {
    committed.current = options;
    workflowsRef.current = workflows;
  });
  const mounted = useRef(false);
  useLayoutEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      queueMicrotask(() => {
        if (!mounted.current) workflows.dispose();
      });
    };
  }, [workflows]);
  return workflows;
}
