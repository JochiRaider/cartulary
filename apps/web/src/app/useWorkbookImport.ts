import { useCallback, useLayoutEffect, useRef, useState } from "react";
import {
  importProfileId,
  importRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import {
  type WorkbookImportBinding,
  WorkbookImportController,
} from "../imports/WorkbookImportController";
import type {
  AppSessionController,
  AppSessionSnapshot,
} from "./appSessionController";

function currentImportClaim(session: AppSessionSnapshot) {
  return (
    session.extensions.kind === "ready" &&
    session.extensions.value.some(
      (profile) =>
        profile.profile_id === importProfileId &&
        profile.claimed &&
        profile.route_families.includes(importRouteFamily),
    )
  );
}

export type WorkbookImportSurfaceBinding = Omit<
  WorkbookImportBinding,
  "scope" | "current"
> & { readonly incidentId: string };

/** The application session owns lifetime; the workbook supplies current incident facts. */
export function useWorkbookImport(options: {
  readonly sessionController: AppSessionController;
  readonly currentIncidentId: () => string;
}) {
  const current = useRef(options);
  current.current = options;
  const surface = useRef<WorkbookImportSurfaceBinding | null>(null);
  const [controller] = useState(() => new WorkbookImportController());
  const synchronize = useRef(() => {});
  synchronize.current = () => {
    const session = current.current.sessionController.getSnapshot(),
      binding = surface.current;
    if (session.state === "anonymous" || session.ended) {
      controller.retire();
      return;
    }
    if (
      !binding ||
      !session.session ||
      !session.lifetime ||
      session.authenticationTransportPending ||
      session.state !== "authenticated"
    ) {
      controller.pause();
      return;
    }
    if (binding.incidentId !== current.current.currentIncidentId()) {
      surface.current = null;
      controller.retire();
      return;
    }
    const scope = {
      incidentId: binding.incidentId,
      actorId: session.session.user_id,
      lifetime: session.lifetime,
    };
    const membership = session.session.memberships.find(
      (m) => m.incident_id === binding.incidentId,
    );
    controller.bind({
      ...binding,
      available: binding.available && currentImportClaim(session),
      scope,
      role: membership ? (binding.role === null ? null : membership.role) : "",
      current: () => {
        const now = current.current.sessionController.getSnapshot();
        return (
          current.current.currentIncidentId() === scope.incidentId &&
          now.state === "authenticated" &&
          !now.authenticationTransportPending &&
          currentImportClaim(now) &&
          now.lifetime === scope.lifetime &&
          now.session?.user_id === scope.actorId
        );
      },
    });
  };
  useLayoutEffect(() => {
    const unsubscribe = options.sessionController.subscribe(() =>
      synchronize.current(),
    );
    return () => {
      unsubscribe();
      controller.retire();
    };
  }, [controller, options.sessionController]);
  const bindWorkbook = useCallback(
    (binding: WorkbookImportSurfaceBinding | null) => {
      surface.current = binding;
      synchronize.current();
    },
    [],
  );
  return { controller, bindWorkbook };
}
