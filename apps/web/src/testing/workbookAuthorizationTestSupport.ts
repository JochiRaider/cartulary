import { loadSession } from "../app/api/authAccountClient";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
export function workbookAuthorizationRecovery(): AuthorizationRecoveryPort {
  return {
    recover: async ({ incidentId, signal }) => {
      const result = await loadSession({ signal });
      if (signal.aborted) return { kind: "cancelled" };
      if (!result.ok)
        return result.status === 401
          ? { kind: "session_lost" }
          : { kind: "unavailable", failure: "transient" };
      const membership = result.payload.data.memberships.find(
        (entry) => entry.incident_id === incidentId,
      );
      return membership === undefined
        ? { kind: "access_lost" }
        : {
            kind: "authorized",
            role: membership.role,
            userId: result.payload.data.user_id,
          };
    },
  };
}
