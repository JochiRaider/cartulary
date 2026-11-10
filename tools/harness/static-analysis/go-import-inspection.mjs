import { spawnSync } from "node:child_process";
import path from "node:path";

// A synchronous child cannot outlive this inspection or its workspace admission.
// Source collection and process execution stay outside the pure policy evaluator.
export function inspectGoImports(collections, options) {
  return new Map([...inspectGoSources(collections, options)].map(([file, result]) => [file, result.imports]));
}

export function inspectGoDeclarations(collections, options) {
  return new Map([...inspectGoSources(collections, options, true)].map(([file, result]) => [file, result.declarations]));
}

export function inspectGoSQL(collections, options) {
  return new Map([...inspectGoSources(collections, options, false, true)].map(([file, result]) => [file, result.sql]));
}

function inspectGoSources(collections, { binary = process.env.CARTULARY_SOURCECHECK_BIN } = {}, includeDeclarations = false, includeSQL = false) {
  const sources = new Map();
  for (const files of collections) for (const file of files) {
    if (!file.relative.endsWith(".go")) continue;
    if (sources.has(file.relative) && sources.get(file.relative) !== file.content) throw new Error(`conflicting import source ${file.relative}`);
    sources.set(file.relative, file.content);
  }
  if (sources.size === 0) return new Map();
  if (!binary || !path.isAbsolute(binary)) throw new Error("Go import parser requires an installed absolute CARTULARY_SOURCECHECK_BIN");
  const files = [...sources].sort(([a], [b]) => a < b ? -1 : a > b ? 1 : 0).map(([path, source]) => ({ path, source }));
  const child = spawnSync(binary, [], { input: JSON.stringify({ files, ...(includeDeclarations ? {include_declarations: true} : {}), ...(includeSQL ? {include_sql: true} : {}) }), encoding: "utf8", timeout: 30000, maxBuffer: 16 * 1024 * 1024 });
  if (child.error || child.signal || child.status !== 0) throw new Error(`Go import inspection failed: ${child.error?.message ?? child.stderr?.trim() ?? child.signal ?? child.status}`);
  let result;
  try { result = JSON.parse(child.stdout); } catch (cause) { throw new Error("malformed Go import inspection response", { cause }); }
  if (!result || Object.keys(result).length !== 1 || !Array.isArray(result.files) || result.files.length !== files.length) throw new Error("incomplete Go import inspection response");
  const imports = new Map();
  for (let i = 0; i < files.length; i++) {
    const record = result.files[i];
    if (!record || Object.keys(record).sort().join(",") !== (includeDeclarations ? "declarations,imports,path" : includeSQL ? "imports,path,sql" : "imports,path") || record.path !== files[i].path || !Array.isArray(record.imports) || record.imports.some((value) => typeof value !== "string")) throw new Error(`invalid Go import inspection response for ${files[i].path}`);
    if (includeSQL && (!Array.isArray(record.sql) || record.sql.some(value => !validSQLDeclaration(value)))) throw new Error(`invalid Go SQL inspection response for ${files[i].path}`);
    if (includeDeclarations && (!Array.isArray(record.declarations) || record.declarations.some((value) => !value || Object.keys(value).sort().join(",") !== "name,source" || typeof value.name !== "string" || typeof value.source !== "string"))) throw new Error(`invalid Go declaration inspection response for ${files[i].path}`);
    imports.set(record.path, record);
  }
  return imports;
}


function validSQLDeclaration(value) {
  if (!value || Object.keys(value).sort().join(",") !== "name,sql" || typeof value.name !== "string" || !value.name || !Array.isArray(value.sql)) return false;
  return value.sql.every(expr => {
    if (!expr || Object.keys(expr).sort().join(",") !== "column,disposition,dynamic,kind,line,text,unresolved") return false;
    if (!["statement", "execution", "fragment", "command", "forward", "copy", "closed", "empty"].includes(expr.kind) || typeof expr.text !== "string") return false;
    if (!Number.isInteger(expr.line) || expr.line < 1 || !Number.isInteger(expr.column) || expr.column < 1) return false;
    if (!Array.isArray(expr.unresolved) || expr.unresolved.some(reason => typeof reason !== "string" || !reason)) return false;
    if (expr.dynamic !== (expr.unresolved.length > 0)) return false;
    const disposition = expr.kind === "forward" ? "transport" : expr.dynamic ? "rejected" : "analyzed";
    if (expr.disposition !== disposition) return false;
    if (expr.kind === "forward" && (expr.unresolved.length !== 1 || !/^(?:callback )?parameter .+/u.test(expr.unresolved[0]))) return false;
    if (["closed", "empty"].includes(expr.kind) && (expr.text !== "" || expr.dynamic)) return false;
    return true;
  });
}
