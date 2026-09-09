import { spawn } from "node:child_process";
import { resolve } from "node:path";

// Exercise a genuinely separate public Make invocation, with no inherited
// execution identity, lease, selection, or jobserver descriptors.
export async function publishIndependentFrontendBuild(measurement: boolean) {
  const env = Object.fromEntries(
    Object.entries(process.env).filter(
      ([key]) =>
        !key.startsWith("CARTULARY_") &&
        !["MAKEFLAGS", "MFLAGS", "MAKELEVEL"].includes(key),
    ),
  );
  return await new Promise<string>((resolveResult, reject) => {
    const child = spawn(
      "make",
      [measurement ? "build-web-measurement" : "build-web"],
      {
        cwd: resolve(import.meta.dirname, "../../../../.."),
        env,
        stdio: ["ignore", "pipe", "pipe"],
      },
    );
    let output = "";
    child.stdout.on("data", (data) => {
      output += data.toString();
    });
    child.stderr.on("data", (data) => {
      output += data.toString();
    });
    child.once("error", reject);
    child.once("close", (code) => {
      if (code !== 0)
        reject(
          new Error(
            `Independent frontend Make build failed with exit ${code}: ${output.trim()}`,
          ),
        );
      else
        resolveResult(
          output.match(/run_root=(\S+)/u)?.[1] ?? "standalone build completed",
        );
    });
  });
}
