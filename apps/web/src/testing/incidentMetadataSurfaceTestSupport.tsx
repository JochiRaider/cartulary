import { useLayoutEffect, useRef, useState } from "react";
import {
  patchIncidentMetadata,
  readIncidentMetadata,
} from "../app/api/incidentMetadataClient";
import { IncidentAdminPanel } from "../app/IncidentAdminPanel";
import { IncidentMetadataPanel } from "../app/IncidentMetadataPanel";
import { IncidentMetadataController } from "../app/incidentMetadataController";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import { metadataActorId } from "./incidentMetadataTestSupport";

export function MetadataTestSurface(
  props: Omit<WorkbookIncidentControlsRendererProps, "activeSection"> & {
    activeSection:
      | WorkbookIncidentControlsRendererProps["activeSection"]
      | null;
  },
) {
  const latest = useRef(props);
  latest.current = props;
  const [controller] = useState(
    () =>
      new IncidentMetadataController({
        read: readIncidentMetadata,
        patch: patchIncidentMetadata,
        isCurrent: (authority) =>
          authority.incidentId === latest.current.incidentId,
        recover: async () => ({
          kind: "authorized",
          role: latest.current.currentIncidentRole ?? "viewer",
          userId: metadataActorId,
        }),
        lost: (reason) => {
          if (reason === "incident") latest.current.onIncidentAccessLost?.();
        },
      }),
  );
  useLayoutEffect(() => {
    controller.setAuthority({
      incidentId: props.incidentId,
      actorId: metadataActorId,
      lifetime: "test",
      role: props.currentIncidentRole ?? "viewer",
    });
    controller.setActive(props.activeSection === "incident-fields");
  });
  useLayoutEffect(() => () => controller.dispose(), [controller]);
  return props.activeSection === null ? null : props.activeSection ===
    "incident-fields" ? (
    <IncidentMetadataPanel controller={controller} />
  ) : (
    <IncidentAdminPanel {...props} activeSection={props.activeSection} />
  );
}
