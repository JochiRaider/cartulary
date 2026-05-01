#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"

SHELLCHECK_VERSION="${SHELLCHECK_VERSION:-0.11.0}"
TOOLBIN_DIR="${TOOLBIN_DIR:-${ROOT_DIR}/tmp/toolbin}"
SHELLCHECK_BIN="${SHELLCHECK_BIN:-${TOOLBIN_DIR}/shellcheck-v${SHELLCHECK_VERSION}}"
ARCHIVE_DIR="${CARTULARY_SHELLCHECK_ARCHIVE_DIR:-${ROOT_DIR}/tmp/shellcheck-archives}"
PLATFORM_OVERRIDE="${CARTULARY_BOOTSTRAP_SHELLCHECK_PLATFORM:-}"

detect_platform() {
  if [[ -n "$PLATFORM_OVERRIDE" ]]; then
    case "$PLATFORM_OVERRIDE" in
      linux-x64 | linux-arm64 | darwin-x64 | darwin-arm64)
        printf '%s\n' "$PLATFORM_OVERRIDE"
        return 0
        ;;
      *)
        echo "unsupported ShellCheck bootstrap platform: ${PLATFORM_OVERRIDE}; supported platforms are linux-x64, linux-arm64, darwin-x64, darwin-arm64" >&2
        return 1
        ;;
    esac
  fi

  local os
  local arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "${os}:${arch}" in
    Linux:x86_64 | Linux:amd64) printf 'linux-x64\n' ;;
    Linux:aarch64 | Linux:arm64) printf 'linux-arm64\n' ;;
    Darwin:x86_64 | Darwin:amd64) printf 'darwin-x64\n' ;;
    Darwin:aarch64 | Darwin:arm64) printf 'darwin-arm64\n' ;;
    *)
      echo "unsupported ShellCheck bootstrap platform: os=${os} arch=${arch}; supported platforms are linux-x64, linux-arm64, darwin-x64, darwin-arm64" >&2
      return 1
      ;;
  esac
}

archive_platform() {
  local platform="$1"

  case "$platform" in
    linux-x64) printf 'linux.x86_64\n' ;;
    linux-arm64) printf 'linux.aarch64\n' ;;
    darwin-x64) printf 'darwin.x86_64\n' ;;
    darwin-arm64) printf 'darwin.aarch64\n' ;;
    *)
      echo "unsupported ShellCheck archive platform: ${platform}" >&2
      return 1
      ;;
  esac
}

expected_sha256() {
  local version="$1"
  local platform="$2"

  case "${version}:${platform}" in
    0.11.0:darwin.aarch64) printf '56affdd8de5527894dca6dc3d7e0a99a873b0f004d7aabc30ae407d3f48b0a79\n' ;;
    0.11.0:darwin.x86_64) printf '3c89db4edcab7cf1c27bff178882e0f6f27f7afdf54e859fa041fca10febe4c6\n' ;;
    0.11.0:linux.aarch64) printf '12b331c1d2db6b9eb13cfca64306b1b157a86eb69db83023e261eaa7e7c14588\n' ;;
    0.11.0:linux.x86_64) printf '8c3be12b05d5c177a04c29e3c78ce89ac86f1595681cab149b65b97c4e227198\n' ;;
    *)
      echo "missing committed ShellCheck checksum for shellcheck-v${version}.${platform}.tar.xz" >&2
      return 1
      ;;
  esac
}

sha256_file() {
  local file="$1"

  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
    return 0
  fi

  echo "sha256sum or shasum is required to verify ShellCheck archive checksums" >&2
  return 1
}

installed_version() {
  local shellcheck_bin="$1"

  "$shellcheck_bin" --version 2>/dev/null | awk -F': ' '$1 == "version" { print $2; exit }'
}

download_archive() {
  local archive="$1"
  local url="$2"

  mkdir -p "$(dirname "$archive")"
  if ! curl -fsSL -o "$archive" "$url"; then
    rm -f "$archive"
    return 1
  fi
}

install_archive() {
  local archive="$1"
  local temp_dir
  local extracted

  mkdir -p "$TOOLBIN_DIR"
  temp_dir="$(mktemp -d "${TOOLBIN_DIR}/shellcheck-install.XXXXXX")"
  trap 'rm -rf "$temp_dir"' RETURN
  tar -xJf "$archive" -C "$temp_dir"
  extracted="${temp_dir}/shellcheck-v${SHELLCHECK_VERSION}/shellcheck"
  if [[ ! -x "$extracted" ]]; then
    echo "ShellCheck archive did not contain executable shellcheck at shellcheck-v${SHELLCHECK_VERSION}/shellcheck" >&2
    return 1
  fi
  cp "$extracted" "$SHELLCHECK_BIN"
  chmod 0755 "$SHELLCHECK_BIN"
  trap - RETURN
  rm -rf "$temp_dir"
}

main() {
  if [[ -x "$SHELLCHECK_BIN" && "$(installed_version "$SHELLCHECK_BIN")" == "$SHELLCHECK_VERSION" ]]; then
    return 0
  fi

  local platform
  local archive_platform_value
  local expected
  local archive_name
  local archive
  local url
  local actual

  platform="$(detect_platform)"
  archive_platform_value="$(archive_platform "$platform")"
  expected="$(expected_sha256 "$SHELLCHECK_VERSION" "$archive_platform_value")"
  archive_name="shellcheck-v${SHELLCHECK_VERSION}.${archive_platform_value}.tar.xz"
  archive="${ARCHIVE_DIR}/${archive_name}"
  url="https://github.com/koalaman/shellcheck/releases/download/v${SHELLCHECK_VERSION}/${archive_name}"

  download_archive "$archive" "$url"

  actual="$(sha256_file "$archive")"
  if [[ "$actual" != "$expected" ]]; then
    rm -f "$archive"
    echo "ShellCheck archive checksum mismatch for ${archive_name}: expected ${expected}, got ${actual}; archive removed before extraction" >&2
    return 1
  fi

  install_archive "$archive"
}

main "$@"
