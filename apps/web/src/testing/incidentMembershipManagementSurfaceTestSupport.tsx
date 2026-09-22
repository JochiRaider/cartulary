import { useLayoutEffect, useRef, useState } from "react";
import {
  listIncidentMembershipPage,
  mutateIncidentMembership,
} from "../app/api/incidentMembershipManagementClient";
import { IncidentMembershipManagementPanel } from "../app/IncidentMembershipManagementPanel";
import { IncidentMembershipManagementController } from "../app/incidentMembershipManagementController";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import { membershipAuthority } from "./incidentMembershipManagementTestSupport";

/** Presentation fixture: lifecycle integration separately uses the real AppSessionController. */
export function MembershipTestSurface(
  props: WorkbookIncidentControlsRendererProps,
) {
  const latest = useRef(props);
  latest.current = props;
  const [controller] = useState(
    () =>
      new IncidentMembershipManagementController({
        list: listIncidentMembershipPage,
        mutate: mutateIncidentMembership,
        isCurrent: (authority) =>
          authority.incidentId === latest.current.incidentId,
        recover: async () => ({
          kind: "authorized",
          role: latest.current.currentIncidentRole ?? "viewer",
          userId: membershipAuthority.actorId,
        }),
        lost: (reason) => {
          if (reason === "incident") latest.current.onIncidentAccessLost?.();
        },
      }),
  );
  useLayoutEffect(() => {
    controller.setAuthority({
      ...membershipAuthority,
      incidentId: props.incidentId,
      role: props.currentIncidentRole ?? "viewer",
    });
    controller.setActive(props.activeSection === "memberships");
  });
  useLayoutEffect(() => () => controller.dispose(), [controller]);
  return props.activeSection === "memberships" ? (
    <IncidentMembershipManagementPanel
      controller={controller}
      density={props.density}
    />
  ) : null;
}
