export function slowestWork(completedWork) {
  return [...completedWork]
    .sort(
      (left, right) =>
        right.duration_ms - left.duration_ms ||
        left.label.localeCompare(right.label),
    )
    .slice(0, 5);
}

function incrementCount(counts, key, amount = 1) {
  if (!key) {
    return;
  }
  counts[key] = (counts[key] ?? 0) + amount;
}

function pressureFixtureClass(resourceClaims) {
  if (resourceClaims.migration_scratch_postgres) {
    return "migration_scratch";
  }
  if (resourceClaims.postgres_clone) {
    return "template_clone";
  }
  if (resourceClaims.postgres_reset) {
    return "package_reset";
  }
  if (resourceClaims.postgres) {
    return "transaction_or_shared_postgres";
  }
  return "none";
}

function pressureAccountingCounts(reporter) {
  return {
    executed: reporter.completedWork.filter((record) => record.kind === "work_unit").length,
    reused: 0,
    skipped: reporter.skippedWork.length,
  };
}

export function buildPressureSummary({ reporter, status, slowest, timing }) {
  const resourceClaimCounts = {};
  const fixtureClassCounts = {};
  const laneDurations = {};
  const targetCounts = {};
  for (const record of reporter.completedWork) {
    if (record.kind !== "work_unit") {
      continue;
    }
    const lane = record.aggregate_target || record.service_session_target || record.label;
    laneDurations[lane] = (laneDurations[lane] ?? 0) + record.duration_ms;
    incrementCount(targetCounts, lane);
    for (const [resource, amount] of Object.entries(record.resource_claims ?? {})) {
      if (amount) {
        incrementCount(resourceClaimCounts, resource, Number(amount));
      }
    }
    incrementCount(fixtureClassCounts, pressureFixtureClass(record.resource_claims ?? {}));
  }
  return {
    schema_id: "cartulary.scheduler_pressure_summary.v1",
    target: reporter.schedule.target,
    scheduler_kind: reporter.schedule.kind,
    status,
    total_work_units: reporter.schedule.totalWorkUnits,
    completed_work_units: reporter.completedCount,
    scheduler_total_duration_ms: timing.scheduler_total_duration_ms,
    target_counts: targetCounts,
    lane_duration_ms: laneDurations,
    resource_claim_counts: resourceClaimCounts,
    fixture_class_counts: fixtureClassCounts,
    slowest_work_units: slowest,
    reused_accounting_counts: pressureAccountingCounts(reporter),
    readiness_attribution_counts: {},
    generated_at: timing.scheduler_completed_at,
  };
}

export function finalizerTimings(completedWork) {
  return completedWork
    .filter((record) => record.kind === "finalizer")
    .map((record) => ({
      label: record.label,
      id: record.id,
      status: record.status,
      duration_ms: record.duration_ms,
      log_file: record.log_file,
    }));
}
