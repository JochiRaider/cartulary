#!/usr/bin/env bash
set -euo pipefail
umask 077
release="$(unset CDPATH && cd -- "$(dirname "$0")/../.." && pwd -P)"
verify="$release/assets/scripts/release-verify.sh"
fail() { echo "cartulary install: $1" >&2; exit 2; }
[[ $# == 1 && "$1" == /* && "$1" != */ && "$(realpath -m -- "$1")" == "$1" ]] || fail 'one canonical absolute destination is required'
destination="$1"
parent="$(dirname -- "$destination")"
[[ -d "$parent" && "$(realpath -e -- "$parent")" == "$parent" && ! -e "$destination" && ! -L "$destination" ]] || fail 'destination must be unoccupied with an existing canonical parent'
"$verify" "$release" files >/dev/null
[[ "$(docker info --format '{{.OSType}}/{{.Architecture}}')" =~ ^linux/(amd64|x86_64)$ ]] || fail 'linux/amd64 Docker Desktop backend is required'
while IFS=$'\t' read -r id reference archive; do
  if tagged="$(docker image inspect --format '{{.Id}}' "$reference" 2>/dev/null)"; then [[ "$tagged" == "$id" ]] || fail 'content-named image reference is occupied by different bytes'; fi
  docker image load --input "$release/$archive" >/dev/null
done < <(jq -r '.images[] | [.id,.reference,.archive] | @tsv' "$release/release-manifest.json")
"$verify" "$release" images >/dev/null
stage="$(mktemp -d "${parent}/.cartulary-install.XXXXXXXX")"
cleanup() { if [[ -n "$stage" && -d "$stage" ]]; then rm -rf -- "$stage"; fi; }
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
mkdir "$stage/release"
cp -a "$release/." "$stage/release/"
for example in "$stage/release/assets/"*.example "$stage/release/assets/.env.example"; do
  name="$(basename -- "$example" .example)"
  cp -- "$example" "$stage/$name"
  chmod 0600 "$stage/$name"
done
mkdir -m 0700 "$stage/runtime"
# Rename on the destination filesystem; no existing installation is replaced.
mv -T --no-clobber -- "$stage" "$destination"
[[ ! -e "$stage" ]] || fail 'destination became occupied during publication'
stage=
printf 'installed release %s\n' "$(sha256sum "$destination/release/release-manifest.json" | cut -d ' ' -f 1)"
