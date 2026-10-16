#!/usr/bin/env node

import path from "node:path";
import { parseStrictJSON, readRetainedArtifactBytes, resolveRetainedArtifactIdentity, validateSchemaSync } from "../contract/index.mjs";
import { atomicLocalFile } from "../runtime/secure-local-files.mjs";
import { buildSourceSnapshot, loadTestCatalog } from "../test-catalog/index.mjs";
import { compileRetainedRun } from "../scheduler/work-graph/index.mjs";
import { auditOwnerEvidence } from "./owner-evidence-audit.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const usage = () => Object.assign(new Error("usage: canonical-evidence-audit-cli.mjs --owner <owner-id> --evidence-roots-file <path>"), { exit_code: 2 });
function parseArgs(argv) {
  const options = { owner: "", manifest: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const key = argv[index] === "--owner" ? "owner" : argv[index] === "--evidence-roots-file" ? "manifest" : null;
    if (!key || options[key] || !argv[index + 1]) throw usage();
    options[key] = argv[++index];
  }
  if (!options.owner || !options.manifest) throw usage();
  return options;
}
let output;
const publish = (report) => {
  validateSchemaSync(report.schema_id, report);
  atomicLocalFile(output, `${JSON.stringify(report, null, 2)}\n`);
};
try {
  const options = parseArgs(process.argv.slice(2));
  const manifestFile = path.resolve(root, options.manifest);
  const manifest = parseStrictJSON(readRetainedArtifactBytes(path.dirname(manifestFile), path.basename(manifestFile), 1024 * 1024).toString("utf8"));
  const sourceDigest = buildSourceSnapshot(root).digest;
  const catalog = loadTestCatalog(root);
  const surface = parseStrictJSON(readRetainedArtifactBytes(root, "tools/task_surface_owner.json", 4 * 1024 * 1024).toString("utf8"));
  const identity = resolveRetainedArtifactIdentity("test-evidence-audit", {
    ...process.env,
    OWNER: options.owner,
    EVIDENCE_ROOTS_FILE: options.manifest,
    // The dispatcher deliberately removes Make inputs after translating them
    // to argv. These two values were explicitly supplied to this CLI.
    CARTULARY_MAKE_INPUT_SOURCES: "OWNER=cli EVIDENCE_ROOTS_FILE=cli",
  });
  output = path.join(identity.run_root, "test-evidence-audit/audit.json");
  const report = await auditOwnerEvidence({ manifest, manifestFile, owner: options.owner, catalog, surface, sourceDigest, compileRun: (value) => compileRetainedRun(root, value) });
  if (buildSourceSnapshot(root).digest !== sourceDigest) throw Object.assign(new Error("executable source changed during audit"), { exit_code: 11 });
  publish(report);
  process.stdout.write(`[EVIDENCE] status=pass owner=${options.owner} rows=${report.accepted.reduce((sum, entry) => sum + entry.row_ids.length, 0)} partitions=${report.accepted.length} unused=${report.unused_inputs.length} artifact=test-evidence-audit/audit.json\n`);
} catch (error) {
  if (output && error.auditReport) {
    try { publish(error.auditReport); } catch (publicationError) { process.stderr.write(`audit result publication failed: ${publicationError.message}\n`); }
  }
  process.stderr.write(`${error.message}\n`);
  process.exitCode = error.exit_code ?? 11;
}
