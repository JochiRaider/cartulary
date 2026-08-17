import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { lstatSync, readFileSync, readlinkSync } from "node:fs";
import path from "node:path";

import { canonicalJSONString } from "../contract/index.mjs";
import { restrictedExecutableInputRoots } from "./restricted-input-boundary.mjs";

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
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
    extension === ".markdown"
  );
}

function repositoryPaths(root, restrictedRoots) {
  return execFileSync(
    "git",
    ["ls-files", "--cached", "--others", "--exclude-standard", "-z"],
    { cwd: root, encoding: "utf8", maxBuffer: 128 * 1024 * 1024 },
  )
    .split("\0")
    .filter(Boolean)
    .map(validateRepositoryPath)
    .filter((relativePath) => !isDocumentationOnlyPath(relativePath, restrictedRoots))
    .filter((relativePath) => {
      try {
        lstatSync(path.join(root, relativePath));
        return true;
      } catch (error) {
        if (error?.code === "ENOENT") return false;
        throw error;
      }
    })
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
  const entries = paths.map((relativePath) => {
    const absolutePath = path.join(resolvedRoot, relativePath);
    const stat = lstatSync(absolutePath);
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
      byte_digest: sha256(readFileSync(absolutePath)),
    };
  });
  const snapshot = { schema_id: "cartulary.source_snapshot_digest.v2", entries };
  return {
    digest: sha256(Buffer.from(canonicalJSONString(snapshot), "utf8")),
    entries,
    file_count: entries.length,
  };
}
