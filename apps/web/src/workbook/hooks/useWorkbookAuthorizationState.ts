import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import type { AuthorizationRecoveryPort } from "../../shared/authorizationRecovery";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";

type WorkbookAuthorizationStateOptions = {
  readonly accountUserId: string | undefined;
  readonly authorizationRecovery: AuthorizationRecoveryPort;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly onSessionLost?: (() => void) | undefined;
};

/** Owns incident authorization presentation; the application owns session acceptance. */
export function useWorkbookAuthorizationState({
  accountUserId,
  authorizationRecovery,
  incidentId,
  onIncidentAccessLost,
  onSessionLost,
}: WorkbookAuthorizationStateOptions) {
  const subject = useMemo(
    () => ({ accountUserId, authorizationRecovery, incidentId }),
    [accountUserId, authorizationRecovery, incidentId],
  );
  const [currentUserId, setCurrentUserId] = useState<string | null>(
    () => accountUserId ?? null,
  );
  const [currentIncidentRole, setCurrentIncidentRole] =
    useState<WorkbookIncidentRole | null>(null);
  const request = useRef<AbortController | null>(null);
  useLayoutEffect(() => {
    setCurrentUserId(subject.accountUserId ?? null);
    setCurrentIncidentRole(null);
    return () => {
      request.current?.abort();
      request.current = null;
    };
  }, [subject]);
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
    request.current?.abort();
    const controller = new AbortController();
    request.current = controller;
    try {
      const result = await subject.authorizationRecovery.recover({
        incidentId: subject.incidentId,
        signal: controller.signal,
      });
      if (controller.signal.aborted || request.current !== controller) return;
      if (result.kind === "authorized") {
        acceptRecoveredAuthorization(result);
        return;
      }
      if (result.kind === "session_lost") {
        onSessionLost?.();
        return;
      }
      if (result.kind !== "access_lost") return;
      setCurrentUserId(null);
      setCurrentIncidentRole("");
      onIncidentAccessLost?.();
    } finally {
      if (request.current === controller) request.current = null;
    }
  }, [
    acceptRecoveredAuthorization,
    subject,
    onIncidentAccessLost,
    onSessionLost,
  ]);
  return {
    acceptRecoveredAuthorization,
    authorizationGeneration: `${currentUserId ?? "anonymous"}:${currentIncidentRole ?? "none"}`,
    currentIncidentRole,
    currentUserId,
    loadSessionRole,
  };
}
