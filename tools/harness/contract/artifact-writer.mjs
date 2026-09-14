import {
  chmodSync,
  closeSync,
  constants,
  createWriteStream,
  fchmodSync,
  fstatSync,
  ftruncateSync,
  lstatSync,
  mkdirSync,
  openSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

export const secureDirMode = 0o700;
const secureFileMode = 0o600;

function normalizeOptions(modeOrOptions, defaultMode) {
  if (typeof modeOrOptions === "number") {
    return { mode: modeOrOptions };
  }
  return { mode: defaultMode, ...(modeOrOptions ?? {}) };
}

function pathIsUnder(parent, child) {
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
  const absolute = path.resolve(targetPath);
  let current = path.parse(absolute).root;
  for (const segment of absolute.slice(current.length).split(path.sep)) {
    current = path.join(current, segment);
    let info;
    try {
      info = lstatSync(current);
    } catch (error) {
      if (error.code === "ENOENT") return;
      throw error;
    }
    if (info.isSymbolicLink()) {
      throw new Error(`refusing to write retained artifact through symlink: ${current}`);
    }
  }
}

export function secureMkdir(dir, modeOrOptions = secureDirMode) {
  const { mode, allowedRoot } = normalizeOptions(modeOrOptions, secureDirMode);
  assertAllowedPath(dir, allowedRoot);
  assertFinalPathIsNotSymlink(dir);
  mkdirSync(dir, { recursive: true, mode });
  const info = lstatSync(dir);
  if (!info.isDirectory() || (process.getuid && info.uid !== process.getuid())) {
    throw new Error(`artifact directory ownership or type mismatch: ${dir}`);
  }
  chmodSync(dir, mode);
  return dir;
}

export function secureWriteFile(file, content, modeOrOptions = secureFileMode) {
  const fd = openSecureArtifact(file, modeOrOptions);
  try {
    writeFileSync(fd, content);
  } finally {
    closeSync(fd);
  }
}

function openSecureArtifact(file, modeOrOptions = secureFileMode, flags = "w") {
  const { mode, allowedRoot } = normalizeOptions(modeOrOptions, secureFileMode);
  assertAllowedPath(file, allowedRoot);
  assertFinalPathIsNotSymlink(file);
  secureMkdir(path.dirname(file), { mode: secureDirMode, allowedRoot });
  const access = { w: 0, wx: constants.O_EXCL, a: constants.O_APPEND, ax: constants.O_APPEND | constants.O_EXCL }[flags];
  if (access === undefined) throw new Error(`unsupported secure artifact open mode: ${flags}`);
  const fd = openSync(file, constants.O_WRONLY | constants.O_CREAT | constants.O_NOFOLLOW | constants.O_NONBLOCK | access, mode);
  try {
    const info = fstatSync(fd);
    if (!info.isFile() || (process.getuid && info.uid !== process.getuid())) {
      throw new Error(`artifact file ownership or type mismatch: ${file}`);
    }
    fchmodSync(fd, mode);
    if (flags.startsWith("w")) ftruncateSync(fd, 0);
    return fd;
  } catch (error) {
    closeSync(fd);
    throw error;
  }
}

export function createSecureWriteStream(file, options = {}) {
  const {
    flags = "w",
    mode = secureFileMode,
    allowedRoot = "",
    ...streamOptions
  } = options;
  const fd = openSecureArtifact(file, { mode, allowedRoot }, flags);
  return createWriteStream(file, { ...streamOptions, fd, autoClose: true });
}
