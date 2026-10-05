// Preserve worktree admission across managed child lifetimes, including detach.
import * as childProcess from "node:child_process";
import { fstatSync, lstatSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

function inheritedAdmission() {
  try {
    if (!/^lock:.*FLOCK\s+ADVISORY\s+READ\s/mu.test(readFileSync("/proc/self/fdinfo/9", "utf8"))) return false;
    const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
    const git = childProcess.execFileSync("git", ["-C", root, "rev-parse", "--absolute-git-dir"], { encoding: "utf8", stdio: ["ignore", "pipe", "ignore"] }).trim();
    const lock = lstatSync(path.join(git, "cartulary-workspace.lock"));
    const descriptor = fstatSync(9);
    return lock.isFile() && lock.dev === descriptor.dev && lock.ino === descriptor.ino;
  } catch { return false; }
}

const admitted = inheritedAdmission();
function optionsWithAdmission(options = {}, ipc = false) {
  if (!admitted) return options;
  const stdio = Array.isArray(options.stdio) ? [...options.stdio] : Array(3).fill(options.stdio ?? (ipc && !options.silent ? "inherit" : "pipe"));
  if (ipc && !stdio.includes("ipc")) stdio.push("ipc");
  while (stdio.length < 10) stdio.push("ignore");
  if (stdio[9] !== "ignore" && stdio[9] !== 9 && stdio[9] != null) throw new Error("managed child descriptor 9 is reserved for workspace admission");
  stdio[9] = 9;
  return { ...options, stdio };
}
export function spawn(command, args = [], options = {}) {
  return childProcess.spawn(command, args, optionsWithAdmission(options));
}
export function spawnSync(command, args = [], options = {}) {
  return childProcess.spawnSync(command, args, optionsWithAdmission(options));
}
export function execFileSync(file, args = [], options = {}) {
  return childProcess.execFileSync(file, args, optionsWithAdmission(options));
}
export function fork(modulePath, args = [], options = {}) {
  return childProcess.fork(modulePath, args, optionsWithAdmission(options, true));
}
