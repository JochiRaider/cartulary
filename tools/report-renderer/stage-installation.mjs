import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { constants, chmodSync, closeSync, fchmodSync, fstatSync, lstatSync, mkdirSync, openSync, readFileSync, readSync, readdirSync, realpathSync, writeSync } from "node:fs";
import path from "node:path";
import Ajv2020 from "ajv/dist/2020.js";

import { parseStrictJSON } from "../harness/contract/index.mjs";

// This is a build-time producer. Compile the Reporting-owned installation
// contract here; do not copy a product contract into the harness evidence registry.
const validateInstallation = new Ajv2020({ strict: true }).compile(JSON.parse(readFileSync(new URL("../../contracts/reporting/runtime-installation.v1.schema.json", import.meta.url), "utf8")));

const sha = (bytes) => createHash("sha256").update(bytes).digest("hex");
const forbidden = (name) => /^(?:docs|readme(?:\..*)?)$/iu.test(name) || /\.(?:md|markdown)$/iu.test(name);
function admittedPath(value) {
  assert.equal(typeof value, "string");
  assert.ok(value.length > 0 && value.length <= 512 && value === value.normalize("NFC"));
  assert.ok(!/[\\:\x00-\x1f\x7f]/u.test(value));
  for (const segment of value.split("/")) assert.ok(segment && segment !== "." && segment !== ".." && !forbidden(segment));
}
function readBounded(file, limit) {
  const fd = openSync(file, constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK);
  try {
    const stat = fstatSync(fd);
    assert.ok(stat.isFile() && stat.nlink === 1 && stat.size <= limit);
    const bytes = Buffer.alloc(stat.size + 1);
    let offset = 0, count;
    while ((count = readSync(fd, bytes, offset, bytes.length - offset, null)) > 0) {
      offset += count;
      assert.ok(offset <= stat.size, "runtime metadata changed during packaging");
    }
    assert.equal(offset, stat.size);
    return bytes.subarray(0, offset);
  } finally { closeSync(fd); }
}
function copyBoundFile(source, destination, expected) {
  const input = openSync(source, constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK);
  let output;
  try {
    const stat = fstatSync(input);
    assert.ok(stat.isFile() && stat.nlink === 1 && stat.size <= 1 << 30);
    output = openSync(destination, "wx", stat.mode & 0o111 ? 0o555 : 0o444);
    fchmodSync(output, stat.mode & 0o111 ? 0o555 : 0o444);
    const buffer = Buffer.alloc(1024 * 1024), digest = createHash("sha256");
    let total = 0, count;
    while ((count = readSync(input, buffer, 0, buffer.length, null)) > 0) {
      total += count;
      assert.ok(total <= stat.size, "runtime input changed during packaging");
      digest.update(buffer.subarray(0, count));
      let offset = 0;
      while (offset < count) offset += writeSync(output, buffer, offset, count - offset);
    }
    assert.equal(total, stat.size);
    assert.equal(digest.digest("hex"), expected, "runtime file content mismatch");
  } finally { closeSync(input); if (output !== undefined) closeSync(output); }
}

// Copy one immutable installation into a private Docker build context. Retired
// roots and unrelated checkout contents never enter that context. This verifies
// package bytes; native execution qualification remains a separate obligation.
function admitReceipt(receiptBytes) {
  const receipt = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(receiptBytes));
  assert.deepEqual(Object.keys(receipt).sort(), ["manifest", "manifest_sha256", "root", "schema_id"]);
  assert.equal(receipt.schema_id, "cartulary.reporting_runtime_receipt.v1");
  assert.match(receipt.manifest_sha256, /^[a-f0-9]{64}$/u);
  assert.equal(receipt.root, `${receipt.manifest_sha256}.root`);
  assert.equal(receipt.manifest, `${receipt.manifest_sha256}.json`);
  return receipt;
}

function admitMetadata(receiptBytes, manifestBytes) {
  const receipt = admitReceipt(receiptBytes);
  assert.equal(sha(manifestBytes), receipt.manifest_sha256);
  const manifest = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(manifestBytes));
  assert.ok(validateInstallation(manifest), "Reporting runtime installation shape invalid");
  const expected = new Map();
  let previous = "";
  for (const file of manifest.files) {
    admittedPath(file.path);
    assert.ok(Buffer.compare(Buffer.from(previous), Buffer.from(file.path)) < 0, "runtime file set is not uniquely ordered");
    previous = file.path;
    expected.set(file.path, file.sha256);
  }
  return { receipt, expected };
}

// Image qualification permits tools only inside the receipt's exact file set.
// This inventory check complements native admission of the actual file bytes.
export function assertPackagedRendererLayout(entries, receiptBytes, manifestBytes) {
  const {receipt,expected}=admitMetadata(receiptBytes,manifestBytes);
  const prefix="opt/cartulary/report-renderer/", root=`${prefix}${receipt.root}/`;
  const required=new Set([`${prefix}installation.json`,`${prefix}${receipt.manifest}`,...[...expected.keys()].map(name=>root+name)]);
  const seen=new Set();
  for(const entry of entries) {
    assert.ok(!seen.has(entry),"duplicate image entry");seen.add(entry);
    if(entry===prefix || entry===root) continue;
    if(entry.startsWith(root)) {
      admittedPath(entry.slice(root.length).replace(/\/$/u,""));
      if(entry.endsWith("/")) continue;
    }
    if(entry.startsWith(prefix)) assert.ok(required.has(entry),"unbound renderer image entry");
    else assert.ok(!/(^|\/)(node|npm|pnpm|vite|node_modules)(\/|$)|(^|\/)(apps\/web|db\/migrations)(\/|$)/u.test(entry),"tooling outside renderer installation");
  }
  for(const name of required) assert.ok(seen.has(name),"renderer image file missing");
  return receipt.manifest_sha256;
}

export function stageRendererInstallation(source, destination) {
  assert.ok(path.isAbsolute(source) && path.normalize(source) === source);
  for (const segment of source.split(path.sep)) assert.ok(!forbidden(segment));
  assert.equal(realpathSync(source), source);
  const receiptBytes = readBounded(path.join(source, "installation.json"), 4096);
  const selected = admitReceipt(receiptBytes);
  const manifestBytes = readBounded(path.join(source, selected.manifest), 32 << 20);
  const { receipt, expected } = admitMetadata(receiptBytes, manifestBytes);
  const runtimeRoot = path.join(source, receipt.root);
  assert.ok(lstatSync(runtimeRoot).isDirectory());
  mkdirSync(destination);
  let count = 0;
  function copyDirectory(from, to, prefix = "") {
    mkdirSync(to);
    for (const entry of readdirSync(from, { withFileTypes: true })) {
      // Reject human material before stat, opening, or directory traversal.
      const relative = prefix ? `${prefix}/${entry.name}` : entry.name;
      admittedPath(relative);
      if (entry.isDirectory()) copyDirectory(path.join(from, entry.name), path.join(to, entry.name), relative);
      else {
        assert.ok(entry.isFile() && expected.has(relative), "undeclared runtime entry");
        copyBoundFile(path.join(from, entry.name), path.join(to, entry.name), expected.get(relative));
        count++;
      }
    }
  }
  copyDirectory(runtimeRoot, path.join(destination, receipt.root));
  assert.equal(count, expected.size, "runtime file set incomplete");
  copyBoundFile(path.join(source, receipt.manifest), path.join(destination, receipt.manifest), receipt.manifest_sha256);
  copyBoundFile(path.join(source, "installation.json"), path.join(destination, "installation.json"), sha(receiptBytes));
  // Build hosts may use a private umask. The image owns these directories as
  // root, so its non-root application still needs read/traverse permission.
  function admitDirectoryModes(directory) {
    for (const entry of readdirSync(directory, { withFileTypes: true })) if (entry.isDirectory()) admitDirectoryModes(path.join(directory, entry.name));
    chmodSync(directory, 0o755);
  }
  admitDirectoryModes(destination);
  return receipt.manifest_sha256;
}
