import { createHash } from "node:crypto";
import { createReadStream } from "node:fs";
import { lstat, mkdir, readFile, readdir, realpath, rename, rm, stat, utimes, writeFile } from "node:fs/promises";
import path from "node:path";

// The renderer installation store (Testing Harness build-artifact profiles,
// G37, P5-3c). Installations are immutable content-addressed roots; the
// receipt is the only mutable pointer, and keys/<input key>.json records
// which installation an input key produced and the seal of its root.

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

// seal fingerprints every entry under a sealed root by its metadata: path,
// mode, size, inode, and modification and change times. Writing, replacing,
// adding, removing or changing the mode of any entry changes it, and nothing
// can set a change time back, so a root that keeps its seal still holds the
// bytes it was sealed with. Reuse checks the seal instead of hashing every
// file of the installation again.
export async function seal(directory) {
  const entries = [];
  const walk = async (current, prefix) => {
    const children = await readdir(current, { withFileTypes: true });
    await Promise.all(children.map(async (entry) => {
      const relative = prefix ? `${prefix}/${entry.name}` : entry.name;
      const absolute = path.join(current, entry.name);
      const info = await lstat(absolute, { bigint: true });
      entries.push([relative, info.mode.toString(8), info.size.toString(), info.ino.toString(), info.mtimeNs.toString(), info.ctimeNs.toString()]);
      if (entry.isDirectory()) await walk(absolute, relative);
    }));
  };
  await walk(directory, "");
  entries.sort((a, b) => Buffer.compare(Buffer.from(a[0]), Buffer.from(b[0])));
  return sha(JSON.stringify(entries));
}

const hostLibraryDirectories = ["/lib/x86_64-linux-gnu", "/usr/lib/x86_64-linux-gnu", "/lib64"];
const digestPattern = /^[a-f0-9]{64}$/u;

// reusableInstallation returns the installation recorded for key when its
// manifest still has its digest, its root still has the seal recorded when
// it was sealed, and every host library it copied still has the copied
// digest; otherwise null. Anything unreadable is not reusable. A reused
// record is touched, so collection keeps what is in use.
export async function reusableInstallation(output, key, { libraryDirectories = hostLibraryDirectories, loader = "/lib64/ld-linux-x86-64.so.2" } = {}) {
  const recordFile = path.join(output, "keys", `${key}.json`);
  let record;
  try { record = JSON.parse(await readFile(recordFile, "utf8")); }
  catch { return null; }
  const identity = record?.manifest_sha256;
  if (!digestPattern.test(identity ?? "") || !digestPattern.test(record.root_seal ?? "")) return null;
  let manifest;
  try { manifest = JSON.parse(await readFile(path.join(output, `${identity}.json`), "utf8")); }
  catch { return null; }
  if (sha(JSON.stringify(manifest)) !== identity || !Array.isArray(manifest.files)) return null;
  try {
    if (await seal(path.join(output, `${identity}.root`)) !== record.root_seal) return null;
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
  const now = new Date();
  await utimes(recordFile, now, now).catch(() => {});
  return identity;
}

// publishReceipt points the receipt at identity, atomically through staging.
export async function publishReceipt(output, identity, staging) {
  const receipt = path.join(staging, "installation.json");
  await writeFile(receipt, JSON.stringify({ schema_id: "cartulary.reporting_runtime_receipt.v1", manifest_sha256: identity, root: `${identity}.root`, manifest: `${identity}.json` }) + "\n");
  await rename(receipt, path.join(output, "installation.json"));
}

// recordKey records that key produced identity, whose root has rootSeal,
// atomically through staging.
export async function recordKey(output, key, identity, rootSeal, staging) {
  await mkdir(path.join(output, "keys"), { recursive: true });
  const record = path.join(staging, "key.json");
  await writeFile(record, JSON.stringify({ input_key: key, manifest_sha256: identity, root_seal: rootSeal }) + "\n");
  await rename(record, path.join(output, "keys", `${key}.json`));
}

const day = 24 * 60 * 60 * 1000;

// collect removes key records unused for 30 days, then installations that
// neither the receipt nor a key record names once an hour has passed since
// they were written, so a process still running an older installation keeps
// it for that hour, and set-aside altered roots after 30 days. It returns the
// names it removed.
export async function collect(output, now = Date.now()) {
  const removed = [];
  const age = async (name) => now - (await stat(path.join(output, name))).mtimeMs;
  const named = new Set();
  try { named.add(JSON.parse(await readFile(path.join(output, "installation.json"), "utf8")).manifest_sha256); }
  catch { /* no receipt yet */ }
  let records = [];
  try { records = await readdir(path.join(output, "keys")); }
  catch (error) { if (error.code !== "ENOENT") throw error; }
  for (const name of records) {
    const file = path.join("keys", name);
    if (await age(file) > 30 * day) {
      await rm(path.join(output, file), { force: true });
      removed.push(file);
      continue;
    }
    try { named.add(JSON.parse(await readFile(path.join(output, file), "utf8")).manifest_sha256); }
    catch { /* an unreadable record names nothing */ }
  }
  for (const name of await readdir(output)) {
    const match = /^([a-f0-9]{64})\.json$/u.exec(name);
    const altered = name.startsWith(".altered-");
    if ((match && !named.has(match[1]) && await age(name) > 60 * 60 * 1000) || (altered && await age(name) > 30 * day)) {
      const root = match ? `${match[1]}.root` : name;
      await rm(path.join(output, root), { recursive: true, force: true });
      if (match) await rm(path.join(output, name), { force: true });
      removed.push(root);
    }
  }
  return removed;
}
