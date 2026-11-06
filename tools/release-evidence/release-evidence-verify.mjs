import path from "node:path";
import { parseStrictJSON, readRetainedArtifactBytes, resolveRetainedArtifactIdentity, validateSchemaSync } from "../harness/contract/index.mjs";
import { atomicLocalFile } from "../harness/runtime/secure-local-files.mjs";
import { buildSourceSnapshot, loadTestCatalog } from "../harness/test-catalog/index.mjs";
import { compileRetainedRun, WorkGraphCompiler } from "../harness/scheduler/work-graph/index.mjs";
import { qualificationRequirements, resolveEvidenceRoots, verifySelectedEvidence } from "./verify-selected-evidence.mjs";

const root = path.resolve(import.meta.dirname, "../..");
try {
  const args = process.argv.slice(2);
  if (args.length !== 2 || args[0] !== "--evidence-roots-file" || !args[1]) throw new Error("usage: make release-evidence-verify EVIDENCE_ROOTS_FILE=<file>");
  const file = path.resolve(root, args[1]);
  const manifest = parseStrictJSON(readRetainedArtifactBytes(path.dirname(file), path.basename(file), 1024 * 1024).toString("utf8"));
  const policy = parseStrictJSON(readRetainedArtifactBytes(root, "contracts/verification/release-qualification.v1.json", 1024 * 1024).toString("utf8"));
  const compiler = new WorkGraphCompiler(root);
  const catalog = loadTestCatalog(root);
  const required = qualificationRequirements(policy, catalog, target => compiler.compileTarget(target));
  const sourceDigest = buildSourceSnapshot(root).digest;
  const report = await verifySelectedEvidence({roots: resolveEvidenceRoots(manifest, file), sourceDigest, catalog,
    required, compileRun: value => compileRetainedRun(root, value)});
  if (buildSourceSnapshot(root).digest !== sourceDigest) throw new Error("source changed during evidence verification");
  const identity = resolveRetainedArtifactIdentity("release-evidence-verify", {...process.env, EVIDENCE_ROOTS_FILE: args[1], CARTULARY_MAKE_INPUT_SOURCES: "EVIDENCE_ROOTS_FILE=cli"});
  validateSchemaSync(report.schema_id, report);
  atomicLocalFile(path.join(identity.run_root, "release-evidence-verify/closure.json"), `${JSON.stringify(report, null, 2)}\n`);
  process.stdout.write(`[RELEASE-EVIDENCE] status=pass runs=${report.accepted.length} package=${report.archive_sha256} artifact=release-evidence-verify/closure.json\n`);
} catch (error) {
  process.stderr.write(`release evidence verification failed: ${error.message}\n`);
  process.exitCode = 11;
}
