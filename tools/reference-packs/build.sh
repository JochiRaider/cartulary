#!/usr/bin/env bash
set -euo pipefail

bash ./tools/harness/readiness/go-toolchain-readiness.sh ensure
"${NODE_BIN:?}" tools/harness/readiness/cryptographic-build-cli.mjs verify
mkdir -p build/bin "${GO_CACHE_DIR:?}" "${GO_MOD_CACHE_DIR:?}" "${GO_TMP_DIR:?}"
producer_work="$(mktemp -d build/.attack-producer.XXXXXXXX)"
trap 'rm -rf -- "$producer_work"' EXIT
for pass in first second; do
  env CGO_ENABLED=0 GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
    "${GO:?}" build -mod=readonly -trimpath -buildvcs=false \
    -o "$producer_work/$pass" ./tools/reference-packs
done
cmp -- "$producer_work/first" "$producer_work/second"
read -r producer_sha _ < <(sha256sum "$producer_work/first")
producer_store=".cartulary/reference-packs/builders/$producer_sha"
mkdir -p "$producer_store"
if [[ -e "$producer_store/reference-pack-producer" ]]; then
  cmp -- "$producer_work/first" "$producer_store/reference-pack-producer"
else
  ln -- "$producer_work/first" "$producer_store/reference-pack-producer"
fi
"$producer_store/reference-pack-producer" --action identity > "$producer_work/identity.json"
if [[ -e "$producer_store/identity.json" ]]; then
  cmp -- "$producer_work/identity.json" "$producer_store/identity.json"
else
  ln -- "$producer_work/identity.json" "$producer_store/identity.json"
fi
printf '{"schema_id":"cartulary.attack_builder_reproduction.v1","first_sha256":"%s","second_sha256":"%s","byte_identical":true}\n' \
  "$producer_sha" "$producer_sha" > "$producer_work/reproducibility.json"
if [[ -e "$producer_store/reproducibility.json" ]]; then
  cmp -- "$producer_work/reproducibility.json" "$producer_store/reproducibility.json"
else
  ln -- "$producer_work/reproducibility.json" "$producer_store/reproducibility.json"
fi
install -m 0755 "$producer_work/first" build/bin/reference-pack-producer
printf 'ATT&CK source producer reproduced: %s; evidence=%s\n' "$producer_sha" "$producer_store"
