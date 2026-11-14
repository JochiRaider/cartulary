import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  inspectExecutableRepositoryPath,
  isDocumentationOnlyPath,
  restrictedExecutableInputRoots,
} from "../test-catalog/index.mjs";

function normalized(relative) {
  if (typeof relative !== "string" || path.isAbsolute(relative) || /[\\\x00-\x1f]/u.test(relative) ||
      relative.split("/").some(part => !part || part === "." || part === "..")) {
    throw new Error("generation input must be a normalized repository-relative path");
  }
  return relative;
}

function sameFile(left, right) {
  return ["dev", "ino", "mode", "size", "mtimeMs", "ctimeMs"].every(key => left[key] === right[key]);
}

export function copyGenerationInputs(root, destination, requested, restrictedRoots = restrictedExecutableInputRoots(root)) {
  const sourceRoot = path.resolve(root), outputRoot = path.resolve(destination);
  if (sourceRoot === outputRoot || outputRoot.startsWith(`${sourceRoot}/`) || sourceRoot.startsWith(`${outputRoot}/`)) {
    throw new Error("generation scratch and source roots must be disjoint");
  }
  if (!Array.isArray(requested) || requested.length === 0) throw new Error("generation input selection is empty");
  for (const relative of requested) {
    normalized(relative);
    if (isDocumentationOnlyPath(relative, restrictedRoots)) throw new Error("generation input is outside executable scope");
  }
  const copied = new Set();
  const copy = relative => {
    // Skip before lstat/open/readdir. Never copy prose and filter it afterward.
    if (isDocumentationOnlyPath(relative, restrictedRoots) || copied.has(relative)) return;
    const admitted = inspectExecutableRepositoryPath(sourceRoot, relative);
    if (!admitted) throw new Error(`required generate-drift input missing: ${relative}`);
    const { absolute, stat } = admitted;
    const output = path.join(outputRoot, relative);
    if (stat.isDirectory()) {
      fs.mkdirSync(output, { recursive: true, mode: 0o700 });
      copied.add(relative);
      for (const name of fs.readdirSync(absolute).sort()) copy(`${relative}/${name}`);
      return;
    }
    if (!stat.isFile()) throw new Error(`generation input must be a regular file: ${relative}`);
    const descriptor = fs.openSync(absolute, fs.constants.O_RDONLY | fs.constants.O_NOFOLLOW | fs.constants.O_NONBLOCK);
    try {
      if (!sameFile(stat, fs.fstatSync(descriptor))) throw new Error(`generation input changed before copying: ${relative}`);
      fs.mkdirSync(path.dirname(output), { recursive: true, mode: 0o700 });
      // The kernel-held descriptor pins the admitted regular file. Copying via
      // that descriptor cannot follow a replacement source-path symlink.
      fs.copyFileSync(`/proc/self/fd/${descriptor}`, output, fs.constants.COPYFILE_EXCL);
      if (!sameFile(stat, fs.fstatSync(descriptor)) || !sameFile(stat, fs.lstatSync(absolute))) {
        throw new Error(`generation input changed while copying: ${relative}`);
      }
      fs.chmodSync(output, stat.mode & 0o777);
    } finally {
      fs.closeSync(descriptor);
    }
    copied.add(relative);
  };
  for (const relative of requested) copy(relative);
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const [root, destination, ...requested] = process.argv.slice(2);
    if (!root || !destination) throw new Error("generation copy requires source and scratch roots");
    copyGenerationInputs(root, destination, requested);
  } catch (error) {
    console.error(error.message);
    process.exitCode = 11;
  }
}
