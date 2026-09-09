import type { ImportWriteAttempt } from "../imports/importRequests";
import type {
  ImportFailure,
  ImportWriteReceipt,
} from "../services/importClient";
import type {
  DiscoveredImportUnit,
  ImportJobResource,
  ImportSelectionReceipt,
} from "../services/importContractAdapter";
import type {
  NetworkFlowImportPreviewResult,
  NetworkFlowTable,
} from "../services/networkFlowContractAdapter";
import {
  type NetworkFlowImportDiscovery,
  type NetworkFlowMappingDraft,
  networkFlowMappingIssues,
} from "./networkFlowImportModel";

export type NetworkFlowImportStage =
  | "idle"
  | "uploading"
  | "discovering"
  | "loading_source"
  | "mapping"
  | "previewing"
  | "approving"
  | "selecting"
  | "submitting_apply"
  | "observing_apply"
  | "handoff"
  | "finished";
export type NetworkFlowImportAttempt = {
  readonly request: ImportWriteAttempt;
  readonly disposition: "pending" | "uncertain" | "rejected" | "accepted";
  readonly failure: ImportFailure | null;
  readonly receipt: ImportWriteReceipt | null;
};
export type NetworkFlowImportJob = {
  readonly attempt: ImportWriteAttempt;
  readonly resource: ImportJobResource;
  readonly current: boolean;
  readonly observing: boolean;
  readonly failure: ImportFailure | null;
};
export type NetworkFlowImportPreview = {
  readonly candidateKey: string;
  readonly authority: number;
  readonly value: NetworkFlowImportPreviewResult;
};
export type NetworkFlowImportApproval = {
  readonly attempt: ImportWriteAttempt;
  readonly candidateKey: string;
  readonly fingerprint: string;
  readonly unit: DiscoveredImportUnit;
};
export type NetworkFlowImportHandoffRequest = {
  readonly incidentId: string;
  readonly sessionId: string;
  readonly unitId: string;
  readonly sourceHash: string;
  readonly fingerprint: string;
  readonly tableId: string;
  readonly signal: AbortSignal;
  readonly current: () => boolean;
};
export type NetworkFlowImportHandoffResult =
  | { readonly kind: "selected"; readonly table: NetworkFlowTable }
  | { readonly kind: "unavailable" | "superseded" }
  | { readonly kind: "failed"; readonly failure: ImportFailure };
export type NetworkFlowImportHandoff = {
  readonly status:
    | "pending"
    | "loading"
    | "selected"
    | "unavailable"
    | "failed";
  readonly tableId: string | null;
  readonly failure: ImportFailure | null;
};
export type NetworkFlowImportState = {
  readonly revision: number;
  readonly access: "active" | "paused" | "unavailable";
  readonly canWrite: boolean;
  readonly closed: boolean;
  readonly presented: boolean;
  readonly workspaceActive: boolean;
  readonly stage: NetworkFlowImportStage;
  readonly discovery: NetworkFlowImportDiscovery | null;
  readonly draft: NetworkFlowMappingDraft | null;
  readonly preview: NetworkFlowImportPreview | null;
  readonly previewStale: boolean;
  readonly previewFailure: ImportFailure | null;
  readonly sourceFailure: ImportFailure | null;
  readonly write: NetworkFlowImportAttempt | null;
  readonly approval: NetworkFlowImportApproval | null;
  readonly selection: ImportSelectionReceipt | null;
  readonly discoveryJob: NetworkFlowImportJob | null;
  readonly applyJob: NetworkFlowImportJob | null;
  readonly cancellation: NetworkFlowImportAttempt | null;
  readonly cancellationChecking: boolean;
  readonly handoff: NetworkFlowImportHandoff | null;
  readonly message: string;
};
export function initialNetworkFlowImportState(): NetworkFlowImportState {
  return {
    revision: 0,
    access: "unavailable",
    canWrite: false,
    closed: false,
    presented: false,
    workspaceActive: false,
    stage: "idle",
    discovery: null,
    draft: null,
    preview: null,
    previewStale: false,
    previewFailure: null,
    sourceFailure: null,
    write: null,
    approval: null,
    selection: null,
    discoveryJob: null,
    applyJob: null,
    cancellation: null,
    cancellationChecking: false,
    handoff: null,
    message: "Choose a NetFlow CSV source.",
  };
}
export function networkFlowImportMappingState(state: NetworkFlowImportState) {
  if (!state.draft) return null;
  if (state.previewFailure) return "validation_preview_failed";
  if (
    state.approval &&
    state.preview?.value.materialized_mapping.source_profile_id !==
      state.draft.sourceProfileId
  )
    return "mapping_required";
  if (networkFlowMappingIssues(state.draft).length) return "mapping_required";
  if (state.stage === "previewing" || state.previewStale)
    return "validation_preview_pending";
  return state.preview ? "validation_preview_ready" : "mapping_ready";
}
/** Only the first three Table 7-B tokens belong to this import seam. */
export function networkFlowImportStatus(state: NetworkFlowImportState) {
  if (state.applyJob) return null;
  const mapping = networkFlowImportMappingState(state);
  if (
    state.sourceFailure ||
    state.previewFailure ||
    state.write?.disposition === "rejected"
  )
    return "validation_failed";
  if (mapping === "mapping_required") return "mapping_required";
  return state.stage === "previewing" ? "validating" : null;
}
export function unresolvedNetworkFlowWrite(state: NetworkFlowImportState) {
  return (
    state.write?.disposition === "pending" ||
    state.write?.disposition === "uncertain"
  );
}
export function networkFlowImportHasWork(state: NetworkFlowImportState) {
  return (
    state.stage !== "idle" ||
    state.write !== null ||
    state.discoveryJob !== null
  );
}
export function networkFlowImportTableMatches(
  table: NetworkFlowTable,
  request: NetworkFlowImportHandoffRequest,
) {
  return (
    table.network_flow_table_id === request.tableId &&
    table.incident_id === request.incidentId &&
    table.table_status === "active" &&
    table.source_import_session_id === request.sessionId &&
    table.source_import_unit_id === request.unitId &&
    table.source_content_sha256 === request.sourceHash &&
    table.mapping_fingerprint === request.fingerprint
  );
}
