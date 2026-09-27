import { randomBytes } from "node:crypto";
import { closeSync, constants, fstatSync, fsyncSync, linkSync, mkdirSync, openSync, readSync, renameSync, unlinkSync, writeFileSync } from "node:fs";
import path from "node:path";

// Anchor each traversal in an open directory. O_NOFOLLOW on just the final file
// is insufficient when a caller can replace an intermediate path with a link.
export function openDirectory(directory, { create = false, privateLeaf = false } = {}) {
  const absolute = path.resolve(directory);
  let descriptor = openSync(path.parse(absolute).root, constants.O_RDONLY | constants.O_DIRECTORY | constants.O_NOFOLLOW);
  try {
    const parts = absolute.split(path.sep).filter(Boolean);
    for (const [index, part] of parts.entries()) {
      const anchored = `/proc/self/fd/${descriptor}/${part}`;
      if (create) try { mkdirSync(anchored, { mode: 0o700 }); } catch (error) { if (error.code !== "EEXIST") throw error; }
      const next = openSync(anchored, constants.O_RDONLY | constants.O_DIRECTORY | constants.O_NOFOLLOW);
      closeSync(descriptor); descriptor = next;
      const info = fstatSync(descriptor);
      if (index === parts.length - 1 && privateLeaf && (info.uid !== process.getuid() || (info.mode & 0o777) !== 0o700)) throw new Error("unsafe local directory");
    }
    return descriptor;
  } catch (error) { closeSync(descriptor); throw error; }
}
export function privateDirectory(directory) {
  const descriptor = openDirectory(directory, { create: true, privateLeaf: true }); closeSync(descriptor);
  return directory;
}
export function readLocalFile(file, { maximum = 65536, privateFile = true } = {}) {
  const parent = openDirectory(path.dirname(file));
  let descriptor;
  try {
    descriptor = openSync(`/proc/self/fd/${parent}/${path.basename(file)}`, constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK);
    const before = fstatSync(descriptor);
    if (!before.isFile() || before.nlink !== 1 || before.uid !== process.getuid() || (before.mode & 0o022) !== 0 || (privateFile && (before.mode & 0o777) !== 0o600) || before.size > maximum) throw new Error("unsafe local file");
    const bytes = Buffer.alloc(before.size + 1);
    let offset = 0;
    while (offset < bytes.length) {
      const count = readSync(descriptor, bytes, offset, bytes.length - offset, offset);
      if (!count) break;
      offset += count;
    }
    const after = fstatSync(descriptor);
    if (offset !== before.size || before.size !== after.size || before.mtimeMs !== after.mtimeMs || before.ctimeMs !== after.ctimeMs || before.ino !== after.ino || after.nlink !== 1) throw new Error("changed local file");
    return bytes.subarray(0, offset);
  } finally { if (descriptor !== undefined) closeSync(descriptor); closeSync(parent); }
}
export function atomicLocalFile(file, bytes, { replace = false } = {}) {
  const parent = openDirectory(path.dirname(file), { create: true, privateLeaf: true });
  const prefix = `/proc/self/fd/${parent}/`;
  const temporary = `${prefix}.publishing-${randomBytes(16).toString("hex")}`;
  const destination = `${prefix}${path.basename(file)}`;
  let descriptor;
  let exists = false;
  try {
    descriptor = openSync(temporary, constants.O_CREAT | constants.O_EXCL | constants.O_WRONLY | constants.O_NOFOLLOW, 0o600); exists = true;
    writeFileSync(descriptor, bytes); fsyncSync(descriptor); closeSync(descriptor); descriptor = undefined;
    if (replace) {
      // Only this private directory's owner may replace a previously published
      // locator. Validate the old inode before replacement; never follow it.
      try { readLocalFile(file, { maximum: 1048576 }); } catch (error) { if (error.code !== "ENOENT") throw error; }
      renameSync(temporary, destination); exists = false;
    } else {
      linkSync(temporary, destination); unlinkSync(temporary); exists = false;
    }
    fsyncSync(parent);
    const currentParent = openDirectory(path.dirname(file), { privateLeaf: true });
    try {
      const original = fstatSync(parent), current = fstatSync(currentParent);
      if (original.dev !== current.dev || original.ino !== current.ino) throw new Error("publication directory changed");
    } finally { closeSync(currentParent); }
  } finally {
    if (descriptor !== undefined) closeSync(descriptor);
    if (exists) unlinkSync(temporary);
    closeSync(parent);
  }
}
