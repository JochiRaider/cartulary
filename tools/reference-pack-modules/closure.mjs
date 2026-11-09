import { isAbsolute, relative, sep } from "node:path";

const app = "github.com/JochiRaider/cartulary";
const format = `${app}/referencepacks/format`;
const inside = (parent, child) => {
  const value = relative(parent, child);
  return value === "" || !value.startsWith(`..${sep}`) && value !== ".." && !isAbsolute(value);
};

// Inspect the compiler's complete dependency graph, including test/embedded
// inputs. A source import search cannot substitute for this closure check.
export function validateModuleClosure(packages, { modulePath, sourceRoot, moduleCache, goRoot, buildCache }) {
  if (!Array.isArray(packages) || packages.length === 0) throw new Error("empty module build closure");
  const admittedDirectories = [sourceRoot, moduleCache, goRoot];
  for (const pkg of packages) {
    if (pkg.Error || pkg.DepsErrors?.length || pkg.Incomplete) throw new Error("incomplete module build closure");
    if (!pkg.Dir || !admittedDirectories.some(root => inside(root, pkg.Dir))) throw new Error("module build reads outside exported source and pinned dependencies");
    if (pkg.Module?.Replace) throw new Error("module dependency replacement is forbidden");
    if (pkg.Module && !pkg.Module.Main && !pkg.Module.Version) throw new Error("unversioned module dependency");
    for (const imported of [pkg.ImportPath, ...(pkg.Imports ?? []), ...(pkg.TestImports ?? []), ...(pkg.XTestImports ?? [])]) {
      const name = imported.split(" [")[0];
      if (name === app || name.startsWith(`${app}/`) && !(name === `${modulePath}.test` && pkg.Name === "main" && pkg.Module?.Main === true || name === modulePath || name.startsWith(`${modulePath}/`) || modulePath !== format && (name === format || name.startsWith(`${format}/`)))) {
        throw new Error("application or sibling producer in module dependency closure");
      }
    }
    for (const group of ["GoFiles", "CgoFiles", "CFiles", "CXXFiles", "MFiles", "HFiles", "FFiles", "SFiles", "SwigFiles", "SwigCXXFiles", "SysoFiles", "EmbedFiles", "TestGoFiles", "XTestGoFiles", "TestEmbedFiles", "XTestEmbedFiles"]) {
      for (const name of pkg[group] ?? []) {
        // `go list -test` writes its synthetic test main into GOCACHE. Admit
        // that compiler-owned output only, never arbitrary absolute inputs.
        const testMain = group === "GoFiles" && pkg.Name === "main" && pkg.Module?.Main === true &&
          (pkg.ImportPath === `${modulePath}.test` || pkg.ImportPath.startsWith(`${modulePath}/`) && pkg.ImportPath.endsWith(".test")) &&
          pkg.GoFiles.length === 1 && buildCache && isAbsolute(name) && inside(buildCache, name) &&
          /^[0-9a-f]{2}\/[0-9a-f]{64}-d$/u.test(relative(buildCache, name).split(sep).join("/"));
        if (testMain) continue;
        if (isAbsolute(name) || name.split(/[\\/]/u).some(part => part === ".." || part === "docs" || /^readme(?:\.|$)/iu.test(part))) throw new Error(`external or prose module build input: ${pkg.ImportPath} ${group}`);
      }
    }
  }
}

export function decodeGoPackages(raw) {
  const documents = []; let start = 0, depth = 0, quoted = false, escape = false;
  for (let index = 0; index < raw.length; index++) {
    const character = raw[index];
    if (quoted) {
      if (escape) escape = false;
      else if (character === "\\") escape = true;
      else if (character === '"') quoted = false;
      continue;
    }
    if (character === '"') { quoted = true; continue; }
    if (character === "{" || character === "[") { if (depth === 0) start = index; depth++; }
    else if (character === "}" || character === "]") {
      depth--;
      if (depth === 0) documents.push(JSON.parse(raw.slice(start, index + 1)));
      if (depth < 0) throw new Error("invalid Go package stream");
    } else if (depth === 0 && !/\s/u.test(character)) throw new Error("invalid Go package stream");
  }
  if (depth || quoted || escape) throw new Error("incomplete Go package stream");
  return documents;
}

// Production executables may consume format readers, never source producers or
// publication/signing packages, including when reached through another module.
export function validateApplicationClosure(packages) {
  if (!Array.isArray(packages) || packages.length === 0) throw new Error("empty application build closure");
  for (const pkg of packages) {
    if (pkg.Error || pkg.DepsErrors?.length || pkg.Incomplete) throw new Error("incomplete application build closure");
    for (const name of [pkg.ImportPath, ...(pkg.Imports ?? [])]) {
      if (name === `${format}/publication` || name.startsWith(`${format}/publication/`) ||
        name === `${app}/internal/modules/reference_data/producer` || name.startsWith(`${app}/internal/modules/reference_data/producer/`) ||
        name.startsWith(`${app}/referencepacks/`) && !(name === format || name.startsWith(`${format}/`))) {
        throw new Error("producer or signing code in application dependency closure");
      }
    }
    for (const name of pkg.EmbedFiles ?? []) {
      const parts = name.split(/[\\/]/u);
      if (isAbsolute(name) || parts.some(part => part === ".." || part === "docs" || /^readme(?:\.|$)/iu.test(part)) ||
        parts.some(part => /^(?:reference-pack-producers|predecessor-contract-v2-projections|predecessor-contract-v2-vectors|attack-native-v2|dataset-v1|signing-context|private-keys)$/u.test(part))) {
        throw new Error("source corpus, private publication input or prose in application embedding closure");
      }
    }
  }
}
