import { readFileSync } from "node:fs";
import path from "node:path";
import { operationBindingNames, readOperationOwner } from "../execution/operations.mjs";

// reviewForwarded is what review preparation forwards from its parent by name:
// the Make variables operation executions name (the operation runner resolves
// them) except public target inputs, which reach review only through its
// request, and Playwright's own browser location. Everything else a recipe
// child receives is a process or tool basic, or set by the recipe
// (TH-HARNESS 5.5). It is its own module so the command entry reaches it
// through the browser facade without the review recipes.
export function reviewForwarded(root) {
  const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
  const publicInputs = new Set(manifest.targets.flatMap((target) => target.input_contract?.inputs ?? []).map((input) => input.name));
  return [...operationBindingNames(readOperationOwner(root)).filter((name) => !publicInputs.has(name)), "PLAYWRIGHT_BROWSERS_PATH"];
}
