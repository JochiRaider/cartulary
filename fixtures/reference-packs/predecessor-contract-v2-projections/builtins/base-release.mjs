import { readFileSync, writeFileSync, renameSync } from "node:fs";
import { createHash } from "node:crypto";

const canonical = value => Array.isArray(value) ? `[${value.map(canonical).join(",")}]` : value !== null && typeof value === "object" ? `{${Object.keys(value).sort().map(key => `${JSON.stringify(key)}:${canonical(value[key])}`).join(",")}}` : JSON.stringify(value);
const sha = value => createHash("sha256").update(value).digest("hex");
const read = file => JSON.parse(readFileSync(file, "utf8"));
const check = (ok, message) => { if (!ok) throw new Error(`Base release: ${message}`); };
const input = read("tools/reference-pack-modules/base-release-input.v3.json");
check(input.schema_id === "cartulary.reference_pack_base_release_input.v3" && input.predecessor_release === "fixtures/reference-packs/predecessor-contract-v2-projections/builtins/release.v2.json" && input.source_profile === "contracts/reference-packs/builtins/source-profile.v2.json", "unsupported release inputs");
const old = read(input.predecessor_release);
check(sha(canonical(old)) === input.predecessor_release_sha256, "predecessor identity changed");
const profile = read(input.source_profile), profileSHA = sha(canonical(profile));
check(profile.pack_contract_version === "cartulary.reference_pack_contract.v3", "source profile contract");
const builderSHA = sha(Buffer.concat([Buffer.from("cartulary.base_pack_builder.v3\0"), Buffer.from(canonical(input)), Buffer.from([0]), readFileSync("tools/reference-pack-modules/base-release.mjs")]));
const packs = old.packs.map(pack => {
  const manifest = JSON.parse(pack.members["manifest.json"]);
  const key = manifest.pack_key;
  check(Object.hasOwn(input.payloads, key) && input.payloads[key] === pack.binding.payload_sha256, "payload identity changed");
  for (const file of manifest.files) {
    const raw = Buffer.from(pack.members[file.path]);
    check(raw.length === file.size_bytes && sha(raw) === file.sha256, "retained logical member changed");
  }
  Object.assign(manifest, {
    schema_id: "cartulary.reference_pack_manifest.v2",
    pack_contract_version: "cartulary.reference_pack_contract.v3",
    pack_version: input.pack_version,
    pack_release_sequence: input.pack_release_sequence,
    built_at: input.built_at,
    source_profile_id: profile.source_profile_id,
    source_profile_sha256: profileSHA,
    source_version: profile.source_version,
    source_artifacts: [{ source_ref: profile.source_profile_id, source_version: profile.source_version, sha256: profileSHA }],
    builder: { builder_id: "cartulary.base_pack_builder", builder_version: "3", builder_source_sha256: builderSHA },
  });
  const raw = canonical(manifest);
  return {
    binding: { ...pack.binding, pack_version: manifest.pack_version, pack_release_sequence: manifest.pack_release_sequence, manifest_sha256: sha(raw) },
    members: { ...pack.members, "manifest.json": raw },
  };
});
check(packs.length === 3 && new Set(packs.map(pack => pack.binding.pack_key)).size === 3, "required registry cardinality");
const bytes = canonical({ schema_id: "cartulary.reference_pack_builtin_release.v1", application_release_id: input.application_release_id, source_profile_sha256: profileSHA, packs });
const output = "contracts/reference-packs/builtins/release.v3.json";
let unchanged = false;
try { unchanged = readFileSync(output, "utf8") === bytes; } catch (error) { if (error.code !== "ENOENT") throw error; }
if (!unchanged) {
  const temporary = `${output}.projection-${process.pid}`;
  writeFileSync(temporary, bytes, { mode: 0o644, flag: "wx" });
  renameSync(temporary, output);
}
