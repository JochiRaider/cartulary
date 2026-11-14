import { createHash } from "node:crypto";
import { closeSync, constants, fstatSync, openSync, readSync, readdirSync } from "node:fs";
import path from "node:path";
import { openDirectory, semanticJSONDigest } from "../harness/contract/index.mjs";
import { isDocumentationOnlyPath, restrictedExecutableInputRoots } from "../harness/test-catalog/index.mjs";

const repoRoot = path.resolve(import.meta.dirname, "../..");
const restrictedRoots = restrictedExecutableInputRoots(repoRoot);
const maximumBytes = 8 * 1024 * 1024;
const patterns = Object.freeze({
  package: /^(?:license|licence|copying|notice)(?:[._-].*)?$/iu,
  font: /^(?:OFL|LICENSE)(?:\.[^.]+)?$/iu,
  first_party: /^LICENSE$/u,
});

function excluded(file) {
  const relative = path.relative(repoRoot, file).split(path.sep).join("/");
  return isDocumentationOnlyPath(relative, restrictedRoots) || relative.split("/").some(part => /\.(?:mdown|mkd)$/iu.test(part));
}

function readInput(directory, name) {
  let descriptor, primary;
  try {
    descriptor = openSync(`/proc/self/fd/${directory}/${name}`, constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK);
    const before = fstatSync(descriptor);
    if (!before.isFile() || before.size > maximumBytes) throw new Error("license input must be a regular file of at most 8 MiB");
    const bytes = Buffer.alloc(before.size + 1);
    let length = 0;
    while (length < bytes.length) {
      const count = readSync(descriptor, bytes, length, bytes.length - length, length);
      if (count === 0) break;
      length += count;
    }
    const after = fstatSync(descriptor);
    if (length !== before.size || ["dev", "ino", "size", "mode", "uid", "nlink", "mtimeMs", "ctimeMs"].some(key => before[key] !== after[key])) {
      throw new Error("license input changed during acquisition");
    }
    return { name, bytes: length, sha256: `sha256:${createHash("sha256").update(bytes.subarray(0, length)).digest("hex")}` };
  } catch (error) { primary = error; throw error; }
  finally {
    if (descriptor !== undefined) {
      try { closeSync(descriptor); }
      catch (error) { if (!primary) throw error; (primary.cleanupFailures ??= []).push(error); }
    }
  }
}

function inputs(sourceDir, pattern) {
  if (sourceDir === null || sourceDir === undefined || sourceDir === "") return [];
  if (typeof sourceDir !== "string" || !path.isAbsolute(sourceDir)) throw new Error("license directory must be absolute");
  const directory = path.resolve(sourceDir);
  // Lexical exclusion precedes even opening an ancestor directory.
  if (excluded(directory)) throw new Error("documentation-only license input");
  let descriptor, primary;
  try {
    try { descriptor = openDirectory(directory); }
    catch (error) { if (error.code === "ENOENT") return []; throw error; }
    const names = readdirSync(`/proc/self/fd/${descriptor}`)
      .filter(name => pattern.test(name) && !excluded(path.join(directory, name))).sort();
    return names.map(name => readInput(descriptor, name));
  } catch (error) { primary = error; throw error; }
  finally {
    if (descriptor !== undefined) {
      try { closeSync(descriptor); }
      catch (error) { if (!primary) throw error; (primary.cleanupFailures ??= []).push(error); }
    }
  }
}

export function licenseEvidence(sourceDir, { metadata = null, kind = "package" } = {}) {
  if (!Object.hasOwn(patterns, kind)) throw new Error("unknown license input kind");
  let files;
  try { files = inputs(sourceDir, patterns[kind]); }
  catch (error) {
    const rejected = new Error("license input rejected", { cause: error });
    if (error.cleanupFailures) rejected.cleanupFailures = error.cleanupFailures;
    throw rejected;
  }
  if (files.length) return {
    evidenceDigest: semanticJSONDigest({ files }), evidenceSource: "package-distributed license file", issues: [],
  };
  return {
    evidenceDigest: null,
    evidenceSource: metadata ? "package metadata license field" : "missing",
    issues: metadata ? ["license_text_missing"] : ["license_text_missing", "license_metadata_missing"],
  };
}
