import path from "node:path";
import { createHash } from "node:crypto";
import { readdirSync, readFileSync, readlinkSync } from "node:fs";
import { parseStrictJSON } from "../contract/strict-json.mjs";
import { atomicLocalFile, openDirectory, readLocalFile } from "./secure-local-files.mjs";
import { closeSync } from "node:fs";
import { readBootClock } from "./boot-clock.mjs";

export const invocationRootKey = "CARTULARY_HARNESS_INVOCATION_ROOT";
export const invocationDigest = (value) => `sha256:${createHash("sha256").update(value).digest("hex")}`;
const privateJSON = (file) => parseStrictJSON(readLocalFile(file, { maximum: 65536 }).toString("utf8"));
const exactKeys = (value, keys) => value && Object.keys(value).sort().join() === [...keys].sort().join();
export function readWrapperOwner(root) {
  const descriptor = openDirectory(root, { privateLeaf: true }); closeSync(descriptor);
  const owner = privateJSON(path.join(root, "owner.json"));
  if (!exactKeys(owner, ["version", "invocation_id", "pid", "start_ticks", "start_ms", "boot_id", "namespace"]) || owner.version !== 1 ||
      !/^[a-f0-9-]{36}$/u.test(owner.invocation_id) || !Number.isSafeInteger(owner.pid) || owner.pid < 1 ||
      !/^[0-9]+$/u.test(owner.start_ticks) || !Number.isSafeInteger(owner.start_ms) || owner.start_ms < 0 ||
      !/^[a-f0-9-]{36}$/u.test(owner.boot_id) || !/^time:\[[0-9]+\]$/u.test(owner.namespace)) throw Error("invalid wrapper owner");
  const stat = readFileSync(`/proc/${owner.pid}/stat`, "utf8").split(") ").at(-1).trim().split(/\s+/u);
  if (["Z", "X"].includes(stat[0]) || stat[19] !== owner.start_ticks ||
      readFileSync("/proc/sys/kernel/random/boot_id", "utf8").trim() !== owner.boot_id || readlinkSync(`/proc/${owner.pid}/ns/time`) !== owner.namespace) throw Error("stale wrapper owner");
  return owner;
}
export function readWrapperGraphs(root, owner) {
  const names = readdirSync(root);
  if (names.length > 66 || names.some((name) => !["owner.json", "completion.json"].includes(name) && !/^graph-[a-f0-9]{64}\.json$/u.test(name))) throw Error("invalid wrapper membership roster");
  const graphs = names.filter((name) => name.startsWith("graph-")).sort().map((name) => {
    const value = privateJSON(path.join(root, name));
    if (!exactKeys(value, ["version", "invocation_digest", "run_root", "run_id", "source_digest", "graph_digest", "phase", "scan_start_ms", "scan_end_ms", "scan_outcome"]) || value.version !== 1 ||
        value.invocation_digest !== invocationDigest(owner.invocation_id) || !path.isAbsolute(value.run_root) || path.resolve(value.run_root) !== value.run_root || path.basename(value.run_root) !== value.run_id ||
        !/^[A-Za-z0-9_.-]{1,96}$/u.test(value.run_id) || !/^sha256:[a-f0-9]{64}$/u.test(value.source_digest) || !/^sha256:[a-f0-9]{64}$/u.test(value.graph_digest) ||
        name !== `graph-${invocationDigest(value.run_root).slice(7)}.json` || !["registered", "scanning", "scanned", "complete"].includes(value.phase) ||
        ![null, "passed", "failed"].includes(value.scan_outcome) || [value.scan_start_ms, value.scan_end_ms].some((ms) => ms !== null && (!Number.isSafeInteger(ms) || ms < owner.start_ms))) throw Error("invalid wrapper graph binding");
    return value;
  });
  if (new Set(graphs.map((graph) => graph.run_id)).size !== graphs.length) throw Error("duplicate invocation run identity");
  return graphs;
}

export function registerWrapperGraph({ runRoot, manifest, environment = process.env }) {
  const unavailable = { digest: null, scanStarted() {}, scanFinished() {}, complete() {} };
  if (manifest?.instrumentation?.mode !== "basic" || !environment[invocationRootKey]) return unavailable;
  try {
    const root = environment[invocationRootKey], owner = readWrapperOwner(root);
    if (readWrapperGraphs(root, owner).length >= 64) return unavailable;
    const value = { version: 1, invocation_digest: invocationDigest(owner.invocation_id), run_root: path.resolve(runRoot), run_id: manifest.run_id,
      source_digest: manifest.source_digest, graph_digest: manifest.graph_digest, phase: "registered", scan_start_ms: null, scan_end_ms: null, scan_outcome: null };
    const file = path.join(root, `graph-${invocationDigest(value.run_root).slice(7)}.json`);
    atomicLocalFile(file, JSON.stringify(value));
    const update = (fields) => {
      try {
        if (value.phase === "complete") return;
        const current = readWrapperOwner(root);
        if (current.invocation_id !== owner.invocation_id) return;
        Object.assign(value, fields); atomicLocalFile(file, JSON.stringify(value), { replace: true, maximumReplacementBytes: 65536 });
      } catch { /* Loss of private wrapper coordination never changes the primary outcome. */ }
    };
    const now = () => {
      const clock = readBootClock();
      if (clock.clock_identity !== invocationDigest(`${owner.boot_id}:${owner.namespace}`)) return null;
      return clock.monotonic_ms;
    };
    return { digest: value.invocation_digest,
      scanStarted() { if (value.phase === "registered") update({ phase: "scanning", scan_start_ms: now() }); },
      scanFinished(passed) { if (value.phase === "scanning") update({ phase: "scanned", scan_end_ms: now(), scan_outcome: passed ? "passed" : "failed" }); },
      complete() { update({ phase: "complete" }); },
    };
  } catch { return unavailable; }
}
