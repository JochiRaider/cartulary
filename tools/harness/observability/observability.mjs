import {
  chmodSync,
  existsSync,
  readFileSync,
  readdirSync,
  statSync,
} from "node:fs";
import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import os from "node:os";
import path from "node:path";

import {
  prettyJSONString,
  repoRoot,
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";
import { buildSourceSnapshot } from "../owner-slice/source-snapshot.mjs";
import { collectServiceTimingContamination } from "../duration-accounting/duration-drift.mjs";

export const harnessScope = "cartulary.harness.execution";
export const observabilityDirName = "harness-observability";

const sourceNames = new Set([
  "harness-invocation-start.json",
  "run-summary.json",
  "scheduler-events.jsonl",
  "sequence-events.jsonl",
  "step-summary.json",
  "target-summary.json",
  "target-timing.json",
  "timing-span.json",
  "tool-run-summary.json",
]);
const diagnosticDir = path.join("_shared", observabilityDirName);

function observabilityPolicy() {
  const manifest = readJSON(path.join(repoRoot, "tools", "task_surface_manifest.json"));
  const policy = manifest.observability_policy;
  return {
    required_targets: [...policy.required_targets].sort((left, right) => left.localeCompare(right)),
    exclusions: [...policy.excluded_targets].sort((left, right) => left.target.localeCompare(right.target)).map((entry) => ({ ...entry })),
    out_of_scope: [...policy.out_of_scope_targets].sort((left, right) => left.target.localeCompare(right.target)).map((entry) => ({ ...entry })),
    measurement_profile_sha256: sha256(canonicalJSON({
      profiles: policy.measurement_profiles,
      bindings: policy.target_measurement_profiles,
    })),
  };
}

export function observabilityRequiredTarget(target) {
  return observabilityPolicy().required_targets.includes(target);
}

export function executionProfileDigests() {
  const catalog = loadTestCatalog(repoRoot);
  const host = {
    architecture: process.arch,
    available_parallelism: os.availableParallelism(),
    logical_cpus: os.cpus().length,
    platform: process.platform,
  };
  return {
    host: sha256(JSON.stringify(host)),
    capacity: sha256(readFileSync(path.join(repoRoot, "tools", "scheduler_resource_registry.json"))),
    workload: sha256(`${catalog.semantic_digest}\u0000${catalog.verification.semantic_digest}`),
    toolchain: sha256(readFileSync(path.join(repoRoot, "tools", "toolchain_pins.json"))),
  };
}

function sha256(value) {
  return createHash("sha256").update(value).digest("hex");
}

function canonicalJSON(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function safeToken(value, fallback = "unknown") {
  const token = String(value ?? "")
    .trim()
    .toLowerCase()
    .replaceAll(/[^a-z0-9_.:-]+/gu, "-")
    .replaceAll(/^-+|-+$/gu, "")
    .slice(0, 128);
  return token || fallback;
}

function safeDiagnostic(error) {
  const name = error instanceof Error ? error.name : "Error";
  return `${safeToken(name, "error")}:diagnostic-generation-failed`.slice(0, 512);
}

function runRelative(runDir, file) {
  const relative = path.relative(runDir, file).replaceAll("\\", "/");
  if (relative === ".." || relative.startsWith("../") || path.isAbsolute(relative)) {
    throw new Error("observability source escaped retained run root");
  }
  return relative;
}

function artifactPath(runDir, relative, label = "artifact") {
  if (
    typeof relative !== "string" ||
    path.isAbsolute(relative) ||
    relative.includes("\\") ||
    path.posix.normalize(relative) !== relative ||
    relative === ".." ||
    relative.startsWith("../")
  ) {
    throw new Error(`${label} has an unsafe retained path`);
  }
  const resolved = path.resolve(runDir, relative);
  if (resolved !== runDir && !resolved.startsWith(`${runDir}${path.sep}`)) {
    throw new Error(`${label} escaped retained run root`);
  }
  return resolved;
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function timestampMs(value, label) {
  const result = Date.parse(String(value ?? ""));
  if (!Number.isFinite(result) || result < 0) {
    throw new Error(`${label} has an invalid clock`);
  }
  return result;
}

function toUnixNano(ms) {
  return (BigInt(Math.round(ms)) * 1_000_000n).toString();
}

function fromUnixNano(value) {
  return Number(BigInt(value) / 1_000_000n);
}

function intervalUnion(intervals) {
  const sorted = intervals
    .filter(([start, end]) => Number.isFinite(start) && Number.isFinite(end) && end > start)
    .sort((left, right) => left[0] - right[0] || left[1] - right[1]);
  let total = 0;
  let current = null;
  for (const interval of sorted) {
    if (!current || interval[0] > current[1]) {
      if (current) total += current[1] - current[0];
      current = [...interval];
    } else {
      current[1] = Math.max(current[1], interval[1]);
    }
  }
  if (current) total += current[1] - current[0];
  return total;
}

function statusFor(value) {
  return value === "pass" || value === 0 ? "OK" : value === "pending" ? "UNSET" : "ERROR";
}

function attributes(values) {
  return Object.entries(values)
    .filter(([, value]) => value !== undefined && value !== null && String(value) !== "")
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([key, value]) => ({ key, value: safeToken(value) }));
}

function sourceFiles(runDir) {
  const files = [];
  function visit(dir) {
    for (const entry of readdirSync(dir, { withFileTypes: true }).sort((left, right) => left.name.localeCompare(right.name))) {
      if (entry.name === observabilityDirName || entry.name === "node_modules" || entry.name === ".git") {
        continue;
      }
      const file = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        visit(file);
      } else if (
        entry.isFile() &&
        (sourceNames.has(entry.name) ||
          (path.basename(dir) === "timing-spans" && entry.name.endsWith(".json")))
      ) {
        files.push(file);
      }
    }
  }
  visit(runDir);
  return files.sort((left, right) => runRelative(runDir, left).localeCompare(runRelative(runDir, right)));
}

function parseJSONLines(file, runDir) {
  return readFileSync(file, "utf8")
    .split(/\r?\n/u)
    .filter(Boolean)
    .map((line, index) => {
      try {
        return JSON.parse(line);
      } catch {
        throw new Error(`${runRelative(runDir, file)}:${index + 1} is malformed JSONL`);
      }
    });
}

function hasRunArtifacts(dir) {
  if (existsSync(path.join(dir, "run-summary.json"))) return true;
  return readdirSync(dir, { withFileTypes: true }).some(
    (entry) => entry.isDirectory() && existsSync(path.join(dir, entry.name, "tool-run-summary.json")),
  );
}

export function resolveRunDir(resultsDir, runID = "", { allowNewest = true } = {}) {
  const selected = path.resolve(repoRoot, resultsDir);
  if (!existsSync(selected)) throw new Error("RESULTS_DIR does not exist");
  if (runID) {
    if (path.basename(selected) === runID && hasRunArtifacts(selected)) {
      return selected;
    }
    const candidate = path.join(selected, runID);
    if (!existsSync(candidate) || !hasRunArtifacts(candidate)) throw new Error("RUN_ID does not identify retained evidence");
    return candidate;
  }
  if (existsSync(path.join(selected, "run-summary.json")) || hasRunArtifacts(selected)) {
    return selected;
  }
  if (existsSync(path.join(selected, "tool-run-summary.json"))) {
    const runDir = path.dirname(selected);
    return runDir;
  }
  if (!allowNewest) {
    throw new Error("RUN_ID is required when RESULTS_DIR names a result root");
  }
  const candidates = readdirSync(selected, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(selected, entry.name))
    .filter(hasRunArtifacts)
    .sort((left, right) => statSync(right).mtimeMs - statSync(left).mtimeMs || right.localeCompare(left));
  if (candidates.length === 0) throw new Error("RESULTS_DIR contains no retained harness run");
  return candidates[0];
}

export function resolveExactRunDir(resultsDir, runID = "") {
  const selected = path.resolve(repoRoot, resultsDir);
  if (!existsSync(selected)) throw new Error("RESULTS_DIR does not exist");
  if (runID) return resolveRunDir(selected, runID, { allowNewest: false });
  if (
    existsSync(observabilityIndexPath(selected)) ||
    existsSync(path.join(selected, "run-summary.json")) ||
    existsSync(path.join(selected, "tool-run-summary.json"))
  ) {
    return selected;
  }
  throw new Error("RUN_ID is required when RESULTS_DIR names a result root");
}

function targetSummaries(runDir) {
  const summaries = [];
  for (const entry of readdirSync(runDir, { withFileTypes: true }).sort((left, right) => left.name.localeCompare(right.name))) {
    if (!entry.isDirectory() || entry.name === "_shared") continue;
    const file = path.join(runDir, entry.name, "target-summary.json");
    if (!existsSync(file)) continue;
    const summary = readJSON(file);
    summaries.push({ file, summary });
  }
  return summaries;
}

function toolInvocationRoot(file) {
  const tool = readJSON(file);
  return {
    target: tool.target,
    file,
    type: "tool",
    summary: {
      ...tool,
      label: tool.target,
      start_time: tool.started_at,
      end_time: tool.completed_at,
    },
  };
}

function invocationRoots(runDir, expectedTarget = "") {
  const runSummaryFile = path.join(runDir, "run-summary.json");
  if (existsSync(runSummaryFile)) {
    const summary = readJSON(runSummaryFile);
    return [{ target: summary.label, summary, file: runSummaryFile, type: "run" }];
  }
  if (expectedTarget) {
    const expectedTargetSummary = path.join(runDir, expectedTarget, "target-summary.json");
    if (existsSync(expectedTargetSummary)) {
      const summary = readJSON(expectedTargetSummary);
      if (summary.target !== expectedTarget) {
        throw new Error("retained top-level target summary identity mismatch");
      }
      return [{ target: summary.target, summary, file: expectedTargetSummary, type: "target" }];
    }
    const expectedToolSummary = path.join(runDir, expectedTarget, "tool-run-summary.json");
    if (existsSync(expectedToolSummary)) {
      const root = toolInvocationRoot(expectedToolSummary);
      if (root.target !== expectedTarget) {
        throw new Error("retained top-level tool summary identity mismatch");
      }
      return [root];
    }
  }
  const summaries = targetSummaries(runDir);
  const referenced = new Set();
  for (const { summary } of summaries) {
    for (const target of summary.children?.expected ?? []) referenced.add(target);
  }
  const roots = summaries.filter(({ summary }) => !referenced.has(summary.target));
  if (roots.length > 0) {
    return roots.map(({ file, summary }) => ({ target: summary.target, summary, file, type: "target" }));
  }
  if (summaries.length > 0) {
    return summaries.map(({ file, summary }) => ({ target: summary.target, summary, file, type: "target" }));
  }
  const tools = readdirSync(runDir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory() && entry.name !== "_shared")
    .map((entry) => path.join(runDir, entry.name, "tool-run-summary.json"))
    .filter(existsSync)
    .map((file) => ({ file, tool: readJSON(file) }));
  if (tools.length !== 1) return [];
  return [toolInvocationRoot(tools[0].file)];
}

function stripDigestPrefix(value) {
  return String(value ?? "").replace(/^sha256:/u, "");
}

function currentCommit() {
  return execFileSync("git", ["rev-parse", "HEAD"], {
    cwd: repoRoot,
    encoding: "utf8",
  }).trim();
}

function currentSourceState() {
  const status = execFileSync(
    "git",
    ["status", "--porcelain=v1", "--untracked-files=all"],
    { cwd: repoRoot, encoding: "utf8", maxBuffer: 16 * 1024 * 1024 },
  );
  return status.trim() === "" ? "clean" : "dirty";
}

function releaseBrowserPolicyProjection(executionTopology) {
  const schedule = executionTopology.service_backed_schedules?.schedules
    ?.find((entry) => entry.target === "release-browser-readiness");
  const stageDefaults = executionTopology.service_backed_schedules?.defaults
    ?.browser_stage_resource_limits ?? {};
  const stages = (executionTopology.browser_e2e_batch?.stages ?? [])
    .filter((stage) => (stage.schedule_tags ?? []).includes("service_backed_release_browser"));
  return {
    browser_stack_capacity: schedule?.resource_limits?.browser_stack,
    stage_capacities: {
      visual: stageDefaults.visual ?? 1,
      accessibility: stageDefaults.a11y ?? 1,
    },
    sessions: stages.flatMap((stage) => (stage.groups ?? []).map((group) => ({
      browser_session_group: group.browser_session_group,
      browser_stage: stage.name,
      runtime_profile_id: group.runtime_profile_id ?? "default",
      browser_session_isolation_reason: group.browser_session_isolation_reason,
    }))).sort((left, right) =>
      left.browser_session_group.localeCompare(right.browser_session_group)),
  };
}

function measurementContract(target, catalog, manifest, executionTopology) {
  const targetEntry = manifest.targets.find((entry) => entry.name === target);
  const binding = manifest.observability_policy.target_measurement_profiles
    .find((entry) => entry.target === target);
  const profile = manifest.observability_policy.measurement_profiles
    .find((entry) => entry.profile_id === binding?.profile_id);
  if (!targetEntry || !binding || !profile) {
    throw new Error(`required observability target ${target} has no retained measurement contract`);
  }
  const canonicalInputs = profile.canonical_inputs ?? {};
  const targetExecutionPolicy = {
    target,
    sequence: manifest.sequences?.[target] ?? null,
    service_backed_schedule:
      executionTopology.service_backed_schedules?.schedules?.find((entry) => entry.target === target) ?? null,
    ...(target === "release-browser-readiness" ? {
      release_browser: releaseBrowserPolicyProjection(executionTopology),
    } : {}),
    ...(target === "backend-unit" ? {
      backend_unit: {
        capture_grouping: {
          dimensions: [
            "package_selection",
            "runtime_binaries",
            "runtime_profile",
            "resource_profile",
            "fixture_profile",
            "fixture_policy",
            "fixture_budget",
            "isolation_policy",
            "evidence_class",
          ],
          raw_selectors: "isolated",
        },
        worker_pool: {
          formula: "min(group_count,clamp(floor(available_parallelism/4),1,8))",
          child_gomaxprocs: "max(1,floor(available_parallelism/workers))",
        },
        report_projection: {
          physical_report_parse: "once_per_physical_report",
          emission: "parallel_host_derived_pool",
        },
      },
    } : {}),
  };
  return {
    target,
    command_id: targetEntry.command_id,
    measurement_profile_id: binding.profile_id,
    canonical_inputs: canonicalInputs,
    observation_eligibility: profile.observation_eligibility,
    performance_gates: [...profile.performance_gates],
    timing_source: "invocation_or_exact_aggregate",
    ...(profile.allowed_policy_transition === undefined
      ? {}
      : { allowed_policy_transition: profile.allowed_policy_transition }),
    workload_evidence_profile_sha256: sha256(
      `${catalog.semantic_digest}\u0000${catalog.verification.semantic_digest}\u0000${targetEntry.command_id}\u0000${canonicalJSON(canonicalInputs)}`,
    ),
    execution_policy: targetExecutionPolicy,
    execution_policy_sha256: sha256(canonicalJSON(targetExecutionPolicy)),
  };
}

function invocationCanonicalInputs(contract) {
  return Object.fromEntries(Object.entries(contract.canonical_inputs).map(([name, expected]) => {
    const supplied = String(process.env[name] ?? "").trim();
    if (expected === "omitted") return [name, supplied === "" ? "omitted" : supplied];
    if (name === "EVIDENCE_ROOTS_FILE" && expected === "retained-module.auth-slice-evidence") {
      return [name, supplied === "" ? "omitted" : expected];
    }
    return [name, supplied === "" ? "omitted" : supplied];
  }));
}

function retainedCapacity() {
  return {
    available_parallelism: Math.max(1, os.availableParallelism()),
    logical_resources: {},
  };
}

function currentExecutionPolicy() {
  const inputs = [
    "tools/task_surface_manifest.json",
    "tools/scheduler_manifest.json",
    "tools/browser_e2e_batch_manifest.json",
    "tools/scheduler_resource_registry.json",
  ];
  return {
    inputs: inputs.map((file) => ({
      file,
      sha256: sha256(readFileSync(path.join(repoRoot, file))),
    })),
  };
}

function contaminationFor(runDir, { status, interrupted, retryCount, sourceState }) {
  const reasons = new Set();
  if (sourceState === "dirty") reasons.add("dirty_source");
  if (status !== "passed") reasons.add("failed_execution");
  if (interrupted) reasons.add("interrupted_execution");
  if (retryCount > 0) reasons.add("retry_observed");
  const serviceContamination = collectServiceTimingContamination(repoRoot, runDir);
  if (serviceContamination.contaminated) {
    if (serviceContamination.reasons.some((reason) => reason.includes("retry"))) {
      reasons.add("retry_observed");
    } else {
      reasons.add("external_activity");
    }
  }
  return [...reasons].sort((left, right) => left.localeCompare(right));
}

function contextPath(runDir) {
  return path.join(runDir, diagnosticDir, "execution-context.json");
}

function invocationStartPath(runDir) {
  return path.join(runDir, "_shared", "harness-invocation-start.json");
}

function loadInvocationStart(runDir, expectedTarget) {
  const file = invocationStartPath(runDir);
  if (!existsSync(file)) return null;
  const retained = readJSON(file);
  validateSchemaSync("cartulary.harness_invocation_start.v1", retained);
  if (retained.run_id !== path.basename(runDir) || retained.target !== expectedTarget) {
    throw new Error("retained harness invocation start identity mismatch");
  }
  return retained;
}

export function captureExecutionContext(runDir, metadata = {}) {
  const resolvedRunDir = path.resolve(runDir);
  const file = contextPath(resolvedRunDir);
  if (existsSync(file)) {
    const retained = readJSON(file);
    validateSchemaSync(retained.schema_id, retained);
    return retained;
  }
  const roots = invocationRoots(resolvedRunDir, metadata.target ?? "");
  if (roots.length !== 1) {
    throw new Error(`retained run must identify exactly one top-level invocation; found ${roots.length}`);
  }
  const root = roots[0];
  const target = metadata.target ?? root.target;
  const rootInterval = intervalForSummary(root.summary, target);
  const invocationStart = loadInvocationStart(resolvedRunDir, target);
  const invocationStartMs = invocationStart
    ? timestampMs(invocationStart.started_at, "harness invocation start")
    : rootInterval.start;
  if (invocationStartMs > rootInterval.start + 5) {
    throw new Error("retained harness invocation start follows the top-level result");
  }
  const interval = { start: invocationStartMs, end: rootInterval.end };
  const catalog = loadTestCatalog(repoRoot);
  const manifest = readJSON(path.join(repoRoot, "tools", "task_surface_manifest.json"));
  const executionTopology = readJSON(path.join(repoRoot, "tools", "execution_topology_manifest.json"));
  const measurementTargets = new Set(
    manifest.observability_policy.target_measurement_profiles.map((binding) => binding.target),
  );
  const measurementContracts = [...measurementTargets]
    .sort((left, right) => left.localeCompare(right))
    .map((observedTarget) => measurementContract(observedTarget, catalog, manifest, executionTopology));
  const workloadContracts = measurementContracts.map((contract) => ({
    target: contract.target,
    command_id: contract.command_id,
    measurement_profile_id: contract.measurement_profile_id,
    canonical_inputs: contract.canonical_inputs,
    workload_evidence_profile_sha256: contract.workload_evidence_profile_sha256,
  }));
  const rootContract = measurementContracts.find((contract) => contract.target === target);
  if (!rootContract) {
    throw new Error(`top-level target ${target} has no retained measurement contract`);
  }
  const capacity = retainedCapacity();
  const invocationInputs = invocationCanonicalInputs(rootContract);
  const hostProfile = {
    architecture: process.arch,
    platform: process.platform,
    release: os.release(),
  };
  const sourceState = currentSourceState();
  const status = metadata.status === "interrupted" || root.summary.status === "interrupted"
    ? "interrupted"
    : root.summary.status === "pass"
      ? "passed"
      : "failed";
  const interrupted = status === "interrupted";
  const retryCount = Number.isInteger(metadata.retryCount) ? metadata.retryCount : 0;
  const contaminationReasons = contaminationFor(resolvedRunDir, {
    status,
    interrupted,
    retryCount,
    sourceState,
  });
  const observedRetryCount = retryCount > 0 || contaminationReasons.includes("retry_observed")
    ? Math.max(1, retryCount)
    : 0;
  const executionPolicy = currentExecutionPolicy();
  const context = {
    schema_id: "cartulary.harness_execution_context.v2",
    run_id: path.basename(resolvedRunDir),
    invocation_id: safeToken(target),
    target,
    command_id: rootContract.command_id,
    canonical_inputs: invocationInputs,
    invocation_boundary_retained: invocationStart !== null,
    invocation_edges: invocationStart?.invocation_edges ?? [],
    measurement_contracts: measurementContracts,
    measurement_contracts_sha256: sha256(canonicalJSON(measurementContracts)),
    workload_contracts_sha256: sha256(canonicalJSON(workloadContracts)),
    commit: currentCommit(),
    source_snapshot_sha256: stripDigestPrefix(buildSourceSnapshot(repoRoot).digest),
    source_state: sourceState,
    host_profile_sha256: sha256(canonicalJSON(hostProfile)),
    toolchain_profile_sha256: sha256(readFileSync(path.join(repoRoot, "tools", "toolchain_pins.json"))),
    available_capacity: capacity,
    capacity_profile_sha256: sha256(canonicalJSON(capacity)),
    workload_evidence_profile_sha256: sha256(
      `${catalog.semantic_digest}\u0000${catalog.verification.semantic_digest}\u0000${rootContract.command_id}\u0000${canonicalJSON(invocationInputs)}`,
    ),
    execution_policy: executionPolicy,
    execution_policy_sha256: sha256(canonicalJSON(executionPolicy)),
    started_at: new Date(interval.start).toISOString(),
    ended_at: new Date(interval.end).toISOString(),
    status,
    interrupted,
    retry_count: observedRetryCount,
    warm_eligibility: contaminationReasons.length === 0 ? "eligible" : "ineligible",
    contamination_reasons: contaminationReasons,
  };
  validateSchemaSync(context.schema_id, context);
  const base = path.dirname(file);
  secureMkdir(base, { allowedRoot: resolvedRunDir });
  writeJSON(file, context, resolvedRunDir);
  return context;
}

export function loadRetainedExecutionContext(runDir) {
  const resolvedRunDir = path.resolve(runDir);
  const file = contextPath(resolvedRunDir);
  if (!existsSync(file)) throw new Error("retained run has no harness execution context");
  const context = readJSON(file);
  if (!new Set([
    "cartulary.harness_execution_context.v1",
    "cartulary.harness_execution_context.v2",
  ]).has(context.schema_id)) {
    throw new Error(`unsupported retained execution context ${context.schema_id}`);
  }
  validateSchemaSync(context.schema_id, context);
  if (context.run_id !== path.basename(resolvedRunDir)) {
    throw new Error("retained execution context run identity mismatch");
  }
  return { context, file, sha256: sha256(readFileSync(file)) };
}

function intervalForSummary(summary, label) {
  const start = timestampMs(summary.start_time, `${label}.start_time`);
  const end = timestampMs(summary.end_time, `${label}.end_time`);
  if (end < start) throw new Error(`${label} has a negative clock interval`);
  return { start, end };
}

function spanID(traceID, occurrence) {
  return sha256(`${traceID}\u0000${occurrence}`).slice(0, 16);
}

function makeSpan({ traceID, occurrence, parentSpanID, links = [], name, phase, status, start, end, attrs = {}, kind = "INTERNAL" }) {
  if (!Number.isFinite(start) || !Number.isFinite(end) || end < start) {
    throw new Error(`${occurrence} has an invalid span interval`);
  }
  return {
    span_id: spanID(traceID, occurrence),
    parent_span_id: parentSpanID,
    links: [...new Set(links)].sort((left, right) => left.localeCompare(right)),
    name: safeToken(name),
    kind,
    phase,
    status: statusFor(status),
    start_time_unix_nano: toUnixNano(start),
    end_time_unix_nano: toUnixNano(end),
    attributes: attributes({
      "harness.phase": phase,
      "harness.status": status,
      ...attrs,
    }),
    source_occurrence: occurrence,
  };
}

function summaryStatus(summary) {
  return summary?.status === "pass" ? "pass" : "fail";
}

function targetSpans(runDir, root, traceID, rootSpan, retainedTargetIDs, invocationEdges, invocationBoundaryRetained) {
  const result = [];
  const allSummaries = targetSummaries(runDir);
  const spanByTarget = new Map([[root.target, rootSpan]]);
  const itemsByTarget = new Map();
  for (const item of allSummaries) {
    if (itemsByTarget.has(item.summary.target)) {
      throw new Error(`duplicate retained target summary ${item.summary.target}`);
    }
    itemsByTarget.set(item.summary.target, item);
  }
  const childrenByTarget = new Map();
  for (const item of allSummaries) {
    const children = (item.summary.children?.expected ?? [])
      .filter((child) => itemsByTarget.has(child));
    childrenByTarget.set(item.summary.target, children);
  }
  for (const edge of invocationEdges) {
    if (edge.parent_target !== root.target && !itemsByTarget.has(edge.parent_target)) continue;
    if (!itemsByTarget.has(edge.child_target)) continue;
    childrenByTarget.set(edge.parent_target, [...new Set([
      ...(childrenByTarget.get(edge.parent_target) ?? []),
      edge.child_target,
    ])].sort((left, right) => left.localeCompare(right)));
  }
  const reaches = (from, to, visiting = new Set()) => {
    if (from === to) return true;
    if (visiting.has(from)) return false;
    visiting.add(from);
    for (const child of childrenByTarget.get(from) ?? []) {
      if (reaches(child, to, visiting)) return true;
    }
    return false;
  };
  const summaries = allSummaries.filter((item) =>
    retainedTargetIDs.has(item.summary.target) &&
    (
      item.summary.target === root.target ||
      !invocationBoundaryRetained ||
      reaches(root.target, item.summary.target)
    ));
  const retainedTargets = summaries.map((item) => item.summary.target);
  const parentByTarget = new Map();
  for (const child of retainedTargets.filter((target) => target !== root.target)) {
    const candidates = retainedTargets
      .filter((candidate) => candidate !== child && reaches(candidate, child))
      .sort((left, right) => left.localeCompare(right));
    const nearest = candidates.filter((candidate) =>
      !candidates.some((other) => other !== candidate && reaches(candidate, other)));
    if (nearest.length > 1) {
      throw new Error(`retained target ${child} has ambiguous explicit summary parents`);
    }
    parentByTarget.set(child, nearest[0] ?? root.target);
  }
  const pending = summaries
    .filter((item) => item.summary.target !== root.target)
    .sort((left, right) => left.summary.target.localeCompare(right.summary.target));
  while (pending.length > 0) {
    let progressed = false;
    for (let index = 0; index < pending.length; ) {
      const item = pending[index];
      const parentTarget = parentByTarget.get(item.summary.target) ?? root.target;
      const parent = spanByTarget.get(parentTarget);
      if (!parent) {
        index += 1;
        continue;
      }
      const interval = intervalForSummary(item.summary, item.summary.target);
      const parentStart = fromUnixNano(parent.start_time_unix_nano);
      const parentEnd = fromUnixNano(parent.end_time_unix_nano);
      if (interval.start < parentStart - 5 || interval.end > parentEnd + 5) {
        throw new Error(`retained target ${item.summary.target} escapes explicit parent ${parentTarget}`);
      }
      const occurrence = `${runRelative(runDir, item.file)}:target`;
      const span = makeSpan({
        traceID,
        occurrence,
        parentSpanID: parent.span_id,
        name: item.summary.target,
        phase: "target",
        status: summaryStatus(item.summary),
        start: interval.start,
        end: interval.end,
        attrs: { "harness.target": item.summary.target },
      });
      spanByTarget.set(item.summary.target, span);
      result.push(span);
      pending.splice(index, 1);
      progressed = true;
    }
    if (!progressed) {
      throw new Error(`retained target summary relationships contain a cycle or missing parent: ${pending.map((item) => item.summary.target).join(",")}`);
    }
  }
  return { spans: result, spanByTarget };
}

function stepSpans(runDir, root, traceID, rootSpan, spanByTarget) {
  const result = [];
  for (const file of sourceFiles(runDir).filter((candidate) => path.basename(candidate) === "step-summary.json")) {
    const summary = readJSON(file);
    const interval = intervalForSummary(summary, runRelative(runDir, file));
    const parent = spanByTarget.get(summary.target) ?? rootSpan;
    result.push(makeSpan({
      traceID,
      occurrence: `${runRelative(runDir, file)}:runner`,
      parentSpanID: parent.span_id,
      name: summary.label ?? summary.step ?? "runner",
      phase: "runner",
      status: summaryStatus(summary),
      start: interval.start,
      end: interval.end,
      attrs: {
        "harness.target": summary.target ?? root.target,
        "harness.evidence_class": summary.runner ?? "runner",
      },
    }));
  }
  return result;
}

function sequenceSpans(runDir, root, traceID, rootSpan) {
  const file = path.join(runDir, root.target, "sequence-events.jsonl");
  if (!existsSync(file)) return { spans: [], dependencies: new Map(), dependencyWaitMs: new Map(), queueWaitMs: 0, resourceWaitIntervals: [] };
  const events = parseJSONLines(file, runDir);
  const starts = new Map();
  const eligible = new Map();
  const spans = [];
  let queueWaitMs = 0;
  const priorSpansByTarget = new Map();
  const dependencies = new Map();
  const dependencyWaitMs = new Map();
  const resourceWaitIntervals = [];
  for (const [index, event] of events.entries()) {
    validateSchemaSync("cartulary.harness_sequence_event.v1", event);
    if (event.seq !== index + 1) {
      throw new Error(`${runRelative(runDir, file)} has a sequence gap`);
    }
    if (event.event === "step_eligible") eligible.set(event.step_index, event);
    if (event.event === "step_started") starts.set(event.step_index, event);
    if (!new Set(["step_finished", "step_skipped"]).has(event.event)) continue;
    const startEvent = starts.get(event.step_index);
    const eligibleEvent = eligible.get(event.step_index);
    if (!eligibleEvent) throw new Error(`${runRelative(runDir, file)} terminal step has no matching eligibility`);
    if (event.event === "step_finished" && !startEvent) {
      throw new Error(`${runRelative(runDir, file)} step_finished has no matching step_started`);
    }
    const eligibleAt = timestampMs(event.eligible_at ?? eligibleEvent.emitted_at, "sequence step eligibility");
    const end = timestampMs(event.ended_at ?? event.emitted_at, "sequence step terminal");
    const start = startEvent
      ? timestampMs(event.started_at ?? startEvent.emitted_at, "sequence step start")
      : end;
    if (start < eligibleAt || end < start) {
      throw new Error(`${runRelative(runDir, file)} has malformed sequence boundaries`);
    }
    const waitDuration = Math.max(0, start - eligibleAt);
    queueWaitMs += waitDuration;
    const resourceClaims = event.resource_claims ?? eligibleEvent.resource_claims ?? {};
    for (const resource of Object.keys(resourceClaims).sort((left, right) => left.localeCompare(right))) {
      if (waitDuration > 0) resourceWaitIntervals.push({ resource, start: eligibleAt, end: start });
    }
    if (waitDuration > 0) {
      const waitReason = event.event === "step_skipped"
        ? "scheduler_stop"
        : Object.keys(resourceClaims).length > 0
          ? "resources"
          : "capacity";
      spans.push(makeSpan({
        traceID,
        occurrence: `${runRelative(runDir, file)}:${eligibleEvent.seq}:sequence_wait`,
        parentSpanID: rootSpan.span_id,
        name: `${event.target}-queue-wait`,
        phase: "scheduler_wait",
        status: "pending",
        start: eligibleAt,
        end: start,
        attrs: {
          "harness.target": event.target,
          "harness.wait_reason": waitReason,
          ...(Object.keys(resourceClaims)[0] ? { "harness.resource_class": Object.keys(resourceClaims).sort()[0] } : {}),
        },
      }));
    }
    if (event.event === "step_skipped") {
      starts.delete(event.step_index);
      eligible.delete(event.step_index);
      continue;
    }
    const predecessorIDs = (event.needs ?? startEvent.needs ?? eligibleEvent.needs ?? [])
      .map((target) => priorSpansByTarget.get(target)?.at(-1)?.span_id)
      .filter(Boolean);
    const span = makeSpan({
      traceID,
      occurrence: `${runRelative(runDir, file)}:${event.seq}:sequence_step`,
      parentSpanID: rootSpan.span_id,
      links: predecessorIDs,
      name: event.target,
      phase: "sequence_step",
      status: event.status,
      start,
      end,
      attrs: { "harness.target": event.target, "harness.work_unit": `step-${event.step_index}` },
    });
    spans.push(span);
    dependencies.set(span.span_id, predecessorIDs);
    dependencyWaitMs.set(span.span_id, waitDuration);
    const targetSpans = priorSpansByTarget.get(event.target) ?? [];
    targetSpans.push(span);
    priorSpansByTarget.set(event.target, targetSpans);
    starts.delete(event.step_index);
    eligible.delete(event.step_index);
  }
  if (starts.size > 0) throw new Error(`${runRelative(runDir, file)} has unfinished sequence steps`);
  if (eligible.size > 0) throw new Error(`${runRelative(runDir, file)} has unterminated eligible sequence steps`);
  if (
    events.filter((event) => event.event === "sequence_started").length !== 1 ||
    events.filter((event) => new Set(["sequence_finished", "sequence_interrupted"]).has(event.event)).length !== 1
  ) {
    throw new Error(`${runRelative(runDir, file)} has an incomplete sequence lifecycle`);
  }
  return { spans, dependencies, dependencyWaitMs, queueWaitMs, resourceWaitIntervals };
}

function retainedTimingSpans(runDir, root, traceID, rootSpan, spanByTarget) {
  const result = [];
  for (const file of sourceFiles(runDir).filter((candidate) => path.basename(path.dirname(candidate)) === "timing-spans")) {
    const timing = readJSON(file);
    if (String(timing.label ?? "").startsWith("run-go-target ")) continue;
    const interval = intervalForSummary(timing, runRelative(runDir, file));
    const target = path.basename(path.dirname(path.dirname(file)));
    const parent = spanByTarget.get(target) ?? rootSpan;
    const phase = timing.bucket === "report_collation" ? "finalizer"
      : timing.bucket === "service_wait" || timing.bucket === "migration" || timing.bucket === "server_startup" || timing.bucket === "frontend_startup" ? "service"
        : timing.bucket === "test_command" ? "runner" : "artifact";
    result.push(makeSpan({
      traceID,
      occurrence: `${runRelative(runDir, file)}:timing`,
      parentSpanID: parent.span_id,
      name: timing.label ?? timing.bucket,
      phase,
      status: timing.status ?? "pass",
      start: interval.start,
      end: interval.end,
      attrs: { "harness.target": target },
    }));
  }
  return result;
}

function schedulerSpans(runDir, root, traceID, rootSpan, spanByTarget) {
  const spans = [];
  let queueWaitMs = 0;
  const resourceWaitIntervals = [];
  const dependencies = new Map();
  const dependencyWaitMs = new Map();
  for (const file of sourceFiles(runDir).filter((candidate) => path.basename(candidate) === "scheduler-events.jsonl")) {
    const events = parseJSONLines(file, runDir);
    if (events.length === 0) continue;
    let previousMonotonic = -1;
    for (const [index, event] of events.entries()) {
      validateSchemaSync("cartulary.scheduler_event.v7", event);
      if (event.seq !== index + 1) throw new Error(`${runRelative(runDir, file)} has a sequence gap`);
      if (event.monotonic_ms < previousMonotonic) throw new Error(`${runRelative(runDir, file)} has a monotonic clock regression`);
      previousMonotonic = event.monotonic_ms;
    }
    const schedulerTarget = events[0].target ?? path.basename(path.dirname(file));
    const parent = spanByTarget.get(schedulerTarget) ?? rootSpan;
    if (events[0].event !== "scheduler-start" || events.at(-1).event !== "scheduler-finish") {
      throw new Error(`${runRelative(runDir, file)} has an incomplete scheduler lifecycle`);
    }
    const anchorWall = timestampMs(events[0].emitted_at, "scheduler lifecycle start") - events[0].monotonic_ms;
    const finalStates = events.at(-1).work_unit_states;
    const spanIDByUnit = new Map();
    for (const state of finalStates) {
      if (state.started_monotonic_ms === null || state.terminal_monotonic_ms === null) {
        if (!new Set(["skipped_dependency", "cancelled", "interrupted"]).has(state.terminal_state)) {
          throw new Error(`${runRelative(runDir, file)} has an unterminated work unit ${state.work_unit_id}`);
        }
        continue;
      }
      if (
        state.eligibility_monotonic_ms === null ||
        state.started_monotonic_ms < state.eligibility_monotonic_ms ||
        state.terminal_monotonic_ms < state.started_monotonic_ms
      ) {
        throw new Error(`${runRelative(runDir, file)} has malformed monotonic boundaries for ${state.work_unit_id}`);
      }
      const startEvent = events.find((event) =>
        new Set(["start", "finalize-start"]).has(event.event) &&
        (event.work_unit_id ?? event.finalizer_id) === state.work_unit_id);
      const isFinalizer = startEvent?.event === "finalize-start";
      const occurrence = `${runRelative(runDir, file)}:${state.manifest_ordinal}:scheduler_work`;
      const id = spanID(traceID, occurrence);
      spanIDByUnit.set(state.work_unit_id, id);
      const waitDuration = state.started_monotonic_ms - state.eligibility_monotonic_ms;
      queueWaitMs += waitDuration;
      if (waitDuration > 0) {
        const resources = new Set(["resources", "earlier_overlapping_ready"]).has(state.wait_reason)
          ? state.blocking_resources.length > 0
            ? state.blocking_resources
            : Object.keys(state.resource_claims)
          : [];
        for (const resource of resources) {
          resourceWaitIntervals.push({
            resource,
            start: anchorWall + state.eligibility_monotonic_ms,
            end: anchorWall + state.started_monotonic_ms,
          });
        }
        spans.push(makeSpan({
          traceID,
          occurrence: `${runRelative(runDir, file)}:${state.manifest_ordinal}:scheduler_wait`,
          parentSpanID: parent.span_id,
          name: `${state.work_unit_id}-wait`,
          phase: "scheduler_wait",
          status: "pending",
          start: anchorWall + state.eligibility_monotonic_ms,
          end: anchorWall + state.started_monotonic_ms,
          attrs: {
            "harness.target": schedulerTarget,
            "harness.work_unit": state.work_unit_id,
            "harness.wait_reason": state.wait_reason ?? (resources.length > 0 ? "resources" : "capacity"),
            ...(resources[0] ? { "harness.resource_class": resources[0] } : {}),
          },
        }));
      }
      spans.push(makeSpan({
        traceID,
        occurrence,
        parentSpanID: parent.span_id,
        name: state.work_unit_id,
        phase: isFinalizer ? "finalizer" : "scheduler_work",
        status: state.terminal_state === "passed" ? "pass" : "fail",
        start: anchorWall + state.started_monotonic_ms,
        end: anchorWall + state.terminal_monotonic_ms,
        attrs: {
          "harness.target": schedulerTarget,
          "harness.work_unit": state.work_unit_id,
          ...(Object.keys(state.resource_claims)[0] ? { "harness.resource_class": Object.keys(state.resource_claims).sort()[0] } : {}),
        },
      }));
      dependencyWaitMs.set(id, waitDuration);
    }
    for (const edge of events.at(-1).dependency_edges) {
      const to = spanIDByUnit.get(edge.to);
      const from = spanIDByUnit.get(edge.from);
      if (!to || !from) continue;
      const retained = dependencies.get(to) ?? [];
      retained.push(from);
      dependencies.set(to, [...new Set(retained)].sort((left, right) => left.localeCompare(right)));
    }
  }
  return { spans, queueWaitMs, resourceWaitIntervals, dependencies, dependencyWaitMs };
}

function directChildUnionMs(spans, rootSpan) {
  const rootStart = fromUnixNano(rootSpan.start_time_unix_nano);
  const rootEnd = fromUnixNano(rootSpan.end_time_unix_nano);
  const intervals = spans
    .filter((span) => span.parent_span_id === rootSpan.span_id)
    .map((span) => [
      Math.max(rootStart, fromUnixNano(span.start_time_unix_nano)),
      Math.min(rootEnd, fromUnixNano(span.end_time_unix_nano)),
    ])
    .filter(([start, end]) => end > start)
    .sort((left, right) => left[0] - right[0] || left[1] - right[1]);
  return intervalUnion(intervals);
}

function dependencyCriticalPath(spans, rootSpan, dependencies = new Map(), dependencyWaitMs = new Map()) {
  const sequence = spans.filter((span) => span.phase === "sequence_step");
  const candidates = sequence.length > 0
    ? sequence
    : spans.filter((span) => span.span_id !== rootSpan.span_id && !new Set(["target", "scheduler_wait"]).has(span.phase));
  if (candidates.length === 0) return { durationMs: 0, spanIDs: [] };
  const byID = new Map(candidates.map((span) => [span.span_id, span]));
  const memo = new Map();
  const visiting = new Set();
  function bestTo(span) {
    if (memo.has(span.span_id)) return memo.get(span.span_id);
    if (visiting.has(span.span_id)) throw new Error("dependency cycle in retained sequence events");
    visiting.add(span.span_id);
    const duration = fromUnixNano(span.end_time_unix_nano) - fromUnixNano(span.start_time_unix_nano) + (dependencyWaitMs.get(span.span_id) ?? 0);
    let best = { durationMs: duration, spanIDs: [span.span_id] };
    for (const dependencyID of dependencies.get(span.span_id) ?? []) {
      const dependency = byID.get(dependencyID);
      if (!dependency) continue;
      const prefix = bestTo(dependency);
      if (prefix.durationMs + duration > best.durationMs) {
        best = { durationMs: prefix.durationMs + duration, spanIDs: [...prefix.spanIDs, span.span_id] };
      }
    }
    visiting.delete(span.span_id);
    memo.set(span.span_id, best);
    return best;
  }
  return candidates
    .map(bestTo)
    .sort((left, right) => right.durationMs - left.durationMs || left.spanIDs.join("").localeCompare(right.spanIDs.join("")))[0];
}

function buildOTLPTrace(bundle) {
  return {
    resourceSpans: [{
      resource: { attributes: [{ key: "service.name", value: { stringValue: "cartulary.harness" } }] },
      scopeSpans: [{
        scope: { name: harnessScope, version: "1" },
        spans: bundle.spans.map((span) => ({
          traceId: bundle.trace_id,
          spanId: span.span_id,
          ...(span.parent_span_id ? { parentSpanId: span.parent_span_id } : {}),
          name: span.name,
          kind: 1,
          startTimeUnixNano: span.start_time_unix_nano,
          endTimeUnixNano: span.end_time_unix_nano,
          attributes: span.attributes.map((item) => ({ key: item.key, value: { stringValue: item.value } })),
          links: span.links.map((spanIDValue) => ({ traceId: bundle.trace_id, spanId: spanIDValue })),
          status: { code: span.status === "ERROR" ? 2 : span.status === "OK" ? 1 : 0 },
        })),
      }],
    }],
  };
}

function buildOTLPMetrics(root, hotspot, traceID) {
  const target = safeToken(root.target);
  const points = [
    ["cartulary.harness.invocation.duration", hotspot.duration_ms],
    ["cartulary.harness.dependency.critical_path", hotspot.actual_dependency_critical_path_ms],
    ["cartulary.harness.scheduler.queue_wait", hotspot.queue_wait_ms],
    ["cartulary.harness.scheduler.resource_blocking", hotspot.resource_blocking_ms],
    ["cartulary.harness.invocation.unattributed", hotspot.unattributed_envelope_ms],
  ];
  return {
    resourceMetrics: [{
      resource: { attributes: [{ key: "service.name", value: { stringValue: "cartulary.harness" } }] },
      scopeMetrics: [{
        scope: { name: harnessScope, version: "1" },
        metrics: points.map(([name, value]) => ({
          name,
          unit: "ms",
          gauge: {
            dataPoints: [{
              attributes: [{ key: "harness.target", value: { stringValue: target } }],
              asDouble: value,
              exemplars: [{ asDouble: value, traceId: traceID }],
            }],
          },
        })),
      }],
    }],
  };
}

function hotspotSummary(root, spans, rootSpan, scheduler, dependencies, dependencyWaitMs) {
  const durationMs = fromUnixNano(rootSpan.end_time_unix_nano) - fromUnixNano(rootSpan.start_time_unix_nano);
  const attributedUnionMs = directChildUnionMs(spans, rootSpan);
  const critical = dependencyCriticalPath(spans, rootSpan, dependencies, dependencyWaitMs);
  const resourceBlockingByResource = Object.fromEntries(
    [...new Set(scheduler.resourceWaitIntervals.map((interval) => interval.resource))]
      .sort((left, right) => left.localeCompare(right))
      .map((resource) => [
        resource,
        intervalUnion(scheduler.resourceWaitIntervals
          .filter((interval) => interval.resource === resource)
          .map((interval) => [interval.start, interval.end])),
      ]),
  );
  const finalizationUnionMs = intervalUnion(
    spans.filter((span) => span.phase === "finalizer").map((span) => [
      fromUnixNano(span.start_time_unix_nano),
      fromUnixNano(span.end_time_unix_nano),
    ]),
  );
  const candidates = spans
    .filter((span) => span.span_id !== rootSpan.span_id)
    .map((span) => ({ span, duration: fromUnixNano(span.end_time_unix_nano) - fromUnixNano(span.start_time_unix_nano) }))
    .sort((left, right) => right.duration - left.duration || left.span.span_id.localeCompare(right.span.span_id))
    .slice(0, 20);
  return {
    schema_id: "cartulary.harness_hotspot_summary.v1",
    invocation_id: safeToken(root.target),
    status: "complete",
    duration_ms: durationMs,
    scheduler_envelope_critical_path_ms: Number(root.summary.critical_path_wall_duration_ms ?? durationMs),
    actual_dependency_critical_path_ms: Math.min(durationMs, critical.durationMs),
    queue_wait_ms: scheduler.queueWaitMs,
    resource_blocking_ms: scheduler.resourceBlockingMs,
    resource_blocking_by_resource: resourceBlockingByResource,
    finalization_union_ms: finalizationUnionMs,
    attributed_union_ms: attributedUnionMs,
    unattributed_envelope_ms: Math.max(0, durationMs - attributedUnionMs),
    critical_path: critical.spanIDs,
    hotspots: candidates.map(({ span, duration }, index) => ({
      rank: index + 1,
      span_id: span.span_id,
      name: span.name,
      phase: span.phase,
      duration_ms: duration,
      queue_wait_ms: span.phase === "scheduler_wait" ? duration : 0,
      resource_blocking_ms: span.phase === "scheduler_wait" ? duration : 0,
    })),
  };
}

function mergeDependencyMaps(...maps) {
  const result = new Map();
  for (const values of maps) {
    for (const [spanIDValue, dependencies] of values) {
      result.set(spanIDValue, [...new Set([...(result.get(spanIDValue) ?? []), ...dependencies])]
        .sort((left, right) => left.localeCompare(right)));
    }
  }
  return result;
}

function buildInvocation(runDir, runID, root, allSourceFiles, contextRecord) {
  const occurrence = `${runRelative(runDir, root.file)}:invocation`;
  const traceID = sha256(`${runID}\u0000${safeToken(root.target)}\u0000${occurrence}`).slice(0, 32);
  const retainedInterval = intervalForSummary(root.summary, root.target);
  const interval = {
    start: timestampMs(contextRecord.context.started_at, "execution context started_at"),
    end: timestampMs(contextRecord.context.ended_at, "execution context ended_at"),
  };
  if (interval.end < interval.start) {
    throw new Error("retained execution context has a negative invocation interval");
  }
  if (retainedInterval.start < interval.start - 5 || retainedInterval.end > interval.end + 5) {
    throw new Error("top-level result escapes the retained invocation boundary");
  }
  const rootSpan = makeSpan({
    traceID,
    occurrence,
    parentSpanID: null,
    name: root.target,
    phase: "target",
    status: summaryStatus(root.summary),
    start: interval.start,
    end: interval.end,
    attrs: { "harness.target": root.target },
  });
  const retainedTargetIDs = new Set(contextRecord.context.measurement_contracts.map((contract) => contract.target));
  retainedTargetIDs.add(root.target);
  const targets = targetSpans(
    runDir,
    root,
    traceID,
    rootSpan,
    retainedTargetIDs,
    contextRecord.context.invocation_edges,
    contextRecord.context.invocation_boundary_retained,
  );
  const sequence = sequenceSpans(runDir, root, traceID, rootSpan);
  const scheduler = schedulerSpans(runDir, root, traceID, rootSpan, targets.spanByTarget);
  scheduler.queueWaitMs += sequence.queueWaitMs;
  scheduler.resourceWaitIntervals = [...scheduler.resourceWaitIntervals, ...sequence.resourceWaitIntervals];
  scheduler.resourceBlockingMs = intervalUnion(
    scheduler.resourceWaitIntervals.map((interval) => [interval.start, interval.end]),
  );
  const runners = stepSpans(runDir, root, traceID, rootSpan, targets.spanByTarget);
  const timings = retainedTimingSpans(runDir, root, traceID, rootSpan, targets.spanByTarget);
  const dependencyMap = mergeDependencyMaps(sequence.dependencies, scheduler.dependencies);
  const dependencyWaitMap = mergeDependencyMaps();
  for (const [spanIDValue, wait] of sequence.dependencyWaitMs) dependencyWaitMap.set(spanIDValue, wait);
  for (const [spanIDValue, wait] of scheduler.dependencyWaitMs) dependencyWaitMap.set(spanIDValue, wait);
  for (const span of scheduler.spans) span.links = dependencyMap.get(span.span_id) ?? span.links;
  const spans = [rootSpan, ...targets.spans, ...sequence.spans, ...scheduler.spans, ...runners, ...timings]
    .sort((left, right) =>
      BigInt(left.start_time_unix_nano) < BigInt(right.start_time_unix_nano) ? -1 :
        BigInt(left.start_time_unix_nano) > BigInt(right.start_time_unix_nano) ? 1 :
          left.span_id.localeCompare(right.span_id));
  const sourceArtifacts = allSourceFiles.map((file) => ({
    path: runRelative(runDir, file),
    sha256: sha256(readFileSync(file)),
  }));
  const bundle = {
    schema_id: "cartulary.harness_trace_bundle.v1",
    scope: harnessScope,
    invocation_id: safeToken(root.target),
    run_id: runID,
    command_id: contextRecord.context.command_id,
    execution_context_sha256: contextRecord.sha256,
    trace_id: traceID,
    root_span_id: rootSpan.span_id,
    status: "complete",
    start_time_unix_nano: rootSpan.start_time_unix_nano,
    end_time_unix_nano: rootSpan.end_time_unix_nano,
    source_artifacts: sourceArtifacts,
    spans,
  };
  const hotspot = hotspotSummary(root, spans, rootSpan, scheduler, dependencyMap, dependencyWaitMap);
  validateSchemaSync(bundle.schema_id, bundle);
  validateSchemaSync(hotspot.schema_id, hotspot);
  return {
    bundle,
    hotspot,
    traceOTLP: buildOTLPTrace(bundle),
    metricsOTLP: buildOTLPMetrics(root, hotspot, traceID),
  };
}

function outputPaths(runDir, invocationID) {
  const base = path.join(runDir, diagnosticDir, invocationID);
  return {
    base,
    traceBundle: path.join(base, "trace-bundle.json"),
    traceOTLP: path.join(base, "trace.otlp.json"),
    metricsOTLP: path.join(base, "metrics.otlp.json"),
    hotspotSummary: path.join(base, "hotspot-summary.json"),
  };
}

function writeJSON(file, value, allowedRoot) {
  secureWriteFile(file, canonicalJSON(value), { allowedRoot });
  chmodSync(file, 0o600);
}

function artifactRef(runDir, file, value) {
  return {
    path: runRelative(runDir, file),
    sha256: sha256(canonicalJSON(value)),
  };
}

function observabilityIndexPath(runDir) {
  return path.join(runDir, diagnosticDir, "observability-index.json");
}

export function reconstructObservability(runDir, {
  write = true,
  contextRecord: suppliedContextRecord,
  policy: suppliedPolicy,
} = {}) {
  const resolvedRunDir = path.resolve(runDir);
  const runID = path.basename(resolvedRunDir);
  const existingContextRecord = suppliedContextRecord ??
    (existsSync(contextPath(resolvedRunDir)) ? loadRetainedExecutionContext(resolvedRunDir) : null);
  const contextRecord = existingContextRecord ?? (write
    ? (() => {
        const context = captureExecutionContext(resolvedRunDir);
        const file = contextPath(resolvedRunDir);
        return { context, file, sha256: sha256(readFileSync(file)) };
      })()
    : loadRetainedExecutionContext(resolvedRunDir));
  const roots = invocationRoots(resolvedRunDir, contextRecord.context.target);
  if (roots.length !== 1) throw new Error(`retained run must have exactly one top-level invocation; found ${roots.length}`);
  if (
    contextRecord.context.target !== roots[0].target ||
    contextRecord.context.invocation_id !== safeToken(roots[0].target)
  ) {
    throw new Error("retained execution context invocation identity mismatch");
  }
  const files = sourceFiles(resolvedRunDir);
  const built = roots
    .sort((left, right) => left.target.localeCompare(right.target))
    .map((root) => ({ root, result: buildInvocation(resolvedRunDir, runID, root, files, contextRecord) }));
  const invocations = built.map(({ root, result }) => {
    const paths = outputPaths(resolvedRunDir, result.bundle.invocation_id);
    return {
      invocation_id: result.bundle.invocation_id,
      target: root.target,
      command_id: contextRecord.context.command_id,
      status: "complete",
      trace_id: result.bundle.trace_id,
      artifacts: {
        trace_bundle: artifactRef(resolvedRunDir, paths.traceBundle, result.bundle),
        trace_otlp: artifactRef(resolvedRunDir, paths.traceOTLP, result.traceOTLP),
        metrics_otlp: artifactRef(resolvedRunDir, paths.metricsOTLP, result.metricsOTLP),
        hotspot_summary: artifactRef(resolvedRunDir, paths.hotspotSummary, result.hotspot),
      },
      source_digests: result.bundle.source_artifacts,
    };
  });
  let retainedPolicy;
  const retainedIndexFile = observabilityIndexPath(resolvedRunDir);
  if (!write && existsSync(retainedIndexFile)) {
    const retainedIndex = readJSON(retainedIndexFile);
    validateSchemaSync("cartulary.harness_observability_index.v1", retainedIndex);
    retainedPolicy = retainedIndex.policy;
  }
  const index = {
    schema_id: "cartulary.harness_observability_index.v1",
    run_id: runID,
    status: "complete",
    scope: harnessScope,
    execution_context: {
      path: runRelative(resolvedRunDir, contextRecord.file),
      sha256: contextRecord.sha256,
    },
    policy: suppliedPolicy ?? retainedPolicy ?? observabilityPolicy(),
    invocations,
  };
  validateSchemaSync(index.schema_id, index);
  if (write) {
    const base = path.join(resolvedRunDir, diagnosticDir);
    secureMkdir(base, { allowedRoot: resolvedRunDir });
    for (const { result } of built) {
      const paths = outputPaths(resolvedRunDir, result.bundle.invocation_id);
      secureMkdir(paths.base, { allowedRoot: resolvedRunDir });
      writeJSON(paths.traceBundle, result.bundle, resolvedRunDir);
      writeJSON(paths.traceOTLP, result.traceOTLP, resolvedRunDir);
      writeJSON(paths.metricsOTLP, result.metricsOTLP, resolvedRunDir);
      writeJSON(paths.hotspotSummary, result.hotspot, resolvedRunDir);
    }
    writeJSON(observabilityIndexPath(resolvedRunDir), index, resolvedRunDir);
  }
  return { runDir: resolvedRunDir, context: contextRecord.context, contextRecord, index, built };
}

function loadArtifactRef(runDir, ref, label) {
  const file = artifactPath(runDir, ref.path, label);
  if (!existsSync(file)) throw new Error(`${label} is missing`);
  const bytes = readFileSync(file);
  if (sha256(bytes) !== ref.sha256) throw new Error(`${label} digest mismatch`);
  return { file, value: JSON.parse(bytes.toString("utf8")) };
}

export function loadRetainedObservability(runDir) {
  const resolvedRunDir = path.resolve(runDir);
  const indexFile = observabilityIndexPath(resolvedRunDir);
  if (!existsSync(indexFile)) throw new Error("retained run has no observability index");
  const index = readJSON(indexFile);
  validateSchemaSync("cartulary.harness_observability_index.v1", index);
  if (index.status !== "complete") throw new Error("retained observability index is incomplete");
  const contextRecord = loadRetainedExecutionContext(resolvedRunDir);
  if (
    index.execution_context.path !== runRelative(resolvedRunDir, contextRecord.file) ||
    index.execution_context.sha256 !== contextRecord.sha256
  ) {
    throw new Error("retained execution context digest mismatch");
  }
  const reconstructed = reconstructObservability(resolvedRunDir, {
    write: false,
    contextRecord,
    policy: index.policy,
  });
  if (canonicalJSON(index) !== canonicalJSON(reconstructed.index)) {
    throw new Error("retained observability index differs from independent reconstruction");
  }
  for (const [invocationIndex, invocation] of index.invocations.entries()) {
    const expected = reconstructed.built[invocationIndex]?.result;
    if (!expected) throw new Error("retained observability invocation count mismatch");
    const artifacts = {
      trace_bundle: expected.bundle,
      trace_otlp: expected.traceOTLP,
      metrics_otlp: expected.metricsOTLP,
      hotspot_summary: expected.hotspot,
    };
    for (const [name, ref] of Object.entries(invocation.artifacts)) {
      const retained = loadArtifactRef(resolvedRunDir, ref, `${invocation.invocation_id}.${name}`);
      if (canonicalJSON(retained.value) !== canonicalJSON(artifacts[name])) {
        throw new Error(`${invocation.invocation_id}.${name} differs from independent reconstruction`);
      }
    }
    for (const source of invocation.source_digests) {
      const file = artifactPath(resolvedRunDir, source.path, "source artifact");
      if (!existsSync(file) || sha256(readFileSync(file)) !== source.sha256) {
        throw new Error(`source artifact digest mismatch: ${source.path}`);
      }
    }
  }
  return reconstructed;
}

export function writePartialObservability(runDir, error) {
  const resolvedRunDir = path.resolve(runDir);
  const runID = path.basename(resolvedRunDir);
  const diagnostic = safeDiagnostic(error);
  const contextRecord = loadRetainedExecutionContext(resolvedRunDir);
  const zeroDigest = "0".repeat(64);
  const index = {
    schema_id: "cartulary.harness_observability_index.v1",
    run_id: runID,
    status: "partial",
    scope: harnessScope,
    execution_context: {
      path: runRelative(resolvedRunDir, contextRecord.file),
      sha256: contextRecord.sha256,
    },
    policy: observabilityPolicy(),
    invocations: [{
      invocation_id: "partial",
      target: "unknown",
      command_id: contextRecord.context.command_id,
      status: "partial",
      trace_id: sha256(`${runID}\u0000partial`).slice(0, 32),
      artifacts: {
        trace_bundle: { path: `${diagnosticDir}/partial/trace-bundle.json`, sha256: zeroDigest },
        trace_otlp: { path: `${diagnosticDir}/partial/trace.otlp.json`, sha256: zeroDigest },
        metrics_otlp: { path: `${diagnosticDir}/partial/metrics.otlp.json`, sha256: zeroDigest },
        hotspot_summary: { path: `${diagnosticDir}/partial/hotspot-summary.json`, sha256: zeroDigest },
      },
      source_digests: [],
      diagnostic,
    }],
    diagnostic,
  };
  validateSchemaSync(index.schema_id, index);
  const base = path.join(resolvedRunDir, diagnosticDir);
  secureMkdir(base, { allowedRoot: resolvedRunDir });
  writeJSON(path.join(base, "observability-index.json"), index, resolvedRunDir);
  return index;
}

export function finalizeObservabilitySafely(runDir, metadata = {}) {
  try {
    captureExecutionContext(runDir, metadata);
    return { status: "complete", result: reconstructObservability(runDir) };
  } catch (error) {
    try {
      writePartialObservability(runDir, error);
    } catch {
      // Diagnostic generation never changes the native test result.
    }
    return { status: "partial", diagnostic: safeDiagnostic(error) };
  }
}

export function deterministicBytes(result) {
  return JSON.stringify({
    index: result.index,
    invocations: result.built.map(({ result: item }) => ({
      bundle: item.bundle,
      hotspot: item.hotspot,
      traceOTLP: item.traceOTLP,
      metricsOTLP: item.metricsOTLP,
    })),
  });
}

export function printObservabilityPerformance(runDir, target = "") {
  const result = loadRetainedObservability(runDir);
  const selected = target
    ? result.built.filter(({ root }) => root.target === target)
    : result.built;
  if (selected.length === 0) throw new Error("selected target has no observability invocation");
  return selected.map(({ root, result: item }) => ({ target: root.target, ...item.hotspot }));
}
