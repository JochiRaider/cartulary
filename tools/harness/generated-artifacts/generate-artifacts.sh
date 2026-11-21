#!/usr/bin/env bash
set -euo pipefail

# Public generation renders every generated artifact once into private scratch,
# validates the render's closure, compares it and atomically publishes only the
# changed outputs (TH-HARNESS-REQ-004, REQ-261). A failed render leaves the
# tree untouched. The step list lives in render-generated-artifacts.sh.
exec "$(dirname -- "${BASH_SOURCE[0]}")/check-generate-drift.sh" --refresh
