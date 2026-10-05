#!/usr/bin/env bash
# Deliberately independent of runtimes, package installs, and retained reports.
set -euo pipefail
export LC_ALL=C
cleanup_execution=0

cleanup_fail() {
  printf 'cleanup: %s\n' "$2" >&2
  if ((cleanup_execution)); then execution_failure; fi
  exit "$1"
}
for utility in git flock stat find sort sha256sum readlink rm chmod; do
  command -v "$utility" >/dev/null || cleanup_fail 2 "configuration_error: missing $utility"
done
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd -- "$root"
[[ $# == 1 && ( "$1" == clean || "$1" == distclean ) ]] || cleanup_fail 2 'usage_error: expected clean or distclean'
scope="$1"
command_id="cartulary.harness.command.${scope}.v2"
trim() {
  local value="$1" previous whitespace
  # Match the existing TrimString contract without a locale/runtime dependency.
  # Whole UTF-8 sequences are essential: a byte character class would also trim
  # bytes from non-whitespace Unicode characters.
  local -a whitespace_sequences=(
    ' ' $'\t' $'\n' $'\v' $'\f' $'\r' $'\302\240' $'\341\232\200'
    $'\342\200\200' $'\342\200\201' $'\342\200\202' $'\342\200\203'
    $'\342\200\204' $'\342\200\205' $'\342\200\206' $'\342\200\207'
    $'\342\200\210' $'\342\200\211' $'\342\200\212' $'\342\200\250'
    $'\342\200\251' $'\342\200\257' $'\342\201\237' $'\343\200\200' $'\357\273\277'
  )
  while :; do
    previous="$value"
    for whitespace in "${whitespace_sequences[@]}"; do
      value="${value#"$whitespace"}"; value="${value%"$whitespace"}"
    done
    [[ "$previous" != "$value" ]] || break
  done
  REPLY="$value"
}
trim "${CARTULARY_CLEANUP_DRY_RUN:-}"
case "$REPLY" in '') preview=0 ;; 1) preview=1 ;; *) cleanup_fail 2 'usage_error: CARTULARY_CLEANUP_DRY_RUN must be empty or 1' ;; esac
trim "${CARTULARY_OUTPUT_MODE:-}"
case "$REPLY" in ''|quiet|summary|verbose|ci) ;; *) cleanup_fail 2 'usage_error: unsupported cleanup output mode' ;; esac
for input in ${CARTULARY_MAKE_COMMAND_LINE_INPUTS:-}; do
  case "$input" in CARTULARY_CLEANUP_DRY_RUN|CARTULARY_OUTPUT_MODE|VERBOSE|CI_VERBOSE|CI) ;; *) cleanup_fail 2 "usage_error: undeclared Make input $input" ;; esac
done
projection=tools/workspace_layout.generated.sh
[[ -f "$projection" && ! -L "$projection" ]] || cleanup_fail 2 'configuration_error: missing workspace projection'
# shellcheck source=tools/workspace_layout.generated.sh
source "$projection"
printf '%s\n' "$CARTULARY_LAYOUT_DIGESTS" | sha256sum --check --status || cleanup_fail 2 'configuration_error: stale workspace projection; run make generate'
git_dir="$(git rev-parse --absolute-git-dir)" || cleanup_fail 2 'configuration_error: Git administrative storage unavailable'
[[ "$(git rev-parse --show-toplevel)" == "$root" ]] || cleanup_fail 2 'configuration_error: cleanup requires the worktree root'
lock="$git_dir/cartulary-workspace.lock"
[[ ! -L "$lock" ]] || cleanup_fail 2 'configuration_error: symlinked admission lock'
if ((preview)); then
  if [[ -e "$lock" ]]; then
    exec 8<"$lock"
    flock --nonblock --exclusive 8 || cleanup_fail 4 'resource_conflict: managed work is active'
  fi
else
  (umask 077; : >>"$lock")
  exec 8<"$lock"
  flock --nonblock --exclusive 8 || cleanup_fail 4 'resource_conflict: managed work is active'
fi

# The success sentinel carries the producer's status across process substitution.
# Enumeration is NUL delimited; no temporary files are created even on failure.
enumerate() {
  local -n enumeration_result="$1"
  shift
  mapfile -d '' -t enumeration_result < <("$@" 2>/dev/null && printf '\001complete\0')
  ((${#enumeration_result[@]})) && [[ "${enumeration_result[-1]}" == $'\001complete' ]] || cleanup_fail 12 'cleanup_error: enumeration failed before mutation'
  unset 'enumeration_result[-1]'
}
declare -a worktrees=() git_records=() tracked=() candidates=() ids=()
enumerate git_records git worktree list --porcelain -z
for record in "${git_records[@]}"; do
  [[ "$record" != 'worktree '* ]] || worktrees+=("${record#worktree }")
done
worktrees+=("$git_dir" "$(git rev-parse --path-format=absolute --git-common-dir)")
enumerate tracked git ls-files -z --cached

check_ancestors() {
  local relative="$1" cursor="$root" remaining="$1"
  while [[ "$remaining" == */* ]]; do
    cursor+="/${remaining%%/*}"
    remaining="${remaining#*/}"
    [[ ! -L "$cursor" ]] || cleanup_fail 12 "cleanup_error: symlinked ancestor of $(printf '%q' "$relative")"
    [[ ! -e "$cursor" || -d "$cursor" ]] || cleanup_fail 12 'cleanup_error: ancestor is not a directory'
  done
}
append_candidate() {
  local id="$1" relative="$2" previous
  check_ancestors "$relative"
  for previous in "${candidates[@]}"; do
    [[ "$relative" != "$previous" && "$relative" != "$previous/"* ]] || return 0
  done
  candidates+=("$relative"); ids+=("$id")
}
for entry in "${CARTULARY_CLEANUP_ENTRIES[@]}"; do
  IFS='|' read -r id tier relative <<<"$entry"
  [[ "$tier" == clean || "$scope" == distclean ]] || continue
  if [[ "$relative" == 'packages/*/'* ]]; then
    check_ancestors 'packages/placeholder'
    if [[ -d packages ]]; then
      declare -a packages=()
      enumerate packages find -P packages -mindepth 1 -maxdepth 1 -print0
      for package in "${packages[@]}"; do
        [[ ! -L "$package" ]] || cleanup_fail 12 'cleanup_error: symlinked workspace package'
        [[ -d "$package" ]] || continue
        append_candidate "$id" "$package/${relative#packages/\*/}"
      done
    fi
  else
    append_candidate "$id" "$relative"
  fi
done
# A later ancestor (e.g. node_modules in distclean) subsumes a clean-tier child.
declare -a plan=() plan_ids=()
for index in "${!candidates[@]}"; do
  relative="${candidates[index]}"; covered=0
  for ancestor in "${candidates[@]}"; do
    if [[ "$relative" == "$ancestor/"* ]]; then covered=1; break; fi
  done
  if ((covered == 0)); then plan+=("$relative"); plan_ids+=("${ids[index]}"); fi
done
((${#plan[@]} <= 180)) || cleanup_fail 2 'configuration_error: cleanup plan exceeds human output budget'

declare -a fingerprints=() directory_paths=() directory_ids=() directory_modes=() plan_states=() root_ids=()
declare -a listing=() mount_records=()
# Linux mountinfo catches bind mounts as well as differing device IDs.
mapfile -t mount_records </proc/self/mountinfo || cleanup_fail 12 'cleanup_error: mount enumeration unavailable'
inspect_tree() {
  local relative="$1" row identity kind mode owner object device root_device mount mountpoint
  local -a fields
  root_device="$(stat -c '%d' -- "$relative")" || cleanup_fail 12 'cleanup_error: cannot inspect candidate'
  for mount in "${mount_records[@]}"; do
    read -r -a fields <<<"$mount"
    printf -v mountpoint '%b' "${fields[4]}"
    [[ "$mountpoint" != "$root/$relative" && "$mountpoint" != "$root/$relative/"* ]] || cleanup_fail 12 'cleanup_error: candidate crosses a mount boundary'
  done
  enumerate listing find -P "$relative" -xdev -printf '%D:%i:%y:%m:%U\0%p\0'
  for ((row=0; row<${#listing[@]}; row+=2)); do
    identity="${listing[row]}"; object="${listing[row+1]}"
    IFS=: read -r device _ kind mode owner <<<"$identity"
    [[ "$device" == "$root_device" ]] || cleanup_fail 12 'cleanup_error: cross-filesystem traversal'
    [[ "${object##*/}" != .git ]] || cleanup_fail 12 "cleanup_error: nested Git metadata in $(printf '%q' "$relative")"
    case "$kind" in
      d)
        if [[ ! -r "$object" || ! -x "$object" ]] || (( (8#$mode & 0500) != 0500 )); then
          cleanup_fail 12 'cleanup_error: unreadable directory'
        fi
        if (( (8#$mode & 0700) != 0700 && owner != EUID && EUID != 0 )); then
          cleanup_fail 12 'cleanup_error: cannot repair directory permissions owned by another user'
        fi
        if [[ "$inspection_phase" == plan ]]; then
          directory_paths+=("$object"); directory_ids+=("${identity%:*:*}"); directory_modes+=("$mode")
        fi
        ;;
      f|l) ;;
      *) cleanup_fail 12 'cleanup_error: unsupported filesystem object' ;;
    esac
  done
  tree_fingerprint="$(printf '%s\0' "${listing[@]}" | sort -z | sha256sum)"
}

inspection_phase=plan
for index in "${!plan[@]}"; do
  relative="${plan[index]}"
  if [[ -e "$relative" || -L "$relative" ]]; then
    parent=.
    [[ "$relative" != */* ]] || parent="${relative%/*}"
    [[ -w "$parent" && -x "$parent" ]] || cleanup_fail 12 "cleanup_error: candidate parent is not writable: $(printf '%q' "$parent")"
  fi
  for worktree in "${worktrees[@]}"; do
    [[ "$worktree" != "$root/$relative" && "$worktree" != "$root/$relative/"* ]] || cleanup_fail 12 "cleanup_error: registered worktree in $(printf '%q' "$relative")"
  done
  for source in "${tracked[@]}"; do
    [[ "$source" != "$relative" && "$source" != "$relative/"* ]] || cleanup_fail 12 "cleanup_error: tracked source overlaps $(printf '%q' "$relative")"
  done
  if [[ -L "$relative" ]]; then
    root_ids+=('')
    plan_states+=(symlink); fingerprints+=("$(stat -c '%d:%i:%f' -- "$relative")")
  elif [[ ! -e "$relative" ]]; then
    root_ids+=('')
    plan_states+=(absent); fingerprints+=('')
  elif [[ -d "$relative" ]]; then
    root_ids+=("$(stat -c '%d:%i' -- "$relative")")
    plan_states+=(directory); inspect_tree "$relative"; fingerprints+=("$tree_fingerprint")
  else
    cleanup_fail 12 "cleanup_error: candidate is not a directory: $(printf '%q' "$relative")"
  fi
done
if ((preview)); then
  for index in "${!plan[@]}"; do
    action=remove-directory
    case "${plan_states[index]}" in symlink) action='unlink' ;; absent) action='no-op' ;; esac
    printf 'DRY-RUN %s %q proof=workspace-layout:%s state=%s\n' "$action" "$root/${plan[index]}" "${plan_ids[index]}" "${plan_states[index]}"
  done
  exit 0
fi

# Revalidate the whole inspected tree before the first permission change. The
# lock excludes cooperating writers; this is not a hostile-filesystem sandbox.
inspection_phase=recheck
for index in "${!plan[@]}"; do
  relative="${plan[index]}"; check_ancestors "$relative"
  case "${plan_states[index]}" in
    absent) [[ ! -e "$relative" && ! -L "$relative" ]] || cleanup_fail 12 'cleanup_error: candidate appeared during validation' ;;
    symlink) [[ -L "$relative" && "$(stat -c '%d:%i:%f' -- "$relative")" == "${fingerprints[index]}" ]] || cleanup_fail 12 'cleanup_error: candidate changed during validation' ;;
    directory)
      [[ -d "$relative" && ! -L "$relative" ]] || cleanup_fail 12 'cleanup_error: directory identity changed'
      inspect_tree "$relative"
      [[ "$tree_fingerprint" == "${fingerprints[index]}" ]] || cleanup_fail 12 'cleanup_error: directory contents changed during validation'
      ;;
  esac
done
completed=0
execution_failure() {
  printf 'cleanup: cleanup_error command=%s completed=%s remaining=%s; deletion is not rolled back\n' "$command_id" "$completed" "$((${#plan[@]}-completed))" >&2
  for ((index=completed; index<${#plan[@]} && index<completed+20; index++)); do printf 'remaining %q\n' "${plan[index]}" >&2; done
  exit 12
}
trap execution_failure ERR
cleanup_execution=1
for index in "${!directory_paths[@]}"; do
  object="${directory_paths[index]}"
  [[ ! -L "$object" && "$(stat -c '%d:%i:d' -- "$object")" == "${directory_ids[index]}" ]] || execution_failure
  if (( (8#${directory_modes[index]} & 0700) != 0700 )); then chmod u+rwx -- "$object"; fi
done
for index in "${!plan[@]}"; do
  relative="${plan[index]}"
  check_ancestors "$relative"
  case "${plan_states[index]}" in
    symlink)
      [[ -L "$relative" && "$(stat -c '%d:%i:%f' -- "$relative")" == "${fingerprints[index]}" ]] || execution_failure
      rm -- "$relative"
      ;;
    directory)
      [[ ! -L "$relative" && "$(stat -c '%d:%i' -- "$relative")" == "${root_ids[index]}" ]] || execution_failure
      rm -rf --one-file-system -- "$relative"
      ;;
  esac
  printf 'CLEANED %q proof=workspace-layout:%s\n' "$root/$relative" "${plan_ids[index]}"
  completed=$((completed+1))
done
