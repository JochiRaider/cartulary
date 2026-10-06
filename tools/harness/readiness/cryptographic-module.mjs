import { createHash } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";

export const sha256 = (bytes) => createHash("sha256").update(bytes).digest("hex");
export function pins(root = process.cwd()) {
  return JSON.parse(readFileSync(path.join(root, "tools/toolchain_pins.json"), "utf8"));
}
export function goOutput(go, args, env = process.env) {
  const result = spawnSync(go, args, { env, encoding: "utf8", maxBuffer: 16 * 1024 * 1024 });
  if (result.error || result.status !== 0) throw new Error("cryptographic module toolchain command failed");
  return result.stdout;
}
export function verifyArchive(go, env = process.env) {
  const policy = pins();
  const identity = JSON.parse(goOutput(go, ["env", "-json", "GOROOT", "GOVERSION", "GOFIPS140"], env));
  if (identity.GOVERSION !== policy.go_toolchain || identity.GOFIPS140 !== policy.cryptographic_module.selector) {
    throw new Error("cryptographic module toolchain identity rejected");
  }
  const digest = sha256(readFileSync(path.join(identity.GOROOT, "lib/fips140", `${policy.cryptographic_module.selector}.zip`)));
  if (digest !== policy.cryptographic_module.archive_sha256) throw new Error("cryptographic module archive digest rejected");
  return { toolchain: identity.GOVERSION, ...policy.cryptographic_module };
}
export function buildReceipt(go, binary, env = process.env) {
  const module = verifyArchive(go, env);
  const info = JSON.parse(goOutput(go, ["version", "-m", "-json", binary], env));
  const selectors = info.Settings.filter((s) => s.Key === "GOFIPS140");
  if (info.GoVersion !== module.toolchain || selectors.length !== 1 || selectors[0].Value !== module.selector) {
    throw new Error("cryptographic executable metadata rejected");
  }
  const receipt = { schema_id: "cartulary.cryptographic_build_receipt.v1", ...module, binary_sha256: sha256(readFileSync(binary)), build_info: info };
  writeFileSync(`${binary}.crypto.json`, `${JSON.stringify(receipt, null, 2)}\n`, { mode: 0o600 });
  return receipt;
}

