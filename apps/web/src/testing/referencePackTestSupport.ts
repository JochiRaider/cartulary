import type { ListReferencePacksResponse } from "@cartulary/protocol-ts/http";
import type {
  ReferencePackJobResource,
  ReferencePackVersion,
} from "../services/referencePacks";
export const referencePackTestActor = "22222222-2222-4222-8222-222222222222";
export const referencePackTestJobId = "11111111-1111-4111-8111-111111111111";
export function referencePackFixture(
  overrides: Partial<ReferencePackVersion> = {},
): ReferencePackVersion {
  return {
    pack_key: "type_registry.host",
    pack_kind: "type_registry",
    pack_version: "1",
    pack_version_state: "verified_available",
    active: false,
    source_identifier: null,
    manifest_sha256: "a".repeat(64),
    payload_sha256: "b".repeat(64),
    pack_contract_version: "cartulary.reference_pack.v1",
    verification_method: "manifest_sha256_v1",
    verification_result: "passed",
    signer_key_id: null,
    previous_active_version: null,
    imported_by_user_id: null,
    imported_at: "2026-01-01T00:00:00Z",
    activated_by_user_id: null,
    activated_at: null,
    ...overrides,
  };
}
export function referencePackListFixture(
  pack_versions: ReferencePackVersion[] = [],
  paging: ListReferencePacksResponse["meta"]["paging"] = {
    limit: 100,
    has_more: false,
    next_cursor: null,
  },
): ListReferencePacksResponse {
  return {
    data: { pack_versions },
    meta: { request_id: "reference-list", paging },
  };
}
export function referencePackJobFixture(
  status: ReferencePackJobResource["status"] = "queued",
  overrides: Partial<ReferencePackJobResource> = {},
): ReferencePackJobResource {
  const terminal = ["succeeded", "failed", "canceled"].includes(status);
  return {
    job_id: referencePackTestJobId,
    scope: { kind: "deployment" },
    status_route: `/api/v1/jobs/${referencePackTestJobId}`,
    status,
    cancelable: status === "queued" || status === "running",
    submitted_by_user_id: referencePackTestActor,
    submitted_at: "2026-01-01T00:00:00Z",
    updated_at: terminal
      ? "2026-01-01T00:00:03Z"
      : status === "queued"
        ? "2026-01-01T00:00:00Z"
        : "2026-01-01T00:00:02Z",
    started_at: status === "queued" ? null : "2026-01-01T00:00:01Z",
    finished_at: terminal ? "2026-01-01T00:00:03Z" : null,
    retained_until: terminal ? "2026-01-09T00:00:03Z" : null,
    progress: { completed: status === "succeeded" ? 1 : 0, total: 1 },
    result_summary:
      status === "succeeded"
        ? {
            code: "reference_packs_refreshed",
            message: "Refreshed",
            resource_refs: [],
          }
        : status === "canceled"
          ? { code: "job_canceled", message: "Canceled" }
          : null,
    error_summary:
      status === "failed"
        ? {
            code: "reference_pack_verification_failed",
            message: "Verification failed",
            retryable: false,
          }
        : null,
    ...overrides,
  };
}
