#!/usr/bin/env bash
set -euo pipefail

fixture_id="${FIXTURE:-}"
case "${fixture_id}" in
  GP-FIX-[0-9][0-9][0-9]) ;;
  "")
    echo "FIXTURE=GP-FIX-NNN is required" >&2
    exit 2
    ;;
  *)
    echo "invalid FIXTURE=${fixture_id}" >&2
    exit 2
    ;;
esac

if [[ -z "${CARTULARY_TEST_RESULTS_DIR:-}" ]]; then
  echo "CARTULARY_TEST_RESULTS_DIR is required" >&2
  exit 2
fi

go_bin="${GO:-go}"
exec env \
  GOCACHE="${GO_CACHE_DIR:?GO_CACHE_DIR is required}" \
  GOMODCACHE="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}" \
  GOTMPDIR="${GO_TMP_DIR:?GO_TMP_DIR is required}" \
  GRAPH_PROJECTION_FIXTURE="${fixture_id}" \
  "${go_bin}" test ./internal/modules/graphprojection \
  -run '^TestGraphProjectionFixtureCandidate$' \
  -count=1
