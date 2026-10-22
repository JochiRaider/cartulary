import path from "node:path";
import { readFileSync } from "node:fs";
import { spawnSync } from "../workspace/child-process.mjs";

// Placement is resolved by the same native proof reader as Go service owners.
// Paths are transient capability data and must never be retained as diagnostics.
function account(root, args) {
  const result = spawnSync(path.join(root, ".cache/cartulary/toolbin/cartulary-harness"), ["account", ...args], { encoding: "utf8", maxBuffer: 4096, timeout: 5000 });
  if (result.status !== 0) throw new Error("invocation accounting placement unavailable");
  return String(result.stdout ?? "").trim();
}

export function invocationContainerParent(root) {
  const membership = readFileSync("/proc/self/cgroup", "utf8");
  if (!membership.includes("/cartulary-account-")) return "";
  const parent = account(root, ["container-parent"]);
  if (!parent.startsWith("/") || parent.includes("\n")) throw new Error("invalid invocation accounting parent");
  return parent;
}

export function verifyInvocationContainer(root, parent, id) {
  if (parent) account(root, ["verify-container", id]);
}
