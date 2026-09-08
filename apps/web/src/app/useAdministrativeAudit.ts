import { useLayoutEffect, useRef, useState } from "react";
import {
  AdministrativeAuditController,
  type AuditPorts,
  auditTransportPorts,
} from "./administrativeAuditController";
import type { AuditAuthority } from "./administrativeAuditModel";

export function useAdministrativeAudit(options: {
  readonly authority: AuditAuthority | null;
  readonly active: boolean;
  readonly isCurrent: AuditPorts["isCurrent"];
  readonly confirmAccess: AuditPorts["confirmAccess"];
  readonly authorizationFailed: AuditPorts["authorizationFailed"];
}) {
  const current = useRef(options);
  current.current = options;
  const [controller] = useState(
    () =>
      new AdministrativeAuditController({
        ...auditTransportPorts,
        isCurrent: (authority) => current.current.isCurrent(authority),
        confirmAccess: (authority, signal, admitted) =>
          current.current.confirmAccess(authority, signal, admitted),
        authorizationFailed: (status, authority) =>
          current.current.authorizationFailed(status, authority),
      }),
  );
  const lifetime = options.authority?.lifetime ?? null;
  const actorId = options.authority?.actorId ?? null;
  useLayoutEffect(() => {
    controller.setAuthority(lifetime && actorId ? { lifetime, actorId } : null);
    controller.setActive(
      options.active && document.visibilityState !== "hidden",
    );
  }, [controller, lifetime, actorId, options.active]);
  const mounted = useRef(false);
  useLayoutEffect(() => {
    mounted.current = true;
    const visibility = () =>
      controller.setActive(
        current.current.active && document.visibilityState !== "hidden",
      );
    document.addEventListener("visibilitychange", visibility);
    return () => {
      mounted.current = false;
      document.removeEventListener("visibilitychange", visibility);
      controller.retire();
      queueMicrotask(() => {
        if (!mounted.current) controller.dispose();
      });
    };
  }, [controller]);
  return controller;
}
