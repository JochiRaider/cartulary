import { useCallback, useEffect, useState } from "react";
import type { AuthorizationRecoveryPort } from "../../shared/authorizationRecovery";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";

type WorkbookAuthorizationStateOptions = {
  readonly accountUserId: string | undefined;
  readonly authorizationRecovery: AuthorizationRecoveryPort;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
};

/** Owns the current incident authorization subject and explicit recovery. */
export function useWorkbookAuthorizationState({
  accountUserId,
  authorizationRecovery,
  incidentId,
  onIncidentAccessLost,
}: WorkbookAuthorizationStateOptions) {
  const [currentUserId, setCurrentUserId] = useState<string | null>(
    () => accountUserId ?? null,
  );
  const [currentIncidentRole, setCurrentIncidentRole] =
    useState<WorkbookIncidentRole | null>(null);

  useEffect(() => {
    if (accountUserId) {
      setCurrentUserId(accountUserId);
    }
  }, [accountUserId]);

  const acceptRecoveredAuthorization = useCallback(
    (result: {
      readonly role: WorkbookIncidentRole;
      readonly userId: string;
    }) => {
      setCurrentUserId(result.userId || null);
      setCurrentIncidentRole(result.role);
    },
    [],
  );
  const loadSessionRole = useCallback(async () => {
    const result = await authorizationRecovery.recover({
      incidentId,
      signal: new AbortController().signal,
    });
    if (result.kind !== "authorized") {
      setCurrentUserId(null);
      setCurrentIncidentRole("");
      onIncidentAccessLost?.();
      return;
    }
    acceptRecoveredAuthorization(result);
  }, [
    acceptRecoveredAuthorization,
    authorizationRecovery,
    incidentId,
    onIncidentAccessLost,
  ]);

  return {
    acceptRecoveredAuthorization,
    authorizationGeneration: `${currentUserId ?? "anonymous"}:${currentIncidentRole ?? "none"}`,
    currentIncidentRole,
    currentUserId,
    loadSessionRole,
  };
}
