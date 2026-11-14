import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { closeSync, fchmodSync, fstatSync, mkdtempSync } from "node:fs";
import path from "node:path";
import { CommandFailure } from "../harness/runtime/command-failure.mjs";
import { atomicLocalFile, openDirectory, readLocalChunks, removePrivateTree, privateDirectory } from "../harness/runtime/secure-local-files.mjs";
import { validatePackageReceipt } from "./package-identity.mjs";

export const packageFileLimit = 64 * 1024 * 1024 * 1024;

export function hashPackageFile(file) {
  const hash = createHash("sha256");
  for (const chunk of readLocalChunks(file, { maximum: packageFileLimit, privateFile: false })) hash.update(chunk);
  return hash.digest("hex");
}

export function preparePackageArchiveDirectory(directory) {
  const descriptor = openDirectory(directory, { create: true });
  try {
    const before = fstatSync(descriptor), mode = before.mode & 0o7777;
    if (before.uid !== process.getuid() || ![0o700, 0o755].includes(mode)) throw new Error("unsafe package archive directory");
    if (mode === 0o755) fchmodSync(descriptor, 0o700);
    const current = openDirectory(directory, { privateLeaf: true });
    try {
      const after = fstatSync(current);
      if (before.ino !== after.ino || before.dev !== after.dev) throw new Error("package archive directory changed");
    } finally { closeSync(current); }
  } finally { closeSync(descriptor); }
}

export function publishPackageArchive(directory, manifestDigest, source) {
  assert.match(manifestDigest, /^[a-f0-9]{64}$/u);
  preparePackageArchiveDirectory(directory);
  const destination = path.join(directory, `${manifestDigest}.tar`);
  let digest;
  function* bytes() {
    const hash = createHash("sha256");
    for (const chunk of readLocalChunks(source, { maximum: packageFileLimit, privateFile: false })) {
      hash.update(chunk); yield chunk;
    }
    digest = hash.digest("hex");
  }
  try { atomicLocalFile(destination, bytes()); }
  catch (error) {
    // Only a fully read input and clean publication collision can reuse a winner.
    if (error.code !== "EEXIST" || !digest || error.cleanupFailures?.length) throw error;
    assert.equal(hashPackageFile(destination), digest, "same manifest cannot publish different archive bytes");
  }
  return { archive: path.basename(destination), archive_sha256: digest };
}

export function withVerifiedPackageArchive(directory, receipt, workspaceRoot, consume) {
  validatePackageReceipt(receipt);
  assert.ok(typeof workspaceRoot === "string" && path.isAbsolute(workspaceRoot) &&
    path.normalize(workspaceRoot) === workspaceRoot && workspaceRoot !== path.parse(workspaceRoot).root,
    "package snapshot requires an explicit canonical workspace");
  privateDirectory(workspaceRoot);
  preparePackageArchiveDirectory(directory);
  let stage, primary;
  try {
    stage = mkdtempSync(path.join(workspaceRoot, "snapshot-"));
    const snapshot = path.join(stage, "package.tar");
    function* bytes() {
      const hash = createHash("sha256");
      for (const chunk of readLocalChunks(path.join(directory, receipt.archive), { maximum: packageFileLimit, privateFile: false })) {
        hash.update(chunk); yield chunk;
      }
      assert.equal(hash.digest("hex"), receipt.archive_sha256, "package archive digest mismatch");
    }
    atomicLocalFile(snapshot, bytes());
    return consume(snapshot);
  } catch (error) { primary = error; throw error; }
  finally {
    if (stage) {
      try { removePrivateTree(stage); }
      catch (cause) {
        const failure = new CommandFailure("package archive cleanup failed", { failure_class: "harness", failure_reason: "cleanup_error" }, { cause });
        if (primary) (primary.cleanupFailures ??= []).push(failure);
        else throw failure;
      }
    }
  }
}
