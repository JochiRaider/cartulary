#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="${ROOT_DIR}/tools/harness/readiness/frontend-install.sh"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "${ROOT_DIR}/tools/harness/test-support/harness-scratch.sh"

scratch="$(cartulary_harness_mktemp_dir "frontend-install-self-hosting.XXXXXX")"
trap 'rm -rf "$scratch"' EXIT

fake_pnpm="${scratch}/pnpm"
install_marker="${scratch}/install-ran"
stamp="${scratch}/frontend-install.stamp"

cat >"${fake_pnpm}" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
case "$*" in
  "config get store-dir") printf '.pnpm-store\n' ;;
  "config get confirmModulesPurge") printf 'false\n' ;;
  install*)
    [[ " $* " == *" --frozen-lockfile "* ]]
    [[ "${CI:-}" == "true" ]]
    : >"${FAKE_INSTALL_MARKER:?}"
    ;;
  *) printf 'unexpected fake pnpm invocation: %s\n' "$*" >&2; exit 99 ;;
esac
EOF
chmod +x "${fake_pnpm}"

env -u RUN_STEP_SCRIPT \
  FAKE_INSTALL_MARKER="${install_marker}" \
  FRONTEND_INSTALL_STAMP="${stamp}" \
  NODE_BIN="${scratch}/node" \
  NODE_VERSION="24.15.0" \
  PNPM="${fake_pnpm}" \
  PNPM_VERSION="10.33.0" \
  PNPM_INSTALL_FLAGS="--reporter=append-only --loglevel=warn" \
  "${SCRIPT}"

[[ -f "${install_marker}" ]] || {
  echo "frontend install did not execute without RUN_STEP_SCRIPT" >&2
  exit 1
}
[[ -f "${stamp}" ]] || {
  echo "frontend install did not publish its readiness stamp" >&2
  exit 1
}
grep -Fq "pnpm_store_dir=.pnpm-store" "${stamp}" || {
  echo "frontend install stamp omitted the pinned store" >&2
  exit 1
}
