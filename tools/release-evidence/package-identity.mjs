import path from "node:path";
import { parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../harness/contract/index.mjs";
import { atomicLocalFile } from "../harness/runtime/secure-local-files.mjs";

export function readPackageJSON(root, relative, schemaID) {
  try {
    const bytes = readRetainedArtifactBytes(root, relative, 16 * 1024 * 1024);
    const value = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes));
    validateSchemaSync(schemaID, value);
    return value;
  } catch (cause) { throw new Error("package evidence admission failed", { cause }); }
}

export function packageRunContext(environment = process.env) {
  const root = environment.CARTULARY_TEST_RESULTS_DIR, runID = environment.CARTULARY_TEST_RUN_ID;
  if (typeof root !== "string" || !path.isAbsolute(root) || typeof runID !== "string" ||
      !/^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$/u.test(runID)) throw new Error("package requires explicit current run context");
  const runRoot = path.join(root, runID);
  const manifest = readPackageJSON(runRoot, "run-manifest.json", "cartulary.harness_run_manifest.v3");
  if (manifest.run_id !== runID) throw new Error("package run identity mismatch");
  return { runRoot, manifest };
}

export function validatePackageReceipt(receipt) {
  validateSchemaSync("cartulary.local_release_receipt.v1", receipt);
  if (receipt.archive !== `${receipt.manifest_sha256}.tar` ||
      JSON.stringify(receipt.images.map(image => image.role)) !== JSON.stringify(["application", "postgres", "seaweedfs"])) {
    throw new Error("package receipt identity invalid");
  }
  return receipt;
}

export function readPackageReceipt(runRoot, relative = "package-release/artifacts/package.json") {
  return validatePackageReceipt(readPackageJSON(runRoot, relative, "cartulary.local_release_receipt.v1"));
}

export function publishPackageEvidence(runRoot, kind, value) {
  if (kind === "package") validatePackageReceipt(value);
  else if (kind === "inspection") {
    validateSchemaSync("cartulary.local_release_inspection.v1", value);
    const { installed_without_build_toolchain: _installed, occupied_destination_rejected: _occupied, ...receipt } = value;
    validatePackageReceipt({ ...receipt, schema_id: "cartulary.local_release_receipt.v1" });
  } else throw new Error("unknown package evidence kind");
  const bytes = `${JSON.stringify(value)}\n`;
  if (Buffer.byteLength(bytes) > 16 * 1024 * 1024) throw new Error("package evidence exceeds byte limit");
  atomicLocalFile(path.join(runRoot, "package-release/artifacts", `${kind}.json`), bytes);
}

/** The producer and every consumer must identify the very same archive. */
export function verifyPackageIdentity(producer, consumers) {
  validatePackageReceipt(producer);
  for (const [name, receipt] of consumers) {
    validatePackageReceipt(receipt);
    for (const key of ["manifest_sha256", "archive_sha256", "archive", "operator_sha256"]) {
      if (receipt[key] !== producer[key]) throw new Error(`${name}: package identity mismatch`);
    }
    if (JSON.stringify(receipt.images) !== JSON.stringify(producer.images)) throw new Error(`${name}: package image identity mismatch`);
  }
  return { manifest_sha256: producer.manifest_sha256, archive_sha256: producer.archive_sha256 };
}
