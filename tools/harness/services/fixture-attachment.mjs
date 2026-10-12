// Existing local-service lifecycle owner, exposed to one verified native borrower.
import path from "node:path";
import { fileURLToPath } from "node:url";
import { attachLocalSession, detachLocalBorrower, resolveServiceSessionMode } from "./local-session.mjs";
import { borrowRuntime } from "../runtime/session-client.mjs";
import { atomicLocalFile, readLocalFile, removePrivateFile } from "../runtime/secure-local-files.mjs";
import { parseStrictJSON } from "../contract/index.mjs";

const root = path.resolve(import.meta.dirname,"../../..");
export function fixtureAttachment(operation, environment = process.env) {
  const runID = environment.CARTULARY_TEST_RUN_ID;
  const runtime = borrowRuntime({repoRoot:root,runRoot:path.join(environment.CARTULARY_TEST_RESULTS_DIR,runID),environment});
  const file = runtime.recoveryPath("borrowed-suite.json");
  if (operation === "attach") {
    const target = environment.CARTULARY_TEST_TARGET;
    const posture = resolveServiceSessionMode({target,environment});
    if (posture.mode !== "attach") throw new Error("attachment mode required");
    let proof;
    const owner = parseStrictJSON(environment.CARTULARY_NATIVE_FIXTURE_OWNER);
    const suite = attachLocalSession({root,binary:environment.CARTULARY_TEST_SERVICES_BIN || environment.TEST_SERVICES_BIN,
      sessionFile:posture.sessionFile,target,runID,suiteRuntime:runtime,borrowerOwner:owner,onContainerProof:(value) => {proof=value;} });
    try {
      atomicLocalFile(file,`${JSON.stringify({run_id:runID,runtime_lease_id:runtime.leaseID,session_file:posture.sessionFile,borrower:suite.borrower_proof})}\n`);
      return {environment:suite.environment,container_proof:proof};
    } catch (error) { suite.close(); throw error; }
  }
  if (operation === "detach") {
    const value = parseStrictJSON(readLocalFile(file,{maximum:1048576}).toString());
    if (Object.keys(value).sort().join(",") !== "borrower,run_id,runtime_lease_id,session_file" || value.run_id !== runID || value.runtime_lease_id !== runtime.leaseID || value.borrower.run_id !== runID) throw new Error("foreign borrower binding");
    detachLocalBorrower({sessionFile:value.session_file,borrower:value.borrower});
    removePrivateFile(file); return {};
  }
  throw new Error("unknown attachment operation");
}
if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try { if (process.argv.length !== 3) throw new Error("invalid arity"); process.stdout.write(`${JSON.stringify(fixtureAttachment(process.argv[2]))}\n`); }
  catch { process.stderr.write("fixture attachment failed\n"); process.exitCode = 11; }
}
