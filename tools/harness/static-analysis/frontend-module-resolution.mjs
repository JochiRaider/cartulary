import path from "node:path";
import ts from "typescript";

// Use the compiler's paths/extends/package rules without discovering a TS
// program. Dependency inspection must not acquire human documents as inputs.
export function createFrontendModuleResolver(root) {
  const directories = new Map();
  const admitted = (file) => {
    const relative = path.relative(root, path.resolve(file));
    return relative !== ".." && !relative.startsWith(`..${path.sep}`) &&
      !path.isAbsolute(relative) && !relative.split(path.sep).includes("docs") &&
      !/\.(?:md|markdown)$/iu.test(relative) && !/^README(?:\.|$)/iu.test(path.basename(relative));
  };
  const host = {
    fileExists: (file) => admitted(file) && ts.sys.fileExists(file),
    readFile: (file) => admitted(file) ? ts.sys.readFile(file) : undefined,
    directoryExists: (directory) => admitted(directory) && ts.sys.directoryExists(directory),
    getCurrentDirectory: () => root,
    useCaseSensitiveFileNames: true,
    readDirectory: () => [],
  };
  function configuration(directory) {
    if (directories.has(directory)) return directories.get(directory);
    const file = path.join(directory,"tsconfig.json");
    let result;
    if (host.fileExists(file)) {
      const loaded = ts.readConfigFile(file,host.readFile);
      if (loaded.error) throw new Error(ts.flattenDiagnosticMessageText(loaded.error.messageText,"\n"));
      const parsed = ts.parseJsonConfigFileContent(loaded.config,host,directory,undefined,file);
      // No program files are collected; absence of inputs is intentional.
      const errors = parsed.errors.filter(error=>error.code!==18003);
      if (errors.length) throw new Error(errors.map(error=>ts.flattenDiagnosticMessageText(error.messageText,"\n")).join("\n"));
      result = {options:parsed.options,cache:ts.createModuleResolutionCache(root,file=>file,parsed.options)};
    } else if (directory !== root) result = configuration(path.dirname(directory));
    else result = {options:{moduleResolution:ts.ModuleResolutionKind.Bundler},cache:ts.createModuleResolutionCache(root,file=>file)};
    directories.set(directory,result);
    return result;
  }
  const resolve = (file,specifier) => {
    if (!admitted(file)) throw new Error("frontend importer is outside executable source scope");
    const {options,cache} = configuration(path.dirname(file));
    return ts.resolveModuleName(specifier,file,options,host,cache).resolvedModule?.resolvedFileName ?? null;
  };
  const packageNames = new Map();
  function packageName(file) {
    if (!file || !admitted(file)) return null;
    let directory=path.dirname(file);
    const visited=[];
    let name=null;
    while (admitted(directory)) {
      if (packageNames.has(directory)) {name=packageNames.get(directory);break;}
      visited.push(directory);
      const manifest=path.join(directory,"package.json");
      if (host.fileExists(manifest)) {name=JSON.parse(host.readFile(manifest)).name ?? null;break;}
      if (directory===root) break;
      directory=path.dirname(directory);
    }
    for (const item of visited) packageNames.set(item,name);
    return name;
  }
  return {resolve,packageName};
}
