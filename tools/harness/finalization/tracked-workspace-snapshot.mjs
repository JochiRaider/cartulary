import { constants, closeSync, fstatSync, mkdirSync, openSync, readFileSync, readlinkSync, symlinkSync, unlinkSync, writeFileSync } from "node:fs";
import path from "node:path";
import { executableRepositoryPaths, inspectExecutableRepositoryPath } from "../test-catalog/index.mjs";

function entry(root, relative) {
  const admitted = inspectExecutableRepositoryPath(root, relative);
  if (!admitted) return null;
  const { absolute, stat } = admitted;
  if (stat.isSymbolicLink()) return { kind: "symlink", value: readlinkSync(absolute) };
  if (!stat.isFile()) throw new Error("snapshot input is not a regular file or symbolic link");
  const descriptor = openSync(absolute, constants.O_RDONLY | constants.O_NOFOLLOW);
  try {
    const opened = fstatSync(descriptor);
    if (stat.dev !== opened.dev || stat.ino !== opened.ino) throw new Error("snapshot input changed during admission");
    return { kind: "file", mode: stat.mode & 0o777, value: readFileSync(descriptor) };
  } finally { closeSync(descriptor); }
}

function equal(left, right) {
  if (!left || !right) return left === right;
  return left.kind === right.kind && left.mode === right.mode &&
    (left.kind === "file" ? left.value.equals(right.value) : left.value === right.value);
}

export function snapshotTrackedWorkspace(root, options) {
  return new Map(executableRepositoryPaths(root, options).map((relative) => [relative, entry(root, relative)]));
}

export function restoreTrackedWorkspace(root, snapshot) {
  const restored = [];
  try {
    for (const [relative, expected] of snapshot) {
      const current = entry(root, relative);
      if (equal(current, expected)) continue;
      const absolute = path.join(root, relative);
      if (current) unlinkSync(absolute);
      if (expected) {
        mkdirSync(path.dirname(absolute), { recursive: true });
        if (expected.kind === "symlink") symlinkSync(expected.value, absolute);
        else writeFileSync(absolute, expected.value, { mode: expected.mode, flag: "wx" });
      }
      restored.push(relative);
    }
    return { status: restored.length ? "restored" : "not_needed", restored_files: restored, error: null };
  } catch {
    return { status: "failed", restored_files: restored, error: "tracked executable input restoration failed" };
  }
}
