import { execFileSync } from "../../workspace/child-process.mjs";
import { accessSync, constants, statSync } from "node:fs";
import path from "node:path";

function nodeRuntimeError(message) {
  return Object.assign(new Error(message), {
    failure_class: "config",
    failure_reason: "configuration_error",
  });
}

export function resolveGraphNodeBinary({ cwd, nodeBin }) {
  if (typeof nodeBin !== "string" || nodeBin.trim() === "") {
    throw nodeRuntimeError("selected NODE_BIN is missing");
  }
  const resolved = path.resolve(cwd, nodeBin);
  try {
    if (!statSync(resolved).isFile()) {
      throw new Error("not a regular file");
    }
    accessSync(resolved, constants.X_OK);
  } catch {
    throw nodeRuntimeError("selected NODE_BIN is not an executable file");
  }
  return resolved;
}

export function assertGraphNodeLaunch(nodeBinary) {
  try {
    execFileSync(nodeBinary, ["--version"], { timeout: 5000, stdio: "ignore" });
  } catch {
    throw nodeRuntimeError("selected NODE_BIN could not be executed");
  }
}
