import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { closeSync, constants, fstatSync, lstatSync, openSync, readSync, readlinkSync } from "node:fs";
import path from "node:path";

import { canonicalJSONString } from "../contract/index.mjs";
import { executableInputPathspec, inspectExecutableRepositoryPath, restrictedExecutableInputRoots } from "./restricted-input-boundary.mjs";

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function sameFile(left, right) {
  return ["dev", "ino", "mode", "size", "mtimeMs", "ctimeMs"].every((key) => left[key] === right[key]);
}

// Reuse one bounded buffer for the whole snapshot. A source digest must not
// require a file-sized allocation or combine bytes from a changing inode.
function fileDigest(file, expected, buffer) {
  const descriptor = openSync(file, constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK);
  try {
    const before = fstatSync(descriptor);
    if (!before.isFile() || !sameFile(expected, before)) throw new Error("source snapshot file changed before hashing");
    const hash = createHash("sha256");
    let offset = 0;
    while (offset < before.size) {
      const count = readSync(descriptor, buffer, 0, Math.min(buffer.length, before.size - offset), offset);
      if (count === 0) throw new Error("source snapshot file truncated during hashing");
      hash.update(buffer.subarray(0, count));
      offset += count;
    }
    if (readSync(descriptor, buffer, 0, 1, offset) !== 0 ||
        !sameFile(before, fstatSync(descriptor)) || !sameFile(before, lstatSync(file))) {
      throw new Error("source snapshot file changed during hashing");
    }
    return `sha256:${hash.digest("hex")}`;
  } finally {
    closeSync(descriptor);
  }
}

function validateRepositoryPath(relativePath) {
  if (
    path.isAbsolute(relativePath) ||
    relativePath.includes("\\") ||
    path.posix.normalize(relativePath) !== relativePath ||
    relativePath.startsWith("../")
  ) {
    throw new Error(`source snapshot path is not normalized: ${relativePath}`);
  }
  return relativePath;
}

function isDocumentationOnlyPath(relativePath, restrictedRoots) {
  const extension = path.posix.extname(relativePath).toLowerCase();
  return (
    restrictedRoots.some(
      (restrictedRoot) =>
        relativePath === restrictedRoot ||
        relativePath.startsWith(`${restrictedRoot}/`),
    ) ||
    extension === ".md" ||
    extension === ".markdown" ||
    /^readme(?:\.|$)/iu.test(path.posix.basename(relativePath))
  );
}

function repositoryPaths(root, restrictedRoots) {
  return execFileSync(
    "git",
    ["ls-files", "--cached", "--others", "--exclude-standard", "-z", "--", ...executableInputPathspec(restrictedRoots)],
    { cwd: root, encoding: "utf8", maxBuffer: 128 * 1024 * 1024 },
  )
    .split("\0")
    .filter(Boolean)
    .map(validateRepositoryPath)
    .filter((relativePath) => !isDocumentationOnlyPath(relativePath, restrictedRoots))
    .filter((relativePath) => inspectExecutableRepositoryPath(root, relativePath) !== null)
    .sort(asciiCompare);
}

export function buildSourceSnapshot(root, options = {}) {
  const resolvedRoot = path.resolve(root);
  const restrictedRoots =
    options.restrictedRoots ?? restrictedExecutableInputRoots(resolvedRoot);
  const paths = repositoryPaths(resolvedRoot, restrictedRoots);
  if (new Set(paths).size !== paths.length) {
    throw new Error("source snapshot contains duplicate repository paths");
  }
  const buffer = Buffer.allocUnsafe(64 * 1024);
  const entries = paths.map((relativePath) => {
    const admitted = inspectExecutableRepositoryPath(resolvedRoot, relativePath);
    if (!admitted) throw new Error("source snapshot input disappeared during admission");
    const { absolute: absolutePath, stat } = admitted;
    if (stat.isSymbolicLink()) {
      return {
        path: relativePath,
        kind: "symlink",
        mode: (stat.mode & 0o7777).toString(8).padStart(4, "0"),
        byte_digest: sha256(Buffer.from(readlinkSync(absolutePath), "utf8")),
      };
    }
    if (!stat.isFile()) {
      throw new Error(`source snapshot path is not a regular file or symlink: ${relativePath}`);
    }
    return {
      path: relativePath,
      kind: "file",
      mode: (stat.mode & 0o7777).toString(8).padStart(4, "0"),
      byte_digest: fileDigest(absolutePath, stat, buffer),
    };
  });
  const snapshot = { schema_id: "cartulary.source_snapshot_digest.v2", entries };
  return {
    digest: sha256(Buffer.from(canonicalJSONString(snapshot), "utf8")),
    entries,
    file_count: entries.length,
  };
}


// Provenance follows the same executable boundary as the source digest. Git's
// pathspec excludes documentation before worktree status inspection.
export function executableProvenance(root) {
  const source_commit = execFileSync("git", ["rev-parse", "HEAD"], { cwd: root, encoding: "utf8" }).trim();
  const status = execFileSync("git", ["status", "--porcelain=v1", "--untracked-files=all", "--", ...executableInputPathspec(restrictedExecutableInputRoots(root))], { cwd: root, encoding: "utf8", maxBuffer: 32 * 1024 * 1024 });
  return { source_commit, source_state: status === "" ? "clean" : "dirty" };
}
