// Disposable package-test producer. Never imported by runtime code or shipped.
import { createHash, createPrivateKey, createPublicKey, sign } from "node:crypto";
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { canonicalJSONString } from "../harness/contract/semantic-json.mjs";

const canonical = (value) => Buffer.from(canonicalJSONString(value));
const digest = (value) => createHash("sha256").update(value).digest("hex");
const repository = "package_smoke.repo";

export function packageFixtures(repoRoot, destination) {
  mkdirSync(destination, { recursive: true });
  const day = new Date();
  day.setUTCHours(0, 0, 0, 0);
  const expiry = (days) => new Date(+day + days * 86400000).toISOString().replace(".000Z", "Z");
  const keys = Array.from({ length: 3 }, (_, i) => {
    const seed = createHash("sha256").update(`cartulary package smoke test key only ${i}`).digest();
    const privateKey = createPrivateKey({ key: Buffer.concat([Buffer.from("302e020100300506032b657004220420", "hex"), seed]), format: "der", type: "pkcs8" });
    const publicHex = createPublicKey(privateKey).export({ format: "der", type: "spki" }).subarray(-32).toString("hex");
    const publicValue = { keytype: "ed25519", scheme: "ed25519", keyval: { public: publicHex } };
    return { privateKey, publicValue, id: digest(canonical(publicValue)) };
  });
  const signed = (value) => canonical({ signed: value, signatures: (value._type === "root" ? keys : keys.slice(0, 1)).map((key) => ({ keyid: key.id, sig: sign(null, canonical(value), key.privateKey).toString("hex") })).sort((a, b) => a.keyid.localeCompare(b.keyid)) });
  const root = (version) => signed({ _type: "root", spec_version: "1.0.35", version, expires: expiry(365), consistent_snapshot: false, keys: Object.fromEntries(keys.map((key) => [key.id, key.publicValue])), roles: { root: { keyids: keys.map((key) => key.id).sort(), threshold: 2 }, ...Object.fromEntries(["targets", "snapshot", "timestamp"].map((role) => [role, { keyids: [keys[0].id], threshold: 1 }])) }, cartulary: { schema_id: "cartulary.reference_pack_tuf_root_binding.v1", trust_repository_id: repository } });
  const bootstrap = canonical({ schema_id: "cartulary.reference_pack_trust_bootstrap.v1", repositories: [{ repository_id: repository, trusted_root: JSON.parse(root(1)), trusted_root_sha256: digest(root(1)) }] });
  writeFileSync(path.join(destination, "trust.json"), bootstrap, { mode: 0o644 });
  const base = JSON.parse(readFileSync(path.join(repoRoot, "contracts/reference-packs/builtins/release.v1.json")));
  const profiles = JSON.parse(readFileSync(path.join(repoRoot, "contracts/reference-pack-fixtures/fixtures/profiles.v1.json")));
  function bundle(name, { key = "enrichment.tor", version = 2, metadataVersion = version, rotation = false, badSignature = false, incompatible = false } = {}) {
    const source = base.packs.find((pack) => pack.binding.pack_key === key) ?? base.packs[0];
    const members = Object.fromEntries(Object.entries(source.members).map(([name, value]) => [name, Buffer.from(value)]));
    const manifest = JSON.parse(members["manifest.json"]);
    if (!key.startsWith("type_registry.")) {
      for (const name of Object.keys(members)) if (name.startsWith("payload/")) delete members[name];
      const profile = profiles.cases.find((profile) => profile.pack_key === key);
      for (const [name, value] of Object.entries(profile.members)) members[name] = Buffer.from(value);
    }
    Object.assign(manifest, { pack_key: key, pack_kind: key.split(".")[0], pack_version: String(version), pack_release_sequence: version + 1, trust_repository_id: repository, source_as_of: expiry(0), built_at: expiry(0), content_profile_id: `cartulary.reference_pack.${key}.v1` });
    if (incompatible) manifest.pack_contract_version = "999";
    manifest.files = Object.keys(members).filter((name) => name !== "manifest.json").sort().map((name) => ({ path: name, role: name.startsWith("notices/") ? "notice" : "payload", media_type: name.startsWith("notices/") ? "text/plain" : "application/x-ndjson", size_bytes: members[name].length, sha256: digest(members[name]) }));
    manifest.content_summary = { kind: "entries", entry_count: members["payload/entries.ndjson"].filter((byte) => byte === 10).length };
    const payload = createHash("sha256").update("cartulary.reference_pack.payload.v1\0");
    for (const file of manifest.files) {
      const length = Buffer.alloc(4); length.writeUInt32BE(Buffer.byteLength(file.path));
      const size = Buffer.alloc(8); size.writeBigUInt64BE(BigInt(file.size_bytes));
      payload.update(length).update(file.path).update(size).update(Buffer.from(file.sha256, "hex"));
    }
    members["manifest.json"] = canonical(manifest);
    members["bundle.json"] = canonical({ schema_id: "cartulary.reference_pack_bundle_hint.v1", trust_repository_id: repository });
    const targets = Object.fromEntries(Object.entries(members).map(([name, value]) => [name, { length: value.length, hashes: { sha256: digest(value) } }]));
    members["metadata/targets.json"] = signed({ _type: "targets", spec_version: "1.0.35", version: metadataVersion, expires: expiry(30), targets, cartulary: { schema_id: "cartulary.reference_pack_tuf_targets_binding.v1", trust_repository_id: repository, pack_key: key, pack_version: String(version), pack_release_sequence: version + 1, manifest_sha256: digest(members["manifest.json"]), payload_sha256: payload.digest("hex") } });
    if (badSignature) {
      const target = JSON.parse(members["metadata/targets.json"]);
      target.signatures[0].sig = "0".repeat(128);
      members["metadata/targets.json"] = canonical(target);
    }
    for (const [role, child, days] of [["snapshot", "targets", 14], ["timestamp", "snapshot", 7]]) {
      const value = members[`metadata/${child}.json`];
      members[`metadata/${role}.json`] = signed({ _type: role, spec_version: "1.0.35", version: metadataVersion, expires: expiry(days), meta: { [`${child}.json`]: { version: metadataVersion, length: value.length, hashes: { sha256: digest(value) } } } });
    }
    if (rotation) members["metadata/2.root.json"] = root(2);
    const temporary = mkdtempSync(path.join(tmpdir(), "cartulary-pack-producer-"));
    try {
      for (const [name, value] of Object.entries(members)) {
        mkdirSync(path.dirname(path.join(temporary, name)), { recursive: true });
        writeFileSync(path.join(temporary, name), value);
      }
      const result = spawnSync("tar", ["--format=ustar", "--owner=0", "--group=0", "--mtime=@0", "-cf", path.join(destination, name), "-C", temporary, "--", ...Object.keys(members).sort()], { encoding: "utf8" });
      if (result.status !== 0) throw new Error("package fixture archive creation failed");
    } finally { rmSync(temporary, { recursive: true, force: true }); }
  }
  bundle("valid.tar");
  bundle("rotated.tar", { version: 3, rotation: true });
  bundle("rollback.tar", { version: 3, metadataVersion: 1 });
  bundle("bad-signature.tar", { version: 5, badSignature: true });
  bundle("collectable.tar", { version: 7 });
  bundle("incompatible.tar", { version: 6, incompatible: true });
  writeFileSync(path.join(destination, "malformed.tar"), "not an archive\n");
}
