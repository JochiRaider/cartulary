import path from "node:path";
import { parseStrictJSON, semanticJSONDigest, validateSchemaSync } from "../../contract/index.mjs";

import { atomicLocalFile, readLocalFile } from "../../runtime/secure-local-files.mjs";

// Coordinator checkpoints and the final row adapter use the same canonical
// publication boundary. The native finalizer preserves already settled rows.
export function publishRowResult(runRoot, result, { preserveEquivalent = false } = {}) {
  validateSchemaSync("cartulary.harness_row_result.v6", result);
  const file = path.join(runRoot, "rows", `${result.row_id}.json`);
  if (preserveEquivalent) {
    try {
      const previous = parseStrictJSON(readLocalFile(file, { allowReplacement: true }).toString("utf8"));
      validateSchemaSync("cartulary.harness_row_result.v6", previous);
      const semantic = ({ started_at, finished_at, wall_duration_ms, ...value }) => semanticJSONDigest(value);
      if (semantic(previous) === semantic(result)) return;
    } catch (error) { if (error.code !== "ENOENT") throw error; }
  }
  atomicLocalFile(file, `${JSON.stringify(result, null, 2)}\n`, { replace: true });
}
