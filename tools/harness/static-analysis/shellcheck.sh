#!/usr/bin/env bash
set -euo pipefail

DEFAULT_ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
ROOT_DIR="${CARTULARY_SHELLCHECK_ROOT:-$DEFAULT_ROOT_DIR}"
SHELLCHECK_BIN="${SHELLCHECK_BIN:-${ROOT_DIR}/tmp/toolbin/shellcheck-v0.11.0}"
LINT_SHELL_STRICT="${LINT_SHELL_STRICT:-0}"
CACHE_ARTIFACT_SCRIPT="$DEFAULT_ROOT_DIR/tools/harness/readiness/cache-artifact.sh"
CACHE_DIR="${CARTULARY_STATIC_ANALYSIS_CACHE_DIR:-$ROOT_DIR/.cache/cartulary/static-analysis}"
CACHE_STAMP="$CACHE_DIR/outputs/lint-shell.ok"

# shellcheck source=tools/harness/generated-artifacts/generated-artifacts.sh
# shellcheck disable=SC1091
source "$DEFAULT_ROOT_DIR/tools/harness/generated-artifacts/generated-artifacts.sh"

resolve_shellcheck_bin() {
  local candidate="$1"

  if [[ "$candidate" != */* ]] && command -v "$candidate" >/dev/null 2>&1; then
    command -v "$candidate"
    return 0
  fi
  if [[ "$candidate" != /* ]]; then
    printf '%s/%s\n' "$ROOT_DIR" "$candidate"
    return 0
  fi
  printf '%s\n' "$candidate"
}

is_excluded_path() {
  local path="$1"

  if cartulary_is_generated_artifact_path "$path"; then
    return 0
  fi

  case "$path" in
    vendor/* | */vendor/* | \
    node_modules/* | */node_modules/* | \
    tmp/* | */tmp/* | \
    .cache/* | */.cache/* | \
    .cartulary/* | */.cartulary/* | \
    .pnpm-store/* | */.pnpm-store/* | \
    coverage/* | */coverage/* | \
    playwright-report/* | */playwright-report/* | \
    reports/* | */reports/* | \
    test-results/* | */test-results/* | \
    dist/* | */dist/* | \
    build/* | */build/* | \
    out/* | */out/*)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

has_shell_shebang() {
  local file="$1"
  local first_line=""

  IFS= read -r first_line <"$file" || true
  [[ "$first_line" =~ ^#!.*(^|[[:space:]/])(bash|sh|dash|ksh)([[:space:]]|$) ]] && return 0
  [[ "$first_line" =~ ^#!.*(^|[[:space:]/])busybox[[:space:]]+sh([[:space:]]|$) ]] && return 0
  return 1
}

discover_shell_files() {
  local rel
  local path
  local shell_files=()

  while IFS= read -r -d '' rel; do
    is_excluded_path "$rel" && continue
    path="${ROOT_DIR}/${rel}"
    [[ -f "$path" && ! -L "$path" ]] || continue
    if [[ "$rel" == *.sh ]] || has_shell_shebang "$path"; then
      shell_files+=("$rel")
    fi
  done < <(git -C "$ROOT_DIR" ls-files -z --cached --others --exclude-standard)

  if [[ "${#shell_files[@]}" -eq 0 ]]; then
    return 0
  fi
  printf '%s\0' "${shell_files[@]}" | LC_ALL=C sort -z
}

shellcheck_bin="$(resolve_shellcheck_bin "$SHELLCHECK_BIN")"
shell_files=()
mapfile -d '' -t shell_files < <(discover_shell_files)
inventory_artifact=""
if [[ -n "${CARTULARY_STEP_ARTIFACT_DIR:-}" ]]; then
  mkdir -p "$CARTULARY_STEP_ARTIFACT_DIR"
  inventory_artifact="${CARTULARY_STEP_ARTIFACT_DIR}/shellcheck-inventory.txt"
fi

if [[ "${#shell_files[@]}" -eq 0 ]]; then
  if [[ -n "$inventory_artifact" ]]; then
    : >"$inventory_artifact"
  fi
  printf '0 files checked\n'
  exit 0
fi

if [[ -n "$inventory_artifact" ]]; then
  printf '%s\n' "${shell_files[@]}" >"$inventory_artifact"
else
  printf '%s\n' "${shell_files[@]}"
fi

if [[ ! -x "$shellcheck_bin" ]]; then
  echo "lint-shell requires an executable SHELLCHECK_BIN at $shellcheck_bin" >&2
  echo "run make shell-lint-toolchain before lint-shell or set SHELLCHECK_BIN to a ready ShellCheck binary" >&2
  exit 1
fi

cd "$ROOT_DIR"

run_shellcheck_direct() {
  set +e
  "$shellcheck_bin" "${shell_files[@]}"
  status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    printf '%s files checked\n' "${#shell_files[@]}"
    exit 0
  fi

  if [[ "$LINT_SHELL_STRICT" == "1" ]]; then
    exit "$status"
  fi

  printf 'lint-shell warning-only: ShellCheck exited with status %s; set LINT_SHELL_STRICT=1 to fail on findings\n' "$status" >&2
  exit 0
}

if [[ "$LINT_SHELL_STRICT" != "1" ]]; then
  run_shellcheck_direct
fi

shellcheck_version="$("$shellcheck_bin" --version 2>/dev/null | tr '\n' ' ' || true)"
file_list="$(mktemp "${TMPDIR:-/tmp}/cartulary-lint-shell.XXXXXX")"
trap 'rm -f "$file_list"' EXIT
printf '%s\0' "${shell_files[@]}" >"$file_list"

cache_args=(
  --schema-id cartulary.cache.static_analysis.v1
  --scope static-analysis
  --profile lint-shell
  --cache-dir "$CACHE_DIR"
  --disable-env CARTULARY_STATIC_ANALYSIS_DISABLE_CACHE
  --force-env CARTULARY_STATIC_ANALYSIS_FORCE
  --input "$DEFAULT_ROOT_DIR/tools/harness/static-analysis/shellcheck.sh"
  --input "$DEFAULT_ROOT_DIR/tools/harness/static-analysis/shellcheck-runner.sh"
  --input "$DEFAULT_ROOT_DIR/tools/harness/generated-artifacts/generated-artifacts.sh"
  --input "$DEFAULT_ROOT_DIR/tools/harness/readiness/cache-artifact.sh"
  --input "$DEFAULT_ROOT_DIR/tools/harness/readiness/cache-policy.sh"
  --output "$CACHE_STAMP"
  --key "shellcheck_bin=$shellcheck_bin"
  --key "shellcheck_version=$shellcheck_version"
  --key "strict=$LINT_SHELL_STRICT"
)
if [[ -f "$shellcheck_bin" ]]; then
  cache_args+=(--input "$shellcheck_bin")
fi
for rel in "${shell_files[@]}"; do
  cache_args+=(--input "$ROOT_DIR/$rel")
done

status=0
env \
  CARTULARY_SHELLCHECK_FILE_LIST="$file_list" \
  CARTULARY_STATIC_CACHE_STAMP="$CACHE_STAMP" \
  SHELLCHECK_BIN="$shellcheck_bin" \
  "$CACHE_ARTIFACT_SCRIPT" "${cache_args[@]}" -- "$DEFAULT_ROOT_DIR/tools/harness/static-analysis/shellcheck-runner.sh" || status=$?

if [[ "$status" -eq 0 ]]; then
  printf '%s files checked\n' "${#shell_files[@]}"
  exit 0
fi

exit "$status"
