#!/usr/bin/env bash
set -euo pipefail

if [[ "${CARTULARY_FRONTEND_TOOLCHAIN_QUIET:-0}" != "1" ]]; then
  cat "${FRONTEND_TOOLCHAIN_STAMP:?FRONTEND_TOOLCHAIN_STAMP is required}"
fi
