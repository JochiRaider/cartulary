import { useLayoutEffect, useRef, useState } from "react";
import { listIncidentMembershipAuditPage } from "../app/api/incidentMembershipAuditClient";
import { IncidentAdminPanel } from "../app/IncidentAdminPanel";
import { IncidentMembershipAuditPanel } from "../app/IncidentMembershipAuditPanel";
import { IncidentMembershipAuditController } from "../app/incidentMembershipAuditController";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";

export function MembershipAuditTestSurface(
  props: WorkbookIncidentControlsRendererProps,
) {
  const current = useRef(props);
  current.current = props;
  const [controller] = useState(
    () =>
      new IncidentMembershipAuditController({
        list: listIncidentMembershipAuditPage,
        isCurrent: (authority) =>
          current.current.incidentId === authority.incidentId &&
          current.current.currentIncidentRole === "admin",
        recover: async () => ({
          kind: "authorized",
          role: "admin",
          userId: "00000000-0000-4000-8000-000000000001",
        }),
        lost: (reason) => {
          if (reason === "incident") current.current.onIncidentAccessLost?.();
        },
      }),
  );
  useLayoutEffect(() => {
    controller.setAuthority(
      props.currentIncidentRole === "admin"
        ? {
            incidentId: props.incidentId,
            actorId: "00000000-0000-4000-8000-000000000001",
            lifetime: "test-session",
          }
        : null,
    );
    controller.setActive(props.activeSection === "membership-audit");
  }, [
    controller,
    props.incidentId,
    props.currentIncidentRole,
    props.activeSection,
  ]);
  useLayoutEffect(() => () => controller.retire(), [controller]);
  return props.activeSection === "membership-audit" ? (
    <IncidentMembershipAuditPanel controller={controller} />
  ) : (
    <IncidentAdminPanel {...props} />
  );
}
