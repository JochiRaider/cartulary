import type { ListReferencePacksResponse } from "@cartulary/protocol-ts/http";
import type {
  ReferencePackJobResource,
  ReferencePackValidationSummary,
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
    health: "verified_available" as const,
    administratively_disabled: false,
    distribution_kind: "operator_imported" as const,
    removed: false,
    pending_work: false,
    reproducibility_pinned: false,
    fallback_from_version: null,
    dependencies: [],
    missing_reason: null,
    last_failure_code: null,
    content_profile_id: "cartulary.reference_pack.type_registry.host.v1",
    content_profile_version: "1",
    pack_release_sequence: 1,
    source_profile_id: "test.fixture.v1",
    source_profile_sha256: "d".repeat(64),
    source_version: "1",
    source_as_of: null,
    license_expression: "MIT",
    redistribution: "allowed",
    trust_repository_id: "test.repo",
    last_verified_at: "2026-05-24T00:00:00Z",
    trust_valid_until: "2027-05-24T00:00:00Z",
    verified_signer_key_ids: ["c".repeat(64)],
    pack_version: "1",
    pack_version_state: "verified_available",
    active: false,
    source_identifier: null,
    manifest_sha256: "a".repeat(64),
    payload_sha256: "b".repeat(64),
    pack_contract_version: "cartulary.reference_pack_contract.v1",
    verification_method: "tuf_1_0_35_offline_bundle_v1",

    previous_active_version: null,
    imported_by_user_id: null,
    imported_at: "2026-01-01T00:00:00Z",
    activated_by_user_id: null,
    activated_at: null,
    ...(overrides.pack_version_state === "disabled"
      ? { administratively_disabled: true }
      : {}),
    ...(["staged", "failed", "missing"].includes(
      overrides.pack_version_state ?? "",
    )
      ? {
          health:
            overrides.pack_version_state as ReferencePackVersion["health"],
        }
      : {}),
    ...(overrides.pack_version_state === "staged"
      ? {
          last_verified_at: null,
          trust_valid_until: null,
          manifest_sha256: null,
          payload_sha256: null,
          pack_contract_version: null,
          content_profile_id: null,
          content_profile_version: null,
          pack_release_sequence: null,
          source_profile_id: null,
          source_profile_sha256: null,
          source_version: null,
          source_as_of: null,
          license_expression: null,
          redistribution: null,
          trust_repository_id: null,
          verification_method: null,
          verified_signer_key_ids: [],
          source_identifier: null,
        }
      : {}),
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

export const referencePackTestValidationRef = `rpvs_${"e".repeat(64)}`;
export function referencePackValidationFixture(): ReferencePackValidationSummary {
  const issue = {
    issue_id: `rpi_${"a".repeat(64)}`,
    check_id: "content_schema",
    severity: "error" as const,
    phase: "content_schema",
    code: "content_schema_invalid",
    reason_code: "content_schema_invalid",
    path: '$.entries[0]["<unknown>"]',
    entry_id: null,
    safe_details: {
      limit_id: null,
      expected_token: "unknown_member",
      actual_token: null,
      related_pack_key: null,
      related_pack_version: null,
    },
  };
  return {
    schema_id: "cartulary.reference_pack_validation_summary.v1",
    result: "failed",
    primary_issue_id: issue.issue_id,
    issues_truncated: false,
    total_issue_count: 1,
    retained_issue_count: 1,
    issues: [issue],
  };
}
export function referencePackRejectedJobFixture(
  summary = referencePackValidationFixture(),
): ReferencePackJobResource {
  return referencePackJobFixture("failed", {
    error_summary: {
      code: "reference_pack_verification_failed",
      message: "Verification failed",
      retryable: false,
      details: {
        reason_code: "content_schema_invalid",
        check_id: "content_schema",
        failed_count: 1,
        primary_issue_id: summary.primary_issue_id,
        validation_summary_ref: referencePackTestValidationRef,
        total_issue_count: summary.total_issue_count,
        retained_issue_count: summary.retained_issue_count,
        issues_truncated: summary.issues_truncated,
      },
    },
  });
}
