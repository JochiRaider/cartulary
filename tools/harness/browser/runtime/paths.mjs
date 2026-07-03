import path from "node:path";

export function browserHarnessScript(repoRoot, scriptName) {
  return path.join(repoRoot, "tools", "harness", "browser", scriptName);
}

export function webserverBatchScript(repoRoot) {
  return browserHarnessScript(repoRoot, "run-playwright-webserver-batch.sh");
}

export function browserKindScript(repoRoot, scriptName) {
  return browserHarnessScript(repoRoot, scriptName);
}
