import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "node:child_process";
import { syncBuiltinESMExports } from "node:module";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { hashPackageFile, packageFileLimit, preparePackageArchiveDirectory, publishPackageArchive, withVerifiedPackageArchive } from "../package-archive.mjs";

const hash = bytes => createHash("sha256").update(bytes).digest("hex");
const receipt = bytes => ({ schema_id: "cartulary.local_release_receipt.v1", manifest_sha256: "a".repeat(64),
 archive_sha256: hash(bytes), archive: `${"a".repeat(64)}.tar`, operator_sha256: "c".repeat(64),
 images: ["application", "postgres", "seaweedfs"].map(role => ({ role, id: `sha256:${"d".repeat(64)}` })),
 inventory_disposition: "content_complete_review_unestablished" });

test("package cache migration is limited to its exact owned legacy leaf", t => {
 const root = createScratch(t, "cartulary-package-cache-");
 const parent = path.join(root, "parent"); fs.mkdirSync(parent); fs.chmodSync(parent, 0o755);
 const cache = path.join(parent, "cache"); fs.mkdirSync(cache); fs.chmodSync(cache, 0o755);
 preparePackageArchiveDirectory(cache);
 assert.equal(fs.statSync(cache).mode & 0o777, 0o700);
 assert.equal(fs.statSync(parent).mode & 0o777, 0o755);
 preparePackageArchiveDirectory(cache);
 const fresh = path.join(parent, "new", "cache"); preparePackageArchiveDirectory(fresh);
 assert.equal(fs.statSync(fresh).mode & 0o777, 0o700);
 for (const mode of [0o777, 0o750, 0o1700]) {
  fs.chmodSync(cache, mode); assert.throws(() => preparePackageArchiveDirectory(cache), /unsafe package archive directory/u);
  assert.equal(fs.statSync(cache).mode & 0o7777, mode);
 }
 fs.chmodSync(cache, 0o755);
 fs.symlinkSync(cache, path.join(parent, "linked"));
 fs.symlinkSync(parent, path.join(root, "linked-parent"));
 assert.throws(() => preparePackageArchiveDirectory(path.join(parent, "linked")));
 assert.throws(() => preparePackageArchiveDirectory(path.join(root, "linked-parent", "cache")));
 assert.equal(fs.statSync(cache).mode & 0o777, 0o755);
 const original = fs.fstatSync;
 try {
  fs.fstatSync = descriptor => ({ ...original(descriptor), uid: process.getuid() + 1 }); syncBuiltinESMExports();
  assert.throws(() => preparePackageArchiveDirectory(cache), /unsafe package archive directory/u);
 } finally { fs.fstatSync = original; syncBuiltinESMExports(); }
 assert.equal(fs.statSync(cache).mode & 0o777, 0o755);
});

test("package publication preserves complete concurrent winners and private modes", t => {
 const root = createScratch(t, "cartulary-package-publish-");
 const bytes = Buffer.from("independent package fixture"), value = receipt(bytes), source = path.join(root, "input.tar");
 fs.writeFileSync(source, bytes, { mode: 0o644 });
 for (const winner of ["none", "same", "different", "io_failure", "cleanup_failure"]) {
  const cache = path.join(root, winner); preparePackageArchiveDirectory(cache);
  const destination = path.join(cache, value.archive);
  const original = { linkSync: fs.linkSync, unlinkSync: fs.unlinkSync };
  let publications = 0;
  try {
   fs.linkSync = (temporary, target) => {
    publications++;
    assert.equal(path.basename(target), value.archive);
    assert.deepEqual(fs.readFileSync(temporary), bytes, "only complete bytes can publish");
    assert.equal(fs.statSync(temporary).mode & 0o777, 0o600);
    if (winner === "io_failure") throw Object.assign(new Error("publication failed"), { code: "EIO" });
    if (winner !== "none") fs.writeFileSync(target, winner === "different" ? "other winner" : bytes, { mode: 0o600 });
    return original.linkSync(temporary, target);
   };
   fs.unlinkSync = file => {
    if (winner === "cleanup_failure") throw new Error("temporary cleanup failed");
    return original.unlinkSync(file);
   };
   syncBuiltinESMExports();
   if (["none", "same"].includes(winner)) assert.deepEqual(publishPackageArchive(cache, value.manifest_sha256, source), { archive: value.archive, archive_sha256: value.archive_sha256 });
   else if (winner === "different") assert.throws(() => publishPackageArchive(cache, value.manifest_sha256, source), /same manifest cannot publish different archive bytes/u);
   else if (winner === "cleanup_failure") assert.throws(() => publishPackageArchive(cache, value.manifest_sha256, source), error => error.code === "EEXIST" && error.cleanupFailures.length === 1);
   else assert.throws(() => publishPackageArchive(cache, value.manifest_sha256, source), { code: "EIO" });
  } finally { Object.assign(fs, original); syncBuiltinESMExports(); }
  assert.equal(publications, 1);
  assert.deepEqual(fs.readFileSync(source), bytes, "input remains borrowed");
  if (winner === "io_failure") assert.deepEqual(fs.readdirSync(cache), []);
  else {
   assert.deepEqual(fs.readFileSync(destination), winner === "different" ? Buffer.from("other winner") : bytes);
   if (winner !== "cleanup_failure") assert.deepEqual(fs.readdirSync(cache), [value.archive]);
   else assert.equal(fs.readdirSync(cache).filter(name => name.startsWith(".publishing-")).length, 1);
  }
 }
 const legacy = path.join(root, "legacy"); preparePackageArchiveDirectory(legacy);
 fs.writeFileSync(path.join(legacy, value.archive), bytes, { mode: 0o644 });
 assert.deepEqual(publishPackageArchive(legacy, value.manifest_sha256, source), { archive: value.archive, archive_sha256: value.archive_sha256 });
 assert.equal(fs.statSync(path.join(legacy, value.archive)).mode & 0o777, 0o644, "existing safe archive is not rewritten");
});

test("archive admission rejects unsafe, corrupt and changing inputs before extraction", t => {
 const root = createScratch(t, "cartulary-package-admission-");
 const bytes = Buffer.alloc(150000, 42), value = receipt(bytes);
 for (const mutation of ["missing", "corrupt", "symlink", "ancestor_link", "hardlink", "directory", "writable", "oversize", "during_read"]) {
  const cache = path.join(root, mutation); preparePackageArchiveDirectory(cache);
  const archive = path.join(cache, value.archive), input = path.join(root, `input-${mutation}`);
  fs.writeFileSync(input, bytes, { mode: 0o644 });
  if (mutation === "symlink") fs.symlinkSync(input, archive);
  else if (mutation === "hardlink") fs.linkSync(input, archive);
  else if (mutation === "directory") fs.mkdirSync(archive);
  else if (mutation !== "missing") fs.copyFileSync(input, archive);
  if (mutation === "corrupt") fs.writeFileSync(archive, "corrupt");
  if (mutation === "writable") fs.chmodSync(archive, 0o666);
  if (mutation === "oversize") fs.truncateSync(archive, packageFileLimit + 1);
  let selected = cache;
  if (mutation === "ancestor_link") { selected = path.join(root, "alias"); fs.symlinkSync(cache, selected); }
  let consumed = 0, changed = false; const original = fs.readSync;
  try {
   if (mutation === "during_read") {
    fs.readSync = (...args) => { const n = original(...args); if (n > 0 && !changed) { changed = true; fs.appendFileSync(archive, "changed"); } return n; };
    syncBuiltinESMExports();
   }
   assert.throws(() => withVerifiedPackageArchive(selected, value, path.join(root, "workspace", mutation), () => { consumed++; }));
  } finally { fs.readSync = original; syncBuiltinESMExports(); }
  assert.equal(consumed, 0, mutation);
  assert.equal(fs.existsSync(path.join(root, "workspace", mutation)) && fs.readdirSync(path.join(root, "workspace", mutation)).length > 0, false, mutation);
  if (!["corrupt", "during_read", "ancestor_link"].includes(mutation)) assert.throws(() => hashPackageFile(archive));
  if (mutation === "during_read") assert.equal(changed, true);
 }
});

test("extraction consumes verified snapshot bytes even after original pathname replacement", t => {
 const root = createScratch(t, "cartulary-package-snapshot-");
 const content = path.join(root, "content"), extracted = path.join(root, "extracted"), cache = path.join(root, "cache");
 fs.mkdirSync(content); fs.mkdirSync(extracted); preparePackageArchiveDirectory(cache);
 fs.writeFileSync(path.join(content, "payload"), "verified fixture bytes");
 const input = path.join(root, "input.tar");
 const made = spawnSync("tar", ["-cf", input, "-C", content, "payload"], { encoding: "utf8", timeout: 10000 });
 assert.ifError(made.error); assert.equal(made.status, 0, made.stderr);
 const bytes = fs.readFileSync(input), value = receipt(bytes);
 fs.copyFileSync(input, path.join(cache, value.archive));
 let seen;
 const result = withVerifiedPackageArchive(cache, value, path.join(root, "workspace"), snapshot => {
  seen = snapshot;
  assert.equal(path.dirname(path.dirname(snapshot)), path.join(root, "workspace"));
  assert.equal(fs.statSync(snapshot).mode & 0o777, 0o600);
  assert.equal(fs.statSync(path.dirname(snapshot)).mode & 0o777, 0o700);
  fs.renameSync(path.join(cache, value.archive), path.join(cache, "old.tar"));
  fs.writeFileSync(path.join(cache, value.archive), "replacement must never reach tar");
  const untar = spawnSync("tar", ["--no-same-owner", "--same-permissions", "-xf", snapshot, "-C", extracted], { encoding: "utf8", timeout: 10000 });
  assert.ifError(untar.error); assert.equal(untar.status, 0, untar.stderr);
  return "consumed";
 });
 assert.equal(result, "consumed");
 assert.equal(fs.readFileSync(path.join(extracted, "payload"), "utf8"), "verified fixture bytes");
 assert.equal(fs.existsSync(path.dirname(seen)), false);
 assert.deepEqual(fs.readFileSync(path.join(cache, "old.tar")), bytes);
 assert.equal(fs.readFileSync(path.join(cache, value.archive), "utf8"), "replacement must never reach tar");
});

test("archive snapshot cleanup failures preserve primary failures and never pass", t => {
 const root = createScratch(t, "cartulary-package-snapshot-cleanup-");
 for (const scenario of ["consumer", "cleanup", "both", "corrupt_cleanup", "acquisition"]) {
  const cache = path.join(root, scenario); preparePackageArchiveDirectory(cache);
  const bytes = Buffer.from("fixture"), value = receipt(bytes);
  fs.writeFileSync(path.join(cache, value.archive), scenario === "corrupt_cleanup" ? "bad" : bytes);
  const original = { rmSync: fs.rmSync, mkdtempSync: fs.mkdtempSync };
  const primary = new Error("consumer failed"); let cleanupCalls = 0, caught;
  try {
   fs.rmSync = (...args) => { cleanupCalls++; if (scenario !== "consumer") throw new Error("private cleanup error"); return original.rmSync(...args); };
   if (scenario === "acquisition") fs.mkdtempSync = () => { throw primary; };
   syncBuiltinESMExports();
   try { withVerifiedPackageArchive(cache, value, path.join(root, "workspace", scenario), () => { if (["consumer", "both"].includes(scenario)) throw primary; }); }
   catch (error) { caught = error; }
  } finally { Object.assign(fs, original); syncBuiltinESMExports(); }
  assert.ok(caught);
  assert.equal(cleanupCalls, scenario === "acquisition" ? 0 : 1);
  if (["consumer", "both", "acquisition"].includes(scenario)) assert.equal(caught, primary);
  if (["both", "corrupt_cleanup"].includes(scenario)) assert.equal(caught.cleanupFailures[0].failure_reason, "cleanup_error");
  if (scenario === "corrupt_cleanup") assert.match(caught.message, /package archive digest mismatch/u);
  if (scenario === "cleanup") { assert.equal(caught.failure_class, "harness"); assert.equal(caught.failure_reason, "cleanup_error"); }
  assert.deepEqual(fs.readdirSync(cache), [value.archive], "snapshot failures never leave temporary cache entries");
  assert.equal(fs.readdirSync(path.join(root, "workspace", scenario)).length, ["consumer", "acquisition"].includes(scenario) ? 0 : 1);
 }
});


test("archive snapshots require an explicit private workspace and preserve borrowed cache and parent", t => {
 const root = createScratch(t, "cartulary-package-workspace-");
 const cache = path.join(root, "cache"), workspace = path.join(root, "workspace");
 preparePackageArchiveDirectory(cache); fs.mkdirSync(workspace, { mode: 0o700 });
 const bytes = Buffer.from("fixture"), value = receipt(bytes); fs.writeFileSync(path.join(cache, value.archive), bytes, { mode: 0o600 });
 fs.writeFileSync(path.join(workspace, "borrowed"), "preserve");
 const alias = path.join(root, "alias"); fs.symlinkSync(workspace, alias);
 let consumed = 0;
 for (const selected of [undefined, "relative", "/", `${workspace}/../workspace`, alias]) {
  assert.throws(() => withVerifiedPackageArchive(cache, value, selected, () => { consumed++; }));
  assert.equal(consumed, 0, "invalid workspace cannot reach the consumer");
 }
 fs.chmodSync(workspace, 0o755);
 assert.throws(() => withVerifiedPackageArchive(cache, value, workspace, () => { consumed++; }));
 assert.equal(consumed, 0, "nonprivate workspace cannot reach the consumer");
 assert.equal(fs.statSync(workspace).mode & 0o777, 0o755, "borrowed parent is not repaired");
 fs.chmodSync(workspace, 0o700);
 withVerifiedPackageArchive(cache, value, workspace, snapshot => assert.deepEqual(fs.readFileSync(snapshot), bytes));
 assert.deepEqual(fs.readdirSync(workspace), ["borrowed"]); assert.equal(fs.readFileSync(path.join(workspace, "borrowed"), "utf8"), "preserve");
 assert.deepEqual(fs.readdirSync(cache), [value.archive]); assert.deepEqual(fs.readFileSync(path.join(cache, value.archive)), bytes);
});
