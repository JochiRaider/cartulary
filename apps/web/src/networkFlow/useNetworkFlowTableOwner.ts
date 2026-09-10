import { useLayoutEffect, useMemo, useRef, useState } from "react";
import { useIncidentCollaborationSession } from "../collaboration/IncidentCollaborationSession";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  networkFlowActivityProfileId,
  networkFlowRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import {
  NetworkFlowTableController,
  type TableTransport,
} from "./NetworkFlowTableController";
import {
  listNetworkFlowTables,
  submitNetworkFlowTableMutation,
} from "./networkFlowClient";
import {
  interpretNetworkFlowCollaborationMessage,
  type NetworkFlowExtensionResourceChange,
} from "./networkFlowCollaborationInterpreter";
import type { TableAuthority } from "./networkFlowTableOperation";

/** Mounted by the workbook, so switching surfaces never owns mutation lifetime. */
export function useNetworkFlowTableOwner(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly actorId: string | null;
  readonly sessionIdentity: string | null;
  readonly role: string | null;
  readonly open: boolean;
  readonly onMutationAdmitted: () => void;
  readonly onLocalChange: (change: NetworkFlowExtensionResourceChange) => void;
}): NetworkFlowTableController {
  const collaboration = useIncidentCollaborationSession();
  const [controller] = useState(() => new NetworkFlowTableController());
  const current = useRef(options);
  current.current = options;
  const transport = useMemo<TableTransport>(
    () => ({
      list: (signal) =>
        listNetworkFlowTables({
          availability: options.availability,
          apiBase: options.apiBase,
          incidentId: options.incidentId,
          signal,
        }),
      submit: (attempt, signal, authorizeDispatch) =>
        submitNetworkFlowTableMutation({
          availability: options.availability,
          apiBase: options.apiBase,
          attempt,
          signal,
          authorizeDispatch,
        }),
    }),
    [options.availability, options.apiBase, options.incidentId],
  );
  useLayoutEffect(() => {
    const authority = (): TableAuthority => {
      const value = current.current;
      return {
        incidentId: value.incidentId,
        actorId: value.actorId,
        sessionIdentity: value.sessionIdentity,
        profileAvailable: value.availability.isRouteAvailable(
          networkFlowActivityProfileId,
          networkFlowRouteFamily,
        ),
        role: value.role,
        open: value.open,
        available:
          value.sessionIdentity !== null &&
          value.availability.isRenderable({
            extensionProfileId: networkFlowActivityProfileId,
            workspaceKey: "network_analysis",
          }) &&
          value.availability.isRouteAvailable(
            networkFlowActivityProfileId,
            networkFlowRouteFamily,
          ),
        availabilityTag: value.availability.authorityTag(),
      };
    };
    controller.bind(
      transport,
      authority,
      (change) => current.current.onLocalChange(change),
      () => current.current.onMutationAdmitted(),
    );
  });
  useLayoutEffect(
    () =>
      collaboration.subscribe((event) => {
        if (event.kind === "message") {
          const change = interpretNetworkFlowCollaborationMessage(
            event.message,
          );
          if (change !== null) void controller.onResourceChange(change);
        } else if (
          event.kind === "authorization_lost" ||
          event.kind === "session_revoked" ||
          event.kind === "incident_closed"
        ) {
          void controller.onResourceChange({
            resourceKind: "*",
            resourceId: "*",
            changeKind: "remove",
            reasonCode: event.kind,
          });
        } else if (event.kind === "reset_required") {
          void controller
            .onResourceChange({
              resourceKind: "*",
              resourceId: "*",
              changeKind: "invalidate",
              reasonCode: event.reason,
            })
            .then(() => collaboration.completeReset(event.generation));
        }
      }),
    [collaboration, controller],
  );
  useLayoutEffect(
    () =>
      options.availability.subscribeAuthority(() =>
        controller.revalidateAuthority(),
      ),
    [controller, options.availability],
  );
  useLayoutEffect(() => () => controller.dispose(), [controller]);
  return controller;
}
