#!/usr/bin/env bash
# shellcheck shell=bash

BROWSER_LIFECYCLE_ADAPTER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BROWSER_LIFECYCLE_ADAPTER_ROOT="$(cd "${BROWSER_LIFECYCLE_ADAPTER_DIR}/../../.." && pwd)"

# shellcheck source=tools/harness/execution/phase-runtime.sh
source "${BROWSER_LIFECYCLE_ADAPTER_ROOT}/tools/harness/execution/phase-runtime.sh"
# shellcheck source=tools/harness/browser/web-e2e-lifecycle.sh
source "${BROWSER_LIFECYCLE_ADAPTER_ROOT}/tools/harness/browser/web-e2e-lifecycle.sh"
# shellcheck source=tools/harness/browser/lifecycle/ports-and-token.sh
source "${BROWSER_LIFECYCLE_ADAPTER_ROOT}/tools/harness/browser/lifecycle/ports-and-token.sh"
# shellcheck source=tools/harness/browser/lifecycle/reset-route.sh
source "${BROWSER_LIFECYCLE_ADAPTER_ROOT}/tools/harness/browser/lifecycle/reset-route.sh"
