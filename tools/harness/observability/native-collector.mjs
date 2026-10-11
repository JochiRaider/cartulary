import { EventEmitter } from "node:events";
import path from "node:path";
import { spawn } from "../workspace/child-process.mjs";
import { ownedProcess, stopOwnedProcess } from "../runtime/owned-process.mjs";
import { parseStrictJSON } from "../contract/index.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";

// The graph owns and reaps this helper. There is no fallback collector or
// independent polling loop in the bridge; all controls go to one native owner.
export function nativeCollector(_url, { workerData }) {
  const events = new EventEmitter();
  const executable = path.resolve(import.meta.dirname, "../../..", workspaceLayout.toolbin, "cartulary-harness-collect");
  const child = spawn(executable, [], { stdio: ["pipe", "pipe", "ignore"] });
  let proof, tail = "", exited = false;
  const completion = new Promise((resolve) => child.once("close", (code) => { exited = true; events.emit("exit", code); resolve(); }));
  child.once("error", (error) => events.emit("error", error));
  child.stdin.on("error", (error) => { if (!exited) events.emit("error", error); });
  child.stdout.setEncoding("utf8");
  child.stdout.on("data", (data) => {
    tail += data;
    if (Buffer.byteLength(tail) > 1048576) { events.emit("error", Error("collector message limit")); child.stdout.destroy(); return; }
    let end;
    while ((end = tail.indexOf("\n")) !== -1) {
      const line = tail.slice(0, end); tail = tail.slice(end + 1);
      try { events.emit("message", parseStrictJSON(line)); }
      catch (error) { events.emit("error", error); }
    }
  });
  try { proof = ownedProcess(child.pid); } catch { child.kill("SIGKILL"); }
  events.postMessage = (message) => { if (exited) throw Error("collector has exited"); child.stdin.write(`${JSON.stringify(message)}\n`); };
  events.terminate = async () => {
    child.stdin.end();
    if (!exited && proof) await stopOwnedProcess(proof, { graceMS: 0, killMS: workerData.policy.stop_ms, exitNotification: completion });
    else if (!exited) child.kill("SIGKILL");
    await completion;
  };
  events.postMessage(workerData);
  return events;
}
