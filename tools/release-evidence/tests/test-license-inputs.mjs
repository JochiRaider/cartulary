import assert from "node:assert/strict";
import fs, { fstatSync, linkSync, mkdirSync, readFileSync, renameSync, symlinkSync, truncateSync, writeFileSync } from "node:fs";
import { createRequire, syncBuiltinESMExports } from "node:module";
import path from "node:path";
import test from "node:test";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { licenseEvidence } from "../license-inputs.mjs";

test("license identity hashes exact bytes and logical names independently of physical location and creation order", t => {
  const root = createScratch(t, "cartulary-license-identity-");
  const expected = {
    evidenceDigest: "sha256:ea6df3fefef139a6ade0f921cee31ad8d02df61906d00f80d80b156ab5d857a2",
    evidenceSource: "package-distributed license file", issues: [],
  };
  for (const [directory, files] of [
    [path.join(root, "first"), [["LICENSE", "MIT fixture"], ["NOTICE.txt", "notice fixture"]]],
    [path.join(root, "second"), [["NOTICE.txt", "notice fixture"], ["LICENSE", "MIT fixture"]]],
  ]) {
    mkdirSync(directory);
    for (const [name, bytes] of files) writeFileSync(path.join(directory, name), bytes);
    symlinkSync("missing-markdown-target", path.join(directory, "LICENSE.md"));
    assert.deepEqual(licenseEvidence(directory), expected);
  }
  const directory = path.join(root, "first");
  const license = path.join(directory, "LICENSE");
  linkSync(license, path.join(root, "hard-linked-license"));
  assert.deepEqual(licenseEvidence(directory), expected, "stable package-store hard links remain valid");
  writeFileSync(license, "MIT fixture\n");
  assert.notEqual(licenseEvidence(directory).evidenceDigest, expected.evidenceDigest, "trailing bytes are evidence");
  writeFileSync(license, "MIT fixture");
  renameSync(path.join(directory, "NOTICE.txt"), path.join(directory, "NOTICE-other.txt"));
  assert.notEqual(licenseEvidence(directory).evidenceDigest, expected.evidenceDigest, "logical filename changes are evidence");
});

test("license selection preserves package notices, font licenses, first-party identity and metadata fallbacks", t => {
  const root = createScratch(t, "cartulary-license-selection-");
  assert.deepEqual(licenseEvidence(root), { evidenceDigest: null, evidenceSource: "missing", issues: ["license_text_missing", "license_metadata_missing"] });
  assert.deepEqual(licenseEvidence(null, { metadata: "MIT" }), { evidenceDigest: null, evidenceSource: "package metadata license field", issues: ["license_text_missing"] });
  assert.deepEqual(licenseEvidence(path.join(root, "absent")), licenseEvidence(root));
  for (const name of ["LICENSE", "licence.txt", "COPYING", "NOTICE.txt"]) {
    const directory = path.join(root, name);
    mkdirSync(directory); writeFileSync(path.join(directory, name), "fixture");
    assert.match(licenseEvidence(directory).evidenceDigest, /^sha256:[a-f0-9]{64}$/u);
  }
  const font = path.join(root, "font"); mkdirSync(font);
  writeFileSync(path.join(font, "OFL.txt"), "font license");
  assert.equal(licenseEvidence(font).evidenceDigest, null);
  const fontDigest = licenseEvidence(font, { kind: "font" }).evidenceDigest;
  assert.ok(fontDigest);
  writeFileSync(path.join(font, "LICENSE.txt"), "second font license");
  assert.notEqual(licenseEvidence(font, { kind: "font" }).evidenceDigest, fontDigest);
  const firstParty = path.join(root, "first-party"); mkdirSync(firstParty);
  writeFileSync(path.join(firstParty, "LICENSE"), "first-party license");
  const firstPartyDigest = licenseEvidence(firstParty, { kind: "first_party" }).evidenceDigest;
  writeFileSync(path.join(firstParty, "NOTICE"), "other input");
  assert.equal(licenseEvidence(firstParty, { kind: "first_party" }).evidenceDigest, firstPartyDigest);
  assert.throws(() => licenseEvidence(root, { kind: "unknown" }), /unknown license input kind/u);
  const require = createRequire(import.meta.url);
  const installed = path.dirname(require.resolve("typescript/package.json"));
  assert.match(licenseEvidence(installed).evidenceDigest, /^sha256:[a-f0-9]{64}$/u, "prepared package-store layout is admitted");
});

test("license acquisition excludes prose before probes and refuses links without reaching targets", t => {
  const root = createScratch(t, "cartulary-license-admission-");
  writeFileSync(path.join(root, "LICENSE"), "fixture");
  for (const name of ["LICENSE.md", "NOTICE.MARKDOWN", "COPYING.mdown", "LICENCE.mkd", "README"]) writeFileSync(path.join(root, name), "must not be read");
  const originals = Object.fromEntries(["openSync", "statSync", "lstatSync", "readFileSync", "readdirSync", "readlinkSync"].map(name => [name, fs[name]]));
  let prohibited = 0, opens = 0;
  try {
    for (const name of Object.keys(originals)) fs[name] = (...args) => {
      const file = args[0];
      if (typeof file === "string" && /(?:^|\/)(?:docs|README(?:\.[^/]+)?)(?:\/|$)|\.(?:md|markdown|mdown|mkd)$/iu.test(file)) {
        prohibited++; throw new Error("prohibited fixture probe");
      }
      if (name === "openSync") {
        opens++;
        assert.ok(args[1] & fs.constants.O_NOFOLLOW, "every input traversal is anchored without following links");
      }
      return originals[name](...args);
    };
    syncBuiltinESMExports();
    assert.ok(licenseEvidence(root).evidenceDigest);
    const before = opens;
    for (const file of [path.resolve(import.meta.dirname, "../../../docs"), path.join(root, "README.directory"), path.join(root, "LICENSE.md")]) {
      assert.throws(() => licenseEvidence(file), /license input rejected/u);
    }
    assert.equal(opens, before, "excluded roots fail before any ancestor probe");
    const linked = path.join(root, "linked"); mkdirSync(linked);
    symlinkSync(path.join(root, "LICENSE.md"), path.join(linked, "LICENSE"));
    assert.throws(() => licenseEvidence(linked), /license input rejected/u);
    const ancestor = path.join(root, "ancestor"); symlinkSync(root, ancestor);
    assert.throws(() => licenseEvidence(ancestor), /license input rejected/u);
    assert.equal(prohibited, 0);
  } finally {
    Object.assign(fs, originals); syncBuiltinESMExports();
  }
});

test("license inputs reject nonregular, oversized and changed files and close failed read descriptors", t => {
  const root = createScratch(t, "cartulary-license-files-");
  const directory = path.join(root, "nonregular"); mkdirSync(directory); mkdirSync(path.join(directory, "LICENSE"));
  assert.throws(() => licenseEvidence(directory), /license input rejected/u);
  const file = path.join(root, "LICENSE"); writeFileSync(file, "");
  truncateSync(file, 8 * 1024 * 1024);
  assert.ok(licenseEvidence(root).evidenceDigest, "the exact bound is supported");
  truncateSync(file, 8 * 1024 * 1024 + 1);
  assert.throws(() => licenseEvidence(root), /license input rejected/u);
  writeFileSync(file, "before");
  const original = fs.readSync;
  let readDescriptor;
  try {
    fs.readSync = (...args) => {
      const count = original(...args);
      if (readDescriptor === undefined) {
        readDescriptor = args[0]; writeFileSync(file, "after mutation");
      }
      return count;
    };
    syncBuiltinESMExports();
    assert.throws(() => licenseEvidence(root), /license input rejected/u);
    assert.throws(() => fstatSync(readDescriptor), { code: "EBADF" });
  } finally { fs.readSync = original; syncBuiltinESMExports(); }
  assert.equal(readFileSync(file, "utf8"), "after mutation");
});
