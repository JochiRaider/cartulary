#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
GO_TMP_DIR="${GO_TMP_DIR:?GO_TMP_DIR is required}"
GOSEC_BIN="${GOSEC_BIN:-$ROOT_DIR/tmp/toolbin/gosec-v2.26.1}"
GOSEC_AUDIT_RUNTIME_RULES="${GOSEC_AUDIT_RUNTIME_RULES:-G118,G122,G301,G302,G303,G304,G305,G306,G307}"
GOSEC_AUDIT_RUNTIME_FLAGS="${GOSEC_AUDIT_RUNTIME_FLAGS:-}"
GOSEC_AUDIT_RUNTIME_PATTERNS="${GOSEC_AUDIT_RUNTIME_PATTERNS:-./cmd/... ./internal/...}"
GOSEC_AUDIT_SUPPORT_RULES="${GOSEC_AUDIT_SUPPORT_RULES:-G122,G301,G302,G303,G304,G305,G306,G307}"
GOSEC_AUDIT_SUPPORT_FLAGS="${GOSEC_AUDIT_SUPPORT_FLAGS:--exclude-generated -no-fail -quiet}"
GOSEC_AUDIT_SUPPORT_PATTERNS="${GOSEC_AUDIT_SUPPORT_PATTERNS:-./tools/...}"
profile_metadata="${CARTULARY_STEP_ARTIFACT_DIR:+${CARTULARY_STEP_ARTIFACT_DIR}/security-profiles.jsonl}"
audit_cpu_limit="${CARTULARY_SEQUENCE_HOST_CPU_LIMIT:-}"
if [[ -z "$audit_cpu_limit" ]]; then
  audit_cpu_limit="$(getconf _NPROCESSORS_ONLN 2>/dev/null || nproc)"
fi
if [[ ! "$audit_cpu_limit" =~ ^[1-9][0-9]*$ ]]; then
  echo "go-gosec-audit requires a positive scheduler CPU limit" >&2
  exit 1
fi

if [[ "$GO_BIN" != */* ]] && command -v "$GO_BIN" >/dev/null 2>&1; then
  GO_BIN="$(command -v "$GO_BIN")"
elif [[ "$GO_BIN" != /* ]]; then
  GO_BIN="$ROOT_DIR/$GO_BIN"
fi

if [[ ! -x "$GO_BIN" ]]; then
  echo "go-gosec-audit requires an executable GO at $GO_BIN" >&2
  exit 1
fi

if [[ "$GOSEC_BIN" != */* ]] && command -v "$GOSEC_BIN" >/dev/null 2>&1; then
  GOSEC_BIN="$(command -v "$GOSEC_BIN")"
elif [[ "$GOSEC_BIN" != /* ]]; then
  GOSEC_BIN="$ROOT_DIR/$GOSEC_BIN"
fi

if [[ ! -x "$GOSEC_BIN" ]]; then
  echo "go-gosec-audit requires an executable GOSEC_BIN at $GOSEC_BIN" >&2
  echo "run make go-security-toolchain before go-gosec-audit or set GOSEC_BIN to a ready gosec binary" >&2
  exit 1
fi

declare -A seen_rules=()
repository_rules=()
for rules_value in "$GOSEC_AUDIT_RUNTIME_RULES" "$GOSEC_AUDIT_SUPPORT_RULES"; do
  IFS=',' read -r -a rule_entries <<<"$rules_value"
  for rule in "${rule_entries[@]}"; do
    if [[ -n "$rule" && -z "${seen_rules[$rule]:-}" ]]; then
      seen_rules["$rule"]=1
      repository_rules+=("$rule")
    fi
  done
done
repository_rules_value="$(IFS=','; printf '%s' "${repository_rules[*]}")"

runtime_patterns=()
read -r -a runtime_patterns <<<"$GOSEC_AUDIT_RUNTIME_PATTERNS"
repository_patterns=("${runtime_patterns[@]}")
read -r -a support_patterns <<<"$GOSEC_AUDIT_SUPPORT_PATTERNS"
for support_pattern in "${support_patterns[@]}"; do
  covered=0
  for runtime_pattern in "${runtime_patterns[@]}"; do
    runtime_prefix="${runtime_pattern%/...}"
    if [[ "$support_pattern" == "$runtime_pattern" || "$support_pattern" == "$runtime_prefix"/* ]]; then
      covered=1
      break
    fi
  done
  if [[ "$covered" -eq 0 ]]; then
    repository_patterns+=("$support_pattern")
  fi
done
repository_patterns_value="${repository_patterns[*]}"

run_profile() {
  local label="$1"
  local rules="$2"
  local flags="$3"
  local patterns_value="$4"

  if [[ -z "$rules" ]]; then
    echo "go-gosec-audit requires at least one ${label} rule entry" >&2
    exit 1
  fi
  if [[ -z "$patterns_value" ]]; then
    echo "go-gosec-audit requires at least one ${label} package pattern" >&2
    exit 1
  fi

  local args=("-include=$rules")
  if [[ -n "$flags" ]]; then
    local flag_args=()
    read -r -a flag_args <<<"$flags"
    args+=("${flag_args[@]}")
  fi

  local patterns=()
  read -r -a patterns <<<"$patterns_value"

  if [[ -n "$profile_metadata" ]]; then
    mkdir -p "$(dirname "$profile_metadata")"
    printf '{"tool":"gosec","target":"go-gosec-audit","profile":"%s","rules":"%s","flags":"%s","patterns":"%s"}\n' \
      "$label" "$rules" "$flags" "$patterns_value" >>"$profile_metadata"
  fi
  printf 'go-gosec-audit advisory %s profile rules=%s patterns=%s\n' "$label" "$rules" "$patterns_value"
  env GOMAXPROCS="$audit_cpu_limit" \
    GOCACHE="$GO_CACHE_DIR" \
    GOMODCACHE="$GO_MOD_CACHE_DIR" \
    GOTMPDIR="$GO_TMP_DIR" \
    PATH="$(dirname "$GO_BIN"):$PATH" \
    "$GOSEC_BIN" "${args[@]}" "${patterns[@]}"
}

cd "$ROOT_DIR"

run_profile \
  "repository" \
  "$repository_rules_value" \
  "$GOSEC_AUDIT_SUPPORT_FLAGS" \
  "$repository_patterns_value"
