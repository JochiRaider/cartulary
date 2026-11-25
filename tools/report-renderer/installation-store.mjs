import { createHash } from "node:crypto";
import { createReadStream } from "node:fs";
import { lstat, mkdir, readFile, readdir, realpath, rename, writeFile } from "node:fs/promises";
import path from "node:path";

// The renderer installation store (Testing Harness build-artifact profiles,
// G37). Installations are immutable content-addressed roots; the receipt is
// the only mutable pointer, and keys/<input key>.json records which
// installation an input key produced.

export const sha = (bytes) => createHash("sha256").update(bytes).digest("hex");

export async function fileDigest(file) {
  const digest = createHash("sha256");
  for await (const bytes of createReadStream(file)) digest.update(bytes);
  return digest.digest("hex");
}

// inventory lists every regular file under directory with its digest, in
// byte order; a link, special entry, hard link or unsafe name is an error.
export async function inventory(directory, prefix = "") {
  const result = [];
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const relative = prefix ? `${prefix}/${entry.name}` : entry.name;
    if (entry.name !== entry.name.normalize("NFC") || /[\x00-\x1f\x7f\\]/u.test(entry.name)) throw new Error("invalid runtime inventory path");
    if (entry.isDirectory()) result.push(...await inventory(path.join(directory, entry.name), relative));
    else if (entry.isFile()) {
      const absolute = path.join(directory, entry.name);
      const info = await lstat(absolute);
      if (info.nlink !== 1) throw new Error("runtime inventory hard link");
      result.push({ path: relative, sha256: await fileDigest(absolute) });
    } else throw new Error("invalid runtime inventory entry");
  }
  return result.sort((a, b) => Buffer.compare(Buffer.from(a.path), Buffer.from(b.path)));
}

const hostLibraryDirectories = ["/lib/x86_64-linux-gnu", "/usr/lib/x86_64-linux-gnu", "/lib64"];

// reusableInstallation returns the installation recorded for key when its
// manifest still has its digest, its root still holds exactly the manifest's
// files, and every host library it copied still has the copied digest;
// otherwise null. Anything unreadable is not reusable.
export async function reusableInstallation(output, key, { libraryDirectories = hostLibraryDirectories, loader = "/lib64/ld-linux-x86-64.so.2" } = {}) {
  let identity;
  try { identity = JSON.parse(await readFile(path.join(output, "keys", `${key}.json`), "utf8")).manifest_sha256; }
  catch { return null; }
  if (!/^[a-f0-9]{64}$/u.test(identity ?? "")) return null;
  let manifest;
  try { manifest = JSON.parse(await readFile(path.join(output, `${identity}.json`), "utf8")); }
  catch { return null; }
  if (sha(JSON.stringify(manifest)) !== identity || !Array.isArray(manifest.files)) return null;
  try {
    if (JSON.stringify(await inventory(path.join(output, `${identity}.root`))) !== JSON.stringify(manifest.files)) return null;
  } catch { return null; }
  for (const entry of manifest.files) {
    let source = null;
    if (entry.path.startsWith("lib/x86_64-linux-gnu/")) {
      for (const directory of libraryDirectories) {
        try { source = await realpath(path.join(directory, path.basename(entry.path))); break; }
        catch (error) { if (error.code !== "ENOENT") throw error; }
      }
      if (!source) return null;
    } else if (entry.path === loader.slice(1)) source = await realpath(loader);
    if (source && await fileDigest(source) !== entry.sha256) return null;
  }
  return identity;
}

// publishReceipt points the receipt at identity, atomically through staging.
export async function publishReceipt(output, identity, staging) {
  const receipt = path.join(staging, "installation.json");
  await writeFile(receipt, JSON.stringify({ schema_id: "cartulary.reporting_runtime_receipt.v1", manifest_sha256: identity, root: `${identity}.root`, manifest: `${identity}.json` }) + "\n");
  await rename(receipt, path.join(output, "installation.json"));
}

// recordKey records that key produced identity, atomically through staging.
export async function recordKey(output, key, identity, staging) {
  await mkdir(path.join(output, "keys"), { recursive: true });
  const record = path.join(staging, "key.json");
  await writeFile(record, JSON.stringify({ input_key: key, manifest_sha256: identity }) + "\n");
  await rename(record, path.join(output, "keys", `${key}.json`));
}
