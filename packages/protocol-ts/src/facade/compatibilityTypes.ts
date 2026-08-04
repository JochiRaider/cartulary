import type {
  EnvelopeMeta,
  ErrorEnvelope,
  ExtensionMappingPreviewEnvelope,
  ExtensionMappingPreviewRequest,
  ExtensionMappingPreviewResource,
  DensityMode as GeneratedDensityMode,
  EvidenceHandleIssueRequest as GeneratedEvidenceHandleIssueRequest,
  ImportSourceColumnMapping,
} from "../generated/core-http-types.js";

export type {
  EnvelopeMeta,
  ErrorEnvelope,
  ExtensionMappingPreviewEnvelope,
  ExtensionMappingPreviewRequest,
  ExtensionMappingPreviewResource,
  ImportSourceColumnMapping,
};

export type ViewCell = {
  readonly value?: unknown;
  readonly [key: string]: unknown;
};

export type ViewRow = {
  readonly cells: Record<string, ViewCell>;
  readonly record_id: string;
  readonly row_version: number;
  readonly [key: string]: unknown;
};

export type ViewMutationData = {
  readonly change_set_id: string;
  readonly row: ViewRow;
  readonly view_schema_id: string;
};

export type DensityMode = GeneratedDensityMode;

export type AccountProfileResource = {
  readonly user_id: string;
  readonly email: string;
  readonly display_name: string;
  readonly user_version: number;
  readonly created_at: string;
  readonly updated_at: string;
};

export type AccountPreferencesResource = {
  readonly user_id: string;
  readonly density_mode: DensityMode | null;
  readonly preferences_version: number;
  readonly created_at: string;
  readonly updated_at: string;
};

export type AccountProfilePatchRequest = {
  readonly base_user_version: number;
  readonly client_txn_id: string;
  readonly display_name: string;
};

export type AccountPreferencesPutRequest = {
  readonly base_preferences_version: number;
  readonly client_txn_id: string;
  readonly density_mode: DensityMode | null;
};

export type AccountProfileEnvelope = {
  readonly data: AccountProfileResource;
  readonly meta: EnvelopeMeta;
};

export type AccountPreferencesEnvelope = {
  readonly data: AccountPreferencesResource;
  readonly meta: EnvelopeMeta;
};

export type ObjectBlobCreateRequest = {
  readonly byte_size: number;
  readonly client_txn_id: string;
  readonly content_type_hint?: string | null;
  readonly filename_hint?: string | null;
  readonly incident_id: string;
  readonly sha256_hex?: string | null;
};

export type ObjectBlobUploadTarget = {
  readonly expires_at: string;
  readonly headers: Record<string, string>;
  readonly href: string;
  readonly method: "PUT";
};

export type ObjectBlobCreateEnvelope = {
  readonly data: {
    readonly accepted_contract: {
      readonly byte_size: number;
      readonly content_type_hint: string | null;
      readonly filename_hint: string | null;
      readonly incident_id: string;
      readonly sha256_hex: string | null;
    };
    readonly incident_id: string;
    readonly object_blob_id: string;
    readonly pending_expires_at: string;
    readonly target_expires_at: string;
    readonly upload_state: "pending";
    readonly upload_target: ObjectBlobUploadTarget;
  };
  readonly meta: EnvelopeMeta;
};

export type EvidenceAttachBlobRequest = {
  readonly base_row_version: number;
  readonly client_txn_id: string;
  readonly object_blob_id: string;
};

export type EvidenceAttachBlobEnvelope = {
  readonly data: ViewMutationData & {
    readonly object_blob_id: string;
  };
  readonly meta: EnvelopeMeta;
};

export type EvidenceHandleIssueRequest = GeneratedEvidenceHandleIssueRequest;

export type EvidenceHandleEnvelope = {
  readonly data: {
    readonly content_type: string;
    readonly disposition: "inline" | "attachment";
    readonly evidence_lifecycle_state: string;
    readonly expires_at: string;
    readonly filename: string;
    readonly handle_kind: "preview" | "download";
    readonly href: string;
    readonly incident_id: string;
    readonly media_class: string;
    readonly method: "GET";
    readonly object_blob_id: string;
    readonly preview_kind?: string;
    readonly record_id: string;
    readonly sha256: string | null;
    readonly single_use: boolean;
    readonly size_bytes: number;
    readonly upload_state: string;
  };
  readonly meta: EnvelopeMeta;
};
