import { readdir, readFile } from "node:fs/promises";
import path from "node:path";

import {
  classifyExecutionFailure,
  normalizeFailureRecord,
  primaryPublicFailure,
} from "../../contract/index.mjs";
import { relToRepo as relToRepoPath } from "../scheduler-reporting.mjs";

const observedFailedWorkUnitLimit = 25;

function relToRepo(repoRoot, value) {
  return relToRepoPath(repoRoot, value);
}

export function observedFailedWorkUnits(completedWork) {
  return completedWork
    .filter((record) => record.status !== 0)
    .slice(0, observedFailedWorkUnitLimit)
    .map((record) => ({
      id: record.id,
      label: record.label,
      aggregate_target: record.aggregate_target,
      kind: record.kind,
      work_unit_type: record.work_unit_type,
      service_session_target: record.service_session_target,
      status: record.status,
      duration_ms: record.duration_ms,
      started_monotonic_ms: record.started_monotonic_ms,
      finished_monotonic_ms: record.finished_monotonic_ms,
      needs: [...(record.needs ?? [])],
      completion_keys: [...(record.completion_keys ?? [])],
      resource_claims: { ...(record.resource_claims ?? {}) },
      log_file: record.log_file,
    }));
}

function schedulerWorkUnitFailureTokens(failed, failedDetail) {
  const tokens = new Set();
  const values = [
    failed,
    failedDetail?.id,
    failedDetail?.label,
    ...(Array.isArray(failedDetail?.completion_keys)
      ? failedDetail.completion_keys
      : []),
  ];
  for (const value of values) {
    if (typeof value !== "string") {
      continue;
    }
    for (const match of value
      .toLowerCase()
      .matchAll(/browser-functional-shard-\d+/g)) {
      tokens.add(match[0]);
    }
    for (const match of value.toLowerCase().matchAll(/functional-shard-\d+/g)) {
      tokens.add(match[0]);
    }
  }
  return [...tokens];
}

function schedulerFailureMatchesWorkUnit(failure, tokens) {
  const haystack = [
    failure?.label,
    failure?.artifact,
    failure?.work_unit,
    failure?.message,
  ]
    .map((value) => (typeof value === "string" ? value.toLowerCase() : ""))
    .join(" ");
  return tokens.some((token) => haystack.includes(token));
}

async function readSchedulerChildFailureRecord({
  failed,
  failedDetail,
  repoRoot,
  scheduleTarget,
  schedulerTargetDir,
}) {
  const childTarget =
    [
      failedDetail?.aggregate_target,
      failedDetail?.target,
      failedDetail?.label,
      failedDetail?.id,
      failed,
    ]
      .map((value) => (typeof value === "string" ? value.trim() : ""))
      .find((value) => value && value !== scheduleTarget) ?? "";
  if (!childTarget || childTarget === scheduleTarget) {
    return null;
  }
  const summaryFile = path.join(
    schedulerTargetDir(repoRoot, childTarget),
    "target-summary.json",
  );
  let summary;
  try {
    summary = JSON.parse(await readFile(summaryFile, "utf8"));
  } catch {
    return null;
  }
  const failureSource = summary?.totals ?? summary;
  const failures = Array.isArray(failureSource?.failures)
    ? failureSource.failures
    : Array.isArray(summary?.failures)
      ? summary.failures
      : [];
  const workUnitTokens = schedulerWorkUnitFailureTokens(failed, failedDetail);
  const matchingFailures =
    workUnitTokens.length === 0
      ? failures
      : failures.filter((failure) =>
          schedulerFailureMatchesWorkUnit(failure, workUnitTokens),
        );
  const propagated = primaryPublicFailure(
    matchingFailures.length > 0 ? matchingFailures : failures,
    {
      failure_class: failureSource?.failure_class ?? summary?.failure_class,
      failure_reason: failureSource?.failure_reason ?? summary?.failure_reason,
      target: childTarget,
      label: failed ?? childTarget,
    },
  );
  if (!propagated) {
    return null;
  }
  return {
    ...propagated,
    kind: propagated.kind || "child_target",
    source: "scheduler",
    target: scheduleTarget,
    child_target: childTarget,
    work_unit: failedDetail?.id ?? failed ?? childTarget,
    label: failed ?? propagated.label ?? childTarget,
    message: propagated.message || `scheduler child target failed: ${childTarget}`,
    artifact: propagated.artifact || relToRepo(repoRoot, summaryFile),
  };
}

async function readServiceSessionFailureRecord({
  failed,
  failedDetail,
  repoRoot,
  scheduleTarget,
  schedulerTargetDir,
}) {
  if (failedDetail?.work_unit_type !== "service_session") {
    return null;
  }
  const serviceTarget =
    [
      failedDetail?.service_session_target,
      failedDetail?.aggregate_target,
      failedDetail?.target,
    ]
      .map((value) => (typeof value === "string" ? value.trim() : ""))
      .find((value) => value !== "") ?? "";
  if (!serviceTarget) {
    return null;
  }
  const runRoot = path.dirname(schedulerTargetDir(repoRoot, scheduleTarget));
  const servicesRoot = path.join(runRoot, "_shared", "test-services");
  let entries;
  try {
    entries = await readdir(servicesRoot, { withFileTypes: true });
  } catch {
    return null;
  }

  for (const entry of entries
    .filter((item) => item.isDirectory())
    .sort((left, right) => left.name.localeCompare(right.name))) {
    const scopeFile = path.join(servicesRoot, entry.name, "service-scope.json");
    let scope;
    try {
      scope = JSON.parse(await readFile(scopeFile, "utf8"));
    } catch {
      continue;
    }
    if (scope?.schema_id !== "cartulary.test_services.scope.v1") {
      continue;
    }
    if (scope?.target !== serviceTarget) {
      continue;
    }
    const failure =
      scope?.failure ??
      (scope?.preflight?.status === "fail" ? scope.preflight : null);
    if (!failure) {
      continue;
    }
    return normalizeFailureRecord({
      failure_class: failure.failure_class ?? "unknown",
      failure_reason: failure.failure_reason ?? "unknown_failure",
      kind: "service_session",
      source: "test_services",
      target: scheduleTarget,
      child_target: serviceTarget,
      work_unit: failedDetail?.id ?? failed ?? serviceTarget,
      label: failed ?? failedDetail?.label ?? serviceTarget,
      message:
        failure.message ||
        `service session startup failed before child target summary: ${serviceTarget}`,
      artifact: relToRepo(repoRoot, scopeFile),
    });
  }
  return null;
}

function schedulerFallbackFailureRecord(record, scheduleTarget) {
  const label = record?.label ?? scheduleTarget;
  const normalized = record?.status === 10
    ? { failure_class: "product", failure_reason: "test_assertion_failure" }
    : record?.status === 11
      ? { failure_class: "artifact", failure_reason: "artifact_error" }
    : record?.status === 12 && record?.kind === "finalizer"
      ? { failure_class: "harness", failure_reason: "cleanup_error" }
      : record?.status === 13
        ? { failure_class: "timing", failure_reason: "timeout_failure" }
        : record?.status === 130 || record?.status === 143 || record?.termination_reason === "cancelled_or_interrupted"
          ? { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" }
          : { failure_class: classifyExecutionFailure(label, scheduleTarget), failure_reason: "unknown_failure" };
  return normalizeFailureRecord({
    ...normalized,
    kind: "scheduler",
    source: "scheduler",
    target: scheduleTarget,
    child_target: record?.aggregate_target ?? null,
    work_unit: record?.id ?? label,
    label,
    message: `scheduler work unit failed: ${label}`,
    artifact: record?.log_file ?? "",
  });
}

export function schedulerTargetFailureRecord({ failed, scheduleTarget }) {
  return normalizeFailureRecord({
    failure_class: classifyExecutionFailure(failed ?? scheduleTarget, scheduleTarget),
    kind: "scheduler",
    source: "scheduler",
    target: scheduleTarget,
    label: failed ?? scheduleTarget,
    message: failed
      ? `scheduler work unit failed: ${failed}`
      : `scheduler target failed: ${scheduleTarget}`,
  });
}

export async function schedulerFailureRecordsForCompletedWork({
  completedWork,
  repoRoot,
  scheduleTarget,
  schedulerTargetDir,
}) {
  const failedRecords = completedWork.filter((record) => record.status !== 0);
  const failures = [];
  for (const record of failedRecords) {
    const propagated =
      (await readServiceSessionFailureRecord({
        failed: record.label,
        failedDetail: record,
        repoRoot,
        scheduleTarget,
        schedulerTargetDir,
      })) ??
      (await readSchedulerChildFailureRecord({
        failed: record.label,
        failedDetail: record,
        repoRoot,
        scheduleTarget,
        schedulerTargetDir,
      }));
    failures.push(propagated ?? schedulerFallbackFailureRecord(record, scheduleTarget));
  }
  return failures;
}
