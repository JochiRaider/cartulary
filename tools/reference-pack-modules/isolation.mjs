import { spawnSync } from "node:child_process";
import { lstatSync, realpathSync, readFileSync, mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import { createHash } from "node:crypto";
import { fileURLToPath } from "node:url";

const pins = JSON.parse(readFileSync(new URL("../toolchain_pins.json", import.meta.url)));
export function command(program, args, options = {}) {
  const { expectedExit = 0, includeStderr = false, ...spawnOptions } = options;
  const result = spawnSync(program, args, { encoding: "utf8", maxBuffer: 256 * 1024 * 1024, ...spawnOptions });
  if (result.error || result.signal || result.status !== expectedExit) throw new Error(`qualification command failed (${result.status ?? result.signal}): ${result.stderr ?? result.error?.message}`);
  return result.stdout + (includeStderr ? result.stderr : "");
}

// Mount complete declared tool runtimes, never host /usr, /lib, /home, a module
// cache, or a checkout. Dynamic library discovery applies only to trusted tools.
export function executableInputs(executable) {
  const real = realpathSync(executable);
  const result = spawnSync("ldd", [real], { encoding: "utf8" });
  if (result.status !== 0 && !/not a dynamic executable|statically linked/u.test(`${result.stdout}${result.stderr}`)) throw new Error("cannot determine tool runtime closure");
  const libraries = [...`${result.stdout ?? ""}`.matchAll(/(?:=>\s*)?(\/[^\s()]+)/gu)].map(match => match[1]);
  return [...new Set([executable, real, ...libraries])].map(destination => ({ source: realpathSync(destination), destination }));
}

// The manifest is the filesystem capability boundary. Nothing is inherited
// from the caller's environment, working directory, HOME, network, or caches.
export function isolated(manifest, options = {}) {
  if (process.platform !== "linux") throw new Error("reference-pack qualification requires Linux filesystem isolation");
  const bwrap = "/usr/bin/bwrap";
  if (command(bwrap, ["--version"]).trim() !== `bubblewrap ${pins.bubblewrap_version}`) throw new Error("reference-pack qualification bubblewrap pin mismatch");
  if (!manifest.cwd || !Array.isArray(manifest.command) || !manifest.command.length) throw new Error("incomplete isolation input manifest");
  const args = ["--unshare-all", "--die-with-parent", "--new-session", "--cap-drop", "ALL", "--clearenv", "--proc", "/proc", "--dev", "/dev", "--tmpfs", "/tmp"];
  const destinations = new Set();
  for (const [writable, inputs] of [[false, manifest.readonly ?? []], [true, manifest.writable ?? []], [false, manifest.readonly_overlays ?? []]]) {
    for (const input of inputs) {
      const { source, destination = source } = typeof input === "string" ? { source: input } : input;
      if (!path.isAbsolute(source) || !path.isAbsolute(destination) || destination === "/" || destinations.has(destination)) throw new Error("invalid or repeated isolation mount");
      if (lstatSync(source).isSymbolicLink()) throw new Error("symlinked isolation mount source");
      destinations.add(destination);
      args.push(writable ? "--bind" : "--ro-bind", source, destination);
    }
  }
  for (const [name, value] of Object.entries(manifest.environment ?? {})) {
    if (!/^[A-Z][A-Z0-9_]*$/u.test(name) || typeof value !== "string") throw new Error("invalid declared environment input");
    args.push("--setenv", name, value);
  }
  args.push("--chdir", manifest.cwd, "--", ...manifest.command);
  // Namespace/mount failures are qualification failures. There is deliberately
  // no host execution fallback, automatic installation or host reconfiguration.
  return command(bwrap, args, { env: {}, expectedExit: manifest.expected_exit ?? 0, ...options });
}

// Immutable evidence records contain declared capabilities, never file contents
// or signing material. The owning harness manifest binds the executable source.
export function retainIsolation(kind, manifest, details = {}) {
  const record = { schema_id: "cartulary.reference_pack_isolation_qualification.v1", kind, status: "pass", isolation: "linux_bubblewrap", network: false, manifest, ...details };
  const bytes = Buffer.from(JSON.stringify(record) + "\n");
  const digest = createHash("sha256").update(bytes).digest("hex");
  const directory = path.resolve(".cartulary/reference-packs/isolation", kind);
  mkdirSync(directory, { recursive: true, mode: 0o700 });
  const destination = path.join(directory, digest + ".json");
  try { writeFileSync(destination, bytes, { flag: "wx", mode: 0o600 }); }
  catch (error) { if (error.code !== "EEXIST" || !readFileSync(destination).equals(bytes)) throw error; }
  process.stderr.write(`Isolation evidence: ${destination}\n`);
  return destination;
}

export function privateDirectories(root, names) {
  return names.map(name => { const value = path.join(root, name); mkdirSync(value, { recursive: true, mode: 0o700 }); return value; });
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  if (process.argv.length !== 3) throw new Error("explicit isolation manifest required");
  process.stdout.write(isolated(JSON.parse(readFileSync(process.argv[2], "utf8"))));
}
