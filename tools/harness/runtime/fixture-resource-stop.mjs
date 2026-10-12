// Specialized physical operations; native owners retain dependency ordering.
import { stopOwnedProcess } from "./owned-process.mjs";
import { recoverVisualRenderer } from "../browser/index.mjs";

export async function stopFixtureResource(kind, target) {
  if (kind === "process") { await stopOwnedProcess(target); return; }
  if (kind === "renderer") { recoverVisualRenderer(target); return; }
  throw new Error("unknown resource operation");
}
