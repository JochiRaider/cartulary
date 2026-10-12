import path from "node:path";
import os from "node:os";
import { readLocalFile } from "../../runtime/secure-local-files.mjs";
import { parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
export function readRegistry(locatorFile) {
  const locator = parseStrictJSON(readLocalFile(locatorFile).toString());
  validateSchemaSync("cartulary.ui_review_session.v1", locator);
  const file = path.join(os.tmpdir(), `cartulary-ui-review-${process.getuid()}`, `${locator.session_id.slice(9)}.json`);
  const record = parseStrictJSON(readLocalFile(file).toString());
  validateSchemaSync("cartulary.harness_review_registry.v1", record);
  return record;
}
