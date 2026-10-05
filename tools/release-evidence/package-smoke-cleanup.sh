#!/usr/bin/env bash
# Sourced by disposable package checks. Only exact, test-owned Compose labels
# and the allocation's unique image tag authorize cleanup.
remove_package_workspace() {
  local owned_project="$1" workspace="$2" helper_image="$3"
  [[ "$owned_project" =~ ^cartularymvpsmoke[0-9]+(destination)?$ || "$owned_project" =~ ^cartularymvprecoverysmk[0-9]+$ ]] || return 1
  case "$(basename "$workspace")" in cartulary-standup-package-smoke.*|cartulary-standup-recovery-smoke.*) ;; *) return 1 ;; esac
  [[ -d "$workspace" && ! -L "$workspace" ]] || return 1
  if [[ -d "$workspace/runtime" ]]; then
    # Recovery deliberately writes private roots as container UID 65532.
    # Restore ownership of this disposable allocation before host deletion.
    docker run --rm --label "com.docker.compose.project=${owned_project}" \
      --user 0:0 --entrypoint sh --mount "type=bind,source=${workspace}/runtime,target=/cleanup" \
      "$helper_image" -c 'chown -R "$1" /cleanup' sh "$(id -u):$(id -g)" >/dev/null 2>&1 || return 1
  fi
  rm -rf "$workspace"
  [[ ! -e "$workspace" ]]
}

cleanup_package_workspace() {
  local result=passed
  remove_package_workspace "$1" "$2" "$3" || result=failed
  printf '{"cleanup":"%s"}\n' "$result" >"$4" || return 1
  [[ "$result" == passed ]]
}

cleanup_package_resources() {
  local owned_project="$1" owned_image="$2" evidence="$3"
  local errors=0 remaining=0 kind ids id
  [[ "$owned_project" =~ ^cartularymvpsmoke[0-9]+(destination)?$ || "$owned_project" =~ ^cartularymvprecoverysmk[0-9]+$ ]] || return 1
  if [[ -n "$owned_image" && "${owned_image##*:}" != "$owned_project" ]]; then return 1; fi
  for kind in container volume network; do
    local listing=(docker "$kind" ls -q --filter "label=com.docker.compose.project=${owned_project}")
    if [[ "$kind" == container ]]; then listing+=(--all); fi
    if ! ids="$("${listing[@]}" 2>/dev/null)"; then errors=$((errors + 1)); continue; fi
    for id in $ids; do
      if [[ ! "$id" =~ ^[a-zA-Z0-9_.-]+$ ]]; then errors=$((errors + 1)); continue; fi
      local removal=(docker "$kind" rm)
      if [[ "$kind" == container ]]; then removal+=(--force); fi
      if ! "${removal[@]}" "$id" >/dev/null 2>&1; then errors=$((errors + 1)); fi
    done
  done
  if [[ -n "$owned_image" ]]; then
    if ! ids="$(docker image ls -q --filter "reference=${owned_image}" 2>/dev/null)"; then
      errors=$((errors + 1))
    elif [[ -n "$ids" ]] && ! docker image rm "$owned_image" >/dev/null 2>&1; then
      errors=$((errors + 1))
    fi
    if ! ids="$(docker image ls -q --filter "reference=${owned_image}" 2>/dev/null)"; then
      errors=$((errors + 1))
    elif [[ -n "$ids" ]]; then remaining=$((remaining + 1)); fi
  fi
  for kind in container volume network; do
    local listing=(docker "$kind" ls -q --filter "label=com.docker.compose.project=${owned_project}")
    if [[ "$kind" == container ]]; then listing+=(--all); fi
    if ! ids="$("${listing[@]}" 2>/dev/null)"; then
      errors=$((errors + 1))
    elif [[ -n "$ids" ]]; then remaining=$((remaining + 1)); fi
  done
  local result=passed
  if ((errors != 0 || remaining != 0)); then result=failed; fi
  mkdir -p "$(dirname "$evidence")" || return 1
  printf '{"project":"%s","cleanup":"%s","attempt_errors":%d,"remaining_resource_groups":%d}\n' \
    "$owned_project" "$result" "$errors" "$remaining" >"$evidence" || return 1
  [[ "$result" == passed ]]
}
