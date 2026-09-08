import { useLayoutEffect, useRef, useState } from "react";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import { metadataActorId } from "../testing/incidentMetadataTestSupport";
import {
  patchIncidentMetadata,
  readIncidentMetadata,
} from "./api/incidentMetadataClient";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { IncidentMetadataPanel } from "./IncidentMetadataPanel";
import { IncidentMetadataController } from "./incidentMetadataController";

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
