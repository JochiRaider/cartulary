#!/usr/bin/env bash
set -euo pipefail
bash ./tools/harness/readiness/go-toolchain-readiness.sh ensure
qualified_go="$(bash ./tools/harness/readiness/go-toolchain-readiness.sh resolve)"
mkdir -p "${GO_TMP_DIR:?}"
runtime_work="$(mktemp -d "$GO_TMP_DIR/reference-pack-runtime.XXXXXXXX")"
runtime_closure="$runtime_work/packages.json"
trap 'rm -rf -- "$runtime_work"' EXIT
env CGO_ENABLED=0 GOENV=off GOFLAGS= GOWORK=off GOPROXY=off \
  GOCACHE="${GO_CACHE_DIR:?}" GOMODCACHE="${GO_MOD_CACHE_DIR:?}" GOTMPDIR="$GO_TMP_DIR" \
  "${GO:?}" list -mod=readonly -deps -json ./cmd/... > "$runtime_closure"
"${NODE_BIN:?}" tools/reference-pack-modules/runtime-qualify.mjs "$runtime_closure" > "$runtime_work/static.json"
env CGO_ENABLED=0 GOENV=off GOFLAGS= GOWORK=off GOPROXY=off \
  GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
  "$GO" list -mod=readonly -deps -json ./tools/contractgen > "$runtime_work/generator-packages.json"
"$NODE_BIN" tools/reference-pack-modules/application-isolation.mjs "$runtime_closure" "$runtime_work" "$qualified_go" "$GO_MOD_CACHE_DIR" "${REFERENCE_PACK_GENERATION_PROFILE:?}"
cat "$runtime_work/static.json"
