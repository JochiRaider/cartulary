import {
  chmodSync,
  createWriteStream,
  existsSync,
  lstatSync,
  mkdirSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

export const secureDirMode = 0o700;
export const secureFileMode = 0o600;

function normalizeOptions(modeOrOptions, defaultMode) {
  if (typeof modeOrOptions === "number") {
    return { mode: modeOrOptions };
  }
  return { mode: defaultMode, ...(modeOrOptions ?? {}) };
}

export function pathIsUnder(parent, child) {
  const relative = path.relative(path.resolve(parent), path.resolve(child));
  return relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative));
}

function assertAllowedPath(targetPath, allowedRoot) {
  if (!allowedRoot) {
    return;
  }
  if (!pathIsUnder(allowedRoot, targetPath)) {
    throw new Error(`artifact path ${targetPath} escapes allowed root ${allowedRoot}`);
  }
}

function assertFinalPathIsNotSymlink(targetPath) {
  if (!existsSync(targetPath)) {
    return;
  }
  if (lstatSync(targetPath).isSymbolicLink()) {
    throw new Error(`refusing to write retained artifact through symlink: ${targetPath}`);
  }
}

export function secureMkdir(dir, modeOrOptions = secureDirMode) {
  const { mode, allowedRoot } = normalizeOptions(modeOrOptions, secureDirMode);
  assertAllowedPath(dir, allowedRoot);
  assertFinalPathIsNotSymlink(dir);
  mkdirSync(dir, { recursive: true, mode });
  chmodSync(dir, mode);
  return dir;
}

export function secureWriteFile(file, content, modeOrOptions = secureFileMode) {
  const { mode, allowedRoot } = normalizeOptions(modeOrOptions, secureFileMode);
  assertAllowedPath(file, allowedRoot);
  assertFinalPathIsNotSymlink(file);
  secureMkdir(path.dirname(file), { mode: secureDirMode, allowedRoot });
  writeFileSync(file, content, { mode });
  chmodSync(file, mode);
}

export function createSecureWriteStream(file, options = {}) {
  const {
    flags = "w",
    mode = secureFileMode,
    allowedRoot = "",
    ...streamOptions
  } = options;
  assertAllowedPath(file, allowedRoot);
  assertFinalPathIsNotSymlink(file);
  secureMkdir(path.dirname(file), { mode: secureDirMode, allowedRoot });
  return createWriteStream(file, { ...streamOptions, flags, mode });
}
