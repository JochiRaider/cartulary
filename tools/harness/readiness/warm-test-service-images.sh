#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"

write_stamp() {
  local stamp_file="$1"
  local bin_path="$2"
  local tmp_file
  tmp_file="${stamp_file}.tmp"
  mkdir -p "$(dirname "$stamp_file")"
  {
    printf 'schema_id=cartulary.test_service_images_ready.v1\n'
    printf 'testservices_bin=%s\n' "$bin_path"
    printf 'created_at=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  } >"$tmp_file"
  mv "$tmp_file" "$stamp_file"
}

if [[ "${1:-}" == "--write-stamp" ]]; then
  if [[ "$#" -ne 3 ]]; then
    echo "usage: warm-test-service-images.sh --write-stamp <stamp> <testservices-bin>" >&2
    exit 2
  fi
  write_stamp "$2" "$3"
  exit 0
fi

if [[ "${1:-}" == "--warm-and-stamp" ]]; then
  if [[ "$#" -ne 3 ]]; then
    echo "usage: warm-test-service-images.sh --warm-and-stamp <stamp> <testservices-bin>" >&2
    exit 2
  fi
  "$3" warm-images
  write_stamp "$2" "$3"
  exit 0
fi

if [[ "$#" -gt 1 ]]; then
  echo "usage: warm-test-service-images.sh [testservices-bin]" >&2
  exit 2
fi

testservices_bin="${1:-${CARTULARY_TEST_SERVICES_BIN:-$ROOT_DIR/tmp/toolbin/cartulary-test-services}}"
stamp="${CARTULARY_TEST_SERVICE_IMAGES_STAMP:-$ROOT_DIR/tmp/test-service-images/warm.stamp}"
cache_dir="${CARTULARY_READINESS_CACHE_DIR:-$ROOT_DIR/.cache/cartulary/readiness}"

images_present() {
  local image
  local images_file
  local missing=0
  if ! command -v docker >/dev/null 2>&1; then
    return 1
  fi
  images_file="$(mktemp)"
  if ! "$testservices_bin" images >"$images_file"; then
    rm -f "$images_file"
    return 1
  fi
  if [[ ! -s "$images_file" ]]; then
    rm -f "$images_file"
    return 1
  fi
  while IFS= read -r image; do
    [[ -n "$image" ]] || continue
    if ! docker image inspect "$image" >/dev/null 2>&1; then
      missing=1
      break
    fi
  done <"$images_file"
  rm -f "$images_file"
  if [[ "$missing" -ne 0 ]]; then
    return 1
  fi
}

if [[ -f "$stamp" ]] && ! images_present; then
  rm -f "$stamp"
fi

"$ROOT_DIR/tools/harness/readiness/cache-artifact.sh" \
  \
  --scope readiness \
  --profile test-service-images \
  --cache-dir "$cache_dir" \
  --disable-env CARTULARY_READINESS_DISABLE_CACHE \
  --force-env CARTULARY_FORCE_REINSTALL \
  --input "$testservices_bin" \
  --input "$ROOT_DIR/tools/testservices/main.go" \
  --input "$ROOT_DIR/internal/testutil/pgtest/pgtest.go" \
  --input "$ROOT_DIR/internal/testutil/s3test/s3test.go" \
  --input "$ROOT_DIR/tools/harness/readiness/warm-test-service-images.sh" \
  --input "$ROOT_DIR/tools/harness/readiness/cache-artifact.sh" \
  --input "$ROOT_DIR/tools/toolchain_pins.json" \
  --output "$stamp" \
  --key "testservices_bin=$testservices_bin" \
  -- \
  "$ROOT_DIR/tools/harness/readiness/warm-test-service-images.sh" --warm-and-stamp "$stamp" "$testservices_bin"
