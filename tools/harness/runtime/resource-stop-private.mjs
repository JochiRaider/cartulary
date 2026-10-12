// Specialized physical operations; Go chooses recovery order and journal release.
import { readSync } from "node:fs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { borrowRuntime } from "./session-client.mjs";
import { stopFixtureResource } from "./fixture-resource-stop.mjs";
import { stopDiagnosticProcesses } from "../browser/ui-review/diagnostic-processes.mjs";
import path from "node:path";
import { lifecyclePolicy } from "../contract/generated/lifecycle-policy.mjs";
const maximum = lifecyclePolicy.limits.frame;
try {
  const buffer = Buffer.alloc(maximum + 1); let size = 0;
  while (size < buffer.length) { const count = readSync(0, buffer, size, buffer.length - size); if (!count) break; size += count; }
  const data = buffer.subarray(0, size);
  if (data.length > maximum || process.argv.length !== 2) throw new Error("invalid private cleanup");
  const resource = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(data));
  validateSchemaSync("cartulary.harness_runtime_resource.v1", resource);
  const runRoot = path.resolve(process.cwd(), process.env.CARTULARY_TEST_RESULTS_DIR, process.env.CARTULARY_TEST_RUN_ID);
  const runtime = borrowRuntime({ repoRoot: process.cwd(), runRoot });
  if (resource.run_id !== runtime.runID || resource.lease_id !== runtime.leaseID) throw new Error("foreign cleanup");
  if (resource.kind === "diagnostic_scope") await stopDiagnosticProcesses(resource.target);
  else if (resource.kind.endsWith("_process")) await stopFixtureResource("process", resource.target);
  else if (resource.kind === "browser_renderer") await stopFixtureResource("renderer", resource.target);
  else if (resource.kind === "object_store_proxy") await stopFixtureResource("proxy", resource.target);
  else throw new Error("unknown physical cleanup");
  runtime.close();
} catch { process.exitCode = 12; }
