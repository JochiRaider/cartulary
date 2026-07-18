#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
SCRATCH_INPUT_MANIFEST="${GENERATE_DRIFT_SCRATCH_INPUT_MANIFEST:-tools/generate_drift_scratch_inputs.json}"

cd "$ROOT_DIR"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "generated artifact drift check must run inside a git work tree" >&2
  exit 1
fi

sqlc_bin="${SQLC_BIN:-$ROOT_DIR/tmp/toolbin/sqlc-v1.30.0}"
if [[ "$sqlc_bin" != /* ]]; then
  sqlc_bin="$ROOT_DIR/$sqlc_bin"
fi
if [[ ! -x "$sqlc_bin" ]]; then
  echo "generate-drift requires an executable SQLC_BIN at $sqlc_bin" >&2
  echo "run make codegen-toolchain before generate-drift or set SQLC_BIN to a ready sqlc binary" >&2
  exit 1
fi

mkdir -p "$ROOT_DIR/tmp"
scratch="$(mktemp -d "$ROOT_DIR/tmp/generate-drift.XXXXXX")"

cleanup() {
  rm -rf "$scratch"
}
trap cleanup EXIT

copy_path() {
  local source="$1"
  local destination="$scratch/$source"

  if [[ ! -e "$ROOT_DIR/$source" ]]; then
    echo "required generate-drift input missing: $source" >&2
    exit 1
  fi

  if [[ -d "$ROOT_DIR/$source" && ! -L "$ROOT_DIR/$source" ]]; then
    mkdir -p "$destination"
    cp -a "$ROOT_DIR/$source/." "$destination/"
  else
    mkdir -p "$(dirname "$destination")"
    cp -a "$ROOT_DIR/$source" "$destination"
  fi
}

manifest_values() {
  local key="$1"
  local node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
  "$node_bin" - "$ROOT_DIR/$SCRATCH_INPUT_MANIFEST" "$key" <<'NODE'
const fs = require("node:fs");
const [manifestPath, key] = process.argv.slice(2);
let manifest;
try {
  manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
} catch (error) {
  console.error(`generate-drift scratch input manifest unreadable: ${error.message}`);
  process.exit(11);
}
const values = manifest[key];
if (!Array.isArray(values) || values.some((entry) => typeof entry !== "string" || entry === "")) {
  console.error(`generate-drift scratch input manifest field ${key} must be a non-empty string array`);
  process.exit(11);
}
for (const value of values) {
  console.log(value);
}
NODE
}

catalog_selector_inputs() {
  local node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
  "$node_bin" - "$ROOT_DIR" <<'NODE'
const fs = require("node:fs");
const path = require("node:path");

const [repoRoot] = process.argv.slice(2);
const ownerRegistry = JSON.parse(
  fs.readFileSync(path.join(repoRoot, "tools/test_catalog_owner.json"), "utf8"),
);
const inputs = new Set();
for (const owner of ownerRegistry.owners ?? []) {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(repoRoot, owner.manifest_path), "utf8"),
  );
  for (const row of manifest.rows ?? []) {
    for (const candidate of [row.selector?.file, row.selector?.package]) {
      if (typeof candidate !== "string" || candidate.trim() === "") continue;
      const normalized = candidate.trim().replace(/^\.\//u, "");
      if (
        path.isAbsolute(normalized) ||
        normalized === ".." ||
        normalized.startsWith("../") ||
        normalized.includes("/../")
      ) {
        console.error(`unsafe catalog selector input ${candidate}`);
        process.exit(11);
      }
      inputs.add(normalized);
    }
  }
}
for (const input of [...inputs].sort((left, right) => left.localeCompare(right))) {
  console.log(input);
}
NODE
}

task_surface_backing_inputs() {
  local node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
  "$node_bin" - "$ROOT_DIR" <<'NODE'
const fs = require("node:fs");
const path = require("node:path");

const [repoRoot] = process.argv.slice(2);
const owner = JSON.parse(
  fs.readFileSync(path.join(repoRoot, "tools/task_surface_owner.json"), "utf8"),
);
const inputs = new Set();
for (const entry of [...(owner.targets ?? []), ...(owner.harness_checks ?? [])]) {
  for (const candidate of entry.backing_scripts ?? []) {
    if (typeof candidate !== "string" || candidate.trim() === "") continue;
    const normalized = candidate.trim().replace(/^\.\//u, "");
    if (
      path.isAbsolute(normalized) ||
      normalized === ".." ||
      normalized.startsWith("../") ||
      normalized.includes("/../")
    ) {
      console.error(`unsafe task-surface backing input ${candidate}`);
      process.exit(11);
    }
    inputs.add(normalized);
  }
}
for (const input of [...inputs].sort((left, right) => left.localeCompare(right))) {
  console.log(input);
}
NODE
}

copy_required_make_includes() {
  local directive include_path rest

  while read -r directive rest; do
    if [[ "$directive" != "include" ]]; then
      continue
    fi

    for include_path in $rest; do
      if [[ "$include_path" == \#* ]]; then
        break
      fi
      if [[ "$include_path" == /* || "$include_path" == *'$'* ]]; then
        continue
      fi
      copy_path "$include_path"
    done
  done <"$ROOT_DIR/Makefile"
}

copy_path "$SCRATCH_INPUT_MANIFEST"

mapfile -t scratch_copy_paths < <(manifest_values copy_paths)
mapfile -t generated_paths < <(manifest_values generated_paths)
mapfile -t scratch_placeholder_dirs < <(manifest_values placeholder_dirs)

for input in "${scratch_copy_paths[@]}"; do
  copy_path "$input"
done
copy_required_make_includes

mapfile -t selector_inputs < <(catalog_selector_inputs)
for input in "${selector_inputs[@]}"; do
  copy_path "$input"
done

mapfile -t task_surface_inputs < <(task_surface_backing_inputs)
for input in "${task_surface_inputs[@]}"; do
  copy_path "$input"
done

for placeholder_dir in "${scratch_placeholder_dirs[@]}"; do
  mkdir -p "$scratch/$placeholder_dir"
done

make -C "$scratch" --no-print-directory generate-artifacts \
	CARTULARY_TEST_TARGET=generate-artifacts \
	SQLC_BIN="$sqlc_bin" \
	GO="${GO:-go}" \
	GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}" \
	GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}" \
	NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-$ROOT_DIR/tmp/node-runtime}" \
	CARTULARY_NODE_ARCHIVE_DIR="${CARTULARY_NODE_ARCHIVE_DIR:-$ROOT_DIR/tmp/node-archives}" \
	NODE_BIN="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" \
	PNPM="${PNPM:-$ROOT_DIR/tmp/node-runtime/bin/pnpm}"

drift=0
for generated_path in "${generated_paths[@]}"; do
  if ! diff -ruN "$ROOT_DIR/$generated_path" "$scratch/$generated_path" >/dev/null; then
    drift=1
    break
  fi
done

if [[ "$drift" -ne 0 ]]; then
  echo "generated artifact drift detected after make generate-artifacts" >&2
  echo "diff excerpt (first 200 lines):" >&2
  for generated_path in "${generated_paths[@]}"; do
    diff -ruN \
      --label "$generated_path" \
      --label "regenerated $generated_path" \
      "$ROOT_DIR/$generated_path" \
      "$scratch/$generated_path" || true
  done | sed -n '1,200p' >&2
  exit 1
fi
