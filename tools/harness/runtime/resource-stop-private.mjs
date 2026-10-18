// Specialized physical operations; Go chooses recovery order and journal release.
import { readSync } from "node:fs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { borrowRuntime } from "./session-client.mjs";
import { recoverVisualRenderer } from "../browser/index.mjs";
import { lifecyclePolicy } from "../contract/generated/lifecycle-policy.mjs";
const maximum = lifecyclePolicy.limits.frame;
try {
  const buffer = Buffer.alloc(maximum + 1); let size = 0;
  while (size < buffer.length) { const count = readSync(0, buffer, size, buffer.length - size); if (!count) break; size += count; }
  const data = buffer.subarray(0, size);
  if (data.length > maximum || process.argv.length !== 2) throw new Error("invalid private cleanup");
  const resource = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(data));
  validateSchemaSync("cartulary.harness_runtime_resource.v2", resource);
  const runRoot = process.env.CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT;
  const runtime = borrowRuntime({ repoRoot: process.cwd(), runRoot });
  if (resource.run_id !== runtime.runID || resource.lease_id !== runtime.leaseID) throw new Error("foreign cleanup");
  if (resource.kind === "browser_renderer") recoverVisualRenderer(resource.target);
  else throw new Error("unknown physical cleanup");
  runtime.close();
} catch { process.exitCode = 12; }
