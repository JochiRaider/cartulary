#!/usr/bin/env bash

WEB_E2E_LIFECYCLE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_E2E_LIFECYCLE_REPO_ROOT="$(cd "${WEB_E2E_LIFECYCLE_DIR}/../../.." && pwd)"
# shellcheck source=scripts/lib/process-lifecycle.sh
source "${WEB_E2E_LIFECYCLE_REPO_ROOT}/scripts/lib/process-lifecycle.sh"
