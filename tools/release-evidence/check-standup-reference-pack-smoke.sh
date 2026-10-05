#!/usr/bin/env bash
set -euo pipefail
export CARTULARY_PACKAGE_QUALIFICATION=reference-pack
exec "$(dirname -- "$0")/check-standup-package-smoke.sh"
