// Independent current-contract vectors. No Go format or producer code supplies
// the canonicalization, digest preimages, signatures, or container expectations.
import { createHash, createPrivateKey, createPublicKey, sign } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import { crc32, gzipSync } from "node:zlib";

const canonical = value => Array.isArray(value) ? `[${value.map(canonical).join(",")}]` : value !== null && typeof value === "object" ? `{${Object.keys(value).sort().map(key => `${JSON.stringify(key)}:${canonical(value[key])}`).join(",")}}` : JSON.stringify(value);
const sha = value => createHash("sha256").update(value).digest("hex");
const input = JSON.parse(readFileSync("tools/reference-pack-modules/current-fixtures-recipes.v1.json"));
if (input.schema_id !== "cartulary.reference_pack_fixture_recipes.v1") throw new Error("fixture recipe version");
const read = name => structuredClone(input.vectors[name]);
const expect = (actual, expected, label) => {
  if (actual !== expected) throw new Error(`fixture expectation changed: ${label}`);
};
const write = (path, value) => {
  const bytes = `${JSON.stringify(value, null, 2)}\n`;
  let old; try { old=readFileSync(path,"utf8"); } catch(error) { if(error.code!=="ENOENT") throw error; }
  if (old !== bytes) writeFileSync(path, bytes);
};
const output = (name, value) => write(`contracts/reference-pack-fixtures/fixtures/${name}`, value);
const set = members => {
  const body = { schema_id: "cartulary.reference_pack_set.v3", members: [...members].sort((a, b) => a.pack_key < b.pack_key ? -1 : 1) };
  const digest = sha(canonical(body));
  return { ...body, pack_set_id: `rpset_${digest}`, pack_set_sha256: digest };
};
const validateReferences = value => {
  for (const entry of value.sets) {
    const expected = set(entry.members);
    expect(entry.pack_set_id, expected.pack_set_id, "reference set identity");
    expect(entry.pack_set_sha256, expected.pack_set_sha256, "reference set digest");
  }
  return value;
};
const vector = read("manifest");
const manifest = vector.manifest; delete vector.manifest;
vector.canonical_manifest = canonical(manifest);
expect(sha(vector.canonical_manifest), vector.manifest_sha256, "manifest");
// Preserve field order in these readable projections; cryptographic preimages use
// the independent canonical encoder above.
output("manifest.v1.json", {schema_id: vector.schema_id, producer: vector.producer, canonical_manifest: vector.canonical_manifest, manifest_sha256: vector.manifest_sha256, payload_sha256: vector.payload_sha256});
const complete = read("complete-set"), full = set(complete.members);
expect(complete.expected_set_id, full.pack_set_id, "complete set");
output("complete-set.v1.json", complete);
output("portable-references.v1.json", validateReferences(read("portable-references")));
const portable = read("portable-content");
validateReferences(portable.references);
output("portable-content.v1.json", portable);

// These public, reproducible synthetic keys are never release authority. The
// generator retains only signed public metadata; no private key is serialized.
const keys = Array.from({ length: 3 }, (_, i) => {
  const seed = createHash("sha256").update(`cartulary contract-v3 test-vector key ${i}; not a release key`).digest();
  const privateKey = createPrivateKey({ key: Buffer.concat([Buffer.from("302e020100300506032b657004220420", "hex"), seed]), format: "der", type: "pkcs8" });
  const publicValue = { keytype: "ed25519", scheme: "ed25519", keyval: { public: createPublicKey(privateKey).export({ format: "der", type: "spki" }).subarray(-32).toString("hex") } };
  return { privateKey, publicValue, id: sha(canonical(publicValue)) };
});
const signed = value => canonical({ signed: value, signatures: (value._type === "root" ? keys : keys.slice(0, 1)).map(key => ({ keyid: key.id, sig: sign(null, Buffer.from(canonical(value)), key.privateKey).toString("hex") })).sort((a, b) => a.keyid < b.keyid ? -1 : 1) });
function zipStored(members) {
  const local = [], central = []; let offset = 0;
  for (const name of Object.keys(members).sort()) {
    const bytes = Buffer.from(members[name]), filename = Buffer.from(name), crc = crc32(bytes);
    const header = Buffer.alloc(30); header.writeUInt32LE(0x04034b50); header.writeUInt16LE(20, 4); header.writeUInt16LE(33, 12);
    header.writeUInt32LE(crc, 14); header.writeUInt32LE(bytes.length, 18); header.writeUInt32LE(bytes.length, 22); header.writeUInt16LE(filename.length, 26);
    local.push(header, filename, bytes);
    const entry = Buffer.alloc(46); entry.writeUInt32LE(0x02014b50); entry.writeUInt16LE(20, 4); entry.writeUInt16LE(20, 6); entry.writeUInt16LE(33, 14);
    entry.writeUInt32LE(crc, 16); entry.writeUInt32LE(bytes.length, 20); entry.writeUInt32LE(bytes.length, 24); entry.writeUInt16LE(filename.length, 28); entry.writeUInt32LE(offset, 42);
    central.push(entry, filename); offset += header.length + filename.length + bytes.length;
  }
  const directory = Buffer.concat(central), end = Buffer.alloc(22), count = Object.keys(members).length;
  end.writeUInt32LE(0x06054b50); end.writeUInt16LE(count, 8); end.writeUInt16LE(count, 10); end.writeUInt32LE(directory.length, 12); end.writeUInt32LE(offset, 16);
  return Buffer.concat([...local, directory, end]);
}
function encode(recipe) {
  const members = Object.fromEntries(Object.entries(recipe.members).map(([path, text]) => [path, Buffer.from(text)]));
  const targets = recipe.metadata.targets;
  targets.cartulary.manifest_sha256 = sha(members["manifest.json"]);
  for (const [name, entry] of Object.entries(targets.targets)) { entry.length = members[name].length; entry.hashes.sha256 = sha(members[name]); }
  members["metadata/targets.json"] = Buffer.from(signed(targets));
  for (const [role, child] of [["snapshot", "targets"], ["timestamp", "snapshot"]]) {
    const value = recipe.metadata[role], bytes = members[`metadata/${child}.json`];
    value.meta[`${child}.json`].length = bytes.length; value.meta[`${child}.json`].hashes.sha256 = sha(bytes);
    members[`metadata/${role}.json`] = Buffer.from(signed(value));
  }
  const bytes = zipStored(members);
  return { bytes, members, manifestSHA: sha(members["manifest.json"]), containerSHA: sha(bytes) };
}
const signedVector = read("signed-container"), released = encode(signedVector.recipe), root = signedVector.root;
delete signedVector.recipe; delete signedVector.root;
root.keys = Object.fromEntries(keys.map(key => [key.id, key.publicValue]));
root.roles = { root: { keyids: keys.map(key => key.id).sort(), threshold: 2 }, ...Object.fromEntries(["targets", "snapshot", "timestamp"].map(role => [role, { keyids: [keys[0].id], threshold: 1 }])) };
expect(released.manifestSHA, signedVector.manifest_sha256, "signed manifest");
expect(released.containerSHA, signedVector.container_sha256, "signed container");
const signedOutput = {};
for (const [name, value] of Object.entries(signedVector)) {
  if (name === "repository_id") signedOutput.bootstrap = signed(root);
  if (name === "expected_signers") signedOutput.container_base64 = released.bytes.toString("base64");
  signedOutput[name] = value;
}
output("signed-container.v1.json", signedOutput);
const portableInput = read("portable-input");
for (const name of ["allowed", "restricted"]) {
  const variant = portableInput[name], result = encode(variant.recipe);
  expect(result.manifestSHA, variant.descriptor.manifest_sha256, `${name} manifest`);
  expect(result.containerSHA, variant.descriptor.container_sha256, `${name} container`);
  expect(result.bytes.length, variant.descriptor.size_bytes, `${name} size`);
  delete variant.recipe;
  portableInput[name] = {container_base64: result.bytes.toString("base64"), ...variant};
}
output("portable-input.v1.json", portableInput);
for (const [name, expected] of Object.entries(input.expectations)) write(`contracts/reference-pack-fixtures/fixture-manifests/${name}.v1.json`, expected);

function tarStored(members) {
  const blocks = [];
  for (const name of Object.keys(members).sort()) {
    const bytes = Buffer.from(members[name]), header = Buffer.alloc(512);
    if (Buffer.byteLength(name) > 100) throw new Error("fixture tar path recipe changed");
    header.write(name, 0, 100);
    for (const [offset, size, value] of [[100, 8, 0o644], [108, 8, 0], [116, 8, 0], [124, 12, bytes.length], [136, 12, 0]]) header.write(value.toString(8).padStart(size - 1, "0") + "\0", offset, size);
    header.fill(32, 148, 156); header[156] = 48; header.write("ustar\0", 257); header.write("00", 263);
    const checksum = header.reduce((sum, byte) => sum + byte, 0);
    header.write(checksum.toString(8).padStart(6, "0") + "\0 ", 148, 8);
    blocks.push(header, bytes, Buffer.alloc((512 - bytes.length % 512) % 512));
  }
  return Buffer.concat([...blocks, Buffer.alloc(1024)]);
}
for (const {name, vector: value, expectation: expected} of input.admission) {
  if (name.startsWith("equivalent_")) {
    let bytes = name.startsWith("equivalent_zip") ? released.bytes : tarStored(released.members);
    if (name.startsWith("equivalent_gzip")) bytes = gzipSync(bytes);
    expect(sha(bytes), expected.expected_container_sha256, name);
    expect(released.manifestSHA, expected.expected_manifest_sha256, name);
    value.bytes_base64 = bytes.toString("base64");
    value.expected_members = Object.fromEntries(Object.entries(released.members).sort(([a], [b]) => a < b ? -1 : 1).map(([path, raw]) => [path, raw.toString("utf8")]));
  }
  // Malformed recipes intentionally retain exact bytes: duplicate keys, BOM,
  // invalid UTF-8, spacing and malformed license/source fields are test inputs.
  output(`admission/${name}`, name.startsWith("equivalent_") ? {schema_id: value.schema_id, kind: value.kind, bytes_base64: value.bytes_base64, expected_members: value.expected_members, operator: value.operator} : value);
  write(`contracts/reference-pack-fixtures/fixture-manifests/admission/${name}`, expected);
}
