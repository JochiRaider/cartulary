import { useLayoutEffect, useMemo, useRef, useState } from "react";
import { useIncidentCollaborationSession } from "../collaboration/IncidentCollaborationSession";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  networkFlowActivityProfileId,
  networkFlowRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import {
  indicatorTargetChange,
  queryCompatibleIndicatorTargets,
} from "../services/networkFlowIndicatorAdapter";
import {
  type IndicatorLinkTransport,
  NetworkFlowIndicatorLinkController,
} from "./NetworkFlowIndicatorLinkController";
import {
  getNetworkFlowBindingSourceRowLimit,
  linkNetworkFlowIndicator,
} from "./networkFlowClient";
import { interpretNetworkFlowCollaborationMessage } from "./networkFlowCollaborationInterpreter";
import type { IndicatorLinkAuthority } from "./networkFlowIndicatorLinkOperation";

/** The workbook, independently of its lazy surface, owns volatile link recovery. */
export function useNetworkFlowIndicatorLinkOwner(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly actorId: string | null;
  readonly sessionIdentity: string | null;
  readonly role: string | null;
  readonly open: boolean;
}): NetworkFlowIndicatorLinkController {
  const collaboration = useIncidentCollaborationSession();
  const [controller] = useState(() => new NetworkFlowIndicatorLinkController());
  const current = useRef(options);
  current.current = options;
  const transport = useMemo<IndicatorLinkTransport>(
    () => ({
      discoverTargets: (query, signal) =>
        queryCompatibleIndicatorTargets(options.apiBase, query, signal),
      sourceLimit: (signal) =>
        getNetworkFlowBindingSourceRowLimit({
          availability: options.availability,
          apiBase: options.apiBase,
          incidentId: options.incidentId,
          signal,
        }),
      submit: (attempt, signal, authorizeDispatch) =>
        linkNetworkFlowIndicator({
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
    const authority = (): IndicatorLinkAuthority => {
      const value = current.current;
      return {
        incidentId: value.incidentId,
        actorId: value.actorId,
        session: value.sessionIdentity ?? "unresolved",
        sessionResolved: value.sessionIdentity !== null,
        role: value.role,
        open: value.open,
        profileAvailable: value.availability.isRouteAvailable(
          networkFlowActivityProfileId,
          networkFlowRouteFamily,
        ),
        available:
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
    controller.bind(transport, authority);
  });
  useLayoutEffect(
    () =>
      options.availability.subscribeAuthority(controller.revalidateAuthority),
    [controller, options.availability],
  );
  useLayoutEffect(
    () =>
      collaboration.subscribe((event) => {
        if (event.kind === "message") {
          const targetChange = indicatorTargetChange(event.message);
          if (targetChange !== null) controller.onTargetChange(targetChange);
          const change = interpretNetworkFlowCollaborationMessage(
            event.message,
          );
          if (change !== null) controller.onResourceChange(change);
        } else if (
          event.kind === "authorization_lost" ||
          event.kind === "session_revoked" ||
          event.kind === "incident_closed"
        ) {
          controller.onResourceChange({
            resourceKind: "*",
            resourceId: "*",
            changeKind: "remove",
            reasonCode: event.kind,
          });
        } else if (event.kind === "reset_required") {
          controller.onResourceChange({
            resourceKind: "*",
            resourceId: "*",
            changeKind: "invalidate",
            reasonCode: event.reason,
          });
          collaboration.completeReset(event.generation);
        }
      }),
    [collaboration, controller],
  );
  useLayoutEffect(() => () => controller.dispose(), [controller]);
  return controller;
}
