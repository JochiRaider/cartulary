#!/usr/bin/env bash
set -euo pipefail

# Explicit setup owns module acquisition; review only validates this state.
if [[ "${CARTULARY_PREPARATION_POLICY:-}" == installed_only ]]; then
  exit 2
fi
if [[ "${1:-}" == --all ]]; then
  bash "$0" --analysis
  exec bash "$0"
fi
mkdir -p "${GO_CACHE_DIR:?}" "${GO_MOD_CACHE_DIR:?}" "${GO_TMP_DIR:?}"
if [[ "${1:-}" == --analysis ]]; then
  root_dir="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
  cd "$root_dir/tools/go-analysis"
  env GOTOOLCHAIN="${GO_TOOLCHAIN:?}" GOTELEMETRY=off \
    GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
    "${GO:?}" mod tidy
elif [[ "$#" -ne 0 ]]; then
  echo 'usage: install-go-dependencies.sh [--analysis|--all]' >&2
  exit 2
else
  # Explicit bootstrap is the dependency-maintenance boundary. Let Go update
  # its managed checksum file and retire dependencies removed from imports.
  module_list="$("${NODE_BIN:?}" --input-type=module -e 'import { independentGoModules, validateIndependentGoModule } from "./tools/harness/test-catalog/go-modules.mjs"; for (const module of independentGoModules(process.cwd())) { validateIndependentGoModule(process.cwd(), module.root); process.stdout.write(module.root + "\n"); }')"
  # Independent producer manifests pin the exact exported format source archive.
  # Snapshot versions carry a source digest; an existing version is never reused
  # for different bytes and no local replace/workspace supplies the dependency.
  module_work="$(mktemp -d "$GO_TMP_DIR/reference-pack-bootstrap.XXXXXXXX")"
  trap 'rm -rf -- "$module_work"' EXIT
  env GOWORK=off GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
    "${GO:?}" build -mod=readonly -trimpath -buildvcs=false -o "$module_work/archive" ./tools/reference-pack-modules
  format_module="github.com/JochiRaider/cartulary/referencepacks/format"
  format_version=""
  format_proxy="file://$(pwd)/.cartulary/reference-packs/modules/proxy"
  while IFS= read -r module_root; do
    [[ -n "$module_root" ]] || continue
    if [[ "$module_root" != referencepacks/format ]]; then
      [[ -n "$format_version" ]] || { echo 'format archive must precede producer dependency preparation' >&2; exit 1; }
      "$GO" -C "$module_root" mod edit "-require=$format_module@$format_version"
    fi
    env GOWORK=off GOTOOLCHAIN="${GO_TOOLCHAIN:?}" GOTELEMETRY=off \
      GOPROXY="$format_proxy,${GOPROXY:-https://proxy.golang.org}" GONOSUMDB="$format_module" \
      GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
      "${GO:?}" -C "$module_root" mod tidy
    env GOWORK=off GOTOOLCHAIN="${GO_TOOLCHAIN:?}" GOTELEMETRY=off \
      GOPROXY="$format_proxy,${GOPROXY:-https://proxy.golang.org}" GONOSUMDB="$format_module" \
      GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
      "${GO:?}" -C "$module_root" mod download all
    # Cache the complete module graph, then retain only the checksums required
    # by this module. Downloading all may add unused dependency-test sums. The
    # exported source identity must describe the final maintained manifest.
    env GOWORK=off GOTOOLCHAIN="${GO_TOOLCHAIN:?}" GOTELEMETRY=off \
      GOPROXY="$format_proxy,${GOPROXY:-https://proxy.golang.org}" GONOSUMDB="$format_module" \
      GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
      "${GO:?}" -C "$module_root" mod tidy
    if [[ "$module_root" == referencepacks/format ]]; then
      format_base="$("${NODE_BIN:?}" --input-type=module -e 'import { independentGoModules } from "./tools/harness/test-catalog/go-modules.mjs"; process.stdout.write(independentGoModules(process.cwd()).find(module => module.kind === "reference_pack_format").version);')"
      "$module_work/archive" -root "$module_root" -module "$format_module" -version "$format_base" -out "$(pwd)/.cartulary/reference-packs/modules" > "$module_work/format.json"
      format_version="$("$NODE_BIN" --input-type=module -e 'import { readFileSync } from "node:fs"; process.stdout.write(JSON.parse(readFileSync(process.argv[1], "utf8")).version);' "$module_work/format.json")"
    fi
  done <<< "$module_list"
  "$GO" mod edit "-require=$format_module@$format_version"
  env GOWORK=off GOTOOLCHAIN="${GO_TOOLCHAIN:?}" GOTELEMETRY=off \
    GOPROXY="$format_proxy,${GOPROXY:-https://proxy.golang.org}" GONOSUMDB="$format_module" \
    GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
    "${GO:?}" mod tidy

fi
env GOTOOLCHAIN="${GO_TOOLCHAIN:?}" GOTELEMETRY=off \
  GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
  "${GO:?}" mod download all
