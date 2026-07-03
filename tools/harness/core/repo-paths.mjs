import path from "node:path";

export function resolveRepoPath(repoRoot, file) {
  return path.isAbsolute(file) ? file : path.join(repoRoot, file);
}

export function relToRepo(repoRoot, file) {
  const relative = path.relative(repoRoot, file).replaceAll("\\", "/");
  if (!relative.startsWith("../") && relative !== "..") {
    return relative === "" ? "." : relative;
  }
  return file.replaceAll("\\", "/");
}
