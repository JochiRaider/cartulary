import { createRequire } from "node:module";
import { lstatSync, readFileSync, readdirSync, realpathSync } from "node:fs";
import path from "node:path";
import { goRunnerBuildContext } from "./go-build-context.mjs";
import { discoverRunnerSources, excludedSourcePath, frontendSourceProject, sourceExclusionPolicy } from "./source-discovery.mjs";

const requireFromWeb = createRequire(new URL("../../../apps/web/package.json", import.meta.url));

const goOperatingSystems = new Set([
  "aix",
  "android",
  "darwin",
  "dragonfly",
  "freebsd",
  "illumos",
  "ios",
  "js",
  "linux",
  "netbsd",
  "openbsd",
  "plan9",
  "solaris",
  "wasip1",
  "windows",
]);
const goArchitectures = new Set([
  "386",
  "amd64",
  "arm",
  "arm64",
  "loong64",
  "mips",
  "mips64",
  "mips64le",
  "mipsle",
  "ppc64",
  "ppc64le",
  "riscv64",
  "s390x",
  "wasm",
]);

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function pathUnder(candidate, parent) {
  return candidate === parent || candidate.startsWith(`${parent}${path.sep}`);
}

function assertNoSymlinkComponents(root, relativePath, label) {
  let current = root;
  for (const segment of relativePath.split(path.posix.sep)) {
    current = path.join(current, segment);
    if (lstatSync(current).isSymbolicLink()) {
      throw new Error(`${label} must not traverse a symbolic link`);
    }
  }
}

function sourcePathExists(root, relativePath, label) {
  try { assertNoSymlinkComponents(root, relativePath, label); return true; }
  catch (error) { if (error.code === "ENOENT") return false; throw error; }
}

function assertSourceBoundary(relative, approvedRoots, label) {
  if (excludedSourcePath(relative)) throw new Error(`${label} is an excluded source path`);
  if (!approvedRoots.some((approved) => relative === approved || relative.startsWith(`${approved}/`))) {
    throw new Error(`${label} is outside approved runner roots`);
  }
}

function containedFile(root, relativePath, approvedRoots, label, scope) {
  if (path.isAbsolute(relativePath) || relativePath.includes("\\")) {
    throw new Error(`${label} must be a normalized repository-relative path`);
  }
  const normalized = path.posix.normalize(relativePath);
  if (normalized !== relativePath || normalized.startsWith("../")) {
    throw new Error(`${label} contains traversal or normalization drift`);
  }
  if (/[*?{}[\]]/u.test(relativePath)) {
    throw new Error(`${label} must not contain a glob`);
  }
  assertSourceBoundary(relativePath, approvedRoots, label);
  const lexical = path.resolve(root, relativePath);
  const resolvedRoot = realpathSync(root);
  if (!sourcePathExists(root, relativePath, label)) throw new Error(`${label} does not exist: ${relativePath}`);
  const stat = lstatSync(lexical);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error(`${label} must resolve to a non-symlink regular file`);
  }
  const resolved = realpathSync(lexical);
  if (!pathUnder(resolved, resolvedRoot)) {
    throw new Error(`${label} escapes the repository`);
  }
  if (
    approvedRoots.length > 0 &&
    !approvedRoots.some((approvedRoot) =>
      pathUnder(resolved, realpathSync(path.resolve(root, approvedRoot))),
    )
  ) {
    throw new Error(`${label} is outside approved runner roots`);
  }
  const source = scope?.sources.get(lexical) ?? readFileSync(lexical, "utf8");
  scope?.sources.set(lexical, source);
  return { lexical, source };
}

function staticString(node, ts, bindings = new Map(), seen = new Set()) {
  while (
    ts.isAsExpression(node) ||
    ts.isSatisfiesExpression(node) ||
    ts.isTypeAssertionExpression(node) ||
    ts.isParenthesizedExpression(node)
  ) {
    node = node.expression;
  }
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) {
    return node.text;
  }
  if (ts.isIdentifier(node) && bindings.has(node.text) && !seen.has(bindings.get(node.text))) {
    const binding = bindings.get(node.text);
    if (!binding.node) return null;
    return staticString(
      binding.node,
      ts,
      binding.bindings,
      new Set([...seen, binding]),
    );
  }
  if (ts.isElementAccessExpression(node)) {
    const collection = staticValue(node.expression, ts, bindings, seen);
    const index = node.argumentExpression
      ? staticValue(node.argumentExpression, ts, bindings, seen)
      : undefined;
    if (Array.isArray(collection) && Number.isInteger(index)) {
      return typeof collection[index] === "string" ? collection[index] : null;
    }
  }
  if (
    ts.isBinaryExpression(node) &&
    node.operatorToken.kind === ts.SyntaxKind.PlusToken
  ) {
    const left = staticString(node.left, ts, bindings, seen);
    const right = staticString(node.right, ts, bindings, seen);
    return left === null || right === null ? null : left + right;
  }
  return null;
}

function staticValue(node, ts, bindings = new Map(), seen = new Set()) {
  while (
    ts.isAsExpression(node) ||
    ts.isSatisfiesExpression(node) ||
    ts.isTypeAssertionExpression(node) ||
    ts.isParenthesizedExpression(node)
  ) {
    node = node.expression;
  }
  if (ts.isIdentifier(node) && bindings.has(node.text) && !seen.has(bindings.get(node.text))) {
    const binding = bindings.get(node.text);
    if (!binding.node) return undefined;
    return staticValue(
      binding.node,
      ts,
      binding.bindings,
      new Set([...seen, binding]),
    );
  }
  const string = staticString(node, ts, bindings, seen);
  if (string !== null) return string;
  if (ts.isNumericLiteral(node)) return Number(node.text);
  if (ts.isPrefixUnaryExpression(node) && [ts.SyntaxKind.MinusToken, ts.SyntaxKind.PlusToken].includes(node.operator)) {
    const value = staticValue(node.operand, ts, bindings, seen);
    if (typeof value === "number") return node.operator === ts.SyntaxKind.MinusToken ? -value : value;
  }
  if (
    ts.isPropertyAccessExpression(node) &&
    ts.isIdentifier(node.expression) &&
    node.expression.text === "Number" &&
    node.name.text === "NaN"
  ) return Number.NaN;
  if (node.kind === ts.SyntaxKind.TrueKeyword) return true;
  if (node.kind === ts.SyntaxKind.FalseKeyword) return false;
  if (node.kind === ts.SyntaxKind.NullKeyword) return null;
  if (ts.isArrayLiteralExpression(node)) {
    return node.elements.map((element) => staticValue(element, ts, bindings, seen));
  }
  if (ts.isObjectLiteralExpression(node)) {
    const result = {};
    for (const property of node.properties) {
      if (!ts.isPropertyAssignment(property)) return undefined;
      const key = property.name && ts.isIdentifier(property.name)
        ? property.name.text
        : property.name && staticString(property.name, ts);
      if (key === null) return undefined;
      const value = staticValue(property.initializer, ts, bindings, seen);
      if (value !== undefined) result[key] = value;
    }
    return result;
  }
  return undefined;
}

function formatEachTitle(template, row, index) {
  const values = Array.isArray(row) ? row : [row];
  let cursor = 0;
  // The pinned runner owns primitive formatting and quoted display. Discovery
  // invokes that pure formatter, never the test module or its fixture code.
  const { format, objDisplay } = requireFromWeb("@vitest/utils/display");
  if (/%(?![sdif#$])/u.test(template)) throw new Error(`unsupported parameter formatter in test title ${JSON.stringify(template)}`);
  return template.replace(/%([sdif#$])|\$([$\w.]+)/gu, (match, token, key) => {
    if (token === "#") return String(index);
    if (token === "$") return String(index + 1);
    if (token) {
      const value = values[cursor++];
      if (value === undefined || (value !== null && typeof value === "object")) throw new Error(`unresolved primitive parameter in test title ${JSON.stringify(template)}`);
      // The pinned runner preserves negative zero for %f before formatting.
      return token === "f" && Object.is(value, -0) ? "-0" : format(`%${token}`, value);
    }
    if (!row || typeof row !== "object" || Array.isArray(row)) return match;
    const value = row[key];
    if (!Object.hasOwn(row, key) || (value !== null && typeof value === "object") || value === undefined)
      throw new Error(`unresolved named parameter ${key} in test title`);
    return objDisplay(value);
  });
}

function expandedEachTitles(expression, template, ts, bindings) {
  if (!ts.isCallExpression(expression)) return null;
  const name = calleeName(expression.expression, ts);
  if (!name.endsWith(".each") || expression.arguments.length !== 1) return null;
  const rows = staticValue(expression.arguments[0], ts, bindings);
  if (!Array.isArray(rows) || rows.length === 0 || rows.includes(undefined)) return null;
  return rows.map((row, index) => formatEachTitle(template, row, index));
}

function calleeName(expression, ts) {
  if (ts.isIdentifier(expression)) {
    return expression.text;
  }
  if (ts.isPropertyAccessExpression(expression)) {
    const owner = calleeName(expression.expression, ts);
    return owner ? `${owner}.${expression.name.text}` : "";
  }
  if (ts.isElementAccessExpression(expression)) {
    const owner = calleeName(expression.expression, ts);
    return owner ? `${owner}.[computed]` : "";
  }
  if (ts.isCallExpression(expression)) {
    return calleeName(expression.expression, ts);
  }
  return "";
}

function registrationSuiteSources(root, file, sourceFile, ts, scope) {
  const suiteSources = [];
  for (const statement of sourceFile.statements) {
    if (
      !ts.isImportDeclaration(statement) ||
      !ts.isStringLiteral(statement.moduleSpecifier) ||
      !statement.moduleSpecifier.text.startsWith(".") ||
      !statement.importClause ||
      !statement.importClause.namedBindings ||
      !ts.isNamedImports(statement.importClause.namedBindings) ||
      !statement.importClause.namedBindings.elements.some((element) =>
        /^register[A-Za-z0-9]+Suite$/u.test(element.name.text) || /^register[A-Za-z0-9]+Suite$/u.test((element.propertyName ?? element.name).text),
      )
    ) {
      continue;
    }
    const importerDirectory = path.posix.dirname(file);
    const importBase = path.posix.normalize(
      path.posix.join(importerDirectory, statement.moduleSpecifier.text),
    );
    if (!importBase.startsWith(`${importerDirectory}/`)) {
      throw new Error(
        `${file}.registration_suite ${statement.moduleSpecifier.text} must remain beneath the selector directory`,
      );
    }
    const candidates = /\.[cm]?[jt]sx?$/u.test(importBase)
      ? [importBase]
      : [`${importBase}.ts`, `${importBase}.tsx`];
    for (const entry of candidates) assertSourceBoundary(entry, [importerDirectory], `${file}.registration_suite`);
    const candidate = candidates.find((entry) => sourcePathExists(root, entry, `${file}.registration_suite`));
    if (candidate === undefined) {
      throw new Error(
        `${file}.registration_suite ${statement.moduleSpecifier.text} does not resolve to TypeScript source`,
      );
    }
    const contained = containedFile(
      root,
      candidate,
      [importerDirectory],
      `${file}.registration_suite`,
      scope,
    );
    for (const element of statement.importClause.namedBindings.elements) {
      const localName = element.name.text;
      const exportName = (element.propertyName ?? element.name).text;
      if (!/^register[A-Za-z0-9]+Suite$/u.test(localName) && !/^register[A-Za-z0-9]+Suite$/u.test(exportName)) continue;
      const calls = [];
      const collect = (node) => {
        if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === localName) calls.push(node);
        ts.forEachChild(node, collect);
      };
      collect(sourceFile);
      const topLevel = sourceFile.statements.filter((entry) => ts.isExpressionStatement(entry) && calls.includes(entry.expression));
      if (calls.length !== 1 || topLevel.length !== 1 || calls[0].arguments.length !== 0)
        throw new Error(`${file}: registration suite ${localName} must have exactly one unconditional top-level zero-argument call`);
      suiteSources.push({ file: candidate, source: contained.source, exportName });
    }
  }
  return suiteSources;
}

export function registrationSuiteFiles(root, file, source) {
  const ts = requireFromWeb("typescript");
  const scriptKind = file.endsWith("x") ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
  const sourceFile = ts.createSourceFile(
    file,
    source,
    ts.ScriptTarget.Latest,
    false,
    scriptKind,
  );
  return registrationSuiteSources(root, file, sourceFile, ts).map(
    (suite) => suite.file,
  );
}

function testTitles(root, file, source, scope, exportName = null) {
  const cacheKey = JSON.stringify([file, exportName]);
  if (scope.titles.has(cacheKey)) return scope.titles.get(cacheKey);
  const ts = requireFromWeb("typescript");
  const scriptKind = file.endsWith("x") ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
  const sourceFile = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true, scriptKind);
  const aliases = new Map(["it", "test", "describe", "suite"].map((name) => [name, name]));
  for (const statement of sourceFile.statements) {
    if (!ts.isImportDeclaration(statement)) continue;
    const bindings = statement.importClause?.namedBindings;
    if (bindings && ts.isNamedImports(bindings)) {
      for (const item of bindings.elements) {
        const imported = (item.propertyName ?? item.name).text;
        if (["it", "test", "describe", "suite"].includes(imported)) aliases.set(item.name.text, imported);
      }
    } else if (bindings && ts.isNamespaceImport(bindings) && statement.moduleSpecifier.text === "vitest") {
      aliases.set(bindings.name.text, "vitest");
    }
  }
  function canonicalName(expression, bindings, seen = new Set()) {
    const parts = calleeName(expression, ts).split(".");
    let canonical;
    const binding = bindings.get(parts[0]);
    if (binding && !seen.has(binding)) {
      seen.add(binding);
      if (binding.node && (ts.isIdentifier(binding.node) || ts.isPropertyAccessExpression(binding.node))) {
        const candidate = canonicalName(binding.node, binding.bindings, seen);
        if (/^(?:it|test|describe|suite|vitest)(?:\.(?:it|test|describe|suite|each|only|skip|todo|concurrent|sequential|serial|parallel|fails))*$/u.test(candidate)) canonical = candidate;
      }
      else if (binding.node && ts.isCallExpression(binding.node)) {
        const factory = canonicalName(binding.node.expression, binding.bindings, seen);
        if (factory === "test.extend") canonical = "test";
      }
      if (!canonical && aliases.has(parts[0])) throw new Error(`${file}: unresolved runner binding ${parts[0]}`);
    } else canonical = aliases.get(parts[0]);
    if (!canonical) return "";
    const resolved = [canonical, ...parts.slice(1)];
    return (resolved.length > 1 ? resolved.filter((part) => part !== "vitest") : resolved).join(".");
  }
  const matches = new Map();
  const fullTitles = new Map();
  const titlePaths = new Map();
  let testIdentity = 0;
  let matchedExport = false;
  function record(candidate, identity, fullTitle, titlePath) {
    if (!titlePaths.has(fullTitle)) titlePaths.set(fullTitle, new Map());
    titlePaths.get(fullTitle).set(JSON.stringify(titlePath), titlePath);
    if (!fullTitles.has(candidate)) fullTitles.set(candidate, new Set());
    fullTitles.get(candidate).add(fullTitle);
    if (!matches.has(candidate)) matches.set(candidate, new Set());
    matches.get(candidate).add(identity);
  }
  function visit(node, ancestors, inherited, allowed) {
    let bindings = inherited;
    if (ts.isImportDeclaration(node) || ts.isTypeNode(node) || ts.isInterfaceDeclaration(node) || ts.isTypeAliasDeclaration(node)) return;
    if (ts.isSourceFile(node) || ts.isBlock(node)) {
      bindings = new Map(inherited);
      for (const statement of node.statements) {
        if (!ts.isVariableStatement(statement)) continue;
        for (const declaration of statement.declarationList.declarations) {
          if (!ts.isIdentifier(declaration.name)) continue;
          bindings.set(declaration.name.text, { node: null, bindings });
          if ((statement.declarationList.flags & ts.NodeFlags.Const) && declaration.initializer)
            bindings.set(declaration.name.text, { node: declaration.initializer, bindings });
        }
      }
      for (const statement of node.statements) visit(statement, ancestors, bindings, allowed);
      return;
    }
    if (ts.isVariableDeclaration(node)) {
      if (!node.initializer) return;
      const runner = canonicalName(node.initializer, bindings);
      if (runner) {
        if (!ts.isIdentifier(node.name) || !ts.isVariableDeclarationList(node.parent ?? {}) || !(node.parent.flags & ts.NodeFlags.Const)) {
          throw new Error(`${file}: runner aliases must use a simple constant binding`);
        }
        // A literal alias or the supported Playwright fixture extension retains
        // the same registration API. Object destructuring and API escapes do not.
        if (ts.isIdentifier(node.initializer) || ts.isPropertyAccessExpression(node.initializer)) return;
      }
      visit(node.initializer, ancestors, bindings, false);
      return;
    }
    if (ts.isFunctionDeclaration(node) && exportName && node.name?.text === exportName) {
      if (matchedExport || !node.modifiers?.some((modifier) => modifier.kind === ts.SyntaxKind.ExportKeyword) || node.parameters.length || !node.body)
        throw new Error(`${file}: unsupported registration suite export ${exportName}`);
      matchedExport = true;
      visit(node.body, ancestors, bindings, true);
      return;
    }
    if (ts.isCallExpression(node)) {
      const name = canonicalName(node.expression, bindings);
      const segments = name.split(".");
      const base = segments[0];
      const isSuite = ["describe", "suite"].includes(base) || (base === "test" && segments[1] === "describe");
      const isTest = ["it", "test"].includes(base) && !isSuite;
      const controls = ["configure", "beforeEach", "afterEach", "beforeAll", "afterAll", "use", "setTimeout", "extend", "step", "info", "fail", "slow"];
      const isControl = segments.some((segment) => controls.includes(segment));
      if ((isSuite || isTest) && !isControl) {
        if (!allowed) throw new Error(`${file}:${sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile)).line + 1}: unsupported conditional or indirect test registration ${name}`);
        const modifiers = ["it", "test", "describe", "suite", "each", "only", "skip", "todo", "concurrent", "sequential", "serial", "parallel", "fails"];
        if (!segments.every((segment) => modifiers.includes(segment))) throw new Error(`${file}: unsupported test registration modifier ${name}`);
        const title = node.arguments[0] ? staticString(node.arguments[0], ts, bindings) : null;
        if (title === null) throw new Error(`${file}: unresolved ${isSuite ? "suite" : "test"} registration title`);
        if (isSuite) {
          if (segments.includes("each")) throw new Error(`${file}: parameterized suite registration is unsupported`);
          const callback = [...node.arguments].reverse().find((argument) => ts.isArrowFunction(argument) || ts.isFunctionExpression(argument));
          if (!callback || callback.parameters.length) throw new Error(`${file}: unsupported suite callback`);
          visit(callback.body, [...ancestors, title], bindings, true);
        } else {
          const expanded = segments.includes("each") ? expandedEachTitles(node.expression, title, ts, bindings) : [title];
          if (expanded === null) throw new Error(`${file}: unresolved parameterized test registration ${JSON.stringify(title)}`);
          for (const expandedTitle of expanded) {
            const fullTitle = [...ancestors, expandedTitle].join(" ");
            const identity = testIdentity++;
            record(expandedTitle, identity, fullTitle, [...ancestors, expandedTitle]);
            record(fullTitle, identity, fullTitle, [...ancestors, expandedTitle]);
          }
        }
        return;
      }
      if (isControl) {
        for (const argument of node.arguments) visit(argument, ancestors, bindings, false);
        return;
      }
    }
    if (ts.isPropertyAccessExpression(node)) {
      // Property names are not references to imported runner bindings (for
      // example a regular expression's .test method).
      visit(node.expression, ancestors, bindings, false);
      return;
    }
    if (ts.isPropertyAssignment(node) || ts.isParameter(node)) {
      if (node.initializer) visit(node.initializer, ancestors, bindings, false);
      return;
    }
    if (ts.isIdentifier(node) && canonicalName(node, bindings))
      throw new Error(`${file}:${sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile)).line + 1}: unsupported indirect use of the test registration API`);
    // Only an expression statement preserves unconditional registration. Calls,
    // helper functions and control flow are inspected to reject hidden tests,
    // never interpreted as if their callbacks necessarily execute once.
    const childAllowed = allowed && ts.isExpressionStatement(node);
    ts.forEachChild(node, (child) => visit(child, ancestors, bindings, childAllowed));
  }
  visit(sourceFile, [], new Map(), exportName === null);
  if (exportName && !matchedExport) throw new Error(`${file}: missing registration suite export ${exportName}`);
  const counts = new Map([...matches].map(([candidate, identities]) => [candidate, identities.size]));
  const registrationSuites = registrationSuiteSources(root, file, sourceFile, ts, scope);
  for (const suite of registrationSuites) {
    const imported = testTitles(root, suite.file, suite.source, scope, suite.exportName);
    for (const [full, paths] of imported.titlePaths) {
      if (!titlePaths.has(full)) titlePaths.set(full, new Map());
      for (const [key, value] of paths) titlePaths.get(full).set(key, value);
    }
    for (const [title, count] of imported) {
      counts.set(title, (counts.get(title) ?? 0) + count);
      if (!fullTitles.has(title)) fullTitles.set(title, new Set());
      for (const full of imported.fullTitles.get(title)) fullTitles.get(title).add(full);
    }
  }
  counts.fullTitles = fullTitles;
  counts.titlePaths = titlePaths;
  counts.hasRegistrationSuites = registrationSuites.length !== 0;
  scope.titles.set(cacheKey, counts);
  return counts;
}


function packageDirectory(root, packagePath, approvedRoots, label) {
  if (!packagePath.startsWith("./") || /[*?{}[\]\\]/u.test(packagePath)) {
    throw new Error(`${label} must be an exact repository package`);
  }
  const relative = packagePath.slice(2);
  if (path.posix.normalize(relative) !== relative || relative.split("/").includes("..")) {
    throw new Error(`${label} contains traversal or normalization drift`);
  }
  assertSourceBoundary(relative, approvedRoots, label);
  const lexical = path.resolve(root, relative);
  if (!sourcePathExists(root, relative, label)) {
    throw new Error(`${label} does not exist: ${packagePath}`);
  }
  const stat = lstatSync(lexical);
  if (!stat.isDirectory() || stat.isSymbolicLink()) {
    throw new Error(`${label} must resolve to a non-symlink directory`);
  }
  const resolved = realpathSync(lexical);
  const resolvedRoot = realpathSync(root);
  if (!pathUnder(resolved, resolvedRoot)) {
    throw new Error(`${label} escapes the repository`);
  }
  if (
    !approvedRoots.some((approvedRoot) =>
      pathUnder(resolved, realpathSync(path.resolve(root, approvedRoot))),
    )
  ) {
    throw new Error(`${label} is outside approved runner roots`);
  }
  return lexical;
}

function fileNameMatchesGoBuildContext(fileName, context) {
  const stem = fileName.replace(/_test\.go$/u, "");
  const segments = stem.split("_");
  const last = segments.at(-1);
  const previous = segments.at(-2);
  if (goOperatingSystems.has(previous) && goArchitectures.has(last)) {
    return previous === context.goos && last === context.goarch;
  }
  if (goOperatingSystems.has(last)) return last === context.goos;
  if (goArchitectures.has(last)) return last === context.goarch;
  return true;
}

function tokenizeGoBuildExpression(expression, label) {
  const tokens = [];
  let cursor = 0;
  while (cursor < expression.length) {
    if (/\s/u.test(expression[cursor])) {
      cursor += 1;
      continue;
    }
    const operator = expression.slice(cursor, cursor + 2);
    if (operator === "&&" || operator === "||") {
      tokens.push(operator);
      cursor += 2;
      continue;
    }
    if (expression[cursor] === "!" || expression[cursor] === "(" || expression[cursor] === ")") {
      tokens.push(expression[cursor]);
      cursor += 1;
      continue;
    }
    const match = /^[A-Za-z0-9_.]+/u.exec(expression.slice(cursor));
    if (!match) throw new Error(`${label} contains invalid Go build constraint syntax`);
    tokens.push(match[0]);
    cursor += match[0].length;
  }
  return tokens;
}

function evaluateGoBuildExpression(expression, context, label) {
  const tokens = tokenizeGoBuildExpression(expression, label);
  if (tokens.some((tag) => tag.startsWith("goexperiment."))) {
    throw new Error(`${label} uses an unsupported experimental Go build context`);
  }
  let cursor = 0;
  const current = () => tokens[cursor];
  const consume = () => tokens[cursor++];
  const parseUnary = () => {
    if (current() === "!") {
      consume();
      return !parseUnary();
    }
    if (current() === "(") {
      consume();
      const value = parseOr();
      if (consume() !== ")") throw new Error(`${label} has an unmatched parenthesis`);
      return value;
    }
    const tag = consume();
    if (!tag || ["&&", "||", ")"].includes(tag)) {
      throw new Error(`${label} contains an incomplete Go build constraint`);
    }
    return context.tags.has(tag);
  };
  const parseAnd = () => {
    let value = parseUnary();
    while (current() === "&&") {
      consume();
      const right = parseUnary();
      value = value && right;
    }
    return value;
  };
  const parseOr = () => {
    let value = parseAnd();
    while (current() === "||") {
      consume();
      const right = parseAnd();
      value = value || right;
    }
    return value;
  };
  if (tokens.length === 0) throw new Error(`${label} has an empty Go build constraint`);
  const result = parseOr();
  if (cursor !== tokens.length) throw new Error(`${label} contains trailing Go build constraint tokens`);
  return result;
}

function sourceMatchesGoBuildContext(source, context, label) {
  const expressions = [...source.matchAll(/^\/\/go:build[ \t]+(.+)$/gmu)].map((match) => match[1].trim());
  if (expressions.length > 1) throw new Error(`${label} contains multiple //go:build lines`);
  if (expressions.length === 1) {
    return evaluateGoBuildExpression(expressions[0], context, label);
  }
  if (/^\/\/ \+build[ \t]+/mu.test(source)) {
    throw new Error(`${label} uses unsupported legacy // +build syntax`);
  }
  return true;
}

function goSymbols(root, directory, scope, discovery = false) {
  const cacheKey = `${discovery ? "discovery" : "selection"}:${directory}`;
  if (scope.packages.has(cacheKey)) return scope.packages.get(cacheKey);
  const context = scope.goContext ??= goRunnerBuildContext(root);
  const activeCounts = new Map();
  const excludedFiles = new Map();
  const excluded = discovery ? (scope.excluded ??= sourceExclusionPolicy(root)) : () => false;
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const relative = path.relative(root, path.join(directory, entry.name)).split(path.sep).join("/");
    if (excluded(relative)) continue;
    if (!entry.isFile() || !entry.name.endsWith("_test.go") || /^[._]/u.test(entry.name)) {
      continue;
    }
    const source = readFileSync(path.join(directory, entry.name), "utf8");
    const active = fileNameMatchesGoBuildContext(entry.name, context) &&
      sourceMatchesGoBuildContext(source, context, entry.name);
    // Comments and literals can contain apparent declarations. Preserve lines
    // while masking them; only source-level declarations may become selectors.
    const declarations = source.replace(/\/\*[\s\S]*?\*\/|\/\/[^\n]*|`[^`]*`|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/gu,
      (literal) => literal.replace(/[^\n]/gu, " "));
    for (const match of declarations.matchAll(/^func\s+(Test[\p{L}\p{Nd}_]*)\s*\(/gmu)) {
      // Go treats the suffix as a test unless its first rune is lowercase.
      // The current operational selector contract admits ASCII names only;
      // unsupported executable identities must be visible, never omitted.
      if (/^Test\p{Ll}/u.test(match[1])) continue;
      if (active && !/^Test[A-Za-z0-9_]*$/u.test(match[1])) {
        throw new Error(`${relative}: unsupported Go test identity ${match[1]}; explicit runner contract admission is required`);
      }
      if (active) activeCounts.set(match[1], (activeCounts.get(match[1]) ?? 0) + 1);
      else {
        const files = excludedFiles.get(match[1]) ?? [];
        files.push(entry.name);
        excludedFiles.set(match[1], files);
      }
    }
  }
  const symbols = { activeCounts, context, excludedFiles };
  scope.packages.set(cacheKey, symbols);
  return symbols;
}

function resolveScopedRowSelector({ root, row, runner, taskSurfaceCommandIDs }, scope) {
  const label = `${row.row_id}.selector`;
  if (row.runner === "go") {
    const directory = packageDirectory(root, row.selector.package, runner.approved_roots, `${label}.package`);
    const symbols = goSymbols(root, directory, scope);
    for (const symbol of row.selector.tests) {
      if ((symbols.activeCounts.get(symbol) ?? 0) === 0 && symbols.excludedFiles.has(symbol)) {
        throw new Error(
          `${label}.tests ${symbol} is excluded from the Go runner build context ` +
          `${symbols.context.label}: ${symbols.excludedFiles.get(symbol).sort(asciiCompare).join(", ")}`,
        );
      }
      if ((symbols.activeCounts.get(symbol) ?? 0) !== 1) {
        throw new Error(`${label}.tests ${symbol} must resolve exactly once`);
      }
    }
    return row.selector.tests.map((symbol) => `go:${row.selector.package}:${symbol}`);
  }
  if (row.runner === "vitest" || row.runner === "playwright") {
    const { source } = containedFile(root, row.selector.file, runner.approved_roots, `${label}.file`, scope);
    const titles = testTitles(root, row.selector.file, source, scope);
    for (const title of row.selector.titles) {
      if ((titles.get(title) ?? 0) !== 1) {
        throw new Error(`${label}.titles ${JSON.stringify(title)} must resolve exactly once`);
      }
    }
    if (row.runner === "playwright") {
      for (const title of row.selector.titles) playwrightTitlePath(titles, title, label);
      if (!runner.project_ids.includes(row.selector.project_id)) {
        throw new Error(`${label}.project_id is not registered`);
      }
      if (!runner.stages.includes(row.selector.stage)) {
        throw new Error(`${label}.stage is not registered`);
      }
      if (row.selector.scenario_ids.length !== row.selector.titles.length) {
        throw new Error(`${label} scenario_ids and titles must have equal length`);
      }
      return row.selector.scenario_ids.map(
        (scenarioID) => `playwright:${row.selector.project_id}:${row.selector.stage}:${scenarioID}`,
      ).concat(row.selector.titles.map((title) =>
        `playwright-source:${row.selector.project_id}:${row.selector.file}:${[...titles.fullTitles.get(title)][0]}`,
      ));
    }
    if (runner.project_ids && !runner.project_ids.includes(row.selector.project_id)) throw new Error(`${label}.project_id is not registered`);
    return row.selector.titles.map((title) => `vitest:${row.selector.project_id ?? ""}:${row.selector.file}:${[...titles.fullTitles.get(title)][0]}`);
  }
  if (row.runner === "shell") {
    if (!taskSurfaceCommandIDs.has(row.selector.command_id)) {
      throw new Error(`${label}.command_id is not registered`);
    }
    return [`shell:${row.selector.command_id}`];
  }
  throw new Error(`${row.row_id}.runner is unsupported`);
}

// Facts belong to one synchronous catalog load. Sharing them across its rows
// avoids rescanning packages and reparsing source; a fresh load observes edits,
// imported registration suites and changed toolchain pins without invalidation
// heuristics or process-global source retention. Path checks remain per selector.
export function createRowSelectorResolver(root) {
  const scope = { packages: new Map(), sources: new Map(), titles: new Map(), goContext: null };
  const resolve = (options) => resolveScopedRowSelector({ ...options, root }, scope);
  resolve.frontendTitles = (file, runner) => {
    const { source } = containedFile(root, file, runner.approved_roots, "source registration", scope);
    const facts = testTitles(root, file, source, scope);
    return new Set([...facts.fullTitles.values()].flatMap((values) => [...values]));
  };
  resolve.assertSourceCoverage = (runners, selected) => {
    const missing = [];
    for (const runner of runners.values()) {
      if (runner.runner === "shell") continue;
      const files = discoverRunnerSources(root, runner);
      if (runner.runner === "go") {
        const packages = [...new Set(files.map((file) => `./${path.posix.dirname(file)}`))];
        for (const packageName of packages) {
          const directory = packageDirectory(root, packageName, runner.approved_roots, "source coverage");
          const { activeCounts } = goSymbols(root, directory, scope, true);
          for (const [symbol, count] of activeCounts) {
            if (symbol === "TestMain") continue;
            if (count !== 1 || !selected.has(`go:${packageName}:${symbol}`)) missing.push(`go:${packageName}:${symbol}`);
          }
        }
      } else {
        for (const file of files) {
          const { source } = containedFile(root, file, runner.approved_roots, "source coverage", scope);
          const titles = testTitles(root, file, source, scope);
          for (const title of new Set([...titles.fullTitles.values()].flatMap((values) => [...values]))) {
            const key = runner.runner === "vitest"
              ? `vitest:${frontendSourceProject(file)}:${file}:${title}`
              : `playwright-source:chromium:${file}:${title}`;
            if (!selected.has(key)) missing.push(key);
          }
        }
      }
    }
    if (missing.length) throw new Error(`${missing.length} source tests are not cataloged exactly once:\n${missing.sort(asciiCompare).join("\n")}`);
  };
  return resolve;
}

export function resolveRowSelector({ root, ...options }) {
  return createRowSelectorResolver(root)(options);
}

export function assertSortedUnique(values, label) {
  const sorted = [...values].sort(asciiCompare);
  if (JSON.stringify(values) !== JSON.stringify(sorted)) {
    throw new Error(`${label} must be ASCII-sorted`);
  }
  if (new Set(values).size !== values.length) {
    throw new Error(`${label} must not contain duplicates`);
  }
}

// Resolve catalog selectors once per invocation to full names, using the same
// source facts as catalog overlap validation. No suffix-based runtime selection.
export function resolveVitestTitles(root, rows) {
  const runner = JSON.parse(readFileSync(path.join(root, "tools/test_runner_registry.json"), "utf8")).runners.find((entry) => entry.runner === "vitest");
  const scope = { sources: new Map(), titles: new Map() };
  return new Map(rows.map((row) => {
    const { source } = containedFile(root, row.selector.file, runner.approved_roots, row.row_id, scope);
    const facts = testTitles(root, row.selector.file, source, scope);
    return [row.row_id, row.selector.titles.map((title) => {
      if (facts.get(title) !== 1) throw new Error(`${row.row_id}: selected Vitest title must resolve exactly once`);
      return [...facts.fullTitles.get(title)][0];
    })];
  }));
}


function playwrightTitlePath(facts, title, label) {
  if (facts.hasRegistrationSuites) throw new Error(`${label}: Playwright imported registration suites are outside the exact test-list profile`);
  const full = [...(facts.fullTitles.get(title) ?? [])];
  const paths = full.length === 1 ? [...facts.titlePaths.get(full[0]).values()] : [];
  if (facts.get(title) !== 1 || paths.length !== 1) throw new Error(`${label}: browser title must resolve exactly once`);
  const selected = paths[0];
  if (selected.some(part => !part || part.trim() !== part || /[\r\n\u203a]/u.test(part))) {
    throw new Error(`${label}: browser title cannot be represented by the pinned exact test-list interface`);
  }
  const prefixMatches = [...facts.titlePaths.values()].flatMap(values => [...values.values()])
    .filter(parts => selected.length <= parts.length && selected.every((part, index) => parts[index] === part));
  if (prefixMatches.length !== 1) throw new Error(`${label}: browser test-list identity overlaps a suite prefix`);
  return selected;
}

// Ephemeral runner selection derived from the existing catalog, never another
// execution inventory. Preserve suite components for Playwright's test-list API.
export function resolvePlaywrightSelection(root, rows) {
  const runner = JSON.parse(readFileSync(path.join(root, "tools/test_runner_registry.json"), "utf8")).runners.find(entry => entry.runner === "playwright");
  const scope = { sources: new Map(), titles: new Map() };
  return rows.flatMap(row => {
    if (row.runner !== "playwright") throw new Error("browser selection requires Playwright rows");
    const { source } = containedFile(root, row.selector.file, runner.approved_roots, row.row_id, scope);
    const facts = testTitles(root, row.selector.file, source, scope);
    return row.selector.titles.map(title => ({
      row_id: row.row_id, project_id: row.selector.project_id, file: row.selector.file,
      title_path: playwrightTitlePath(facts, title, row.row_id),
    }));
  });
}
