import { useEffect } from "react";
import { useIncidentCollaborationSession } from "../collaboration/IncidentCollaborationSession";
import { networkAnalysisSheetRef } from "./networkFlowClient";
import {
  interpretNetworkFlowCollaborationMessage,
  type NetworkFlowExtensionResourceChange,
} from "./networkFlowCollaborationInterpreter";

export type { NetworkFlowExtensionResourceChange } from "./networkFlowCollaborationInterpreter";

export function useNetworkFlowExtensionEvents({
  apiBase: _apiBase,
  enabled,
  incidentId: _incidentId,
  onResourceChange,
}: {
  readonly apiBase?: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onResourceChange: (
    change: NetworkFlowExtensionResourceChange,
  ) => Promise<void> | void;
}) {
  const session = useIncidentCollaborationSession();

  useEffect(() => {
    if (!enabled) {
      return;
    }
    session.publishPresence({
      sheet_ref: networkAnalysisSheetRef(),
      mode: "viewing",
    });
    return session.subscribe((event) => {
      if (
        event.kind === "authorization_lost" ||
        event.kind === "session_revoked" ||
        event.kind === "incident_closed"
      ) {
        void onResourceChange({
          changeKind: "remove",
          reasonCode:
            event.kind === "incident_closed"
              ? "incident_closed"
              : "authorization_lost",
          resourceId: "*",
        });
        return;
      }
      if (event.kind === "reset_required") {
        void Promise.resolve(
          onResourceChange({
            changeKind: "invalidate",
            reasonCode: event.reason,
            resourceId: "*",
          }),
        ).then(() => session.completeReset(event.generation));
        return;
      }
      if (event.kind !== "message") {
        return;
      }
      const change = interpretNetworkFlowCollaborationMessage(event.message);
      if (change !== null) {
        void onResourceChange(change);
      }
    });
  }, [enabled, onResourceChange, session]);
}
