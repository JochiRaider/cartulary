import { closeSync, constants, fstatSync, mkdirSync, openSync } from "node:fs";
import path from "node:path";

// Anchor each traversal in an open directory. O_NOFOLLOW on just the final file
// is insufficient when a caller can replace an intermediate path with a link.
export function openDirectory(directory, { create = false, privateLeaf = false } = {}) {
  const absolute = path.resolve(directory);
  let descriptor = openSync(path.parse(absolute).root, constants.O_RDONLY | constants.O_DIRECTORY | constants.O_NOFOLLOW);
  const owned = new Set([descriptor]);
  let primary;
  try {
    const parts = absolute.split(path.sep).filter(Boolean);
    for (const [index, part] of parts.entries()) {
      const anchored = `/proc/self/fd/${descriptor}/${part}`;
      let next;
      try { next = openSync(anchored, constants.O_RDONLY | constants.O_DIRECTORY | constants.O_NOFOLLOW); }
      catch (error) {
        if (!create || error.code !== "ENOENT") throw error;
        try { mkdirSync(anchored, { mode: 0o700 }); } catch (error) { if (error.code !== "EEXIST") throw error; }
        // A concurrent creator must pass the same no-follow directory open.
        next = openSync(anchored, constants.O_RDONLY | constants.O_DIRECTORY | constants.O_NOFOLLOW);
      }
      owned.add(next);
      owned.delete(descriptor);
      closeSync(descriptor); descriptor = next;
      if (index === parts.length - 1 && privateLeaf) {
        const info = fstatSync(descriptor);
        if (info.uid !== process.getuid() || (info.mode & 0o777) !== 0o700) throw new Error("unsafe local directory");
      }
    }
    owned.delete(descriptor);
    return descriptor;
  } catch (error) { primary = error; throw error; }
  finally {
    const failures = [];
    for (const current of owned) {
      try { closeSync(current); } catch (error) { failures.push(error); }
    }
    if (failures.length) {
      if (primary) (primary.cleanupErrors ??= []).push(...failures);
      else throw new AggregateError(failures, "directory descriptor cleanup failed");
    }
  }
}
