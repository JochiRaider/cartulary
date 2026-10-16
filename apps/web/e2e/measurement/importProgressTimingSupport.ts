import {
  type JobProgressCompletion,
  type JobProgressPaintInput,
  observeJobProgressPaintInBrowser,
} from "./jobProgressPaint";
export type ImportProgressCompletion = JobProgressCompletion;

import {
  cartularyImportApplyPerformanceContract as applyContract,
  type CartularyImportApplyPredicateId,
  type CartularyImportDiscoveryPredicateId,
  cartularyLargeGridFixture,
  cartularyImportDiscoveryPerformanceContract as discoveryContract,
} from "@cartulary/ui-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { ac043FixtureDigest, resetWorkbookClientTiming } from "./timingSupport";

export type ImportSourceFormat =
  (typeof discoveryContract.predicates)[number]["kind"];
export async function measureImportProgress(
  purpose: "discovery" | "apply",
  page: Page,
  input: Omit<
    JobProgressPaintInput,
    "startMark" | "admissionMark" | "dispatchMark" | "timeoutMs" | "jobPurpose"
  >,
  action: () => Promise<void>,
) {
  const contract = purpose === "discovery" ? discoveryContract : applyContract;
  await resetWorkbookClientTiming(page);
  const observation = page
    .evaluate(observeJobProgressPaintInBrowser, {
      ...input,
      jobPurpose: purpose,
      startMark: contract.start_mark,
      admissionMark: contract.admission_mark,
      dispatchMark: contract.dispatch_mark,
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

export async function attachImportProgressObservation(
  purpose: "discovery" | "apply",
  info: TestInfo,
  predicateId:
    | CartularyImportDiscoveryPredicateId
    | CartularyImportApplyPredicateId,
  target?: { format: ImportSourceFormat; recordId: string; rowVersion: number },
  completion?: ImportProgressCompletion,
  failureReason?: string,
) {
  const contract = purpose === "discovery" ? discoveryContract : applyContract;
  if (!contract.predicates.some((value) => value.predicate_id === predicateId))
    throw new Error("Import predicate does not match the operation");
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
            schema_id:
              purpose === "discovery"
                ? "cartulary.import_discovery_measurement_proof.v1"
                : "cartulary.import_apply_measurement_proof.v1",
            kind: purpose === "discovery" ? "import_discovery" : "import_apply",
            ...(purpose === "apply"
              ? {
                  selected_units: applyContract.workload.selected_units,
                  mapped_column_ordinal:
                    applyContract.workload.mapped_column_ordinal,
                  mapped_field_key: applyContract.workload.mapped_field_key,
                  unmapped_column_ordinal:
                    applyContract.workload.unmapped_column_ordinal,
                }
              : {}),
            source_format: target.format,
            job_id: completion.jobId,
            job_status: completion.jobStatus,
            record_id: target.recordId,
            row_version: target.rowVersion,
            view_schema_id: contract.view_schema_id,
            field_key: contract.workload.field_key,
            data_rows: contract.workload.data_rows,
            columns: contract.workload.columns,
            header_rows: 1,
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
