#!/usr/bin/env node

import { createHash } from "node:crypto";
import { existsSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../..");

const benchmarkManifestSchemaID = "cartulary.benchmark_manifest.v1";
const benchmarkProfileID = "cartulary.perf.desktop_ref.v1";
const defaultBenchmarkManifestPath = path.join(
  repoRoot,
  ".cartulary",
  "benchmark",
  "benchmark_manifest.json",
);

const exactProfileValues = {
  benchmark_manifest_schema_id: benchmarkManifestSchemaID,
  benchmark_profile_id: benchmarkProfileID,
  browser_engine: "chromium",
  browser_build: "134.0.6998.35",
  browser_mode: "headed",
  browser_extensions: "none",
  browser_viewport_css_px: "1440x900",
  browser_device_scale_factor: 1,
  browser_zoom_percent: 100,
  client_runner_id: "aws.ec2.c7i.2xlarge",
  client_os_image_id: "cartulary.bench.ubuntu_24_04_client.2026q1",
  client_reserved_vcpu: 8,
  client_reserved_memory_gib: 16,
  client_power_mode: "performance",
  app_runner_id: "aws.ec2.c7i.2xlarge",
  app_os_image_id: "cartulary.bench.ubuntu_24_04_app.2026q1",
  app_reserved_vcpu: 8,
  app_reserved_memory_gib: 16,
  postgres_runner_id: "aws.ec2.i4i.2xlarge",
  postgres_os_image_id: "cartulary.bench.ubuntu_24_04_postgres.2026q1",
  postgres_reserved_vcpu: 8,
  postgres_reserved_memory_gib: 32,
  postgres_storage_class: "instance_store_nvme",
  object_store_runner_id: "aws.ec2.c7i.xlarge",
  object_store_os_image_id: "cartulary.bench.ubuntu_24_04_object.2026q1",
  object_store_reserved_vcpu: 4,
  object_store_reserved_memory_gib: 8,
  object_store_storage_class: "gp3_ssd",
  client_to_app_link_mbps: 1000,
  client_to_app_rtt_ms_max: 2,
  client_to_app_loss_percent: 0,
  client_to_app_jitter_ms_max: 1,
  app_to_postgres_rtt_ms_max: 1,
  app_to_object_store_rtt_ms_max: 1,
  traffic_trace_id: "cartulary.perf.live_updates_25sessions.v1",
  seed: 20260405,
  warmup_passes: 1,
  authenticated_session_state: "complete",
  incident_open_state: "open",
  surface_warm_state: "loaded",
  benchmark_harness_id: "cartulary.bench.harness.playwright.v1",
  benchmark_harness_version: "2026.04.0",
};

const requiredFields = [
  ...Object.keys(exactProfileValues),
  "criterion_ids",
  "measurement_predicate_ids",
  "fixture_ids",
  "run_started_at",
  "run_completed_at",
  "sample_count",
  "artifact_bundle_sha256",
  "security_controls_state",
];

const p95MeasurementPredicates = new Set([
  "perf.selection_change.v1",
  "perf.focus_change.v1",
  "perf.typing_ack.v1",
  "perf.timeline_blank_row_create.v1",
  "perf.view_change.first_useful_viewport.v1",
  "perf.view_change.stable_viewport.v1",
  "perf.evidence_inspector.metadata_shell.v1",
]);

const requiredSecurityControls = [
  "authentication",
  "session_handling",
  "csrf_protection",
  "sanitization",
  "safe_preview_restrictions",
  "integrity_checks",
];

function main() {
  const manifestPath = resolveManifestPath();
  if (!existsSync(manifestPath)) {
    if (manifestPath === defaultBenchmarkManifestPath) {
      console.log("benchmark claim manifest absent: no claim-bearing benchmark publication requested");
      return;
    }
    throw new Error(`benchmark manifest missing: ${manifestPath}`);
  }
  const manifest = readJSON(manifestPath);
  const manifestDir = path.dirname(manifestPath);
  const errors = validateBenchmarkManifest(manifest, manifestDir);

  if (errors.length > 0) {
    for (const error of errors) {
      console.error(`benchmark claim manifest invalid: ${error}`);
    }
    process.exit(1);
  }

  console.log(`benchmark claim manifest valid: ${path.relative(repoRoot, manifestPath)}`);
}

function resolveManifestPath() {
  const configured =
    process.argv[2] ??
    process.env.BENCHMARK_MANIFEST ??
    defaultBenchmarkManifestPath;
  return path.resolve(path.isAbsolute(configured) ? configured : path.join(repoRoot, configured));
}

function readJSON(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch (error) {
    throw new Error(`failed to read JSON ${file}: ${error.message}`);
  }
}

function validateBenchmarkManifest(manifest, manifestDir) {
  const errors = [];

  if (manifest.claim_bearing === false) {
    errors.push("claim_bearing=false is ordinary measurement metadata, not a benchmark claim");
  }

  for (const field of requiredFields) {
    if (!Object.hasOwn(manifest, field)) {
      errors.push(`missing required field ${field}`);
    }
  }

  for (const [field, expected] of Object.entries(exactProfileValues)) {
    if (Object.hasOwn(manifest, field) && manifest[field] !== expected) {
      errors.push(`${field} must equal ${JSON.stringify(expected)}, got ${JSON.stringify(manifest[field])}`);
    }
  }

  for (const field of ["criterion_ids", "measurement_predicate_ids", "fixture_ids"]) {
    if (!Array.isArray(manifest[field]) || manifest[field].length === 0) {
      errors.push(`${field} must be a non-empty array`);
    } else if (manifest[field].some((value) => typeof value !== "string" || value.trim() === "")) {
      errors.push(`${field} must contain only non-empty strings`);
    }
  }

  if (!isISOTime(manifest.run_started_at)) {
    errors.push("run_started_at must be an ISO timestamp");
  }
  if (!isISOTime(manifest.run_completed_at)) {
    errors.push("run_completed_at must be an ISO timestamp");
  }
  if (isISOTime(manifest.run_started_at) && isISOTime(manifest.run_completed_at)) {
    if (Date.parse(manifest.run_completed_at) < Date.parse(manifest.run_started_at)) {
      errors.push("run_completed_at must not be before run_started_at");
    }
  }

  validateSampleCounts(manifest, errors);
  validateSecurityControls(manifest, errors);
  validateArtifactBundleHash(manifest, manifestDir, errors);

  return errors;
}

function isISOTime(value) {
  return typeof value === "string" && Number.isFinite(Date.parse(value));
}

function validateSampleCounts(manifest, errors) {
  const sampleCount = manifest.sample_count;
  if (!Number.isInteger(sampleCount) || sampleCount < 1) {
    errors.push("sample_count must be a positive integer");
  }

  const sampleCounts =
    manifest.sample_counts !== undefined && typeof manifest.sample_counts === "object"
      ? manifest.sample_counts
      : {};

  for (const predicate of manifest.measurement_predicate_ids ?? []) {
    if (!p95MeasurementPredicates.has(predicate)) {
      continue;
    }
    const predicateSampleCount =
      Number.isInteger(sampleCounts[predicate]) ? sampleCounts[predicate] : sampleCount;
    if (!Number.isInteger(predicateSampleCount) || predicateSampleCount < 100) {
      errors.push(`${predicate} requires at least 100 completed operations for p95 publication`);
    }
  }
}

function validateSecurityControls(manifest, errors) {
  const controls = manifest.security_controls_state;
  if (controls === null || typeof controls !== "object" || Array.isArray(controls)) {
    errors.push("security_controls_state must be an object");
    return;
  }

  for (const control of requiredSecurityControls) {
    if (!Object.hasOwn(controls, control)) {
      errors.push(`security_controls_state.${control} is missing`);
      continue;
    }
    if (!isEnabledControlValue(controls[control])) {
      errors.push(`security_controls_state.${control} must be enabled`);
    }
  }
}

function isEnabledControlValue(value) {
  return value === true || value === "enabled";
}

function validateArtifactBundleHash(manifest, manifestDir, errors) {
  if (typeof manifest.artifact_bundle_sha256 !== "string" || !/^[a-f0-9]{64}$/.test(manifest.artifact_bundle_sha256)) {
    errors.push("artifact_bundle_sha256 must be a lowercase SHA-256 hex digest");
    return;
  }

  if (typeof manifest.artifact_bundle_path === "string" && manifest.artifact_bundle_path.trim() !== "") {
    const bundlePath = resolveRelativePath(manifest.artifact_bundle_path, manifestDir);
    if (!existsSync(bundlePath) || !statSync(bundlePath).isFile()) {
      errors.push(`artifact_bundle_path is not a regular file: ${manifest.artifact_bundle_path}`);
      return;
    }
    const digest = createHash("sha256").update(readFileSync(bundlePath)).digest("hex");
    if (digest !== manifest.artifact_bundle_sha256) {
      errors.push(`artifact_bundle_sha256 mismatch for ${manifest.artifact_bundle_path}`);
    }
    return;
  }

  if (
    typeof manifest.artifact_bundle_metadata_path === "string" &&
    manifest.artifact_bundle_metadata_path.trim() !== ""
  ) {
    const metadataPath = resolveRelativePath(manifest.artifact_bundle_metadata_path, manifestDir);
    if (!existsSync(metadataPath) || !statSync(metadataPath).isFile()) {
      errors.push(`artifact_bundle_metadata_path is not a regular file: ${manifest.artifact_bundle_metadata_path}`);
      return;
    }
    const metadata = readJSON(metadataPath);
    const metadataDigest =
      metadata.artifact_bundle_sha256 ?? metadata.sha256 ?? metadata.digest_sha256 ?? null;
    if (metadataDigest !== manifest.artifact_bundle_sha256) {
      errors.push(`artifact bundle metadata hash mismatch for ${manifest.artifact_bundle_metadata_path}`);
    }
    return;
  }

  errors.push("artifact_bundle_path or artifact_bundle_metadata_path is required");
}

function resolveRelativePath(value, baseDir) {
  if (path.isAbsolute(value)) {
    return value;
  }
  const manifestRelative = path.join(baseDir, value);
  if (existsSync(manifestRelative)) {
    return manifestRelative;
  }
  return path.join(repoRoot, value);
}

try {
  main();
} catch (error) {
  console.error(`benchmark claim check failed: ${error.message}`);
  process.exit(1);
}
