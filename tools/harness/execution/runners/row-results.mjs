import path from "node:path";
import { validateSchemaSync } from "../../contract/index.mjs";

import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";

// Coordinator checkpoints and the final row adapter use the same canonical
// publication boundary. The native finalizer preserves already settled rows.
export function publishRowResult(runRoot, result) {
  validateSchemaSync("cartulary.harness_row_result.v6", result);
  atomicLocalFile(path.join(runRoot, "rows", `${result.row_id}.json`), `${JSON.stringify(result, null, 2)}\n`, { replace: true });
}
