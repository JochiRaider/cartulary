import { useLayoutEffect, useRef } from "react";
import type { NetworkFlowTableController } from "./NetworkFlowTableController";
import { NetworkFlowRequestError } from "./networkFlowErrors";

/** Table lifecycle has one event owner; presentation clears only affected resources. */
export function useNetworkFlowCollaborationController(options: {
  readonly controller: NetworkFlowTableController;
  readonly activeTableId: string | null;
  readonly clearResources: () => void;
  readonly clearActiveTable: () => void;
  readonly onMessage: (message: string) => void;
  readonly onProtectedStateLoss: (error: NetworkFlowRequestError) => void;
}) {
  const current = useRef(options);
  current.current = options;
  useLayoutEffect(
    () =>
      options.controller.subscribeChanges((change) => {
        const value = current.current;
        if (change.resourceKind === "*" && change.changeKind === "remove") {
          const closed = change.reasonCode === "incident_closed";
          const session = change.reasonCode === "session_revoked";
          value.clearResources();
          value.onProtectedStateLoss(
            new NetworkFlowRequestError({
              code: closed
                ? "incident_closed"
                : session
                  ? "session_required"
                  : "authorization_denied",
              retryAction: "do_not_retry",
              retryable: false,
              status: closed ? 409 : session ? 401 : 403,
              safeMessage: closed
                ? "This incident is closed. Network Analysis is unavailable."
                : session
                  ? "Session recovery is required. Protected state is hidden."
                  : "Network Analysis access changed. Protected data was cleared.",
            }),
          );
        } else if (change.changeKind === "remove") {
          if (change.resourceId === value.activeTableId)
            value.clearActiveTable();
          value.onMessage("A Network Flow table was deleted.");
        } else if (change.reasonCode !== "renamed")
          value.onMessage("Network Analysis data changed.");
      }),
    [options.controller],
  );
}
