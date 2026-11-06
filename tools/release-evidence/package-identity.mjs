import { validateSchemaSync } from "../harness/contract/index.mjs";

/** The producer and every consumer must identify the very same archive. */
export function verifyPackageIdentity(producer, consumers) {
  validateSchemaSync("cartulary.local_release_receipt.v1", producer);
  for (const [name, receipt] of consumers) {
    validateSchemaSync("cartulary.local_release_receipt.v1", receipt);
    for (const key of ["manifest_sha256", "archive_sha256", "archive", "operator_sha256"]) {
      if (receipt[key] !== producer[key]) throw new Error(`${name}: package identity mismatch`);
    }
    if (JSON.stringify(receipt.images) !== JSON.stringify(producer.images)) throw new Error(`${name}: package image identity mismatch`);
  }
  return { manifest_sha256: producer.manifest_sha256, archive_sha256: producer.archive_sha256 };
}
