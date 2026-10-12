// Fixed physical-browser operations for the native fixture owner. Allocation,
// sharing, clone ordinals, dependency ordering and recovery decisions live in Go.
import path from "node:path";
import { fileURLToPath } from "node:url";
import { borrowRuntime } from "../runtime/session-client.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { acquireProcess } from "./acquire-process.mjs";
import { normalizeFailureRecord } from "../contract/index.mjs";
import { browserAcquisitionLaunchArguments, recordAcquisitionProcess, createBrowserAcquisition, createAcquisitionLaunch, closeAcquisitionLaunch,
  readBrowserAcquisition, settleBrowserAcquisition } from "./browser-acquisition.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
export async function fixtureLifecycle(operation, args, environment = process.env) {
  const runtime = borrowRuntime({ repoRoot: root,
    runRoot: path.join(environment.CARTULARY_TEST_RESULTS_DIR, environment.CARTULARY_TEST_RUN_ID), environment });
  const sessionRoot = environment.CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT;
  if (operation === "prepare" && args.length === 1) {
    const file = createBrowserAcquisition({ runtime, sessionRoot,
      leaseFile: path.join(sessionRoot, "stack.lease"), suiteLease: args[0] });
    return { file, launch_id: createAcquisitionLaunch(file, "producer") };
  }
  if (operation === "start" && args.length === 2) {
    const owner=readBrowserAcquisition(args[0],runtime),abort=new AbortController();
    const stop=()=>abort.abort(Object.assign(new Error("cancelled"),{failure_class:"interrupted",failure_reason:"cancelled_or_interrupted"}));
    process.on("SIGTERM",stop);process.on("SIGINT",stop);
    try { await acquireProcess(process.execPath,browserAcquisitionLaunchArguments(args[0],args[1],path.join(root,"tools/harness/browser/start-web-e2e.sh"),
      ["--session-start","--env-file",path.join(owner.session_root,"stack.env"),"--lease-file",owner.ready_lease]),
      {cwd:root,environment,signal:abort.signal,onChildProcess:(pid)=>{recordAcquisitionProcess(args[0],args[1],pid);return()=>{};},onReaped:()=>closeAcquisitionLaunch(args[0],args[1])}); }
    finally { process.removeListener("SIGTERM",stop);process.removeListener("SIGINT",stop); }
    return {};
  }
  if (["settle", "recover", "inspect"].includes(operation)) {
    if (args.length !== 1) throw new Error("invalid lifecycle arity");
    const owner = readBrowserAcquisition(args[0], runtime);
    if (operation === "inspect") return { suite_lease: owner.suite_lease };
    else await settleBrowserAcquisition({ file: args[0], runtime, root, environment, recovery: operation === "recover", recoveryResultsRoot: environment.CARTULARY_BROWSER_RECOVERY_RESULTS_DIR });
    return {};
  }
  if (operation === "validate" && args.length === 1) {
    const attached = parseStrictJSON(readLocalFile(args[0], { maximum: 1048576 }).toString());
    if (!attached || Array.isArray(attached) || Object.entries(attached).some(([key,value]) => !/^[A-Z][A-Z0-9_]*$/u.test(key) || typeof value !== "string")) throw new Error("invalid stack environment");
    const file = attached.CARTULARY_WEB_E2E_STACK_JSON_FILE;
    const suite = environment.CARTULARY_TEST_SUITE_ID, session = environment.CARTULARY_BROWSER_SESSION_GROUP;
    if (![suite,session].every((value) => typeof value === "string" && /^[A-Za-z0-9_.-]+$/u.test(value) && value !== "." && value !== "..")) throw new Error("invalid stack identity");
    const expected = path.join(environment.CARTULARY_TEST_RESULTS_DIR,environment.CARTULARY_TEST_RUN_ID,"_shared/test-services",suite,"browser-sessions",session,"stack-v8.json");
    if (file !== expected) throw new Error("foreign stack attachment");
    const stack = parseStrictJSON(readLocalFile(file, { maximum: 1048576 }).toString());
    validateSchemaSync("cartulary.web_e2e_stack.v8", stack);
    const admission = parseStrictJSON(readLocalFile(path.join(path.dirname(file), "service-admission.json"), { maximum: 1048576 }).toString());
    validateSchemaSync("cartulary.test_services.browser_admission.v1", admission);
    if ([stack,admission].some((value) => value.suite_id !== suite || value.browser_session_id !== session)) throw new Error("stack identity mismatch");
    return attached;
  }
  throw new Error("unknown fixture lifecycle operation");
}
if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try { process.stdout.write(`${JSON.stringify(await fixtureLifecycle(process.argv[2], process.argv.slice(3)))}\n`); }
  catch(error) {
    const failure=normalizeFailureRecord({failure_class:error.failure_class??"harness",failure_reason:error.failure_reason??"fixture_error"});
    process.stdout.write(`${JSON.stringify({failure_class:failure.failure_class,failure_reason:failure.failure_reason})}\n`);
    process.stderr.write("fixture lifecycle rejected or failed\n");process.exitCode=failure.failure_class==="artifact"?11:12;
  }
}
