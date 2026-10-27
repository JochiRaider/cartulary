import { spawnSync } from "node:child_process";
import path from "node:path";

// A synchronous child cannot outlive this inspection or its workspace admission.
// Source collection and process execution stay outside the pure policy evaluator.
export function inspectGoImports(collections, { binary = process.env.CARTULARY_SOURCECHECK_BIN } = {}) {
  const sources = new Map();
  for (const files of collections) for (const file of files) {
    if (!file.relative.endsWith(".go")) continue;
    if (sources.has(file.relative) && sources.get(file.relative) !== file.content) throw new Error(`conflicting import source ${file.relative}`);
    sources.set(file.relative, file.content);
  }
  if (sources.size === 0) return new Map();
  if (!binary || !path.isAbsolute(binary)) throw new Error("Go import parser requires an installed absolute CARTULARY_SOURCECHECK_BIN");
  const files = [...sources].sort(([a], [b]) => a < b ? -1 : a > b ? 1 : 0).map(([path, source]) => ({ path, source }));
  const child = spawnSync(binary, [], { input: JSON.stringify({ files }), encoding: "utf8", timeout: 30000, maxBuffer: 16 * 1024 * 1024 });
  if (child.error || child.signal || child.status !== 0) throw new Error(`Go import inspection failed: ${child.error?.message ?? child.stderr?.trim() ?? child.signal ?? child.status}`);
  let result;
  try { result = JSON.parse(child.stdout); } catch (cause) { throw new Error("malformed Go import inspection response", { cause }); }
  if (!result || Object.keys(result).length !== 1 || !Array.isArray(result.files) || result.files.length !== files.length) throw new Error("incomplete Go import inspection response");
  const imports = new Map();
  for (let i = 0; i < files.length; i++) {
    const record = result.files[i];
    if (!record || Object.keys(record).sort().join(",") !== "imports,path" || record.path !== files[i].path || !Array.isArray(record.imports) || record.imports.some((value) => typeof value !== "string")) throw new Error(`invalid Go import inspection response for ${files[i].path}`);
    imports.set(record.path, record.imports);
  }
  return imports;
}
