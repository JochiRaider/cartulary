#!/usr/bin/env bash
set -euo pipefail

make_bin="${MAKE_BIN:-${MAKE:-make}}"

"$make_bin" --no-print-directory toolchain-drift
"$make_bin" --no-print-directory codegen-toolchain
"$make_bin" --no-print-directory go-lint-toolchain
"$make_bin" --no-print-directory shell-lint-toolchain
"$make_bin" --no-print-directory go-security-toolchain
if [[ "${CI:-}" == "1" ]]; then
  "$make_bin" --no-print-directory frontend-install-ci
else
  "$make_bin" --no-print-directory frontend-install
fi
