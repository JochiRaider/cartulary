import { mkdirSync, mkdtempSync, rmSync } from "node:fs";
import path from "node:path";
import { createRunnerContext } from "../contract/index.mjs";

const repoRoot = path.resolve(import.meta.dirname, "../../..");

export function createScratch(t, prefix) {
  const parent = createRunnerContext({ repoRoot }).scratchRoot;
  mkdirSync(parent, { recursive: true });
  const directory = mkdtempSync(path.join(parent, prefix));
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  return directory;
}
