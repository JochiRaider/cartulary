#!/usr/bin/env bash
# Installed verifier: only baseline WSL tools and Docker; no checkout or compiler.
set -euo pipefail
release="${1:?release directory required}"
mode="${2:-files}"
fail() { echo 'cartulary release: integrity, inventory, receipt, or platform admission failed' >&2; exit 2; }
[[ "$release" == /* && "$(realpath -e -- "$release")" == "$release" && -d "$release" ]] || fail
manifest="$release/release-manifest.json"
[[ -f "$manifest" && ! -L "$manifest" && "$(stat -c %s "$manifest")" -le 16777216 ]] || fail
jq -e '
  (keys == ["assets","binaries","images","platform","schema_id"]) and
  .schema_id == "cartulary.local_release_manifest.v1" and .platform == "linux/amd64" and
  (.assets | length > 0 and length <= 4096) and
  (.binaries | keys == ["migrate","operator","server"]) and
  (.images | map(.role) == ["application","postgres","seaweedfs"]) and
  (all(.assets[]; (keys == ["mode","path","sha256","size_bytes"]) and
    (.path | test("^[A-Za-z0-9_.-]+(/[A-Za-z0-9_.-]+)*$")) and
    (.path | split("/") | all(. != "." and . != "..")) and
    (.sha256 | test("^[a-f0-9]{64}$")) and
    (.size_bytes | type == "number" and . >= 0 and floor == .) and
    (.mode == "644" or .mode == "755"))) and
  ((.assets | map(.path)) == (.assets | map(.path) | sort | unique)) and
  all(.images[]; (keys == ["archive","config_sha256","id","inventory","reference","role"]) and
    (.id | test("^sha256:[a-f0-9]{64}$")) and (.config_sha256 | test("^[a-f0-9]{64}$")) and
    (.reference | test("^[a-z0-9./:_@-]+$")) and
    .archive == ("images/" + .role + ".tar") and .inventory == ("inventories/" + .role + ".syft.json")) and
  all(.binaries | to_entries[]; (.value | keys == ["receipt","sha256"]) and (.value.sha256 | test("^[a-f0-9]{64}$")) and .value.receipt == ("receipts/" + .key + ".crypto.json"))
' "$manifest" >/dev/null || fail
[[ -z "$(find "$release" -type l -o -type f -links +1 -o \( ! -type d ! -type f \))" ]] || fail
while IFS=$'\t' read -r member bytes digest permissions; do
  file="$release/$member"
  [[ -f "$file" && "$(stat -c %s -- "$file")" == "$bytes" && "$(stat -c %a -- "$file")" == "$permissions" ]] || fail
  [[ "$(sha256sum -- "$file" | cut -d ' ' -f 1)" == "$digest" ]] || fail
done < <(jq -r '.assets[] | [.path,.size_bytes,.sha256,.mode] | @tsv' "$manifest")
# Exact file membership, including the manifest itself; no unbound executable.
expected="$( { jq -r '.assets[].path' "$manifest"; echo release-manifest.json; } | LC_ALL=C sort)"
actual="$(find "$release" -type f -printf '%P\n' | LC_ALL=C sort)"
[[ "$expected" == "$actual" ]] || fail
while IFS=$'\t' read -r digest receipt; do
  jq -e --arg digest "$digest" '.schema_id == "cartulary.cryptographic_build_receipt.v1" and .binary_sha256 == $digest and (.selector | length > 0)' "$release/$receipt" >/dev/null || fail
done < <(jq -r '.binaries | to_entries[] | [.value.sha256,.value.receipt] | @tsv' "$manifest")
jq -e '.schema_id == "cartulary.image_inventory_tool.v1" and .module == "github.com/anchore/syft" and .version == "v1.44.0" and (.binary_sha256 | test("^[a-f0-9]{64}$"))' "$release/inventories/scanner.json" >/dev/null || fail
jq -e '.schema_id == "cartulary.distribution_notices.v1" and .review_disposition == "unestablished" and (.images | length == 3)' "$release/inventories/notices.json" >/dev/null || fail
while IFS=$'\t' read -r config_digest inventory archive; do
  [[ -f "$release/$archive" ]] || fail
  jq -e --arg id "sha256:$config_digest" '.descriptor.name == "syft" and (.descriptor.version == "1.44.0" or .descriptor.version == "[not provided]") and .source.type == "image" and .source.metadata.imageID == $id and (.artifacts | length > 0)' "$release/$inventory" >/dev/null || fail
done < <(jq -r '.images[] | [.config_sha256,.inventory,.archive] | @tsv' "$manifest")
[[ "$mode" == files || "$mode" == images ]] || fail
if [[ "$mode" == images ]]; then
  [[ "$(docker info --format '{{.OSType}}/{{.Architecture}}')" =~ ^linux/(amd64|x86_64)$ ]] || fail
  while IFS=$'\t' read -r id reference; do
    [[ "$(docker image inspect --format '{{.Id}} {{.Os}}/{{.Architecture}}' "$id")" == "$id linux/amd64" ]] || fail
    # A content-named tag is immutable even when the local daemon is contaminated.
    if tagged="$(docker image inspect --format '{{.Id}}' "$reference" 2>/dev/null)"; then [[ "$tagged" == "$id" ]] || fail; fi
  done < <(jq -r '.images[] | [.id,.reference] | @tsv' "$manifest")
fi
sha256sum "$manifest" | cut -d ' ' -f 1
