import { createHash, X509Certificate } from "node:crypto";
import { chmodSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile } from "../runtime/secure-local-files.mjs";

const hash = (bytes) => createHash("sha256").update(bytes).digest("hex");

export function browserTrustTool(root) {
  const pin = JSON.parse(readFileSync(path.join(root, "tools/toolchain_pins.json"), "utf8")).browser_trust_tool;
  if (!pin || pin.package !== "libnss3-tools" || pin.platform !== "linux/amd64" ||
      !/^2:[0-9.]+-[A-Za-z0-9.]+$/u.test(pin.version) ||
      !/^https:\/\/security\.ubuntu\.com\/ubuntu\/pool\/main\/n\/nss\/libnss3-tools_[A-Za-z0-9._-]+_amd64\.deb$/u.test(pin.url) ||
      !Number.isSafeInteger(pin.bytes) || pin.bytes < 1 || pin.bytes > 4 * 1024 * 1024 ||
      !/^[a-f0-9]{64}$/u.test(pin.sha256)) throw new Error("invalid browser trust tool pin");
  return { pin, file: path.join(root, workspaceLayout.playwright, "trust-tools", `${pin.sha256}.deb`) };
}

export function readBrowserTrustTool(root) {
  const tool = browserTrustTool(root);
  const bytes = readLocalFile(tool.file, { maximum: tool.pin.bytes });
  if (bytes.length !== tool.pin.bytes || hash(bytes) !== tool.pin.sha256) throw new Error("browser trust tool identity mismatch; run make playwright-install");
  return { ...tool, bytes };
}

export async function prepareBrowserTrustTool(root) {
  const tool = browserTrustTool(root);
  try { return readBrowserTrustTool(root); }
  catch (error) { if (error.code !== "ENOENT") throw error; }
  const response = await fetch(tool.pin.url, { redirect: "error", signal: AbortSignal.timeout(30000) });
  if (!response.ok) throw new Error("browser trust tool download failed");
  const chunks = []; let size = 0;
  for await (const chunk of response.body) {
    size += chunk.length;
    if (size > tool.pin.bytes) throw new Error("browser trust tool download exceeds pin");
    chunks.push(chunk);
  }
  const bytes = Buffer.concat(chunks);
  if (size !== tool.pin.bytes || hash(bytes) !== tool.pin.sha256) throw new Error("browser trust tool archive mismatch");
  privateDirectory(path.dirname(tool.file));
  try { atomicLocalFile(tool.file, bytes); }
  catch (error) { if (error.code !== "EEXIST") throw error; }
  return readBrowserTrustTool(root);
}

export function readBrowserTrustBundle(file) {
  const bytes = readLocalFile(file, { maximum: 256 * 1024 });
  const pem = bytes.toString("utf8");
  const certificates = pem.match(/-----BEGIN CERTIFICATE-----[\s\S]*?-----END CERTIFICATE-----/gu) ?? [];
  if (!certificates.length || certificates.length > 8 || pem.replace(/-----BEGIN CERTIFICATE-----[\s\S]*?-----END CERTIFICATE-----/gu, "").trim()) throw new Error("invalid browser fixture trust bundle");
  const roots = certificates.map((certificate) => {
    const parsed = new X509Certificate(certificate);
    if (!parsed.ca || Date.now() < Date.parse(parsed.validFrom) || Date.now() > Date.parse(parsed.validTo)) throw new Error("invalid browser fixture authority");
    return { pem: `${certificate}\n`, sha256: hash(parsed.raw) };
  });
  if (new Set(roots.map((entry) => entry.sha256)).size !== roots.length) throw new Error("duplicate browser fixture authority");
  return { bytes, sha256: hash(bytes), roots };
}

export function admittedBrowserTrust(environment) {
  const bundle = readBrowserTrustBundle(environment.CARTULARY_WEB_E2E_TLS_ROOT_CERTIFICATE ?? "");
  if (`sha256:${bundle.sha256}` !== environment.CARTULARY_WEB_E2E_TLS_TRUST_SHA256) throw new Error("browser trust does not match the admitted stack");
  return bundle;
}

// Only public roots and the pinned tool enter the disposable container. The NSS
// database belongs to its unprivileged user; no host trust store is modified.
export function stageBrowserTrust({ root, environment, containerID, run }) {
  const bundle = admittedBrowserTrust(environment);
  const tool = readBrowserTrustTool(root);
  const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-browser-trust-"));
  try {
    // Copy the admitted bytes, never reopen a mutable cache entry for Docker.
    writeFileSync(path.join(scratch, "tools.deb"), tool.bytes, { mode: 0o644 });
    chmodSync(path.join(scratch, "tools.deb"), 0o644);
    run(["cp", path.join(scratch, "tools.deb"), `${containerID}:/home/pwuser/tools.deb`]);
    writeFileSync(path.join(scratch, "trust.pem"), bundle.bytes, { mode: 0o644 });
    chmodSync(path.join(scratch, "trust.pem"), 0o644);
    run(["cp", path.join(scratch, "trust.pem"), `${containerID}:/home/pwuser/trust.pem`]);
    const certutil = "/home/pwuser/trust-tools/usr/bin/certutil";
    const database = "sql:/home/pwuser/.pki/nssdb";

    for (const [index, certificate] of bundle.roots.entries()) {
      const name = `root-${index}.pem`;
      writeFileSync(path.join(scratch, name), certificate.pem, { mode: 0o644 });
      chmodSync(path.join(scratch, name), 0o644);
      run(["cp", path.join(scratch, name), `${containerID}:/home/pwuser/${name}`]);

    }
    return {
      attestation: { schema_id: "cartulary.browser_tls_trust.v1", tool_sha256: tool.pin.sha256, trust_bundle_sha256: bundle.sha256, authorities_sha256: bundle.roots.map((entry) => entry.sha256) },
      initialize() {
        run(["exec", containerID, "dpkg-deb", "--extract", "/home/pwuser/tools.deb", "/home/pwuser/trust-tools"]);
        run(["exec", containerID, "mkdir", "-p", "-m", "700", "/home/pwuser/.pki/nssdb"]);
        run(["exec", containerID, certutil, "-N", "--empty-password", "-d", database]);
        for (const [index, certificate] of bundle.roots.entries()) {
          run(["exec", containerID, certutil, "-A", "-d", database, "-n", certificate.sha256, "-t", "C,,", "-i", `/home/pwuser/root-${index}.pem`]);
        }
      },
    };
  } finally { rmSync(scratch, { recursive: true, force: true }); }
}

if (process.argv[1] && fileURLToPath(import.meta.url) === path.resolve(process.argv[1])) {
  if (process.argv[2] !== "prepare") throw new Error("unsupported browser trust command");
  await prepareBrowserTrustTool(path.resolve(import.meta.dirname, "../../.."));
}
