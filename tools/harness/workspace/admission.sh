#!/usr/bin/env bash
# Worktree admission is independent of host capacity accounting.
set -euo pipefail

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
    fi
    exec "$@"
    ;;
  *) admission_fail 'usage_error: expected verify or run' ;;
esac
