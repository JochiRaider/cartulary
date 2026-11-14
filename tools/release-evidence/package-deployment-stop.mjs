// Private ordinary-cleanup worker. Recovery calls the same physical adapter.
import path from "node:path";
import { parseStrictJSON } from "../harness/contract/index.mjs";
import { borrowRuntime } from "../harness/runtime/session-client.mjs";
import { cleanupPackageDeployment } from "./package-deployment.mjs";
const repoRoot = path.resolve(import.meta.dirname, "../..");
let runtime;
try {
 if (process.argv.length !== 3 || Buffer.byteLength(process.argv[2]) > 1024 * 1024) throw new Error("invalid private cleanup");
 const selected = parseStrictJSON(process.argv[2]);
 runtime = borrowRuntime({ repoRoot, runRoot: process.env.CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT });
 const admitted = runtime.inspectResource(selected);
 if (admitted.kind !== "package_deployment") throw new Error("invalid private cleanup kind");
 cleanupPackageDeployment(admitted);
} catch { process.exitCode = 12; }
finally { try { runtime?.close(); } catch { process.exitCode = 12; } }
