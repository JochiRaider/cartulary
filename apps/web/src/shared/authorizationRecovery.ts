import type { WorkbookIncidentRole } from "./workbookShellContracts";

export type AuthorizationRecoveryResult =
  | {
      readonly kind: "authorized";
      readonly role: WorkbookIncidentRole;
      readonly userId: string;
    }
  | { readonly kind: "access_lost" }
  | { readonly kind: "cancelled" }
  | { readonly kind: "session_lost" }
  | {
      readonly kind: "unavailable";
      readonly failure: "transient" | "contract";
    };

export interface AuthorizationRecoveryPort {
  recover(input: {
    readonly incidentId: string;
    readonly signal: AbortSignal;
  }): Promise<AuthorizationRecoveryResult>;
}
