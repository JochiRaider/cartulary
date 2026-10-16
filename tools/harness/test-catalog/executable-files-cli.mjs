import path from "node:path";
import { executableRepositoryPaths, inspectExecutableRepositoryPath, restrictedExecutableInputRoots } from "./restricted-input-boundary.mjs";

// Internal file admission shared by static-analysis consumers. Read the policy
// from the tool's repository even when inspecting an isolated fixture checkout.
const policyRoot = path.resolve(import.meta.dirname, "../../..");
const args = process.argv.slice(2);
const lines = args[0] === "--lines";
if (lines) args.shift();
const [root, ...includes] = args;
if (!root || includes.length === 0) throw new Error("repository root and include pathspec required");
const files = executableRepositoryPaths(root, {
  restrictedRoots: restrictedExecutableInputRoots(policyRoot),
  untracked: true,
  includes,
}).filter((relative) => {
  const admitted = inspectExecutableRepositoryPath(root, relative);
  if (!admitted) return false;
  if (admitted.stat.isSymbolicLink()) throw new Error("static analysis input traverses a symbolic link");
  return admitted.stat.isFile();
});
if (lines) {
  // Make's prerequisite/input syntax cannot represent whitespace-bearing paths.
  // Fail explicitly instead of silently splitting one capability into several.
  if (files.some(file => /\s/u.test(file))) throw new Error("build input path contains whitespace");
  for (const include of includes) {
    const normalized = include.replace(/^\.\//u, "").replace(/\/$/u, "");
    if (!files.some(file => file === normalized || file.startsWith(`${normalized}/`)))
      throw new Error("build input root is missing, empty or outside executable scope");
  }
}
const separator = lines ? "\n" : "\0";
if (files.length) process.stdout.write(`${files.join(separator)}${separator}`);
