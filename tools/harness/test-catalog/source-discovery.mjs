import { lstatSync, readFileSync, readdirSync, realpathSync } from "node:fs";
import path from "node:path";

// Discovery is source inspection, never runner execution. Prune by lexical
// identity before stat/read so excluded trees are not executable inputs.
const excludedNames = new Set([
  ".git", ".cache", ".cartulary", "coverage", "dist", "dist-measurement", "docs", "generated",
  "node_modules", "test-results", "testdata", "vendor", "vendored",
]);

export function excludedSourcePath(relative) {
  return relative.split("/").some((name) => excludedNames.has(name) || name.startsWith(".frontend-publish-")) ||
    /(?:^|\/)README(?:\.[^/]*)?$|\.(?:md|markdown)$/iu.test(relative);
}

export function sourceExclusionPolicy(root) {
  const policy = JSON.parse(readFileSync(path.join(root, "tools/generated_artifact_policy.json"), "utf8"));
  const generatedRoots = policy.generated_roots.map((entry) => entry.path);
  const generatedFiles = new Set(policy.generated_files.map((entry) => entry.path));
  return (relative) => excludedSourcePath(relative) ||
    generatedFiles.has(relative) || generatedRoots.some((entry) => relative === entry || relative.startsWith(`${entry}/`));
}

export function discoverRunnerSources(root, runner) {
  const excluded = sourceExclusionPolicy(root);
  const files = [];
  const canonicalRoot = realpathSync(root);
  function walk(relative) {
    if (excluded(relative)) return;
    // Go ignores source files and recursive package directories beginning with
    // a dot or underscore. Do not admit them by an incidental filename suffix.
    if (runner.runner === "go" && relative.split("/").some((part) => /^[._]/u.test(part))) return;
    const absolute = path.join(root, relative);
    const stat = lstatSync(absolute);
    if (stat.isSymbolicLink()) throw new Error(`source discovery must not traverse a symbolic link: ${relative}`);
    const resolved = realpathSync(absolute);
    if (!resolved.startsWith(`${canonicalRoot}${path.sep}`)) throw new Error(`source discovery escapes repository: ${relative}`);
    if (stat.isDirectory()) {
      for (const entry of readdirSync(absolute).sort()) walk(`${relative}/${entry}`);
    } else if (stat.isFile()) {
      if (runner.runner === "go" && relative.endsWith("_test.go")) files.push(relative);
      if (runner.runner === "playwright" && /^apps\/web\/e2e\/.*\.spec\.tsx?$/u.test(relative)) files.push(relative);
      if (runner.runner === "vitest" && (
        /^apps\/web\/src\/.*\.(?:test|spec)\.tsx?$/u.test(relative) ||
        /^apps\/web\/e2e\/.*\.test\.ts$/u.test(relative) ||
        /^packages\/[^/]+\/src\/.*\.test\.tsx?$/u.test(relative)
      )) files.push(relative);
    }
  }
  for (const approved of runner.approved_roots) {
    if (!/^(?:cmd|internal|apps|packages|tools)(?:\/[A-Za-z0-9._@+-]+)*$/u.test(approved) ||
      approved.split("/").some((segment) => segment === "." || segment === "..") || excluded(approved)) {
      throw new Error(`invalid discovery root: ${approved}`);
    }
    let ancestor = root;
    for (const segment of approved.split("/")) {
      ancestor = path.join(ancestor, segment);
      if (lstatSync(ancestor).isSymbolicLink()) throw new Error(`source discovery must not traverse a symbolic link: ${approved}`);
    }
    walk(approved);
  }
  return [...new Set(files)].sort();
}

export function frontendSourceProject(file) {
  if (file.startsWith("packages/report-renderer/src/")) return "renderer-node";
  return file.startsWith("apps/web/e2e/") ? "harness-node" : "browser-unit";
}
