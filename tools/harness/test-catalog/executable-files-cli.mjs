import path from "node:path";
import { executableRepositoryPaths, inspectExecutableRepositoryPath, restrictedExecutableInputRoots } from "./restricted-input-boundary.mjs";

// Internal file admission shared by static-analysis consumers. Read the policy
// from the tool's repository even when inspecting an isolated fixture checkout.
const policyRoot = path.resolve(import.meta.dirname, "../../..");
const [root, ...includes] = process.argv.slice(2);
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
if (files.length) process.stdout.write(`${files.join("\0")}\0`);
