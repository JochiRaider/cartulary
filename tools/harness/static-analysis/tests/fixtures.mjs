import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";
import { createRunnerContext } from "../../contract/index.mjs";

export const repository = path.resolve(import.meta.dirname, "../../../..");

export function createFixture(t) {
  const scratch = createRunnerContext({ repoRoot: repository }).scratchRoot;
  mkdirSync(scratch, { recursive: true });
  const root = mkdtempSync(path.join(scratch, "static-analysis-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  return {
    root,
    write(file, contents) {
      const target = path.join(root, file);
      mkdirSync(path.dirname(target), { recursive: true });
      writeFileSync(target, typeof contents === "string" ? contents : JSON.stringify(contents));
    },
    readJSON(file) {
      return JSON.parse(readFileSync(path.join(root, file), "utf8"));
    },
  };
}
