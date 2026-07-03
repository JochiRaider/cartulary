#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
CHECKER="$ROOT_DIR/tools/release-evidence/check-benchmark-claim.mjs"
cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
  done
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_passes() {
  local label="$1"
  shift

  local output
  if ! output="$("$@" 2>&1)"; then
    fail "$label: expected success, got output: $output"
  fi
  printf '%s' "$output"
}

assert_fails() {
  local label="$1"
  shift

  local output
  local status
  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected failure"
  fi
  printf '%s' "$output"
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/benchmark-claim-check.XXXXXX")"
cleanup_paths+=("$tmp_dir")
artifact="$tmp_dir/artifact-bundle.tgz"
printf '%s\n' "benchmark artifact bundle" >"$artifact"
artifact_hash="$("$NODE_BIN" - "$artifact" <<'EOF'
const { createHash } = require("node:crypto");
const { readFileSync } = require("node:fs");
process.stdout.write(createHash("sha256").update(readFileSync(process.argv[2])).digest("hex"));
EOF
)"
valid_manifest="$tmp_dir/benchmark_manifest.json"

default_manifest="$ROOT_DIR/.cartulary/benchmark/benchmark_manifest.json"
if [[ ! -e "$default_manifest" ]]; then
  default_absent_output="$(assert_passes "default absent benchmark manifest" "$NODE_BIN" "$CHECKER")"
  assert_contains "$default_absent_output" "no claim-bearing benchmark publication requested" "default absent benchmark output"
fi

custom_missing="$tmp_dir/missing-benchmark-manifest.json"
custom_missing_output="$(assert_fails "custom missing benchmark manifest" "$NODE_BIN" "$CHECKER" "$custom_missing")"
assert_contains "$custom_missing_output" "benchmark manifest missing" "custom missing benchmark output"

"$NODE_BIN" - "$valid_manifest" "$artifact" "$artifact_hash" <<'EOF'
const { writeFileSync } = require("node:fs");
const [manifestPath, artifactPath, artifactHash] = process.argv.slice(2);
writeFileSync(
  manifestPath,
  `${JSON.stringify(
    {
      benchmark_manifest_schema_id: "cartulary.benchmark_manifest.v1",
      benchmark_profile_id: "cartulary.perf.desktop_ref.v1",
      criterion_ids: ["AC-043"],
      measurement_predicate_ids: ["perf.typing_ack.v1"],
      fixture_ids: ["fixture_a"],
      traffic_trace_id: "cartulary.perf.live_updates_25sessions.v1",
      seed: 20260405,
      warmup_passes: 1,
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
      authenticated_session_state: "complete",
      incident_open_state: "open",
      surface_warm_state: "loaded",
      benchmark_harness_id: "cartulary.bench.harness.playwright.v1",
      benchmark_harness_version: "2026.04.0",
      run_started_at: "2026-04-24T12:00:00Z",
      run_completed_at: "2026-04-24T12:30:00Z",
      sample_count: 100,
      artifact_bundle_sha256: artifactHash,
      artifact_bundle_path: artifactPath,
      security_controls_state: {
        authentication: true,
        session_handling: true,
        csrf_protection: true,
        sanitization: true,
        safe_preview_restrictions: true,
        integrity_checks: true
      }
    },
    null,
    2,
  )}\n`,
);
EOF

valid_output="$(assert_passes "valid benchmark manifest" "$NODE_BIN" "$CHECKER" "$valid_manifest")"
assert_contains "$valid_output" "benchmark claim manifest valid" "valid benchmark manifest output"

mutate_manifest() {
  local output="$1"
  local expression="$2"
  "$NODE_BIN" - "$valid_manifest" "$output" "$expression" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const [input, output, expression] = process.argv.slice(2);
const manifest = JSON.parse(readFileSync(input, "utf8"));
Function("manifest", expression)(manifest);
writeFileSync(output, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
}

missing_profile="$tmp_dir/missing-profile.json"
mutate_manifest "$missing_profile" 'delete manifest.benchmark_profile_id;'
missing_profile_output="$(assert_fails "missing benchmark profile" "$NODE_BIN" "$CHECKER" "$missing_profile")"
assert_contains "$missing_profile_output" "missing required field benchmark_profile_id" "missing benchmark profile output"

low_samples="$tmp_dir/low-samples.json"
mutate_manifest "$low_samples" 'manifest.sample_count = 99;'
low_samples_output="$(assert_fails "low sample count" "$NODE_BIN" "$CHECKER" "$low_samples")"
assert_contains "$low_samples_output" "requires at least 100 completed operations" "low sample count output"

headless="$tmp_dir/headless.json"
mutate_manifest "$headless" 'manifest.browser_mode = "headless";'
headless_output="$(assert_fails "headless benchmark claim" "$NODE_BIN" "$CHECKER" "$headless")"
assert_contains "$headless_output" 'browser_mode must equal "headed"' "headless benchmark output"

disabled_security="$tmp_dir/disabled-security.json"
mutate_manifest "$disabled_security" 'manifest.security_controls_state.csrf_protection = false;'
disabled_security_output="$(assert_fails "disabled security control" "$NODE_BIN" "$CHECKER" "$disabled_security")"
assert_contains "$disabled_security_output" "security_controls_state.csrf_protection must be enabled" "disabled security output"

wrong_hash="$tmp_dir/wrong-hash.json"
mutate_manifest "$wrong_hash" 'manifest.artifact_bundle_sha256 = "0".repeat(64);'
wrong_hash_output="$(assert_fails "wrong artifact hash" "$NODE_BIN" "$CHECKER" "$wrong_hash")"
assert_contains "$wrong_hash_output" "artifact_bundle_sha256 mismatch" "wrong artifact hash output"

ordinary_measurement="$tmp_dir/ordinary-measurement.json"
mutate_manifest "$ordinary_measurement" 'manifest.claim_bearing = false;'
ordinary_output="$(assert_fails "ordinary measurement cannot pass claim check" "$NODE_BIN" "$CHECKER" "$ordinary_measurement")"
assert_contains "$ordinary_output" "ordinary measurement metadata" "ordinary measurement output"
