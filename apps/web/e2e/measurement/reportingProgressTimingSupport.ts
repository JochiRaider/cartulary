import {
  type CartularyReportingProgressPredicateId,
  cartularyLargeGridFixture,
  cartularyReportingProgressContract as contract,
} from "@cartulary/ui-contracts";
import type { Page, TestInfo } from "@playwright/test";
import {
  type JobProgressCompletion,
  type JobProgressPaintInput,
  observeJobProgressPaintInBrowser,
} from "./jobProgressPaint";
import { ac043FixtureDigest, resetWorkbookClientTiming } from "./timingSupport";

export type ReportingProgressPurpose = "snapshot" | "report";
export async function measureReportingProgress(
  purpose: ReportingProgressPurpose,
  page: Page,
  input: Omit<
    JobProgressPaintInput,
    "jobPurpose" | "startMark" | "dispatchMark" | "admissionMark" | "timeoutMs"
  >,
  action: () => Promise<void>,
) {
  await resetWorkbookClientTiming(page);
  const observation = page
    .evaluate(observeJobProgressPaintInBrowser, {
      ...input,
      jobPurpose: purpose,
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
  const result = await observation;
  if ("error" in result) throw result.error;
  return result.value;
}

export async function attachReportingProgressObservation(
  info: TestInfo,
  predicateId: CartularyReportingProgressPredicateId,
  target?: { recordId: string; rowVersion: number; snapshotId: string | null },
  completion?: JobProgressCompletion,
  failureReason?: string,
) {
  const predicate = contract.predicates.find(
    (value) => value.predicate_id === predicateId,
  );
  if (!predicate) throw new Error("Reporting predicate is not admitted");
  const elapsed = completion?.elapsed ?? null;
  const observation = {
    schema_id: "cartulary.frontend_measurement_observation.v4",
    fixture_profile_id: contract.fixture_profile_id,
    criterion_id: contract.criterion_id,
    predicate_id: predicateId,
    fixture_id: cartularyLargeGridFixture.fixtureId,
    fixture_digest: ac043FixtureDigest(),
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
        {
          fact_id: "analyst_sessions",
          value: cartularyLargeGridFixture.analystSessions,
        },
        {
          fact_id: "background_sessions",
          value: cartularyLargeGridFixture.backgroundSessions,
        },
      ],
      rates: [
        {
          fact_id: "background_updates_per_second",
          value: cartularyLargeGridFixture.backgroundUpdatesPerSecond,
        },
      ],
      conditions: [
        { fact_id: "presence_enabled", value: true },
        { fact_id: "target_row_excluded", value: true },
      ],
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
      completion && target
        ? {
            schema_id: "cartulary.reporting_progress_measurement_proof.v1",
            kind: "reporting_progress",
            job_purpose: predicate.kind,
            source_snapshot_id: target.snapshotId,
            job_id: completion.jobId,
            job_status: completion.jobStatus,
            record_id: target.recordId,
            row_version: target.rowVersion,
            view_schema_id: contract.view_schema_id,
            field_key: contract.workload.field_key,
            cancel_enabled: true,
            grid_edit_accepted: true,
            draft_uncommitted: true,
            consecutive_frames: 2,
            binding_stable: true,
            geometry_unchanged: true,
            scroll_unchanged: true,
            action_count: 1,
            viewport_width: 1920,
            viewport_height: 1080,
            device_scale_factor: 1,
          }
        : null,
    ...(failureReason ? { failure_reason: failureReason } : {}),
  };
  await info.attach(`${observation.schema_id}.${predicateId}`, {
    body: JSON.stringify(observation),
    contentType: "application/json",
  });
}
