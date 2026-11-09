#!/usr/bin/env bash
set -euo pipefail

bash ./tools/harness/readiness/go-toolchain-readiness.sh ensure
qualified_go="$(bash ./tools/harness/readiness/go-toolchain-readiness.sh resolve)"
"${NODE_BIN:?}" tools/harness/readiness/cryptographic-build-cli.mjs verify
module_info="$("$NODE_BIN" --input-type=module -e 'import { independentGoModules, validateIndependentGoModule } from "./tools/harness/test-catalog/go-modules.mjs"; const selected = independentGoModules(process.cwd()).find(entry => entry.root === `referencepacks/${process.env.PRODUCER}`); if (!selected) throw new Error("unregistered module"); validateIndependentGoModule(process.cwd(), selected.root); process.stdout.write([selected.root, selected.module_path, selected.version].join("\t"));')"
IFS=$'\t' read -r module_root module_path module_version <<< "$module_info"
mkdir -p "${GO_TMP_DIR:?}" "${GO_CACHE_DIR:?}" "${GO_MOD_CACHE_DIR:?}" .cartulary/reference-packs/modules
module_work="$(mktemp -d "$GO_TMP_DIR/reference-pack-builder.XXXXXXXX")"
trap 'rm -rf -- "$module_work"' EXIT
# The helper itself is standard-library-only and is not included in exports.
env CGO_ENABLED=0 GOWORK=off GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
  "${GO:?}" build -mod=readonly -trimpath -buildvcs=false -o "$module_work/archive" ./tools/reference-pack-modules
module_store="$(pwd)/.cartulary/reference-packs/modules"
"$module_work/archive" -root "$module_root" -module "$module_path" -version "$module_version" -out "$module_store" > "$module_work/descriptor.json"
archive_info="$("$NODE_BIN" --input-type=module -e 'import { readFileSync } from "node:fs"; const d = JSON.parse(readFileSync(process.argv[1], "utf8")); process.stdout.write([d.archive_path, d.version, d.source_sha256].join("\t"));' "$module_work/descriptor.json")"
IFS=$'\t' read -r archive_path archive_version source_sha <<< "$archive_info"
mkdir "$module_work/source"
"$module_work/archive" -action extract -archive "$archive_path" -module "$module_path" -version "$archive_version" -out "$module_work/source"
mkdir -p build/bin
"$NODE_BIN" tools/reference-pack-modules/isolated-go.mjs producer "$module_work/source" "$module_work" "$GO_MOD_CACHE_DIR" "$qualified_go" "$PRODUCER"
cmp -- "$module_work/binaries/first" "$module_work/binaries/second"
read -r producer_sha _ < <(sha256sum "$module_work/binaries/first")
producer_store=".cartulary/reference-packs/builders/$producer_sha"
mkdir -p "$producer_store"
if [[ -e "$producer_store/reference-pack-producer" ]]; then
  cmp -- "$module_work/binaries/first" "$producer_store/reference-pack-producer"
else
  ln -- "$module_work/binaries/first" "$producer_store/reference-pack-producer"
fi
"$producer_store/reference-pack-producer" --action identity > "$module_work/identity.json"
if [[ -e "$producer_store/identity.json" ]]; then
  cmp -- "$module_work/identity.json" "$producer_store/identity.json"
else
  ln -- "$module_work/identity.json" "$producer_store/identity.json"
fi
printf '{"schema_id":"cartulary.reference_pack_builder_reproduction.v1","first_sha256":"%s","second_sha256":"%s","byte_identical":true}\n' \
  "$producer_sha" "$producer_sha" > "$module_work/reproducibility.json"
if [[ -e "$producer_store/reproducibility.json" ]]; then
  cmp -- "$module_work/reproducibility.json" "$producer_store/reproducibility.json"
else
  ln -- "$module_work/reproducibility.json" "$producer_store/reproducibility.json"
fi
install -m 0755 "$module_work/binaries/first" build/bin/reference-pack-${PRODUCER}
printf 'Independent reference pack producer reproduced: %s; evidence=%s\n' "$producer_sha" "$producer_store"
