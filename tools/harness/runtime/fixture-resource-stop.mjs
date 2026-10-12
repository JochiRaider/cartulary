// Physical stop adapters only. The native owner chooses dependency order and
// acknowledges durable recovery records after these operations succeed.
import path from "node:path";
import { spawnSync } from "../workspace/child-process.mjs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { stopOwnedProcess } from "./owned-process.mjs";
import { recoverVisualRenderer } from "../browser/index.mjs";
import { readLocalFile, atomicLocalFile, removePrivateFile } from "./secure-local-files.mjs";
import { randomUUID } from "node:crypto";

export async function stopFixtureResource(kind,target,environment=process.env) {
  if (kind === "process") { await stopOwnedProcess(target); return; }
  if (kind === "renderer") { recoverVisualRenderer(target); return; }
  if (kind !== "proxy" || typeof target !== "string") throw new Error("unknown resource operation");
  const runtimeRoot=environment.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT;
  const relative=path.relative(path.join(runtimeRoot,"object-store-proxy"),target);
  if (relative.startsWith("..") || path.isAbsolute(relative) || relative.split(path.sep).length !== 2) throw new Error("foreign proxy proof");
  const bytes=readLocalFile(target,{maximum:1048576}),proof=parseStrictJSON(bytes.toString());
  if (!["cartulary.local_object_store_proxy_lease.v2","cartulary.local_object_store_proxy_start_attempt.v2"].includes(proof.schema_id)) throw new Error("unknown proxy proof");
  validateSchemaSync(proof.schema_id,proof);
  const config=proof.configuration,snapshot=`${target}.cleanup-${randomUUID()}`;
  atomicLocalFile(snapshot,bytes);
  try {
    const result=spawnSync(path.join(runtimeRoot,"object-store-proxy","s3corsproxy"),["stop","--state-file",snapshot,
      "--listen",config.listen,"--upstream",config.upstream_origin,"--origin",config.allowed_origin,
      "--tls-certificate",config.tls.certificate_path,"--tls-private-key",config.tls.private_key_path,"--tls-root-certificate",config.tls.root_certificate_path],
    {env:environment,stdio:"ignore",timeout:30000,killSignal:"SIGKILL"});
    if (result.error || result.status!==0) throw new Error("proxy stop failed");
  } finally { removePrivateFile(snapshot); }
}
