import {
  cartularyEvidenceInspectorContract,
  cartularyEvidenceVerificationContract as contract,
} from "@cartulary/ui-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { evidenceInspectorFixtureDigest } from "./evidenceInspectorTimingSupport";
import {
  type JobProgressCompletion,
  type JobProgressPaintInput,
  observeJobProgressPaintInBrowser,
} from "./jobProgressPaint";
import { resetWorkbookClientTiming } from "./timingSupport";

export async function measureEvidenceVerification(
  page: Page,
  input: Omit<
    JobProgressPaintInput,
    "jobPurpose" | "startMark" | "dispatchMark" | "admissionMark" | "timeoutMs"
  >,
  action: () => Promise<void>,
) {
  await resetWorkbookClientTiming(page);
  const pending = page
    .evaluate(observeJobProgressPaintInBrowser, {
      ...input,
      jobPurpose: "evidence_verification" as const,
      startMark: contract.start_mark,
      dispatchMark: contract.dispatch_mark,
      admissionMark: contract.admission_mark,
      timeoutMs: 5_000,
    })
    .then(
      (value) => ({ value }),
      (error) => ({ error }),
    );
  await action();
  const result = await pending;
  if ("error" in result) throw result.error;
  return result.value;
}

export async function attachEvidenceVerificationObservation(
  info: TestInfo,
  target?: { recordId: string; rowVersion: number; sourceRecordId: string },
  completion?: JobProgressCompletion,
  failureReason?: string,
) {
  const elapsed = completion?.elapsed ?? null;
  const observation = {
    schema_id: "cartulary.frontend_measurement_observation.v4",
    fixture_profile_id: contract.fixture_profile_id,
    criterion_id: contract.criterion_ids[0],
    predicate_id: contract.predicate_id,
    fixture_id: cartularyEvidenceInspectorContract.fixture.fixture_id,
    fixture_digest: evidenceInspectorFixtureDigest(),
    measurement_policy_id: contract.measurement_policy_id,
    threshold_ms: contract.threshold_ms,
    warmup_samples: 0,
    measured_samples: completion ? 1 : 0,
    statistics: { kind: "single", elapsed_ms: elapsed },
    outcome:
      process.env.CARTULARY_BROWSER_RESOURCE_PROFILE_ID !==
      "browser_measurement_quiet"
        ? "environment_not_qualified"
        : failureReason || elapsed === null
          ? "incomplete"
          : elapsed <= contract.threshold_ms
            ? "passed"
            : "threshold_failed",
    traffic: {
      counts: [
        { fact_id: "analyst_sessions", value: 1 },
        { fact_id: "background_sessions", value: 0 },
      ],
      rates: [],
      conditions: [],
    },
    samples: completion
      ? [
          {
            sample_index: 0,
            warmup: false,
            total_ms: completion.elapsed,
            stages_ms: [],
          },
        ]
      : [],
    predicate_evidence:
      completion && target && completion.blobId
        ? {
            schema_id: "cartulary.evidence_verification_measurement_proof.v1",
            kind: "evidence_verification",
            job_purpose: "evidence_verification",
            job_id: completion.jobId,
            job_status: completion.jobStatus,
            object_blob_id: completion.blobId,
            source_record_id: target.sourceRecordId,
            record_id: target.recordId,
            row_version: target.rowVersion,
            view_schema_id: contract.view_schema_id,
            field_key: contract.workload.field_key,
            clipboard_driver: contract.workload.clipboard_driver,
            uploaded_bytes: contract.workload.object_bytes,
            preview_preflight_records:
              contract.workload.preview_preflight_records,
            inspector_open: false,
            attachment_details_expanded: false,
            cancel_enabled: true,
            grid_edit_accepted: true,
            draft_uncommitted: true,
            consecutive_frames: 2,
            binding_stable: true,
            geometry_unchanged: true,
            scroll_unchanged: true,
            action_count: 1,
            viewport_width: contract.viewport.width,
            viewport_height: contract.viewport.height,
            device_scale_factor: contract.viewport.device_scale_factor,
          }
        : null,
    ...(failureReason ? { failure_reason: failureReason } : {}),
  };
  await info.attach(`${observation.schema_id}.${contract.predicate_id}`, {
    body: JSON.stringify(observation),
    contentType: "application/json",
  });
}
