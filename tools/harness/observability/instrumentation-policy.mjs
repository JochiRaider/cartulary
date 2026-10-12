import { readFileSync } from "node:fs";
import { semanticJSONDigest, validateSchemaSync } from "../contract/index.mjs";

export function instrumentationPolicy() {
  const policy = JSON.parse(readFileSync(new URL("../../harness_instrumentation_policy.json", import.meta.url)));
  validateSchemaSync(policy.schema_id, policy);
  return { policy, digest: semanticJSONDigest(policy) };
}
