import { Buffer } from "node:buffer";
import {
  createHash,
  createPrivateKey,
  createPublicKey,
  sign,
} from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

// Functional fixture producer only. Independent signature vectors are authored
// separately under contracts/reference-pack-fixtures/fixtures. These keys have no use
// outside a disposable browser deployment.
const repository = "browser.fixture.repo";
const digest = (bytes) => createHash("sha256").update(bytes).digest("hex");
const canonical = (value) => {
  const ordered = (v) =>
    Array.isArray(v)
      ? v.map(ordered)
      : v !== null && typeof v === "object"
        ? Object.fromEntries(
            Object.keys(v)
              .sort()
              .map((k) => [k, ordered(v[k])]),
          )
        : v;
  return Buffer.from(JSON.stringify(ordered(value)));
};
const signer = (ordinal) => {
  const seed = createHash("sha256")
    .update(`cartulary browser reference pack test key only ${ordinal}`)
    .digest();
  const key = createPrivateKey({
    key: Buffer.concat([
      Buffer.from("302e020100300506032b657004220420", "hex"),
      seed,
    ]),
    format: "der",
    type: "pkcs8",
  });
  const publicKey = {
    keytype: "ed25519",
    scheme: "ed25519",
    keyval: {
      public: createPublicKey(key)
        .export({ format: "der", type: "spki" })
        .subarray(-32)
        .toString("hex"),
    },
  };
  return { key, publicKey, id: digest(canonical(publicKey)) };
};
const keys = [0, 1, 2].map(signer);
const signed = (value, signers = keys.slice(0, 1)) =>
  canonical({
    signed: value,
    signatures: signers
      .map(({ id, key }) => ({
        keyid: id,
        sig: sign(null, canonical(value), key).toString("hex"),
      }))
      .sort((a, b) => (a.keyid < b.keyid ? -1 : a.keyid > b.keyid ? 1 : 0)),
  });
const expires = (days) =>
  new Date(Math.floor(Date.now() / 86400000) * 86400000 + days * 86400000)
    .toISOString()
    .replace(".000Z", "Z");

function bootstrap() {
  const roles = {
    root: { keyids: keys.map(({ id }) => id).sort(), threshold: 2 },
  };
  for (const role of ["timestamp", "snapshot", "targets"])
    roles[role] = { keyids: [keys[0].id], threshold: 1 };
  const root = signed(
    {
      _type: "root",
      spec_version: "1.0.35",
      version: 1,
      expires: expires(365),
      consistent_snapshot: false,
      keys: Object.fromEntries(
        keys.map(({ id, publicKey }) => [id, publicKey]),
      ),
      roles,
      cartulary: {
        schema_id: "cartulary.reference_pack_tuf_root_binding.v1",
        trust_repository_id: repository,
      },
    },
    keys,
  );
  return canonical({
    schema_id: "cartulary.reference_pack_trust_bootstrap.v1",
    repositories: [
      {
        repository_id: repository,
        trusted_root: JSON.parse(root),
        trusted_root_sha256: digest(root),
      },
    ],
  });
}

export function referencePackBundle({ invalidManifest = false } = {}) {
  const release = JSON.parse(
    readFileSync(
      new URL(
        "../../../../contracts/reference-packs/builtins/release.v1.json",
        import.meta.url,
      ),
    ),
  );
  const base = release.packs.find(
    (pack) => pack.binding.pack_key === "type_registry.host",
  );
  // A stable version and manifest make exact replay safe across browser rows.
  const version = invalidManifest ? "browser-invalid-schema" : "browser-1";
  const manifest = {
    ...JSON.parse(base.members["manifest.json"]),
    pack_version: version,
    trust_repository_id: repository,
  };
  manifest.source_as_of = manifest.built_at;
  if (invalidManifest)
    manifest.private_hostile_member = "private diagnostic sentinel";
  const members = Object.fromEntries(
    Object.entries(base.members).map(([path, value]) => [
      path,
      Buffer.from(value),
    ]),
  );
  members["manifest.json"] = canonical(manifest);
  members["bundle.json"] = canonical({
    schema_id: "cartulary.reference_pack_bundle_hint.v1",
    trust_repository_id: repository,
  });
  const descriptor = (bytes) => ({
    length: bytes.length,
    hashes: { sha256: digest(bytes) },
  });
  members["metadata/targets.json"] = signed({
    _type: "targets",
    spec_version: "1.0.35",
    version: 1,
    expires: expires(30),
    targets: Object.fromEntries(
      Object.entries(members).map(([path, bytes]) => [path, descriptor(bytes)]),
    ),
    cartulary: {
      schema_id: "cartulary.reference_pack_tuf_targets_binding.v1",
      trust_repository_id: repository,
      pack_key: base.binding.pack_key,
      pack_version: version,
      pack_release_sequence: manifest.pack_release_sequence,
      manifest_sha256: digest(members["manifest.json"]),
      payload_sha256: base.binding.payload_sha256,
    },
  });
  for (const [role, child, days] of [
    ["snapshot", "targets", 14],
    ["timestamp", "snapshot", 7],
  ]) {
    members[`metadata/${role}.json`] = signed({
      _type: role,
      spec_version: "1.0.35",
      version: 1,
      expires: expires(days),
      meta: {
        [`${child}.json`]: {
          version: 1,
          ...descriptor(members[`metadata/${child}.json`]),
        },
      },
    });
  }
  const chunks = [];
  for (const name of Object.keys(members).sort()) {
    const bytes = members[name];
    const header = Buffer.alloc(512);
    header.write(name, 0, 100);
    const octal = (value, offset, width) =>
      header.write(
        `${value.toString(8).padStart(width - 1, "0")}\0`,
        offset,
        width,
      );
    octal(0o644, 100, 8);
    octal(0, 108, 8);
    octal(0, 116, 8);
    octal(bytes.length, 124, 12);
    octal(0, 136, 12);
    header.fill(32, 148, 156);
    header.write("0", 156);
    header.write("ustar\0", 257);
    header.write("00", 263);
    header.write(
      `${header
        .reduce((sum, byte) => sum + byte, 0)
        .toString(8)
        .padStart(6, "0")}\0 `,
      148,
      8,
    );
    chunks.push(
      header,
      bytes,
      Buffer.alloc((512 - (bytes.length % 512)) % 512),
    );
  }
  chunks.push(Buffer.alloc(1024));
  return {
    key: base.binding.pack_key,
    version,
    upload: {
      name: "Reference-pack.tar",
      mimeType: "application/x-tar",
      buffer: Buffer.concat(chunks),
    },
  };
}

if (
  process.argv[1] &&
  import.meta.url === pathToFileURL(process.argv[1]).href
) {
  if (process.argv.length !== 3)
    throw new Error("Expected one private bootstrap output path");
  writeFileSync(process.argv[2], bootstrap(), { mode: 0o600 });
}
