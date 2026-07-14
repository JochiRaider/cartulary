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
  ) => void;
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
        event.kind === "session_revoked"
      ) {
        onResourceChange({
          changeKind: "remove",
          reasonCode: "authorization_lost",
          resourceId: "*",
        });
        return;
      }
      if (event.kind === "reset_required") {
        onResourceChange({
          changeKind: "invalidate",
          reasonCode: event.reason,
          resourceId: "*",
        });
        return;
      }
      if (event.kind !== "message") {
        return;
      }
      const change = interpretNetworkFlowCollaborationMessage(event.message);
      if (change !== null) {
        onResourceChange(change);
      }
    });
  }, [enabled, onResourceChange, session]);
}
