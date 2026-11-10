import * as fs from "node:fs";
import path from "node:path";

export function validateRepositoryRelativePath(value, label = "path") {
  if (typeof value !== "string" || !value || value.trim() !== value) throw new Error(`${label} must be a non-empty, unpadded string`);
  if (value.includes("\\")) throw new Error(`${label} must use repository-relative forward slashes`);
  if (path.posix.isAbsolute(value) || /^[A-Za-z]:\//u.test(value)) throw new Error(`${label} must be repository-relative`);
  if (/[*?[\]{}]/u.test(value)) throw new Error(`${label} must be an exact path without glob syntax`);
  if (value.split("/").some(part => !part || part === "." || part === "..") || path.posix.normalize(value) !== value) {
    throw new Error(`${label} must be normalized without empty, '.' or '..' segments`);
  }
  return value;
}

// Filesystem admission is shared; traversal, source scopes and policy authority
// stay with each collector. The injected IO boundary is used for deterministic
// failure qualification, including when the test runner has elevated privileges.
export function createSourceAdmission(root, io = fs) {
  const canonicalRoot = io.realpathSync(path.resolve(root));
  function access(relative, label, operation) {
    try { return operation(); }
    catch (cause) {
      const detail = cause.code === "ENOENT" ? "source path does not exist (ENOENT)" : cause.message;
      throw new Error(`${label}: ${relative}: ${detail}`, { cause });
    }
  }
  function resolve(relative, kind, label) {
    validateRepositoryRelativePath(relative, label);
    let absolute = canonicalRoot;
    const parts = relative.split("/");
    for (const [index, part] of parts.entries()) {
      absolute = path.join(absolute, part);
      const stat = access(relative, label, () => io.lstatSync(absolute));
      if (stat.isSymbolicLink()) throw new Error(`${label}: linked source ${relative} contains a symlink component`);
      const expected = index < parts.length - 1 || kind === "directory" ? stat.isDirectory() : stat.isFile();
      if (!expected) throw new Error(`${label}: source ${relative} must resolve to a regular ${index < parts.length - 1 ? "directory" : kind}`);
    }
    return absolute;
  }
  return {
    root: canonicalRoot,
    directory(relative, label = "source root") { return resolve(relative, "directory", label); },
    entries(relative, label = "source directory") {
      const absolute = resolve(relative, "directory", label);
      const entries = access(relative, label, () => io.readdirSync(absolute, { withFileTypes: true }));
      for (const entry of entries) {
        if (entry.isSymbolicLink()) throw new Error(`${label}: linked source ${relative}/${entry.name} is a symlink`);
        if (!entry.isFile() && !entry.isDirectory()) throw new Error(`${label}: nonregular source ${relative}/${entry.name}`);
      }
      return entries.sort((a,b)=>a.name.localeCompare(b.name));
    },
    file(relative, label = "source file") {
      const absolute = resolve(relative, "file", label);
      const content = access(relative, label, () => io.readFileSync(absolute, "utf8"));
      if (relative.endsWith(".go") && /^\s*\/\/go:linkname\b/mu.test(content)) throw new Error(`${label}: linked Go declaration ${relative}`);
      return { absolute, relative, content };
    },
  };
}
