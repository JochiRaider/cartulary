import { closeSync, mkdirSync } from "node:fs";
import path from "node:path";
import { borrowRuntime } from "../runtime/session-client.mjs";
import { atomicLocalFile, openDirectory, readLocalFile, removePrivateTree } from "../runtime/secure-local-files.mjs";
import { readBrowserAcquisition } from "./browser-acquisition.mjs";

const repository = path.resolve(import.meta.dirname, "../../..");
const rejected = (phase = "binding") => new Error(`browser private state requires its owned runtime binding (${phase})`);

// The native owner and acquisition record supply authority. Environment paths
// only select that exact binding; they cannot create a new state namespace.
function stateDirectory() {
  let runtime;
  let phase = "environment";
  try {
    const env = process.env;
    if (env.CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED !== "1" ||
        !env.CARTULARY_TEST_RESULTS_DIR || !env.CARTULARY_TEST_RUN_ID ||
        !env.CARTULARY_BROWSER_ACQUISITION_FILE) throw rejected();
    phase = "runtime";
    runtime = borrowRuntime({ repoRoot: repository,
      runRoot: path.resolve(repository, env.CARTULARY_TEST_RESULTS_DIR, env.CARTULARY_TEST_RUN_ID) });
    phase = "acquisition";
    const owner = readBrowserAcquisition(env.CARTULARY_BROWSER_ACQUISITION_FILE, runtime);
    phase = "binding";
    const directory = path.join(owner.session_root, "runtime-root", "playwright-state");
    if (env.CARTULARY_WEB_E2E_RUNTIME_ROOT !== path.dirname(directory) ||
        env.CARTULARY_PLAYWRIGHT_STATE_DIR !== directory ||
        env.CARTULARY_TEST_SUITE_ID !== owner.suite_id) throw rejected();
    const relative = path.relative(runtime.root, directory);
    if (!relative || relative.startsWith("..") || path.isAbsolute(relative)) throw rejected();
    phase = "permissions";
    let current = runtime.root;
    for (const part of ["", ...relative.split(path.sep)]) {
      current = path.join(current, part);
      closeSync(openDirectory(current, { privateLeaf: true }));
    }
    return directory;
  } catch {
    throw rejected(phase);
  } finally { runtime?.close(); }
}

function stateFile(name) {
  if (!/^[A-Za-z0-9][A-Za-z0-9_.-]*$/u.test(name)) throw rejected();
  return path.join(stateDirectory(), name);
}

export function readBrowserState(name) {
  const file = stateFile(name);
  try { return readLocalFile(file, { maximum: 1048576 }).toString("utf8"); }
  catch (error) {
    if (error.code === "ENOENT") return null;
    throw new Error("browser private state is unreadable");
  }
}

export function writeBrowserState(name, contents) {
  const file = stateFile(name);
  if (typeof contents !== "string") throw new Error("browser private state must be serialized before publication");
  try { atomicLocalFile(file, contents, { replace: true }); }
  catch { throw new Error("browser private state publication failed"); }
}

export async function withBrowserSetupLock(callback) {
  const lock = path.join(stateDirectory(), "global-setup.lock");
  const deadline = Date.now() + 120000;
  while (true) {
    try { mkdirSync(lock, { mode: 0o700 }); break; }
    catch (error) {
      if (error.code !== "EEXIST") throw new Error("browser setup lock acquisition failed");
      // A preexisting lock must be private and cannot redirect the waiter.
      try { closeSync(openDirectory(lock, { privateLeaf: true })); }
      catch (observed) {
        // Its owner may have retired the lock between mkdir and inspection.
        if (observed.code !== "ENOENT") throw new Error("browser setup lock is unsafe");
      }
      if (Date.now() >= deadline) throw new Error("browser setup lock timed out");
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
  }
  let primary;
  let failed = false;
  let result;
  try { result = await callback(); }
  catch (error) { primary = error; failed = true; }
  try { removePrivateTree(lock); }
  catch {
    const cleanup = new Error("browser setup lock cleanup failed");
    if (failed) throw new AggregateError([primary, cleanup], "browser setup and cleanup failed");
    throw cleanup;
  }
  if (failed) throw primary;
  return result;
}
