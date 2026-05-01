#!/usr/bin/env bash

WEB_E2E_LIFECYCLE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/process-lifecycle.sh
source "${WEB_E2E_LIFECYCLE_DIR}/process-lifecycle.sh"
