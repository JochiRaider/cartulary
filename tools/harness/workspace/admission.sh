#!/usr/bin/env bash
# Worktree admission is independent of host capacity accounting.
set -euo pipefail

# This clock is diagnostic only. Canonical scheduling keeps its own epoch.
admission_boot_ms() {
  local admission_uptime admission_fraction
  read -r admission_uptime _ </proc/uptime || return 1
  admission_fraction="${admission_uptime#*.}00"
  printf '%s' "$((10#${admission_uptime%%.*} * 1000 + 10#${admission_fraction:0:3}))"
}
admission_entry_ms=''
if [[ "${1:-}" == run && "${HARNESS_DIAGNOSTICS:-off}" == basic ]]; then
  admission_entry_ms="$(admission_boot_ms)" || admission_entry_ms=''
fi

admission_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
admission_fail() { printf 'workspace admission: %s\n' "$1" >&2; exit "${2:-2}"; }
for admission_utility in git flock stat readlink; do
  command -v "$admission_utility" >/dev/null || admission_fail "configuration_error: missing $admission_utility"
done
admission_git="$(git -C "$admission_root" rev-parse --absolute-git-dir)" || admission_fail 'configuration_error: Git administrative storage unavailable'
admission_lock="$admission_git/cartulary-workspace.lock"

admission_valid() {
  [[ ! -L "$admission_lock" && -f "$admission_lock" ]] || return 1
  [[ "$(stat -Lc '%d:%i' /proc/self/fd/9 2>/dev/null)" == "$(stat -c '%d:%i' "$admission_lock")" ]] || return 1
  local admission_line
  while IFS= read -r admission_line; do
    [[ "$admission_line" =~ ^lock:.*FLOCK[[:space:]]+ADVISORY[[:space:]]+READ[[:space:]] ]] && return 0
  done <"/proc/$$/fdinfo/9"
  return 1
}

admission_supervise() {
  local admission_private admission_boot admission_namespace admission_uuid admission_stat
  local admission_child admission_status admission_signal=0 admission_end
  admission_private="$(umask 077; mktemp -d /tmp/cartulary-invocation.XXXXXXXXXX)" || exec "$@"
  if ! read -r admission_boot </proc/sys/kernel/random/boot_id ||
     ! read -r admission_uuid </proc/sys/kernel/random/uuid ||
     ! read -r admission_stat <"/proc/$$/stat" ||
     ! admission_namespace="$(readlink "/proc/$$/ns/time")"; then
    rmdir -- "$admission_private" || true
    exec "$@"
  fi
  admission_stat="${admission_stat##*) }"
  local -a admission_fields
  read -r -a admission_fields <<<"$admission_stat"
  if ! (umask 077; printf '{"version":1,"invocation_id":"%s","pid":%s,"start_ticks":"%s","start_ms":%s,"boot_id":"%s","namespace":"%s"}\n' \
    "$admission_uuid" "$$" "${admission_fields[19]}" "$admission_entry_ms" "$admission_boot" "$admission_namespace" >"$admission_private/owner.json"); then
    rm -rf -- "$admission_private" || true
    exec "$@"
  fi
  export CARTULARY_HARNESS_INVOCATION_ROOT="$admission_private"
  # Give the child its own process group so cancellation reaches the managed
  # command and descendants. Job control also preserves SIGINT for the child.
  set -m
  "$@" <&0 &
  admission_child=$!
  set +m
  trap 'admission_signal=2; kill -INT -- "-$admission_child" 2>/dev/null || true' INT
  trap 'admission_signal=15; kill -TERM -- "-$admission_child" 2>/dev/null || true' TERM
  while true; do
    if wait "$admission_child"; then admission_status=0; else admission_status=$?; fi
    kill -0 "$admission_child" 2>/dev/null || break
  done
  if ((admission_signal)); then
    # An interrupted wait can return before the child's own trap chooses its
    # exit. Bash retains the child's status, including after it has been reaped.
    if wait "$admission_child"; then admission_status=0; else admission_status=$?; fi
  fi
  trap '' INT TERM
  admission_end="$(admission_boot_ms)" || admission_end=''
  if [[ -n "$admission_end" ]]; then
    if (umask 077; printf '{"end_ms":%s,"exit_code":%s,"signal":%s,"boot_id":"%s","namespace":"%s"}\n' \
      "$admission_end" "$admission_status" "$admission_signal" "$admission_boot" "$admission_namespace" >"$admission_private/completion.json") && [[ -x "$admission_root/.cache/cartulary/toolbin/cartulary-harness-diag" ]]; then
      "$admission_root/.cache/cartulary/toolbin/cartulary-harness-diag" --terminal-receipt "$admission_private" || true
    fi
  fi
  rm -rf -- "$admission_private" || true
  exit "$admission_status"
}

case "${1:-}" in
  verify) if admission_valid 2>/dev/null; then printf 'shared\n'; else exit 1; fi ;;
  run)
    shift
    [[ "${1:-}" == -- ]] && shift
    (($# > 0)) || admission_fail 'usage_error: missing managed command'
    if ! admission_valid 2>/dev/null; then
      [[ ! -L "$admission_lock" ]] || admission_fail 'configuration_error: symlinked admission lock'
      # Never truncate or replace this inode: live descendants may still own it.
      (umask 077; : >>"$admission_lock")
      exec 9<"$admission_lock"
      flock --nonblock --shared 9 || admission_fail 'resource_conflict: cleanup is active' 4
      if [[ -n "$admission_entry_ms" ]]; then admission_supervise "$@"; fi
    fi
    exec "$@"
    ;;
  *) admission_fail 'usage_error: expected verify or run' ;;
esac
