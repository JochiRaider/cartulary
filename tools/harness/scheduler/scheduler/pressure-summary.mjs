import { collectGoShardPlanFromRows } from "../../backend/backend-shard-plan.mjs";
import { collectTargetPlanRows } from "../../backend/backend-target-plan.mjs";

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

function fixtureClassForPolicy(policy) {
  switch (policy) {
    case "migration_scratch":
      return "migration_scratch";
    case "template_clone":
      return "template_clone";
    case "package_reset":
      return "package_reset";
    case "transaction":
    case "group_clone":
      return "transaction_or_shared_postgres";
    default:
      return "none";
  }
}

function uniqueSorted(values) {
  return Array.from(new Set(values.filter((value) => value !== ""))).sort((left, right) =>
    left.localeCompare(right),
  );
}

function phaseNamesFromRows(rows) {
  return uniqueSorted(rows.map((row) => row.manifest_phase ?? ""));
}

function shardIndex(root) {
  const rows = collectTargetPlanRows(root);
  const plans = [collectGoShardPlanFromRows(root, rows)];
  for (const phase of phaseNamesFromRows(rows)) {
    plans.push(collectGoShardPlanFromRows(root, rows, { phase }));
  }
  const index = new Map();
  for (const plan of plans) {
    for (const shard of plan.shards ?? []) {
      index.set(`${shard.target}\u001f${shard.name}`, shard);
    }
  }
  return index;
}

function incrementPressure(aggregates, key, durationMs) {
  const current = aggregates.get(key) ?? { work_unit_count: 0, duration_ms: 0 };
  current.work_unit_count += 1;
  current.duration_ms += durationMs;
  aggregates.set(key, current);
}

function apportionedDurations(record, items) {
  const totalWeight = items.reduce((sum, item) => sum + Math.max(0, Number(item.weight_ms) || 0), 0);
  if (items.length === 0) {
    return [];
  }
  if (totalWeight <= 0) {
    const equal = record.duration_ms / items.length;
    return items.map(() => equal);
  }
  return items.map(
    (item) => record.duration_ms * (Math.max(0, Number(item.weight_ms) || 0) / totalWeight),
  );
}

function pressureSort(left, right) {
  return (
    left.target.localeCompare(right.target) ||
    (left.row_id ?? "").localeCompare(right.row_id ?? "") ||
    left.execution_family.localeCompare(right.execution_family) ||
    left.fixture_class.localeCompare(right.fixture_class)
  );
}

function proofSort(left, right) {
  return (
    left.target.localeCompare(right.target) ||
    left.row_id.localeCompare(right.row_id) ||
    left.execution_family.localeCompare(right.execution_family) ||
    (left.symbol ?? "").localeCompare(right.symbol ?? "") ||
    left.fixture_policy.localeCompare(right.fixture_policy)
  );
}

function parsePressureKey(key, fields) {
  const values = key.split("\u001f");
  return Object.fromEntries(fields.map((field, index) => [field, values[index] ?? ""]));
}

function buildFixturePressureAggregates(reporter) {
  const completedGoShards = reporter.completedWork.filter(
    (record) => record.kind === "work_unit" && record.work_unit_type === "go_shard" && record.status === 0,
  );
  if (completedGoShards.length === 0) {
    return {
      row_fixture_pressure: [],
      execution_family_fixture_pressure: [],
      fixture_proof_records: [],
    };
  }
  const byShard = shardIndex(reporter.repoRoot);
  const rowAggregates = new Map();
  const familyAggregates = new Map();
  const proofRecords = new Map();
  for (const record of completedGoShards) {
    const target = record.aggregate_target || record.service_session_target || "";
    const shardName = record.shard || "";
    const shard = byShard.get(`${target}\u001f${shardName}`);
    if (!shard) {
      continue;
    }
    const items = shard.items ?? [];
    const durations = apportionedDurations(record, items);
    items.forEach((item, index) => {
      const fixturePolicy = item.postgres_fixture_policy || "none";
      const fixtureClass = fixtureClassForPolicy(fixturePolicy);
      const durationMs = durations[index] ?? 0;
      const rowID = String(item.id ?? "").split(":")[0];
      const executionFamily = item.aggregate_name || shard.aggregate_name || "";
      const rowKey = [item.target || target, rowID, executionFamily, fixtureClass].join("\u001f");
      const familyKey = [item.target || target, executionFamily, fixtureClass].join("\u001f");
      incrementPressure(rowAggregates, rowKey, durationMs);
      incrementPressure(familyAggregates, familyKey, durationMs);
      const proof = item.postgres_fixture_proof ?? {};
      if (!proof.proof_status) {
        return;
      }
      const proofRecord = {
        target: item.target || target,
        row_id: rowID,
        execution_family: executionFamily,
        ...(item.symbol ? { symbol: item.symbol } : {}),
        fixture_policy: fixturePolicy,
        proof_kind: proof.proof_kind ?? fixturePolicy,
        proof_status: proof.proof_status,
        ...(proof.proof_ref ? { proof_ref: proof.proof_ref } : {}),
        reason: proof.reason || item.postgres_fixture_reason || `Fixture policy ${fixturePolicy}`,
        dirty_tables: proof.dirty_tables ?? item.postgres_fixture_budget?.dirty_tables ?? [],
      };
      const proofKey = [
        proofRecord.target,
        proofRecord.row_id,
        proofRecord.execution_family,
        proofRecord.symbol ?? "",
        proofRecord.fixture_policy,
      ].join("\u001f");
      proofRecords.set(proofKey, proofRecord);
    });
  }
  return {
    row_fixture_pressure: Array.from(rowAggregates.entries())
      .map(([key, value]) => ({
        ...parsePressureKey(key, ["target", "row_id", "execution_family", "fixture_class"]),
        work_unit_count: value.work_unit_count,
        duration_ms: value.duration_ms,
      }))
      .sort(pressureSort),
    execution_family_fixture_pressure: Array.from(familyAggregates.entries())
      .map(([key, value]) => ({
        ...parsePressureKey(key, ["target", "execution_family", "fixture_class"]),
        work_unit_count: value.work_unit_count,
        duration_ms: value.duration_ms,
      }))
      .sort(pressureSort),
    fixture_proof_records: Array.from(proofRecords.values()).sort(proofSort),
  };
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
  const fixturePressure = buildFixturePressureAggregates(reporter);
  return {
    schema_id: "cartulary.scheduler_pressure_summary.v2",
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
    row_fixture_pressure: fixturePressure.row_fixture_pressure,
    execution_family_fixture_pressure: fixturePressure.execution_family_fixture_pressure,
    fixture_proof_records: fixturePressure.fixture_proof_records,
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
