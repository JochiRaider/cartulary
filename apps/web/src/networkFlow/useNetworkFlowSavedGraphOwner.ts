import { useLayoutEffect, useMemo, useRef, useState } from "react";
import { useIncidentCollaborationSession } from "../collaboration/IncidentCollaborationSession";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  networkFlowActivityProfileId,
  networkFlowRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import {
  getNetworkFlowSavedGraph,
  getNetworkFlowSavedGraphResult,
  getSavedGraphObservationLimit,
  listNetworkFlowSavedGraphs,
  queryNetworkFlowSavedGraphContributors,
  submitNetworkFlowSavedGraphMutation,
} from "./networkFlowClient";
import { interpretNetworkFlowCollaborationEvent } from "./networkFlowCollaborationInterpreter";
import type { SavedGraphTransport } from "./SavedGraphController";
import { SavedGraphController } from "./SavedGraphController";
import { readSavedGraphJob } from "./savedGraphObservation";
import type { SavedGraphAuthority } from "./savedGraphOperation";

/** Mounted by the workbook, so switching surfaces never owns mutation lifetime. */
export function useNetworkFlowSavedGraphOwner(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly actorId: string | null;
  readonly sessionIdentity: string | null;
  readonly role: string | null;
  readonly open: boolean;
}): SavedGraphController {
  const collaboration = useIncidentCollaborationSession();
  const [controller] = useState(() => new SavedGraphController());
  const current = useRef(options);
  current.current = options;
  const transport = useMemo<SavedGraphTransport>(
    () => ({
      observationLimit: (signal) =>
        getSavedGraphObservationLimit({
          availability: options.availability,
          apiBase: options.apiBase,
          incidentId: options.incidentId,
          signal,
        }),
      readJob: (target, signal) =>
        readSavedGraphJob(options.apiBase, target, signal),
      navigation: {
        result: (graphViewId, signal) =>
          getNetworkFlowSavedGraphResult({
            availability: options.availability,
            apiBase: options.apiBase,
            incidentId: options.incidentId,
            graphViewId,
            signal,
          }),
        contributors: (graph, selector, cursorToken, signal) =>
          queryNetworkFlowSavedGraphContributors({
            availability: options.availability,
            apiBase: options.apiBase,
            incidentId: options.incidentId,
            graphViewId: graph.graph_view_id,
            projectionResultId:
              graph.selected_result_binding?.projection_result_id ?? "",
            selector,
            cursorToken,
            signal,
          }),
      },
      list: (signal: AbortSignal) =>
        listNetworkFlowSavedGraphs({
          availability: options.availability,
          apiBase: options.apiBase,
          incidentId: options.incidentId,
          signal,
        }),
      get: (graphViewId: string, signal: AbortSignal) =>
        getNetworkFlowSavedGraph({
          availability: options.availability,
          apiBase: options.apiBase,
          incidentId: options.incidentId,
          graphViewId,
          signal,
        }),
      submit: (
        attempt: Parameters<
          typeof submitNetworkFlowSavedGraphMutation
        >[0]["attempt"],
        signal: AbortSignal,
        authorizeDispatch: () => void,
      ) =>
        submitNetworkFlowSavedGraphMutation({
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
    const authority = (): SavedGraphAuthority => {
      const value = current.current;
      return {
        incidentId: value.incidentId,
        actorId: value.actorId,
        session: value.sessionIdentity ?? "unresolved",
        sessionResolved: value.sessionIdentity !== null,
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
    controller.bind(transport, authority);
  });
  useLayoutEffect(
    () =>
      collaboration.subscribe((event) => {
        const change = interpretNetworkFlowCollaborationEvent(event);
        if (change !== null) {
          void controller.onResourceChange(change);
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
