// Confined byte admission only; owner manifests and assertions stay with owners.
import { createHash } from "node:crypto";
import {
  chmodSync,
  closeSync,
  constants,
  fstatSync,
  lstatSync,
  mkdirSync,
  openSync,
  readdirSync,
  readSync,
  realpathSync,
  rmSync,
  statSync,
  writeSync,
} from "node:fs";
import path from "node:path";

const segmentPattern = /^[A-Za-z0-9._@+-]+$/u;
export type FileRecord = {
  logical_path: string;
  size_bytes: number;
  sha256: string;
};

export function digest(data: Uint8Array | string): string {
  return createHash("sha256").update(data).digest("hex");
}

export function safePath(value: string): boolean {
  return (
    value.length > 0 &&
    value
      .split("/")
      .every(
        (part) => part !== "." && part !== ".." && segmentPattern.test(part),
      )
  );
}

export function secureDirectory(directory: string, create = false): void {
  if (!path.isAbsolute(directory))
    throw new Error("Fixture directory must be absolute");
  let current = path.parse(directory).root;
  for (const part of path
    .relative(current, directory)
    .split(path.sep)
    .filter(Boolean)) {
    current = path.join(current, part);
    if (create) {
      try {
        mkdirSync(current, { mode: 0o700 });
      } catch (error) {
        if ((error as NodeJS.ErrnoException).code !== "EEXIST") throw error;
      }
    }
    const info = lstatSync(current);
    if (!info.isDirectory() || info.isSymbolicLink())
      throw new Error("Unsafe fixture directory");
  }
}

export function openRegular(root: string, logicalPath: string): number {
  if (!safePath(logicalPath)) throw new Error("Unsafe fixture path");
  const filename = path.join(root, logicalPath);
  secureDirectory(path.dirname(filename));
  if (!lstatSync(filename).isFile())
    throw new Error("Fixture input is not a regular file");
  const fd = openSync(
    filename,
    constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK,
  );
  try {
    const info = fstatSync(fd);
    // Verify the opened identity before reading, including raced parent replacement.
    secureDirectory(path.dirname(filename));
    const named = statSync(filename);
    if (
      !info.isFile() ||
      realpathSync(filename) !== filename ||
      info.dev !== named.dev ||
      info.ino !== named.ino
    ) {
      throw new Error("Fixture input changed during admission");
    }
    return fd;
  } catch (error) {
    closeSync(fd);
    throw error;
  }
}

export function copyVerified(root: string, file: FileRecord, inputs: string): void {
  const source = openRegular(root, file.logical_path);
  try {
    if (
      !Number.isSafeInteger(file.size_bytes) ||
      fstatSync(source).size !== file.size_bytes
    ) {
      throw new Error("Fixture size mismatch");
    }
    const destination = path.join(inputs, file.logical_path);
    secureDirectory(path.dirname(destination), true);
    const target = openSync(destination, "wx", 0o600);
    const hash = createHash("sha256");
    let total = 0;
    try {
      const buffer = Buffer.alloc(64 * 1024);
      for (;;) {
        const count = readSync(
          source,
          buffer,
          0,
          Math.min(buffer.length, file.size_bytes - total + 1),
          null,
        );
        if (!count) break;
        total += count;
        if (total > file.size_bytes)
          throw new Error("Fixture grew during admission");
        hash.update(buffer.subarray(0, count));
        let written = 0;
        while (written < count) {
          const n = writeSync(target, buffer, written, count - written);
          if (!n) throw new Error("Fixture materialization made no progress");
          written += n;
        }
      }
    } finally {
      closeSync(target);
    }
    if (total !== file.size_bytes || hash.digest("hex") !== file.sha256)
      throw new Error("Fixture hash mismatch");
  } finally {
    closeSync(source);
  }
}

// The executing unit's evidence directory, unit-artifacts/<unit> under the
// selected run (TH-HARNESS-REQ-874). Fixture copies and test evidence a unit
// writes belong there, never at the run root's top level. The Go form is
// fixtureartifacts.SelectedUnitRoot.
export function selectedUnitRoot(
  repoRoot: string,
  environment: NodeJS.ProcessEnv = process.env,
): string {
  const id = environment.CARTULARY_TEST_RUN_ID ?? "";
  const results = environment.CARTULARY_TEST_RESULTS_DIR;
  if (!safePath(id) || id.includes("/") || !results)
    throw new Error("Fixture execution requires a selected harness run");
  const run = path.resolve(repoRoot, results, id);
  const unit = environment.CARTULARY_STEP_ARTIFACT_DIR ?? "";
  if (
    !path.isAbsolute(unit) ||
    path.normalize(unit) !== unit ||
    path.dirname(unit) !== path.join(run, "unit-artifacts")
  )
    throw new Error(
      "Fixture execution requires the executing unit's artifact directory",
    );
  return unit;
}

// Releases a materialization after a passing comparison. Sealed directories
// are reopened without following links; the Go form is
// fixtureartifacts.ReleaseTree.
export function releaseTree(directory: string): void {
  const pending = [directory];
  while (pending.length > 0) {
    const current = pending.pop() ?? "";
    const info = lstatSync(current, { throwIfNoEntry: false });
    if (!info || !info.isDirectory()) continue;
    chmodSync(current, 0o700);
    for (const entry of readdirSync(current, { withFileTypes: true }))
      if (entry.isDirectory()) pending.push(path.join(current, entry.name));
  }
  rmSync(directory, { recursive: true, force: true });
}

export function seal(directory: string): void {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const filename = path.join(directory, entry.name);
    if (entry.isDirectory()) seal(filename);
    else chmodSync(filename, 0o400);
  }
  chmodSync(directory, 0o500);
}
