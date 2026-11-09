import { readFileSync, writeFileSync, existsSync, realpathSync } from "node:fs";
import path from "node:path";
import { command, isolated, privateDirectories, retainIsolation } from "./isolation.mjs";
import { decodeGoPackages } from "./closure.mjs";
import { pins, verifyArchive } from "../harness/readiness/cryptographic-module.mjs";

// Discovery reads pinned module metadata only. Actual builds run with exactly
// those dependency archives and source trees visible and no shared build cache.
export function goInputs(go, source, work, moduleCache) {
  go = realpathSync(go);
  const [cache, temporary, home] = privateDirectories(work, ["cache", "tmp", "home"]);
  const env = { CGO_ENABLED: "0", GOENV: "off", GOFLAGS: "", GOWORK: "off", GOPROXY: "off", GOSUMDB: "off", GOTOOLCHAIN: "local", GOCACHE: cache, GOMODCACHE: moduleCache, GOTMPDIR: temporary, HOME: home };
  env.GOFIPS140 = pins().cryptographic_module.selector;
  env.GODEBUG = "fips140=on";
  verifyArchive(go, env);
  const modules = decodeGoPackages(command(go, ["-C", source, "list", "-mod=readonly", "-m", "-json", "all"], { env }));
  const root = command(go, ["env", "GOROOT"], { env }).trim();
  const readonly = [root];
  for (const module of modules) {
    if (module.Replace) throw new Error("local or remote module replacements forbidden in qualification");
    if (module.Main) continue;
    if (!module.Version) throw new Error(`unpinned module dependency: ${module.Path}`);
    const escaped = module.Path.replace(/[A-Z]/gu, value => `!${value.toLowerCase()}`);
    const goMod = module.GoMod ?? path.join(moduleCache, "cache/download", escaped, "@v", module.Version + ".mod");
    if (!existsSync(goMod)) throw new Error(`unavailable pinned module metadata: ${module.Path}@${module.Version}`);
    readonly.push(goMod);
    // A graph-only dependency may have no extracted source. Mount only the
    // selected version's retained archive metadata and existing extraction.
    const prefix = goMod.slice(0, -4);
    for (const suffix of [".info", ".zip", ".ziphash"]) if (existsSync(prefix + suffix)) readonly.push(prefix + suffix);
    const directory = module.Dir ?? path.join(moduleCache, escaped + "@" + module.Version);
    if (existsSync(directory)) readonly.push(directory);
  }
  return { readonly: [...new Set(readonly)], writable: [cache, temporary, home], environment: { ...env, GOROOT: root }, cwd: source, go, root, cache, dependencies: modules.filter(m => !m.Main).map(m => ({ module: m.Path, version: m.Version })) };
}

if (process.argv[1]?.endsWith("/isolated-go.mjs")) {
  const [mode, source, work, moduleCache, go, ...extra] = process.argv.slice(2);
  if (!["module", "producer"].includes(mode) || !go) throw new Error("explicit independent Go qualification inputs required");
  const inputs = goInputs(go, source, work, moduleCache);
  const manifest = { ...inputs, readonly: [...inputs.readonly, source] };
  const run = args => isolated({ ...manifest, command: [inputs.go, "-C", source, ...args] });
  if (mode === "module") {
    run(["build", "-mod=readonly", "-trimpath", "-buildvcs=false", "./..."]);
    writeFileSync(path.join(work, "packages.json"), run(["list", "-mod=readonly", "-deps", "-test", "-json", "./..."]));
  } else {
    const [output] = privateDirectories(work, ["binaries"]); manifest.writable.push(output);
    for (const pass of ["first", "second"]) run(["build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-o", path.join(output, pass), `./cmd/${extra[0]}`]);
    if (!readFileSync(path.join(output, "first")).equals(readFileSync(path.join(output, "second")))) throw new Error("isolated producer builds differ");
  }
  retainIsolation(mode, manifest, { source, dependencies: inputs.dependencies, byte_identical: mode === "producer" });
  writeFileSync(path.join(work, "isolation.json"), JSON.stringify({ schema_id: "cartulary.reference_pack_isolation.v1", status: "pass", ...manifest, command: [inputs.go, "build/list"], shared_cache: false, network: false }));
}
