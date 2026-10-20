#!/usr/bin/env bash
# Deliberately independent of runtimes, package installs, and retained reports.
set -euo pipefail
set +m
export LC_ALL=C
cleanup_execution=0
completed=0
phase=configuration
active_root=none
command_id=unresolved
launching=0
pending_signal=0
declare -a plan=() plan_ids=()
declare -A children=()

# Each utility is launched in this invocation's own child session. The lock is
# inherited by every child, including a child surviving forced parent death.
stop_children() {
  local child
  for child in "${!children[@]}"; do
    kill -TERM -- "-$child" 2>/dev/null || kill -TERM "$child" 2>/dev/null || true
    kill -KILL -- "-$child" 2>/dev/null || kill -KILL "$child" 2>/dev/null || true
  done
  for child in "${!children[@]}"; do wait "$child" 2>/dev/null || true; done
  children=()
}
report_incomplete() {
  local remaining_index
  printf 'cleanup: %s command=%s completed=%s remaining=%s phase=%s root=%s; deletion is not rolled back\n' \
    "$1" "$command_id" "$completed" "$((${#plan[@]}-completed))" "$phase" "$active_root" >&2
  for ((remaining_index=completed; remaining_index<${#plan[@]} && remaining_index<completed+20; remaining_index++)); do
    printf 'remaining proof=workspace-layout:%s\n' "${plan_ids[remaining_index]}" >&2
  done
}

cleanup_fail() {
  trap - ERR
  stop_children
  printf 'cleanup: %s\n' "$2" >&2
  if ((cleanup_execution)); then report_incomplete cleanup_error; fi
  exit "$1"
}
interrupted() {
  # A signal between fork and registration must not orphan the new child.
  # Complete that small publication boundary before settling owned utilities.
  if ((launching)); then pending_signal="$1"; return 0; fi
  trap '' INT TERM
  trap - ERR
  stop_children
  report_incomplete cancelled_or_interrupted
  exit "$1"
}
trap 'interrupted 130' INT
trap 'interrupted 143' TERM
trap stop_children EXIT
for utility in git flock stat find sha256sum readlink rm chmod setsid; do
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
output_mode="${REPLY:-summary}"
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

# Process substitutions supply anonymous pipes, never filesystem scratch.
# Save and wait for their exact producer: EOF alone is not successful traversal.
finish_launch() {
  launching=0
  if ((pending_signal)); then interrupted "$pending_signal"; fi
}
open_stream() {
  launching=1
  exec {stream_fd}< <(exec setsid "$@" 2>/dev/null)
  stream_pid=$!
  children["$stream_pid"]=1
  finish_launch
}
finish_stream() {
  local descriptor="$1" child="$2" status=0
  exec {descriptor}<&-
  wait "$child" 2>/dev/null || status=$?
  unset 'children[$child]'
  return "$status"
}
run_child() {
  local child status=0
  launching=1
  setsid "$@" >/dev/null 2>&1 &
  child=$!
  children["$child"]=1
  finish_launch
  wait "$child" 2>/dev/null || status=$?
  unset 'children[$child]'
  return "$status"
}
enumerate() {
  local -n enumeration_result="$1"
  shift
  local descriptor child record=''
  enumeration_result=()
  open_stream "$@"; descriptor=$stream_fd; child=$stream_pid
  while IFS= read -r -d '' record <&"$descriptor"; do
    enumeration_result+=("$record"); record=''
  done
  [[ -z "$record" ]] || cleanup_fail 12 'cleanup_error: truncated enumeration record'
  finish_stream "$descriptor" "$child" || cleanup_fail 12 'cleanup_error: enumeration failed'
}
stat_records() {
  local -n stat_result="$1"
  shift
  local descriptor child record='' expected=$#
  stat_result=()
  open_stream stat --printf '%d:%i:%a:%u:%f\0' -- "$@"
  descriptor=$stream_fd; child=$stream_pid
  while IFS= read -r -d '' record <&"$descriptor"; do
    stat_result+=("$record"); record=''
    ((${#stat_result[@]} <= expected)) || cleanup_fail 12 'cleanup_error: surplus identity inspection'
  done
  if [[ -n "$record" ]] || ((${#stat_result[@]} != expected)); then
    cleanup_fail 12 'cleanup_error: incomplete identity inspection'
  fi
  finish_stream "$descriptor" "$child" || cleanup_fail 12 'cleanup_error: identity inspection failed'
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
for index in "${!candidates[@]}"; do
  relative="${candidates[index]}"; covered=0
  for ancestor in "${candidates[@]}"; do
    if [[ "$relative" == "$ancestor/"* ]]; then covered=1; break; fi
  done
  if ((covered == 0)); then plan+=("$relative"); plan_ids+=("${ids[index]}"); fi
done
((${#plan[@]} <= 180)) || cleanup_fail 2 'configuration_error: cleanup plan exceeds human output budget'

declare -a plan_states=() root_ids=() repair_needed=() facts=() mount_records=()
# Reserve mandatory action output before emitting any optional progress.
action_bytes=0
for index in "${!plan[@]}"; do
  printf -v action_line 'DRY-RUN remove-directory %q proof=workspace-layout:%s state=directory\n' "$root/${plan[index]}" "${plan_ids[index]}"
  action_bytes=$((action_bytes+${#action_line}))
done
((action_bytes <= 18000)) || cleanup_fail 2 'configuration_error: cleanup plan exceeds human byte budget'
progress_lines=0
progress_bytes=0
progress() {
  local line
  [[ "$output_mode" != quiet ]] || return 0
  ((progress_lines < 20)) || return 0
  printf -v line 'CLEANUP phase=%s roots=%s completed=%s elapsed_s=%s\n' "$phase" "${#plan[@]}" "$completed" "$SECONDS"
  ((progress_bytes+${#line} <= 2000)) || return 0
  printf '%s' "$line"
  progress_lines=$((progress_lines+1)); progress_bytes=$((progress_bytes+${#line}))
}
# Linux mountinfo catches bind mounts as well as differing device IDs.
check_mounts() {
  local relative="$1" mount mountpoint
  local -a fields
  mapfile -t mount_records </proc/self/mountinfo || cleanup_fail 12 'cleanup_error: mount enumeration unavailable'
  for mount in "${mount_records[@]}"; do
    read -r -a fields <<<"$mount"
    printf -v mountpoint '%b' "${fields[4]}"
    [[ "$mountpoint" != "$root/$relative" && "$mountpoint" != "$root/$relative/"* ]] || cleanup_fail 12 'cleanup_error: candidate crosses a mount boundary'
  done
}
inspect_tree() {
  local relative="$1" descriptor child record needs_repair=0
  local -a denied_owner=(-false)
  if ((EUID != 0)); then denied_owner=(! -uid "$EUID"); fi
  check_mounts "$relative"
  # Predicates run in find, not in a shell loop over millions of ordinary files.
  # Only repair notices and the first rejection cross the pipe.
  open_stream find -P "$relative" -xdev \
    \( -name .git -printf 'nested Git metadata\0' -quit \) -o \
    \( ! -type d ! -type f ! -type l -printf 'unsupported filesystem object\0' -quit \) -o \
    \( -type d \( ! -readable -o ! -executable -o ! -perm -0500 \) -printf 'unreadable directory\0' -quit \) -o \
    \( -type d ! -perm -0700 "${denied_owner[@]}" -printf 'unrepairable directory owner\0' -quit \) -o \
    \( -type d ! -perm -0700 -printf 'repair\0' \)
  descriptor=$stream_fd; child=$stream_pid
  record=''
  while IFS= read -r -d '' record <&"$descriptor"; do
    case "$record" in
      repair) ;;
      'nested Git metadata'|'unsupported filesystem object'|'unreadable directory'|'unrepairable directory owner')
        cleanup_fail 12 "cleanup_error: $record" ;;
      *) cleanup_fail 12 'cleanup_error: invalid inspection record' ;;
    esac
    needs_repair=1
    record=''
  done
  [[ -z "$record" ]] || cleanup_fail 12 'cleanup_error: truncated inspection record'
  finish_stream "$descriptor" "$child" || cleanup_fail 12 'cleanup_error: enumeration failed'
  repair_needed+=("$needs_repair")
}
root_fact() {
  stat_records facts "$1"
  [[ "${facts[0]}" =~ ^[0-9]+:[0-9]+:[0-7]+:[0-9]+:[0-9a-f]+$ ]] || cleanup_fail 12 'cleanup_error: invalid root identity'
  IFS=: read -r fact_device fact_inode _ _ fact_type <<<"${facts[0]}"
  root_identity="$fact_device:$fact_inode:$((16#$fact_type & 0170000))"
}
check_root() {
  local item="$1" relative="${plan[$1]}"
  check_ancestors "$relative"
  case "${plan_states[item]}" in
    absent) [[ ! -e "$relative" && ! -L "$relative" ]] || cleanup_fail 12 'cleanup_error: candidate appeared during validation' ;;
    directory|symlink)
      root_fact "$relative"
      [[ "$root_identity" == "${root_ids[item]}" ]] || cleanup_fail 12 'cleanup_error: root identity changed'
      if [[ "${plan_states[item]}" == directory ]]; then check_mounts "$relative"; fi
      ;;
  esac
}

phase=validating
progress
for index in "${!plan[@]}"; do
  relative="${plan[index]}"
  active_root="${plan_ids[index]}"
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
    root_fact "$relative"
    (( (16#$fact_type & 0170000) == 0120000 )) || cleanup_fail 12 'cleanup_error: root type changed'
    root_ids+=("$root_identity"); repair_needed+=(0); plan_states+=(symlink)
  elif [[ ! -e "$relative" ]]; then
    root_ids+=(''); repair_needed+=(0); plan_states+=(absent)
  elif [[ -d "$relative" ]]; then
    root_fact "$relative"
    (( (16#$fact_type & 0170000) == 0040000 )) || cleanup_fail 12 'cleanup_error: root type changed'
    root_ids+=("$root_identity"); plan_states+=(directory); inspect_tree "$relative"
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
  phase=preview_complete
  progress
  exit 0
fi

# The complete plan is validated before any candidate permissions change.
for index in "${!plan[@]}"; do check_root "$index"; done
declare -a batch_paths=() batch_ids=() batch_facts=()
batch_bytes=0
repair_batch() {
  ((${#batch_paths[@]})) || return 0
  local item device inode mode owner kind
  check_root "$index"
  stat_records batch_facts "${batch_paths[@]}"
  for item in "${!batch_paths[@]}"; do
    [[ "${batch_facts[item]}" =~ ^[0-9]+:[0-9]+:[0-7]+:[0-9]+:[0-9a-f]+$ ]] || cleanup_fail 12 'cleanup_error: invalid repair identity'
    IFS=: read -r device inode mode owner kind <<<"${batch_facts[item]}"
    if [[ "$device:$inode:$mode:$owner" != "${batch_ids[item]}" ]] || (( (16#$kind & 0170000) != 0040000 )); then
      cleanup_fail 12 'cleanup_error: repair identity changed'
    fi
    check_ancestors "${batch_paths[item]}"
  done
  run_child chmod u+rwx -- "${batch_paths[@]}" || cleanup_fail 12 'cleanup_error: permission repair failed'
  batch_paths=(); batch_ids=(); batch_facts=(); batch_bytes=0
}
repair_tree() {
  local relative="$1" descriptor child identity object
  check_root "$index"
  open_stream find -P "$relative" -xdev -type d ! -perm -0700 -printf '%D:%i:%m:%U\0%p\0'
  descriptor=$stream_fd; child=$stream_pid
  identity=''
  while IFS= read -r -d '' identity <&"$descriptor"; do
    IFS= read -r -d '' object <&"$descriptor" || cleanup_fail 12 'cleanup_error: truncated repair pathname'
    [[ "$identity" =~ ^[0-9]+:[0-9]+:[0-7]+:[0-9]+$ && ( "$object" == "$relative" || "$object" == "$relative/"* ) ]] || cleanup_fail 12 'cleanup_error: invalid repair record'
    ((${#object}+1 <= 32768)) || cleanup_fail 12 'cleanup_error: repair pathname exceeds batch bound'
    if ((${#batch_paths[@]} >= 128 || batch_bytes+${#object}+1 > 32768)); then repair_batch; fi
    batch_paths+=("$object"); batch_ids+=("$identity"); batch_bytes=$((batch_bytes+${#object}+1))
    identity=''
  done
  [[ -z "$identity" ]] || cleanup_fail 12 'cleanup_error: truncated repair identity'
  finish_stream "$descriptor" "$child" || cleanup_fail 12 'cleanup_error: repair enumeration failed'
  repair_batch
}
trap 'cleanup_fail 12 "cleanup_error: command failed"' ERR
cleanup_execution=1
phase=repairing
progress
for index in "${!plan[@]}"; do
  active_root="${plan_ids[index]}"
  if ((repair_needed[index])); then repair_tree "${plan[index]}"; fi
done
phase=deleting
progress
for index in "${!plan[@]}"; do
  relative="${plan[index]}"
  active_root="${plan_ids[index]}"
  check_root "$index"
  case "${plan_states[index]}" in
    symlink)
      run_child rm -- "$relative" || cleanup_fail 12 'cleanup_error: unlink failed'
      ;;
    directory)
      run_child rm -rf --one-file-system -- "$relative" || cleanup_fail 12 'cleanup_error: removal failed'
      ;;
  esac
  [[ ! -e "$relative" && ! -L "$relative" ]] || cleanup_fail 12 'cleanup_error: removed root still present'
  printf 'CLEANED %q proof=workspace-layout:%s\n' "$root/$relative" "${plan_ids[index]}"
  completed=$((completed+1))
done
phase=complete
active_root=none
progress
