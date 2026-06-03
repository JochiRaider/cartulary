#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"

NODE_VERSION="${NODE_VERSION:-24.15.0}"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
ARCHIVE_DIR="${CARTULARY_NODE_ARCHIVE_DIR:-${ROOT_DIR}/tmp/node-archives}"
PLATFORM_OVERRIDE="${CARTULARY_BOOTSTRAP_NODE_PLATFORM:-}"

detect_platform() {
  if [[ -n "${PLATFORM_OVERRIDE}" ]]; then
    case "${PLATFORM_OVERRIDE}" in
      linux-x64|linux-arm64|darwin-x64|darwin-arm64)
        printf '%s\n' "${PLATFORM_OVERRIDE}"
        return 0
        ;;
      *)
        echo "unsupported Node bootstrap platform: ${PLATFORM_OVERRIDE}; supported platforms are linux-x64, linux-arm64, darwin-x64, darwin-arm64" >&2
        return 1
        ;;
    esac
  fi

  local os
  local arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "${os}:${arch}" in
    Linux:x86_64|Linux:amd64) printf 'linux-x64\n' ;;
    Linux:aarch64|Linux:arm64) printf 'linux-arm64\n' ;;
    Darwin:x86_64|Darwin:amd64) printf 'darwin-x64\n' ;;
    Darwin:aarch64|Darwin:arm64) printf 'darwin-arm64\n' ;;
    *)
      echo "unsupported Node bootstrap platform: os=${os} arch=${arch}; supported platforms are linux-x64, linux-arm64, darwin-x64, darwin-arm64" >&2
      return 1
      ;;
  esac
}

expected_sha256() {
  local version="$1"
  local platform="$2"

  case "${version}:${platform}" in
    24.15.0:darwin-arm64) printf 'af5cfaeafe603aaf7599f287fd9d100bb41f16794f49788fa59dd3f25546930f\n' ;;
    24.15.0:darwin-x64) printf '5d627245b9f53cb2512cc21b7aa6aad693106affadd91e0c8f42d600fb7ba444\n' ;;
    24.15.0:linux-arm64) printf 'f3d5a797b5d210ce8e2cb265544c8e482eaedcb8aa409a8b46da7e8595d0dda0\n' ;;
    24.15.0:linux-x64) printf '472655581fb851559730c48763e0c9d3bc25975c59d518003fc0849d3e4ba0f6\n' ;;
    *)
      echo "missing committed Node checksum for node-v${version}-${platform}.tar.xz" >&2
      return 1
      ;;
  esac
}

sha256_file() {
  local file="$1"

  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "${file}" | awk '{print $1}'
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "${file}" | awk '{print $1}'
    return 0
  fi

  echo "sha256sum or shasum is required to verify Node archive checksums" >&2
  return 1
}

download_archive() {
  local archive="$1"
  local url="$2"

  mkdir -p "$(dirname "${archive}")"
  if ! curl -fsSL -o "${archive}" "${url}"; then
    rm -f "${archive}"
    return 1
  fi
}

install_archive() {
  local archive="$1"
  local temp_runtime

  mkdir -p "$(dirname "${NODE_RUNTIME_DIR}")"
  temp_runtime="$(mktemp -d "${NODE_RUNTIME_DIR}.tmp.XXXXXX")"
  trap 'rm -rf "${temp_runtime}"' RETURN
  tar -xJf "${archive}" -C "${temp_runtime}" --strip-components=1
  rm -rf "${NODE_RUNTIME_DIR}"
  mv "${temp_runtime}" "${NODE_RUNTIME_DIR}"
  trap - RETURN
}

main() {
  local node_bin="${NODE_RUNTIME_DIR}/bin/node"
  if [[ "${CARTULARY_FORCE_REINSTALL:-0}" != "1" && -x "${node_bin}" && "$("${node_bin}" -v)" == "v${NODE_VERSION}" ]]; then
    return 0
  fi

  local platform
  platform="$(detect_platform)"

  local expected
  expected="$(expected_sha256 "${NODE_VERSION}" "${platform}")"

  local archive_name="node-v${NODE_VERSION}-${platform}.tar.xz"
  local archive="${ARCHIVE_DIR}/${archive_name}"
  local url="https://nodejs.org/dist/v${NODE_VERSION}/${archive_name}"

  local actual
  if [[ -f "${archive}" ]]; then
    actual="$(sha256_file "${archive}")"
    if [[ "${actual}" != "${expected}" ]]; then
      rm -f "${archive}"
    fi
  fi
  if [[ ! -f "${archive}" ]]; then
    download_archive "${archive}" "${url}"
  fi

  actual="$(sha256_file "${archive}")"
  if [[ "${actual}" != "${expected}" ]]; then
    rm -f "${archive}"
    echo "Node archive checksum mismatch for ${archive_name}: expected ${expected}, got ${actual}; archive removed before extraction" >&2
    return 1
  fi

  install_archive "${archive}"
}

main "$@"
