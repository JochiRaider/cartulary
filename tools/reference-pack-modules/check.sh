#!/usr/bin/env bash
set -euo pipefail

bash ./tools/harness/readiness/go-toolchain-readiness.sh ensure
qualified_go="$(bash ./tools/harness/readiness/go-toolchain-readiness.sh resolve)"
"${NODE_BIN:?}" tools/harness/readiness/cryptographic-build-cli.mjs verify
module_info="$("$NODE_BIN" --input-type=module -e 'import { independentGoModules, validateIndependentGoModule } from "./tools/harness/test-catalog/go-modules.mjs"; const selected = independentGoModules(process.cwd()).find(entry => entry.root === "referencepacks/" + process.env.MODULE); if (!selected) throw new Error("unregistered module"); validateIndependentGoModule(process.cwd(), selected.root); process.stdout.write([selected.root, selected.module_path, selected.version].join("\t"));')"
IFS=$'\t' read -r module_root module_path module_version <<< "$module_info"
mkdir -p "${GO_TMP_DIR:?}" "${GO_CACHE_DIR:?}" "${GO_MOD_CACHE_DIR:?}" .cartulary/reference-packs/modules
module_work="$(mktemp -d "$GO_TMP_DIR/reference-pack-module.XXXXXXXX")"
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
# The compiler sees only exported source, selected pinned dependencies and a
# private writable cache. Host source and network access are not capabilities.
"$NODE_BIN" tools/reference-pack-modules/isolated-go.mjs module "$module_work/source" "$module_work" "$GO_MOD_CACHE_DIR" "$qualified_go"
module_goroot="$("$GO" env GOROOT)"
"$NODE_BIN" tools/reference-pack-modules/qualify.mjs "$module_work/descriptor.json" "$module_work/packages.json" "$module_work/source" "$GO_MOD_CACHE_DIR" "$module_goroot" "$module_work/cache" "$module_store" "$module_work/isolation.json"
printf 'Independent module qualified: %s; source=%s\n' "$module_path" "$source_sha"
