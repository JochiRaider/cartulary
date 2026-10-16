import { createHash } from "node:crypto";
import path from "node:path";

import { canonicalJSONString, parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../contract/index.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";
import { validateExecutionContext } from "./canonical-context.mjs";

const evidenceSchemaID = "cartulary.harness_performance_evidence_roots.v3";
export const baselineSchemaID = "cartulary.harness_public_target_duration_baselines.v3";

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sameJSON(left, right) {
  return canonicalJSONString(left) === canonicalJSONString(right);
}

function digestJSON(value) {
  return createHash("sha256").update(canonicalJSONString(value)).digest("hex");
}

function withoutPrefix(value) {
  if (!/^sha256:[a-f0-9]{64}$/u.test(value)) throw new Error(`invalid retained digest ${value}`);
  return value.slice("sha256:".length);
}

function readJSON(file) {
  return parseStrictJSON(readRetainedArtifactBytes(path.dirname(file), path.basename(file), 16 * 1024 * 1024).toString("utf8"));
}

export function median(values) {
  if (values.length === 0) throw new Error("median requires at least one value");
  const sorted = [...values].sort((left, right) => left - right);
  const middle = Math.floor(sorted.length / 2);
  return sorted.length % 2 === 0
    ? (sorted[middle - 1] + sorted[middle]) / 2
    : sorted[middle];
}

export function nearestRankP90(values) {
  if (values.length === 0) throw new Error("p90 requires at least one value");
  const sorted = [...values].sort((left, right) => left - right);
  return sorted[Math.ceil(sorted.length * 0.9) - 1];
}

export function performanceRoster(surface) {
  const required = surface.observability_policy.required_targets;
  const targets = new Map(surface.targets.map((entry) => [entry.name, entry]));
  const profiles = new Map(
    surface.observability_policy.measurement_profiles.map((entry) => [entry.profile_id, entry]),
  );
  const bindings = new Map();
  for (const binding of surface.observability_policy.target_measurement_profiles) {
    if (bindings.has(binding.target)) throw new Error(`duplicate measurement-profile binding ${binding.target}`);
    bindings.set(binding.target, binding);
  }
  if (new Set(required).size !== required.length) throw new Error("required performance roster contains duplicates");
  const sorted = [...required].sort(compareASCII);
  if (!sameJSON(required, sorted)) throw new Error("required performance roster must be sorted");
  return new Map(sorted.map((target) => {
    const command = targets.get(target);
    const binding = bindings.get(target);
    const profile = profiles.get(binding?.profile_id);
    if (!command?.command_id || !profile) throw new Error(`${target} has no complete performance binding`);
    const improvement = profile.performance_gates.some((gate) => gate.includes("improvement"));
    return [target, {
      command_id: command.command_id,
      measurement_profile_id: profile.profile_id,
      canonical_inputs: profile.canonical_inputs,
      input_defaults: Object.fromEntries((command.input_contract?.inputs ?? [])
        .filter((input) => input.default !== null && input.default !== undefined)
        .map((input) => [input.name, String(input.default)])),
      observation_eligibility: profile.observation_eligibility,
      gate: improvement ? "required_improvement" : "no_regression",
      allowed_policy_transition:
        binding.allowed_policy_transition ?? profile.allowed_policy_transition,
    }];
  }));
}

function resolveRoot(baseDirectory, reference) {
  return path.isAbsolute(reference)
    ? path.resolve(reference)
    : path.resolve(baseDirectory, reference);
}

function rootReference(repositoryRoot, runRoot, fallback) {
  const relative = path.relative(repositoryRoot, runRoot).replaceAll("\\", "/");
  return relative !== "" && relative !== ".." && !relative.startsWith("../")
    ? relative
    : fallback;
}

async function readCanonicalRoot(repositoryRoot, baseDirectory, reference, target, expectedCacheMode, retained) {
  const runRoot = resolveRoot(baseDirectory, reference);
  if (!retained.has(runRoot)) retained.set(runRoot, (async () => {
    const canonical = await validateCanonicalRun(runRoot);
    const context = readJSON(path.join(runRoot, "execution-context.json"));
    const cleanup = readJSON(path.join(runRoot, "cleanup-results.json"));
    const bindings = validateExecutionContext(context, canonical, cleanup);
    const scan = readJSON(path.join(runRoot, "retained-secret-scan.json"));
    validateSchemaSync("cartulary.harness_retained_secret_scan.v1", scan);
    if (scan.policy_digest !== canonical.manifest.instrumentation.policy_digest) throw new Error("retained scan policy differs from execution");
    return { ...canonical, ...bindings, context };
  })());
  const canonical = await retained.get(runRoot);
  const { manifest, summary, targetSummaries, terminal } = canonical;
  const targetSummary = targetSummaries.get(target);
  if (!targetSummary) throw new Error(`${reference} has no retained ${target} projection`);
  if (manifest.source_state !== "clean") throw new Error(`${reference} has dirty_source`);
  if (manifest.cache_mode !== expectedCacheMode) throw new Error(`${reference} cache mode ${manifest.cache_mode} does not match ${expectedCacheMode}`);
  if (summary.status !== "pass" || targetSummary.status !== "pass" || targetSummary.unit_ids.some((id) => terminal.get(id)?.status !== "passed")) {
    throw new Error(`${reference} does not contain successful ${target} evidence`);
  }
  return { runRoot, root: rootReference(repositoryRoot, runRoot, reference), manifest, summary, targetSummary,
    systemProfileDigest: canonical.systemProfileDigest, eventCount: canonical.eventCount, inclusive_wall_ms: targetSummary.inclusive_wall_ms };
}

function semanticDeclaredInputs(inputs) {
  const ignored = new Set([
    "HARNESS_DIAGNOSTICS",
    "CARTULARY_HARNESS_CACHE_MODE",
    "CARTULARY_HARNESS_CAPACITY_OVERRIDE",
    "CARTULARY_OUTPUT_MODE",
  ]);
  return Object.fromEntries(
    Object.entries(inputs).filter(([name]) => !ignored.has(name)).sort(([left], [right]) => compareASCII(left, right)),
  );
}

function assertCanonicalInputs(target, expected, actual, defaults = {}) {
  const normalized = { ...actual };
  for (const [name, value] of Object.entries(expected)) {
    if (value === "omitted") {
      if (normalized[name] !== undefined) throw new Error(`${target} input ${name} must be omitted`);
      delete normalized[name];
    } else if (normalized[name] !== value) {
      throw new Error(`${target} input ${name} does not match its canonical binding`);
    } else {
      delete normalized[name];
    }
  }
  for (const [name, value] of Object.entries(normalized)) {
    if (Object.hasOwn(defaults, name) && value === defaults[name]) delete normalized[name];
  }
  if (Object.keys(normalized).length > 0) {
    throw new Error(`${target} has non-canonical inputs ${Object.keys(normalized).sort(compareASCII).join(",")}`);
  }
}

function observation(record, target, timingSource, providerTarget) {
  if (timingSource === "public_invocation_envelope") {
    if (target !== providerTarget || record.manifest.target !== target) {
      throw new Error(`${target} invocation-envelope timing requires a direct provider`);
    }
    throw new Error(`${target} has no qualified complete public invocation envelope; graph duration is not public invocation duration`);
  }
  if (timingSource === "canonical_unit_interval_union") return record.inclusive_wall_ms;
  if (timingSource === "canonical_timing_bucket_union") {
    const accounting = record.targetSummary.timing_accounting;
    return ["setup_ms", "fixture_ms", "execution_ms", "collation_ms", "wrapper_ms", "unattributed_ms"]
      .reduce((total, field) => total + accounting[field], 0);
  }
  throw new Error(`${target} has unsupported timing source ${timingSource}`);
}

function assertSameWindow(target, records) {
  const identity = records[0];
  if (new Set(records.map((record) => record.manifest.run_id)).size !== records.length ||
      new Set(records.map((record) => record.runRoot)).size !== records.length) throw new Error(`${target} performance window repeats an observation`);
  for (let index = 1; index < records.length; index++) {
    const prior = records[index - 1];
    const end = Date.parse(prior.manifest.started_at) + prior.summary.wall_duration_ms;
    if (!Number.isFinite(end) || Date.parse(records[index].manifest.started_at) < end) throw new Error(`${target} performance window is overlapping or out of order`);
  }
  for (const record of records.slice(1)) {
    if (record.systemProfileDigest !== identity.systemProfileDigest) throw new Error(`${target} performance window mixes host profiles`);
    for (const field of ["source_commit", "source_digest", "system_digest", "toolchain_digest", "graph_digest"]) {
      if (record.manifest[field] !== identity.manifest[field]) {
        throw new Error(`${target} performance window mixes ${field}`);
      }
    }
    if (!sameJSON(record.manifest.instrumentation, identity.manifest.instrumentation)) throw new Error(`${target} instrumentation policy differs within window`);
    if (!sameJSON(semanticDeclaredInputs(record.manifest.declared_inputs), semanticDeclaredInputs(identity.manifest.declared_inputs))) {
      throw new Error(`${target} performance window mixes canonical inputs`);
    }
    if (record.manifest.target !== identity.manifest.target) {
      throw new Error(`${target} performance window mixes provider targets`);
    }
    if (record.targetSummary.command_id !== identity.targetSummary.command_id) {
      throw new Error(`${target} performance window mixes command IDs`);
    }
    if (record.targetSummary.workload_digest !== identity.targetSummary.workload_digest) {
      throw new Error(`${target} performance window mixes workload evidence`);
    }
  }
}

function timingAccounting(records) {
  const field = (name, source = "timing_accounting") =>
    median(records.map((record) => record.targetSummary[source][name]));
  return {
    critical_path_p50_ms: median(records.map((record) => record.targetSummary.actual_dependency_critical_path_ms)),
    resource_blocking_p50_ms: field("resource_blocking_ms"),
    setup_p50_ms: field("setup_ms"),
    fixture_p50_ms: field("fixture_ms"),
    execution_p50_ms: field("execution_ms"),
    collation_p50_ms: field("collation_ms"),
    wrapper_p50_ms: field("wrapper_ms"),
    unattributed_p50_ms: field("unattributed_ms"),
    process_count_p50: field("process_count"),
  };
}

function sourceWindows(rows) {
  const grouped = new Map();
  for (const row of rows) {
    const key = `${row.source_commit}:${row.source_snapshot_sha256}`;
    const entry = grouped.get(key) ?? {
      source_commit: row.source_commit,
      source_snapshot_sha256: row.source_snapshot_sha256,
      targets: [],
    };
    entry.targets.push(row.target);
    grouped.set(key, entry);
  }
  return [...grouped.values()]
    .map((entry) => ({ ...entry, targets: entry.targets.sort(compareASCII) }))
    .sort((left, right) =>
      compareASCII(left.source_commit, right.source_commit) ||
      compareASCII(left.source_snapshot_sha256, right.source_snapshot_sha256),
    );
}

function assertUnique(items, key, label) {
  const values = items.map((item) => item[key]);
  if (new Set(values).size !== values.length) throw new Error(`${label} contains duplicate ${key}`);
}

async function buildQualifiedBaseline({
  repositoryRoot,
  baseDirectory,
  surface,
  windows,
  bindings,
  internalWindows = [],
  internalBindings = [],
  rejectedRoots = [],
  role = "reference",
}) {
  const roster = performanceRoster(surface);
  const retained = new Map();
  assertUnique(windows, "window_id", `${role} windows`);
  assertUnique(bindings, "target", `${role} bindings`);
  const byWindow = new Map(windows.map((window) => [window.window_id, window]));
  const referenced = new Set();
  const buildRows = async (selectedBindings, selectedWindows, diagnostic) => {
    const selectedByWindow = new Map(selectedWindows.map((window) => [window.window_id, window]));
    return Promise.all(selectedBindings.map(async (binding) => {
      const window = selectedByWindow.get(binding.window_id);
      if (!window) throw new Error(`${role} ${binding.target} references unknown window ${binding.window_id}`);
      referenced.add(window.window_id);
      const policy = roster.get(binding.target);
      if (!diagnostic && !policy) throw new Error(`${role} contains non-public target ${binding.target}`);
      if (diagnostic && policy) throw new Error(`${binding.target} is public and cannot be an internal diagnostic`);
      const effective = policy ?? {
        command_id: null,
        measurement_profile_id: "internal_diagnostic",
        canonical_inputs: {},
        observation_eligibility: "direct_or_aggregate_exact_once",
        gate: "diagnostic_only",
      };
      if (effective.observation_eligibility === "direct_only" && window.provider_target !== binding.target) {
        throw new Error(`${binding.target} requires a direct performance provider`);
      }
      const cold = await readCanonicalRoot(repositoryRoot, baseDirectory, window.cold_root, binding.target, "cold", retained);
      const warmup = await readCanonicalRoot(repositoryRoot, baseDirectory, window.warmup_root, binding.target, "normal", retained);
      const samples = await Promise.all(window.measured_roots.map((rootRef) =>
        readCanonicalRoot(repositoryRoot, baseDirectory, rootRef, binding.target, "normal", retained),
      ));
      const records = [cold, warmup, ...samples];
      if (records.some((record) => record.manifest.target !== window.provider_target)) {
        throw new Error(`${binding.target} window provider does not match retained manifests`);
      }
      assertSameWindow(binding.target, records);
      assertCanonicalInputs(
        binding.target,
        effective.canonical_inputs,
        semanticDeclaredInputs(cold.manifest.declared_inputs),
        roster.get(window.provider_target)?.input_defaults,
      );
      if (effective.command_id && cold.targetSummary.command_id !== effective.command_id) {
        throw new Error(`${binding.target} retained command ID does not match the owner binding`);
      }
      const samplesMs = samples.map((record) =>
        observation(record, binding.target, binding.timing_source, window.provider_target),
      );
      const p50 = median(samplesMs);
      const p90 = nearestRankP90(samplesMs);
      const mad = median(samplesMs.map((sample) => Math.abs(sample - p50)));
      if (mad > p50 * 0.1) {
        throw new Error(samples.length === 5 ? `${binding.target} requires a sixth observation because MAD exceeds ten percent of p50` : `${binding.target} remains unstable after six observations`);
      }
      const executionPolicy = {
        target: binding.target,
        provider_target: window.provider_target,
        graph_digest: cold.manifest.graph_digest,
        workload_digest: cold.targetSummary.workload_digest,
        unit_ids: cold.targetSummary.unit_ids,
        instrumentation: cold.manifest.instrumentation,
        declared_inputs: semanticDeclaredInputs(cold.manifest.declared_inputs),
      };
      return {
        target: binding.target,
        gate: diagnostic ? "diagnostic_only" : effective.gate,
        command_id: cold.targetSummary.command_id,
        measurement_profile_id: effective.measurement_profile_id,
        canonical_inputs: effective.canonical_inputs,
        timing_source: binding.timing_source,
        source_commit: cold.manifest.source_commit,
        source_snapshot_sha256: withoutPrefix(cold.manifest.source_digest),
        system_profile_sha256: withoutPrefix(cold.systemProfileDigest),
        toolchain_profile_sha256: withoutPrefix(cold.manifest.toolchain_digest),
        workload_evidence_profile_sha256: withoutPrefix(cold.targetSummary.workload_digest),
        execution_policy: executionPolicy,
        execution_policy_sha256: digestJSON(executionPolicy),
        ...(effective.allowed_policy_transition
          ? { allowed_policy_transition: effective.allowed_policy_transition }
          : {}),
        sample_provider_target: window.provider_target,
        cold_root: cold.root,
        cold_ms: observation(cold, binding.target, binding.timing_source, window.provider_target),
        warmup_root: warmup.root,
        sample_roots: samples.map((record) => record.root),
        sample_count: samples.length,
        samples_ms: samplesMs,
        p50_ms: p50,
        p90_ms: p90,
        mad_ms: mad,
        timing_accounting: timingAccounting(samples),
      };
    })).then((rows) => rows.sort((left, right) => compareASCII(left.target, right.target)));
  };
  const targets = await buildRows(bindings, windows, false);
  const diagnostics = await buildRows(internalBindings, internalWindows, true);
  const expected = [...roster.keys()];
  if (!sameJSON(targets.map((row) => row.target), expected)) {
    throw new Error(`${role} does not contain the exact ${expected.length}-target public roster`);
  }
  const unused = [...byWindow.keys()].filter((windowID) => !referenced.has(windowID));
  if (unused.length > 0) throw new Error(`${role} contains unused windows ${unused.sort(compareASCII).join(",")}`);
  const totalP50 = targets.reduce((total, row) => total + row.p50_ms, 0);
  const baseline = {
    schema_id: baselineSchemaID,
    status: "qualified",
    timing_authority: "canonical_unit_events",
    sample_rule: "one_cold_one_discarded_warmup_five_warm_sixth_only_if_unstable",
    targets,
    internal_diagnostics: diagnostics,
    source_windows: sourceWindows(targets),
    public_entrypoint_portfolio: {
      target_count: expected.length,
      targets: expected,
      total_p50_ms: totalP50,
    },
    rejected_roots: [...rejectedRoots].sort((left, right) => compareASCII(left.root, right.root)),
  };
  validateSchemaSync(baseline.schema_id, baseline);
  assertBaselineClosure(baseline, surface);
  return baseline;
}

export function assertBaselineClosure(baseline, surface) {
  validateSchemaSync(baselineSchemaID, baseline);
  const roster = performanceRoster(surface);
  const expected = [...roster.keys()];
  const names = baseline.targets.map((row) => row.target);
  if (!sameJSON(names, expected)) throw new Error("baseline target rows do not close the public roster");
  const portfolio = baseline.public_entrypoint_portfolio;
  if (portfolio.target_count !== names.length || !sameJSON(portfolio.targets, names)) {
    throw new Error("baseline roster, row count, and portfolio bindings do not close");
  }
  const total = baseline.targets.reduce((sum, row) => sum + row.p50_ms, 0);
  if (portfolio.total_p50_ms !== total) throw new Error("baseline portfolio sum does not close");
  if (!sameJSON(baseline.source_windows, sourceWindows(baseline.targets))) {
    throw new Error("baseline source-window roster does not close");
  }
  for (const row of [...baseline.targets, ...baseline.internal_diagnostics]) {
    const policy = roster.get(row.target);
    if (policy) {
      if (row.gate !== policy.gate || row.measurement_profile_id !== policy.measurement_profile_id ||
          !sameJSON(row.canonical_inputs, policy.canonical_inputs) || row.allowed_policy_transition !== policy.allowed_policy_transition) {
        throw new Error(`${row.target} baseline differs from the reviewed measurement policy`);
      }
      assertReviewedSuccessor(row.target, row.command_id, policy.command_id);
    }
    const roots = [row.cold_root, row.warmup_root, ...row.sample_roots].map((root) => path.normalize(root));
    if (new Set(roots).size !== roots.length) throw new Error(`${row.target} baseline repeats an observation`);
    const p50 = median(row.samples_ms);
    const p90 = nearestRankP90(row.samples_ms);
    const mad = median(row.samples_ms.map((sample) => Math.abs(sample - p50)));
    if (mad > p50 * 0.1) throw new Error(row.sample_count === 5
      ? `${row.target} requires a sixth observation because MAD exceeds ten percent of p50`
      : `${row.target} remains unstable after six observations`);
    if (
      row.sample_count !== row.samples_ms.length ||
      row.sample_count !== row.sample_roots.length ||
      row.p50_ms !== p50 ||
      row.p90_ms !== p90 ||
      row.mad_ms !== mad ||
      row.execution_policy_sha256 !== digestJSON(row.execution_policy)
    ) {
      throw new Error(`baseline statistics or digests do not close for ${row.target}`);
    }
  }
}

function evaluateGate(referenceRow, candidateRow, before, after) {
  const referenceP50 = median(before);
  const referenceP90 = nearestRankP90(before);
  const referenceMAD = median(before.map((sample) => Math.abs(sample - referenceP50)));
  const candidateP50 = median(after);
  const candidateP90 = nearestRankP90(after);
  const candidateMAD = median(after.map((sample) => Math.abs(sample - candidateP50)));
  const variabilityBand = 3 * Math.max(referenceMAD, candidateMAD, 1);
  const p90Passed = candidateP90 <= referenceP90 + variabilityBand;
  const passed = referenceRow.gate === "required_improvement"
    ? referenceP50 - candidateP50 > variabilityBand && p90Passed
    : p90Passed;
  return {
    passed,
    variabilityBand,
    referenceP50,
    referenceP90,
    referenceMAD,
    candidateP50,
    candidateP90,
    candidateMAD,
  };
}

function assertReviewedSuccessor(target, before, after) {
  if (before === after) return;
  const pattern = /^(cartulary\.harness\.command\.[a-z][a-z0-9_]*\.v)([1-9][0-9]*)$/u;
  const prior = pattern.exec(before);
  const next = pattern.exec(after);
  if (!prior || !next || prior[1] !== next[1] || Number(next[2]) !== Number(prior[2]) + 1) {
    throw new Error(`${target} command ID is not the immediate reviewed successor`);
  }
}

export function compareQualifiedBaselines(reference, candidate, surface) {
  assertBaselineClosure(reference, surface);
  assertBaselineClosure(candidate, surface);
  const referenceRoots = new Set([...reference.targets, ...reference.internal_diagnostics]
    .flatMap((row) => [row.cold_root, row.warmup_root, ...row.sample_roots]).map((root) => path.normalize(root)));
  if ([...candidate.targets, ...candidate.internal_diagnostics].some((row) =>
    [row.cold_root, row.warmup_root, ...row.sample_roots].some((root) => referenceRoots.has(path.normalize(root))))) {
    throw new Error("reference and candidate reuse an observation");
  }
  const before = new Map(reference.targets.map((row) => [row.target, row]));
  const after = new Map(candidate.targets.map((row) => [row.target, row]));
  const rows = [];
  const failures = [];
  for (const target of [...before.keys()]) {
    const referenceRow = before.get(target);
    const candidateRow = after.get(target);
    if (!candidateRow) throw new Error(`candidate omits ${target}`);
    for (const field of [
      "gate",
      "measurement_profile_id",
      "timing_source",
      "system_profile_sha256",
      "toolchain_profile_sha256",
      "workload_evidence_profile_sha256",
    ]) {
      if (referenceRow[field] !== candidateRow[field]) {
        throw new Error(`${target} reference and candidate mismatch ${field}`);
      }
    }
    if (!sameJSON(referenceRow.canonical_inputs, candidateRow.canonical_inputs)) {
      throw new Error(`${target} reference and candidate mismatch canonical inputs`);
    }
    assertReviewedSuccessor(target, referenceRow.command_id, candidateRow.command_id);
    const policyChanged = !sameJSON(referenceRow.execution_policy, candidateRow.execution_policy);
    if (policyChanged) {
      if (
        !referenceRow.allowed_policy_transition ||
        referenceRow.allowed_policy_transition !== candidateRow.allowed_policy_transition
      ) {
        throw new Error(`${target} has an undeclared execution-policy transition`);
      }
    }
    if (referenceRow.sample_count !== candidateRow.sample_count) {
      throw new Error(`${target} reference and candidate sample counts differ`);
    }
    const result = evaluateGate(referenceRow, candidateRow, referenceRow.samples_ms, candidateRow.samples_ms);
    const highMAD = result.referenceMAD > result.referenceP50 * 0.1 ||
      result.candidateMAD > result.candidateP50 * 0.1;
    const leaveOneOutChanges = referenceRow.samples_ms.some((_, index) => {
      const reducedReference = referenceRow.samples_ms.filter((__, sampleIndex) => sampleIndex !== index);
      const reducedCandidate = candidateRow.samples_ms.filter((__, sampleIndex) => sampleIndex !== index);
      return evaluateGate(referenceRow, candidateRow, reducedReference, reducedCandidate).passed !== result.passed;
    });
    if (referenceRow.sample_count === 5 && (highMAD || leaveOneOutChanges)) {
      throw new Error(`${target} requires one matched sixth observation`);
    }
    if (referenceRow.sample_count === 6 && (highMAD || leaveOneOutChanges)) {
      throw new Error(`${target} remains unstable after six observations`);
    }
    const structuralImproved = [
      "critical_path_p50_ms",
      "resource_blocking_p50_ms",
      "process_count_p50",
    ].some((field) => candidateRow.timing_accounting[field] < referenceRow.timing_accounting[field]);
    const passed = result.passed &&
      (referenceRow.gate !== "required_improvement" || structuralImproved);
    if (!passed) failures.push(target);
    rows.push({
      target,
      gate: referenceRow.gate,
      reference_p50_ms: result.referenceP50,
      reference_p90_ms: result.referenceP90,
      reference_mad_ms: result.referenceMAD,
      candidate_p50_ms: result.candidateP50,
      candidate_p90_ms: result.candidateP90,
      candidate_mad_ms: result.candidateMAD,
      variability_band_ms: result.variabilityBand,
      structural_improvement: structuralImproved,
      status: passed ? "pass" : "fail",
    });
  }
  return {
    rows,
    failures,
    portfolio: {
      target_count: rows.length,
      reference_total_p50_ms: reference.public_entrypoint_portfolio.total_p50_ms,
      candidate_total_p50_ms: candidate.public_entrypoint_portfolio.total_p50_ms,
      delta_p50_ms:
        candidate.public_entrypoint_portfolio.total_p50_ms -
        reference.public_entrypoint_portfolio.total_p50_ms,
    },
    rejected_roots: {
      reference: reference.rejected_roots,
      candidate: candidate.rejected_roots,
    },
  };
}

export async function buildFromManifest(repositoryRoot, manifestFile, surface) {
  const manifest = readJSON(manifestFile);
  validateSchemaSync(evidenceSchemaID, manifest);
  const baseDirectory = path.dirname(manifestFile);
  const reference = await buildQualifiedBaseline({
    repositoryRoot,
    baseDirectory,
    surface,
    windows: manifest.reference_windows,
    bindings: manifest.reference_bindings,
    internalWindows: manifest.reference_internal_windows ?? [],
    internalBindings: manifest.reference_internal_bindings ?? [],
    rejectedRoots: manifest.reference_rejected_roots ?? [],
    role: "reference",
  });
  if (manifest.mode === "baseline") return { manifest, reference };
  const candidate = await buildQualifiedBaseline({
    repositoryRoot,
    baseDirectory,
    surface,
    windows: manifest.candidate_windows,
    bindings: manifest.candidate_bindings,
    internalWindows: manifest.candidate_internal_windows ?? [],
    internalBindings: manifest.candidate_internal_bindings ?? [],
    rejectedRoots: manifest.candidate_rejected_roots ?? [],
    role: "candidate",
  });
  return {
    manifest,
    reference,
    candidate,
    comparison: compareQualifiedBaselines(reference, candidate, surface),
  };
}
