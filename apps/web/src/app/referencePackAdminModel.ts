export type DeploymentAdminSession = {
  readonly is_deployment_admin: boolean;
};

export type ReferencePackVersion = {
  pack_key: string;
  pack_version: string;
  pack_kind: string;
  pack_version_state: string;
  active: boolean;
  source_identifier: string | null;
  manifest_sha256: string;
  payload_sha256: string;
  pack_contract_version: string;
  verification_method: string;
  verification_result: string;
  signer_key_id: string | null;
  previous_active_version: string | null;
  imported_by_user_id: string | null;
  imported_at: string;
  activated_by_user_id: string | null;
  activated_at: string | null;
};

export type ReferencePackJobResource = {
  job_id: string;
  status: string;
  cancelable: boolean;
  progress: {
    completed: number;
    total: number | null;
  };
  result_summary: { code: string } | null;
  error_summary: { code: string; details?: Record<string, unknown> } | null;
};

export type ReferencePackPaging = {
  limit?: number;
  has_more: boolean;
  next_cursor: string | null;
};

export type ReferencePackListEnvelope = {
  data: {
    pack_versions: ReferencePackVersion[];
  };
  meta?: {
    paging?: ReferencePackPaging;
  };
};

export type ReferencePackJobEnvelope = {
  data: ReferencePackJobResource;
};

export type ReferencePackQuery = {
  active: string;
  packVersionState: string;
  search: string;
  verificationResult: string;
};

export const terminalReferencePackJobStates = new Set([
  "succeeded",
  "failed",
  "canceled",
]);

export const defaultReferencePackPaging: ReferencePackPaging = {
  limit: 100,
  has_more: false,
  next_cursor: null,
};
