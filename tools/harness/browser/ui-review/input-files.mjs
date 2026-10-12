import { createHash } from "node:crypto";
import path from "node:path";
import { readLocalFile } from "../../runtime/secure-local-files.mjs";
import { restrictedExecutableInputRoots } from "../../test-catalog/index.mjs";
import { repoRoot } from "./policy.mjs";
import { ReviewFailure } from "./contract.mjs";

export const digest = (bytes) => createHash("sha256").update(bytes).digest("hex");
export const jsonBytes = (value) => Buffer.from(`${JSON.stringify(value)}\n`);
export function inputPath(file, extension) {
  if (typeof file !== "string" || !file || /^[a-z][a-z0-9+.-]*:/iu.test(file) || /[\\\u0000-\u001f]/u.test(file) || file.split("/").includes("..")) throw new ReviewFailure("invalid_request");
  const absolute = path.resolve(repoRoot, file);
  const relative = path.relative(repoRoot, absolute).split(path.sep).join("/");
  const restricted = restrictedExecutableInputRoots(repoRoot);
  if (restricted.some((root) => relative === root || relative.startsWith(`${root}/`)) || /(?:^|\/)readme(?:\.[^/]*)?$|\.(?:md|markdown|mdown)$/iu.test(absolute) || (extension && path.extname(absolute).toLowerCase() !== extension)) throw new ReviewFailure("input_boundary");
  return absolute;
}
export function readInput(file, options) {
  try { return readLocalFile(inputPath(file, options.extension), options); }
  catch (cause) { if (cause instanceof ReviewFailure) throw cause; throw new ReviewFailure("unsafe_artifact", { cause }); }
}
