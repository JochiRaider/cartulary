#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
GENERATED_PATHS=(
  "internal/gen/contracts"
  "internal/gen/sql"
  "packages/protocol-ts/src/generated"
)
SCRATCH_CONTROL_INPUTS=(
  "Makefile"
  "sqlc.yaml"
  "go.mod"
  "go.sum"
  "scripts/list-build-inputs.sh"
  "scripts/lib"
  "tools/execution_topology_manifest.json"
  "tools/scheduler_resource_registry.json"
  "tools/service_backed_make_target_duration_baselines.json"
  "tools/task_surface_manifest.json"
  "tools/task_surface.generated.mk"
  "tools/check_schedule_manifest.json"
  "tools/service_backed_schedule_manifest.json"
  "tools/browser_e2e_batch_manifest.json"
  "tools/contractgen"
)
SCRATCH_CODEGEN_INPUTS=(
  "contracts"
  "db/migrations"
  "db/queries"
)
SCRATCH_PLACEHOLDER_DIRS=(
  "apps/web"
  "cmd/migrate"
  "cmd/server"
  "internal/app"
  "internal/modules"
  "internal/platform"
  "internal/platform/postgres"
  "internal/testutil/pgtest"
  "internal/testutil/s3test"
  "internal/testutil/suiteservices"
  "packages"
  "tools/testservices"
)

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

  mkdir -p "$(dirname "$destination")"
  cp -a "$ROOT_DIR/$source" "$destination"
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

for input in "${SCRATCH_CONTROL_INPUTS[@]}" "${SCRATCH_CODEGEN_INPUTS[@]}" "${GENERATED_PATHS[@]}"; do
  copy_path "$input"
done
copy_required_make_includes

for placeholder_dir in "${SCRATCH_PLACEHOLDER_DIRS[@]}"; do
  mkdir -p "$scratch/$placeholder_dir"
done

make -C "$scratch" --no-print-directory generate-artifacts \
  SQLC_BIN="$sqlc_bin" \
  GO="${GO:-go}" \
  GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}" \
  GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}" \
  NODE_BIN="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" \
  PNPM="${PNPM:-$ROOT_DIR/tmp/node-runtime/bin/pnpm}"

drift=0
for generated_path in "${GENERATED_PATHS[@]}"; do
  if ! diff -ruN "$ROOT_DIR/$generated_path" "$scratch/$generated_path" >/dev/null; then
    drift=1
    break
  fi
done

if [[ "$drift" -ne 0 ]]; then
  echo "generated artifact drift detected after make generate-artifacts" >&2
  echo "diff excerpt (first 200 lines):" >&2
  for generated_path in "${GENERATED_PATHS[@]}"; do
    diff -ruN \
      --label "$generated_path" \
      --label "regenerated $generated_path" \
      "$ROOT_DIR/$generated_path" \
      "$scratch/$generated_path" || true
  done | sed -n '1,200p' >&2
  exit 1
fi
