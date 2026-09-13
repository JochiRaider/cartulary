import { useEffect } from "react";
import { useIncidentCollaborationSession } from "../collaboration/IncidentCollaborationSession";
import { networkAnalysisSheetRef } from "./networkFlowClient";
import {
  interpretNetworkFlowCollaborationEvent,
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
      if (event.kind === "reset_required") {
        void Promise.resolve(
          onResourceChange({
            resourceKind: "*",
            changeKind: "invalidate",
            reasonCode: event.reason,
            resourceId: "*",
          }),
        ).then(() => session.completeReset(event.generation));
        return;
      }
      const change = interpretNetworkFlowCollaborationEvent(event);
      if (change !== null) {
        void onResourceChange(change);
      }
    });
  }, [enabled, onResourceChange, session]);
}
