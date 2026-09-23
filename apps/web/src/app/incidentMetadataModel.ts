import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import type {
  IncidentMetadataResource,
  PatchIncidentRequest,
} from "./api/incidentMetadataContracts";

export type { IncidentMetadataResource } from "./api/incidentMetadataContracts";
export const metadataFields = [
  "description",
  "severity",
  "tlp",
  "current_phase",
  "primary_external_case_ref",
] as const;
export type IncidentMetadataField = (typeof metadataFields)[number];
export type MetadataValues = Record<IncidentMetadataField, string>;
export type MetadataRevisions = Record<IncidentMetadataField, number>;
export const metadataLabels: Record<IncidentMetadataField, string> = {
  description: "Description",
  severity: "Severity",
  tlp: "TLP",
  current_phase: "Current phase",
  primary_external_case_ref: "Primary external case",
};
export const metadataTLPs = [
  "TLP:CLEAR",
  "TLP:GREEN",
  "TLP:AMBER",
  "TLP:AMBER+STRICT",
  "TLP:RED",
] as const satisfies readonly NonNullable<PatchIncidentRequest["tlp"]>[];
export type MetadataAuthority = Readonly<{
  incidentId: string;
  actorId: string;
  lifetime: string;
  role: WorkbookIncidentRole;
  apiBase?: string | undefined;
}>;
export const metadataBinding = (a: MetadataAuthority) =>
  JSON.stringify([a.incidentId, a.actorId, a.lifetime, a.apiBase ?? ""]);
export const metadataValues = (
  r: IncidentMetadataResource,
): MetadataValues => ({
  description: r.description ?? "",
  severity: r.severity ?? "",
  tlp: r.tlp ?? "",
  current_phase: r.current_phase ?? "",
  primary_external_case_ref: r.primary_external_case_ref ?? "",
});
const metadataRevisions = (): MetadataRevisions => ({
  description: 0,
  severity: 0,
  tlp: 0,
  current_phase: 0,
  primary_external_case_ref: 0,
});
export type MetadataDraft = Readonly<{
  base: IncidentMetadataResource;
  values: MetadataValues;
  revisions: MetadataRevisions;
  revision: number;
  resetRevision: number;
  reviewRequired: boolean;
}>;
export const newMetadataDraft = (
  base: IncidentMetadataResource,
): MetadataDraft => ({
  base,
  values: metadataValues(base),
  revisions: metadataRevisions(),
  revision: 0,
  resetRevision: 0,
  reviewRequired: false,
});
export const changedMetadataFields = (draft: MetadataDraft) => {
  const saved = metadataValues(draft.base);
  return metadataFields.filter((field) => draft.values[field] !== saved[field]);
};
export const metadataDirty = (draft: MetadataDraft | null) =>
  draft !== null && changedMetadataFields(draft).length > 0;

/** Raw text reaches the existing normalization owner. Exact empty input is an explicit clear. */
export function buildIncidentMetadataPatch(
  draft: MetadataDraft,
): Readonly<PatchIncidentRequest> {
  const changed = changedMetadataFields(draft);
  const text = (field: Exclude<IncidentMetadataField, "tlp">) =>
    draft.values[field] === "" ? null : draft.values[field];
  const tlp = metadataTLPs.find((token) => token === draft.values.tlp);
  if (draft.values.tlp !== "" && !tlp) throw new Error("Invalid TLP selection");
  return Object.freeze({
    base_incident_version: draft.base.incident_version,
    ...(changed.includes("description")
      ? { description: text("description") }
      : {}),
    ...(changed.includes("severity") ? { severity: text("severity") } : {}),
    ...(changed.includes("tlp") ? { tlp: tlp ?? null } : {}),
    ...(changed.includes("current_phase")
      ? { current_phase: text("current_phase") }
      : {}),
    ...(changed.includes("primary_external_case_ref")
      ? { primary_external_case_ref: text("primary_external_case_ref") }
      : {}),
  });
}
export const metadataProblemCodes = [
  "invalid_incident_patch",
  "incident_version_conflict",
  "incident_closed",
  "incident_not_found",
  "authorization_denied",
  "authentication_required",
  "csrf_failed",
  "internal_error",
  "invalid_public_contract_response",
  "unknown_public_error",
] as const;
export type MetadataProblem = Readonly<{
  code: (typeof metadataProblemCodes)[number];
  field?: IncidentMetadataField | undefined;
  reason?:
    | "field_too_long"
    | "control_character_not_allowed"
    | "invalid_value"
    | undefined;
}>;
export type MetadataAttempt = Readonly<{
  id: number;
  authority: MetadataAuthority;
  base: IncidentMetadataResource;
  revision: number;
  revisions: MetadataRevisions;
  values: MetadataValues;
  fields: readonly IncidentMetadataField[];
  payload: Readonly<PatchIncidentRequest>;
}>;
export type MetadataOperation =
  | { readonly kind: "idle" }
  | {
      readonly kind: "pending";
      readonly attempt: MetadataAttempt;
      readonly stage: "authorization" | "write";
    }
  | {
      readonly kind: "confirmed";
      readonly attempt: MetadataAttempt;
      readonly resource: IncidentMetadataResource;
    }
  | {
      readonly kind: "rejected";
      readonly attempt: MetadataAttempt;
      readonly problem: MetadataProblem;
    }
  | {
      readonly kind: "conflicted" | "uncertain";
      readonly attempt: MetadataAttempt;
    };
export type MetadataState = Readonly<{
  authority: MetadataAuthority | null;
  active: boolean;
  access: "checking" | "ready" | "unavailable";
  resource: IncidentMetadataResource | null;
  draft: MetadataDraft | null;
  read: "idle" | "loading" | "refreshing" | "ready" | "failed";
  readGeneration: number;
  operation: MetadataOperation;
  transportPending: boolean;
  reviewNotice: string;
  fieldErrors: Partial<
    Record<IncidentMetadataField, { revision: number; message: string }>
  >;
  departure: boolean;
}>;
export const initialMetadataState = (): MetadataState => ({
  authority: null,
  active: false,
  access: "checking",
  resource: null,
  draft: null,
  read: "idle",
  readGeneration: 0,
  operation: { kind: "idle" },
  transportPending: false,
  reviewNotice: "",
  fieldErrors: {},
  departure: false,
});
export const metadataEditRole = (role: WorkbookIncidentRole | undefined) =>
  role === "reviewer" || role === "admin";
export function metadataFieldError(problem: MetadataProblem) {
  if (problem.reason === "field_too_long")
    return "Shorten this value and save again.";
  if (problem.reason === "control_character_not_allowed")
    return "Remove unsupported control characters and save again.";
  if (problem.field === "tlp") return "Select a listed TLP value or Unset.";
  return "Use valid text within this field's stated limit, without unsupported control characters.";
}
