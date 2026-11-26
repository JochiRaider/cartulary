import path from "node:path";
import { parseStrictJSON, readRetainedArtifactBytes, resolveRetainedArtifactIdentity, validateSchemaSync } from "../harness/contract/index.mjs";
import { atomicLocalFile } from "../harness/runtime/secure-local-files.mjs";
import { buildSourceSnapshot, loadTestCatalog } from "../harness/test-catalog/index.mjs";
import { loadExecutionTopology } from "../harness/generated-artifacts/execution-topology.mjs";
import { resolveReleaseRoots, verifyReleaseEvidence } from "./verify-release-evidence.mjs";
import { inputSourcesEnvironment } from "../harness/contract/run-context.mjs";

const root = path.resolve(import.meta.dirname, "../..");
try {
  const args = process.argv.slice(2);
  if (args.length !== 2 || args[0] !== "--evidence-roots-file" || !args[1]) throw new Error("usage: make release-evidence-verify EVIDENCE_ROOTS_FILE=<file>");
  const file = path.resolve(root, args[1]);
  const manifest = parseStrictJSON(readRetainedArtifactBytes(path.dirname(file), path.basename(file), 1024 * 1024).toString("utf8"));
  const policy = parseStrictJSON(readRetainedArtifactBytes(root, "contracts/verification/release-qualification.v2.json", 1024 * 1024).toString("utf8"));
  const catalog = loadTestCatalog(root);
  const sourceDigest = buildSourceSnapshot(root).digest;
  const report = await verifyReleaseEvidence({roots: resolveReleaseRoots(manifest, file), sourceDigest, catalog,
    surface: loadExecutionTopology({ root }).taskSurface, policy, repoRoot: root});
  if (buildSourceSnapshot(root).digest !== sourceDigest) throw new Error("source changed during evidence verification");
  const identity = resolveRetainedArtifactIdentity("release-evidence-verify", {...process.env, EVIDENCE_ROOTS_FILE: args[1], ...inputSourcesEnvironment(["EVIDENCE_ROOTS_FILE=cli"])});
  validateSchemaSync(report.schema_id, report);
  atomicLocalFile(path.join(identity.run_root, "release-evidence-verify/closure.json"), `${JSON.stringify(report, null, 2)}\n`);
  const runs = Object.values(report.roots).map(value => `${value.target}=${value.run_id}`).join(" ");
  process.stdout.write(`[RELEASE-EVIDENCE] status=pass ${runs} package=${report.archive_sha256} artifact=release-evidence-verify/closure.json\n`);
} catch (error) {
  process.stderr.write(`release evidence verification failed: ${error.message}\n`);
  process.exitCode = 11;
}
