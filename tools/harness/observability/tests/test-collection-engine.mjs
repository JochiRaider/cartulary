import assert from "node:assert/strict";
import test from "node:test";
const settle = async () => { for (let i=0;i<30;i++) await Promise.resolve(); };
import { spawn } from "node:child_process";
import { once } from "node:events";
import { ownedProcess, stopOwnedProcess } from "../../runtime/owned-process.mjs";

test("exit notifications wake cleanup without replacing live identity proof", async (t) => {
  const child = spawn(process.execPath, ["-e", 'process.on("SIGTERM",()=>process.stdout.write("stopping\\n"));process.stdin.on("data",()=>process.exit(0));process.stdout.write("ready\\n");setInterval(()=>{},1000)'], { stdio: ["pipe", "pipe", "ignore"] });
  t.after(() => { if (child.exitCode === null) child.kill("SIGKILL"); });
  const closed = once(child, "close");
  await once(child.stdout, "data");
  const proof = ownedProcess(child.pid);
  let settled = false;
  const stopping = once(child.stdout, "data");
  const completion = stopOwnedProcess(proof, { graceMS: 1000, exitNotification: Promise.resolve() });
  completion.then(() => { settled = true; });
  await stopping;
  await settle();
  assert.equal(settled, false, "an early event must not establish quiescence");
  child.stdin.end("finish");
  await completion; await closed;
  assert.equal(child.exitCode, 0);
  await stopOwnedProcess(proof, { exitNotification: closed });
});
