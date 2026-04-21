#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

mapfile -t phase_tests < <(
  rg --no-filename '^func (TestPhase[0-9][A-Za-z0-9_]*)\(' internal cmd/server -g '*_test.go' \
    | sed -E 's/^func (TestPhase[0-9][A-Za-z0-9_]*)\(.*/\1/'
)

invalid=()
for name in "${phase_tests[@]}"; do
  case "$name" in
    TestPhase0_*)
      if [[ ! "$name" =~ (_U_0_|_I_0_|_E_0_) ]]; then
        invalid+=("$name")
      fi
      ;;
    TestPhase1_*)
      if [[ ! "$name" =~ (^TestPhase1_.*_(U|I)_1_[0-9]{2}$|^TestPhase1_.*_ProcessSmoke$) ]]; then
        invalid+=("$name")
      fi
      ;;
    TestPhase2_*)
      if [[ ! "$name" =~ (_U_2_|_I_2_|^TestPhase2_ProcessSmoke_) ]]; then
        invalid+=("$name")
      fi
      ;;
    TestPhase3_*)
      if [[ ! "$name" =~ (_U_3_|_I_3_) ]]; then
        invalid+=("$name")
      fi
      ;;
    TestPhase4_*)
      if [[ ! "$name" =~ (_U_4_|_I_4_) ]]; then
        invalid+=("$name")
      fi
      ;;
    *)
      invalid+=("$name")
      ;;
  esac
done

if [[ ${#invalid[@]} -gt 0 ]]; then
  {
    echo "Phase test names must carry their layer token so the repo keeps a stable naming contract alongside the executable phase manifests."
    echo "Invalid names:"
    printf '  %s\n' "${invalid[@]}"
  } >&2
  exit 1
fi
