import { isSheetRef, type SheetRef } from "../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";

export type PreferenceKind = "home" | "default";
export type PreferenceAuthority = Readonly<{
  incidentId: string;
  actorId: string;
  lifetime: string;
  role: WorkbookIncidentRole;
  apiBase?: string | undefined;
}>;
type PreferenceResource = Readonly<{
  incident_id: string;
  created_at: string;
  updated_at: string;
}>;
export type HomePreference = PreferenceResource &
  Readonly<{ user_id: string; home_sheet_ref: SheetRef | null }>;
export type DefaultPreference = PreferenceResource &
  Readonly<{ updated_by_user_id: string; default_sheet_ref: SheetRef | null }>;
export type PreferenceResources = {
  home: HomePreference;
  default: DefaultPreference;
};
export type PreferenceProblem = Readonly<{
  code:
    | "invalid_mutation_payload"
    | "invalid_query_request"
    | "invalid_path_parameter"
    | "authentication_required"
    | "authorization_denied"
    | "csrf_failed"
    | "incident_not_found"
    | "internal_error"
    | "invalid_public_contract_response"
    | "transport"
    | "timeout"
    | "access_unavailable"
    | "unknown_public_error";
}>;
export type PreferenceResult<T> =
  | { readonly kind: "accepted"; readonly value: T }
  | {
      readonly kind: "rejected" | "uncertain";
      readonly status: number;
      readonly problem: PreferenceProblem;
    };
export type PreferenceAttempt = Readonly<{
  id: number;
  kind: PreferenceKind;
  authority: PreferenceAuthority;
  target: SheetRef | null;
}>;
export type PreferenceOperation =
  | { readonly kind: "idle" }
  | { readonly kind: "reviewed"; readonly attempt: PreferenceAttempt }
  | {
      readonly kind: "pending";
      readonly attempt: PreferenceAttempt;
      readonly stage: "authorization" | "write";
    }
  | {
      readonly kind: "confirmed";
      readonly attempt: PreferenceAttempt;
      readonly resource: HomePreference | DefaultPreference;
    }
  | {
      readonly kind: "rejected" | "uncertain";
      readonly attempt: PreferenceAttempt;
      readonly problem: PreferenceProblem;
    };
export type PreferenceSlot<K extends PreferenceKind = PreferenceKind> =
  Readonly<{
    resource: PreferenceResources[K] | null;
    read: "idle" | "loading" | "refreshing" | "ready" | "failed";
    problem: PreferenceProblem | null;
    operation: PreferenceOperation;
    transportPending: boolean;
    observation: number | null;
  }>;
export type PreferenceSurface = Readonly<{
  savedViewLabels?: readonly Readonly<{ id: string; label: string }>[];
  sheetRef: SheetRef;
  label: string | null;
  available: boolean;
}>;
export type PreferenceSnapshot = Readonly<{
  inspectionActive: boolean;
  authority: PreferenceAuthority | null;
  access: "checking" | "ready" | "unavailable";
  home: PreferenceSlot<"home">;
  default: PreferenceSlot<"default">;
  surface: PreferenceSurface | null;
  departure: boolean;
  announcement: Readonly<{ id: number; text: string }> | null;
}>;
export const emptyPreferenceSlot = <
  K extends PreferenceKind,
>(): PreferenceSlot<K> => ({
  resource: null,
  read: "idle",
  problem: null,
  operation: { kind: "idle" },
  transportPending: false,
  observation: null,
});
export function preferencePointer(
  resource: HomePreference | DefaultPreference,
): SheetRef | null {
  return "home_sheet_ref" in resource
    ? resource.home_sheet_ref
    : resource.default_sheet_ref;
}
export function preferencePointersEqual(
  a: SheetRef | null,
  b: SheetRef | null,
): boolean {
  if (a === null || b === null) return a === b;
  if (a.kind !== b.kind) return false;
  if (a.kind === "extension_workspace" && b.kind === "extension_workspace")
    return (
      a.extension_profile_id === b.extension_profile_id &&
      a.workspace_key === b.workspace_key
    );
  return "id" in a && "id" in b && a.id === b.id;
}
export function frozenPreferencePointer(
  value: SheetRef | null,
): SheetRef | null {
  return value === null ? null : Object.freeze({ ...value });
}
export function validPreferenceResource<K extends PreferenceKind>(
  value: unknown,
  kind: K,
  authority: Pick<PreferenceAuthority, "incidentId" | "actorId">,
): value is PreferenceResources[K] {
  if (!value || typeof value !== "object" || Array.isArray(value)) return false;
  const r = value as Record<string, unknown>;
  const pointer = r[kind === "home" ? "home_sheet_ref" : "default_sheet_ref"];
  return (
    r.incident_id === authority.incidentId &&
    (kind === "home"
      ? r.user_id === authority.actorId
      : typeof r.updated_by_user_id === "string" &&
        r.updated_by_user_id !== "") &&
    typeof r.created_at === "string" &&
    Number.isFinite(Date.parse(r.created_at)) &&
    typeof r.updated_at === "string" &&
    Number.isFinite(Date.parse(r.updated_at)) &&
    (pointer === null || isSheetRef(pointer))
  );
}
export function preferenceProblemText(problem: PreferenceProblem): string {
  switch (problem.code) {
    case "invalid_mutation_payload":
      return "The surface reference was rejected. Check the current authorized surface before trying again.";
    case "csrf_failed":
      return "The request was rejected by the session protection check. Refresh current access before trying again.";
    case "authorization_denied":
      return "Current incident permissions do not allow this update.";
    case "authentication_required":
      return "The current session must be checked.";
    case "incident_not_found":
      return "Current incident access must be checked.";
    case "timeout":
      return "The observation deadline expired. This does not establish server cancellation.";
    case "access_unavailable":
      return "Current access could not be checked. Local recovery is retained.";
    case "invalid_public_contract_response":
      return "The server response could not be validated.";
    default:
      return "The preference request could not be completed. Refresh the current value.";
  }
}

export type PreferenceWorkbookBinding = Readonly<{
  incidentId: string;
  actorId: string | null;
  apiBase: string | undefined;
  surface: PreferenceSurface | null;
  onAuthorizationRecovered: (result: {
    role: WorkbookIncidentRole;
    userId: string;
  }) => void;
}>;
