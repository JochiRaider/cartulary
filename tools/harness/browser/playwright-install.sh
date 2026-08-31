#!/usr/bin/env bash
set -euo pipefail

stamp="${PLAYWRIGHT_INSTALL_STAMP:?PLAYWRIGHT_INSTALL_STAMP is required}"
run_step="${RUN_STEP_SCRIPT:?RUN_STEP_SCRIPT is required}"
node_runtime_dir="${NODE_RUNTIME_DIR:?NODE_RUNTIME_DIR is required}"
node_bin="${NODE_BIN:?NODE_BIN is required}"
pnpm="${PNPM:?PNPM is required}"
node_version="${NODE_VERSION:?NODE_VERSION is required}"
pnpm_version="${PNPM_VERSION:?PNPM_VERSION is required}"
visual_renderer_image="mcr.microsoft.com/playwright@sha256:eac9b0a5312cdab40ee8c2429df5bf19bffdccf8f3bf3c42268e173f97541645"
visual_renderer_platform="linux/amd64"

detect_playwright_host_platform_override() {
  local arch
  local os_id
  local os_version
  local suffix

  if [[ -n "${PLAYWRIGHT_HOST_PLATFORM_OVERRIDE:-}" ]]; then
    printf '%s\n' "$PLAYWRIGHT_HOST_PLATFORM_OVERRIDE"
    return 0
  fi

  [[ "$(uname -s)" == "Linux" && -r /etc/os-release ]] || return 0
  arch="$(uname -m)"
  case "$arch" in
    x86_64) suffix="x64" ;;
    aarch64 | arm64) suffix="arm64" ;;
    *) return 0 ;;
  esac
  os_id="$(awk -F= '$1 == "ID" { gsub(/"/, "", $2); print $2 }' /etc/os-release)"
  os_version="$(awk -F= '$1 == "VERSION_ID" { gsub(/"/, "", $2); print $2 }' /etc/os-release)"
  if [[ "$os_id" == "ubuntu" && "$os_version" == 26.* ]]; then
    printf 'ubuntu24.04-%s\n' "$suffix"
  fi
}

playwright_host_platform_override="$(detect_playwright_host_platform_override)"

ubuntu_playwright_tool_packages=(
  xvfb
  fonts-noto-color-emoji
  fonts-unifont
  libfontconfig1
  libfreetype6
  xfonts-cyrillic
  xfonts-scalable
  fonts-liberation
  fonts-ipafont-gothic
  fonts-wqy-zenhei
  fonts-tlwg-loma-otf
  fonts-freefont-ttf
)

ubuntu_chromium_packages=(
  libasound2t64
  libatk-bridge2.0-0t64
  libatk1.0-0t64
  libatspi2.0-0t64
  libcairo2
  libcups2t64
  libdbus-1-3
  libdrm2
  libgbm1
  libglib2.0-0t64
  libnspr4
  libnss3
  libpango-1.0-0
  libx11-6
  libxcb1
  libxcomposite1
  libxdamage1
  libxext6
  libxfixes3
  libxkbcommon0
  libxrandr2
)

run_apt_get() {
  if [[ "$EUID" -eq 0 ]]; then
    apt-get "$@"
    return
  fi
  sudo -n apt-get "$@"
}

can_install_apt_packages() {
  command -v apt-get >/dev/null 2>&1 || return 1
  if [[ "$EUID" -eq 0 ]]; then
    return 0
  fi
  command -v sudo >/dev/null 2>&1 || return 1
  sudo -n true >/dev/null 2>&1
}

install_ubuntu_apt_fallback() {
  if [[ "${CARTULARY_PLAYWRIGHT_APT_FALLBACK:-1}" == "0" ]]; then
    echo "Playwright apt fallback disabled by CARTULARY_PLAYWRIGHT_APT_FALLBACK=0." >&2
    return 1
  fi
  if ! can_install_apt_packages; then
    echo "Playwright apt fallback skipped: apt-get requires root or passwordless sudo." >&2
    return 1
  fi

  echo "Falling back to Ubuntu apt packages for Playwright Chromium dependencies." >&2
  export DEBIAN_FRONTEND=noninteractive
  run_apt_get update
  run_apt_get install -y --no-install-recommends \
    "${ubuntu_playwright_tool_packages[@]}" \
    "${ubuntu_chromium_packages[@]}"
}

find_chromium_binary() {
  local browser_root="${PLAYWRIGHT_BROWSERS_PATH:-${HOME}/.cache/ms-playwright}"
  find "$browser_root" \
    \( -path '*/chrome-headless-shell-linux64/chrome-headless-shell' -o \
    -path '*/chrome-linux/chrome' \) \
    -type f -print 2>/dev/null | LC_ALL=C sort -V | tail -n 1
}

missing_shared_libraries() {
  local binary="$1"
  if ! command -v ldd >/dev/null 2>&1; then
    return 0
  fi
  ldd "$binary" | awk '/not found/ { print $1 }' | LC_ALL=C sort -u
}

run_playwright_cli() {
  local env_args=()
  if [[ -n "$playwright_host_platform_override" ]]; then
    env_args+=("PLAYWRIGHT_HOST_PLATFORM_OVERRIDE=$playwright_host_platform_override")
  fi
  env "${env_args[@]}" "$pnpm" --dir apps/web exec playwright "$@"
}

verify_chromium_native_deps() {
  local browser_binary
  local missing=()
  browser_binary="$(find_chromium_binary)"
  if [[ -z "$browser_binary" ]]; then
    echo "Could not find an installed Playwright Chromium binary to verify." >&2
    return 1
  fi

  mapfile -t missing < <(missing_shared_libraries "$browser_binary")
  if [[ "${#missing[@]}" -eq 0 ]]; then
    return 0
  fi

  echo "Playwright Chromium is missing native shared libraries:" >&2
  printf '  %s\n' "${missing[@]}" >&2
  echo "Install the Playwright OS dependency set, or rerun make playwright-install as root/passwordless sudo." >&2
  return 1
}

run_playwright_install() {
  if [[ -n "$playwright_host_platform_override" ]]; then
    echo "Using PLAYWRIGHT_HOST_PLATFORM_OVERRIDE=$playwright_host_platform_override for Playwright install." >&2
  fi

  if can_install_apt_packages; then
    if run_playwright_cli install --with-deps chromium; then
      verify_chromium_native_deps
      return
    fi

    echo "Playwright's --with-deps installer failed; attempting repo fallback and dependency verification." >&2
    install_ubuntu_apt_fallback || true
  else
    echo "Skipping Playwright --with-deps installer: apt-get requires root or passwordless sudo." >&2
  fi

  run_playwright_cli install chromium
  verify_chromium_native_deps
}

install_visual_renderer_image() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "Docker is required for the pinned frontend visual renderer." >&2
    return 1
  fi
  docker pull --platform "$visual_renderer_platform" "$visual_renderer_image" >/dev/null
  local architecture
  local os
  architecture="$(docker image inspect "$visual_renderer_image" --format '{{.Architecture}}')"
  os="$(docker image inspect "$visual_renderer_image" --format '{{.Os}}')"
  if [[ "$architecture" != "amd64" || "$os" != "linux" ]]; then
    echo "Visual renderer image platform mismatch: expected linux/amd64, got ${os}/${architecture}." >&2
    return 1
  fi
}

if [[ "${CARTULARY_PLAYWRIGHT_INSTALL_CHILD:-0}" == "1" ]]; then
  run_playwright_install
  install_visual_renderer_image
  exit $?
fi

mkdir -p "$(dirname "$stamp")"
"$run_step" "playwright-install" -- \
  env PATH="${node_runtime_dir}/bin:${PATH}" \
    COREPACK_HOME="${node_runtime_dir}/corepack" \
    CARTULARY_PLAYWRIGHT_INSTALL_CHILD=1 \
    PLAYWRIGHT_INSTALL_STAMP="$stamp" \
    RUN_STEP_SCRIPT="$run_step" \
    NODE_RUNTIME_DIR="$node_runtime_dir" \
    NODE_BIN="$node_bin" \
    PNPM="$pnpm" \
    NODE_VERSION="$node_version" \
    PNPM_VERSION="$pnpm_version" \
    "$0"
printf 'node_path=%s\nnode_version=v%s\npnpm_path=%s\npnpm_version=%s\nplaywright_install_args=%s\nnative_dependency_strategy=%s\nplaywright_host_platform_override=%s\nvisual_renderer_image=%s\nvisual_renderer_platform=%s\n' \
  "$node_bin" \
  "$node_version" \
  "$pnpm" \
  "$pnpm_version" \
  "install --with-deps chromium" \
  "playwright_with_deps,ubuntu_apt_fallback,ldd_verify" \
  "${playwright_host_platform_override:-}" \
  "$visual_renderer_image" \
  "$visual_renderer_platform" >"$stamp"
