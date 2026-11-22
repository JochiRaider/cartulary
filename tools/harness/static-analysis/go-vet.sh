#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"

# shellcheck source=tools/harness/generated-artifacts/generated-artifacts.sh
# shellcheck disable=SC1091
source "$ROOT_DIR/tools/harness/generated-artifacts/generated-artifacts.sh"

# shellcheck source=tools/harness/static-analysis/go-lint-modules.sh
# shellcheck disable=SC1091
source "$ROOT_DIR/tools/harness/static-analysis/go-lint-modules.sh"

cartulary_lint_go_modules "$ROOT_DIR" cartulary_filter_authored_go_packages "${GO:-go}" vet
