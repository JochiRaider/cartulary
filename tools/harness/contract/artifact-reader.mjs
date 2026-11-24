import { closeSync, constants, fstatSync, openSync, readSync } from "node:fs";
import path from "node:path";
import { openDirectory } from "./secure-directory.mjs";

export function assertMachineEvidencePath(file) {
  const segments = path.resolve(file).split(path.sep).filter(Boolean);
  if (segments.some((part) => /^(?:docs|readme(?:\..*)?)$/iu.test(part) || /\.(?:md|markdown)$/iu.test(part))) {
    throw new Error("documentation is not an executable evidence input");
  }
}

// Linux descriptor-relative traversal pins each admitted directory before
// opening the next component. Renaming an ancestor cannot redirect the read.
// This boundary accepts retained machine evidence, never documentation.
function withRetainedArtifact(root, relative, consume) {
  if (typeof relative !== "string" || path.isAbsolute(relative) ||
      relative.includes("\\") || relative.includes("\0") ||
      relative.split("/").some((part) => !part || part === "." || part === "..")) {
    throw new Error("retained artifact path must be normalized and relative");
  }
  const absoluteRoot = path.resolve(root);
  const file = path.join(absoluteRoot, relative);
  assertMachineEvidencePath(file);
  const descriptors = [];
  const noFollow = constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK;
  let primaryFailure;
  try {
    descriptors.push(openDirectory(path.dirname(file)));
    descriptors.push(openSync(`/proc/self/fd/${descriptors[0]}/${path.basename(file)}`, noFollow));
    const descriptor = descriptors.at(-1);
    const before = fstatSync(descriptor, { bigint: true });
    if (!before.isFile() || before.nlink !== 1n) {
      throw new Error("retained artifact is not an unaliased regular file");
    }
    return consume(descriptor, before);
  } catch (error) {
    primaryFailure = error;
    throw error;
  } finally {
    const failures = [];
    for (const descriptor of descriptors.reverse()) {
      try { closeSync(descriptor); } catch (error) { failures.push(error); }
    }
    if (failures.length > 0) {
      if (primaryFailure) primaryFailure.cleanupErrors = failures;
      else throw new AggregateError(failures, "retained artifact descriptor cleanup failed");
    }
  }
}

// Presence checks use the same path admission as reads, without loading large
// event streams a second time. A later consumer must still open securely.
export function assertRetainedArtifactFile(root, relative) {
  withRetainedArtifact(root, relative, () => {});
}

export function readRetainedArtifactBytes(root, relative, maxBytes) {
  if (!Number.isSafeInteger(maxBytes) || maxBytes < 1) {
    throw new Error("retained artifact byte bound is invalid");
  }
  return withRetainedArtifact(root, relative, (descriptor, before) => {
    if (before.size > BigInt(maxBytes)) throw new Error("retained artifact is not bounded by the admitted byte limit");
    // One extra byte detects growth without reading beyond the declared bound.
    const buffer = Buffer.alloc(Number(before.size) + 1);
    let read = 0;
    while (read < buffer.length) {
      const count = readSync(descriptor, buffer, read, buffer.length - read, read);
      if (count === 0) break;
      read += count;
    }
    assertUnchanged(descriptor, before, read);
    return buffer.subarray(0, read);
  });
}

// A NUL byte marks binary content, the test Git applies. Classification
// streams, so a binary file of any size returns null at its first NUL; only
// text, which the caller keeps, must fit the bound.
export function readRetainedTextBytes(root, relative, maxBytes) {
  if (!Number.isSafeInteger(maxBytes) || maxBytes < 1) {
    throw new Error("retained artifact byte bound is invalid");
  }
  return withRetainedArtifact(root, relative, (descriptor, before) => {
    const chunk = Buffer.alloc(64 * 1024);
    const parts = [];
    let read = 0;
    for (;;) {
      const count = readSync(descriptor, chunk, 0, chunk.length, read);
      if (count === 0) break;
      const bytes = chunk.subarray(0, count);
      if (bytes.includes(0)) return null;
      read += count;
      if (read > maxBytes) throw new Error("retained text is not bounded by the admitted byte limit");
      parts.push(Buffer.from(bytes));
    }
    assertUnchanged(descriptor, before, read);
    return Buffer.concat(parts, read);
  });
}

function assertUnchanged(descriptor, before, read) {
  const after = fstatSync(descriptor, { bigint: true });
  if (BigInt(read) !== before.size || after.size !== before.size || after.nlink !== 1n ||
      after.mtimeNs !== before.mtimeNs || after.ctimeNs !== before.ctimeNs) {
    throw new Error("retained artifact changed during read");
  }
}
