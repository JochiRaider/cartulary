export function isNetworkFlowAuthorizationLoss(message: string): boolean {
  return (
    message.includes("authorization_denied") ||
    message.includes("incident_not_found") ||
    message.includes("session_required")
  );
}
