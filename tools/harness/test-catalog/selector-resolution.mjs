import { createRequire } from "node:module";
import { existsSync, lstatSync, readFileSync, readdirSync, realpathSync } from "node:fs";
import path from "node:path";

const sourceTitleCache = new Map();
const goBuildContextCache = new Map();

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

function containedFile(root, relativePath, approvedRoots, label) {
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
  const lexical = path.resolve(root, relativePath);
  const resolvedRoot = realpathSync(root);
  if (!existsSync(lexical)) {
    throw new Error(`${label} does not exist: ${relativePath}`);
  }
  assertNoSymlinkComponents(root, relativePath, label);
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
  return { lexical, source: readFileSync(lexical, "utf8") };
}

function staticString(node, ts, bindings = new Map(), seen = new Set()) {
  while (
    ts.isAsExpression(node) ||
    ts.isTypeAssertionExpression(node) ||
    ts.isParenthesizedExpression(node)
  ) {
    node = node.expression;
  }
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) {
    return node.text;
  }
  if (ts.isIdentifier(node) && bindings.has(node.text) && !seen.has(node.text)) {
    return staticString(
      bindings.get(node.text),
      ts,
      bindings,
      new Set([...seen, node.text]),
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
    ts.isTypeAssertionExpression(node) ||
    ts.isParenthesizedExpression(node)
  ) {
    node = node.expression;
  }
  if (ts.isIdentifier(node) && bindings.has(node.text) && !seen.has(node.text)) {
    return staticValue(
      bindings.get(node.text),
      ts,
      bindings,
      new Set([...seen, node.text]),
    );
  }
  const string = staticString(node, ts, bindings, seen);
  if (string !== null) return string;
  if (ts.isNumericLiteral(node)) return Number(node.text);
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
  let result = template.replace(/%([sdifjo])/gu, (_match, token) => {
    const value = values[cursor];
    cursor += 1;
    if (token === "j") return JSON.stringify(value);
    if (["d", "i", "f"].includes(token)) return String(Number(value));
    if (token === "o") return JSON.stringify(value);
    return String(value);
  });
  result = result.replace(/%#/gu, String(index));
  result = result.replace(/%\$/gu, String(index + 1));
  if (row && typeof row === "object" && !Array.isArray(row)) {
    result = result.replace(/\$([A-Za-z_][A-Za-z0-9_]*)/gu, (match, key) =>
      Object.hasOwn(row, key)
        ? typeof row[key] === "string"
          ? `'${row[key].replaceAll("'", "\\'")}'`
          : String(row[key])
        : match,
    );
  }
  return result;
}

function expandedEachTitles(expression, template, ts, bindings) {
  if (!ts.isCallExpression(expression)) return null;
  const name = calleeName(expression.expression, ts);
  if (!name.endsWith(".each") || expression.arguments.length !== 1) return null;
  let dataset = expression.arguments[0];
  while (
    ts.isAsExpression(dataset) ||
    ts.isTypeAssertionExpression(dataset) ||
    ts.isParenthesizedExpression(dataset)
  ) dataset = dataset.expression;
  if (!ts.isArrayLiteralExpression(dataset)) return null;
  const rows = dataset.elements.map((element) => staticValue(element, ts, bindings));
  if (rows.includes(undefined)) return null;
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
  if (ts.isCallExpression(expression)) {
    return calleeName(expression.expression, ts);
  }
  return "";
}

function registrationSuiteSources(root, file, sourceFile, ts) {
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
        /^register[A-Za-z0-9]+Suite$/u.test(element.name.text),
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
    const candidate = candidates.find((entry) => existsSync(path.resolve(root, entry)));
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
    );
    suiteSources.push({ file: candidate, source: contained.source });
  }
  return suiteSources;
}

export function registrationSuiteFiles(root, file, source) {
  const requireFromWeb = createRequire(path.join(root, "apps", "web", "package.json"));
  const ts = requireFromWeb("typescript");
  const scriptKind = file.endsWith("x") ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
  const sourceFile = ts.createSourceFile(
    file,
    source,
    ts.ScriptTarget.Latest,
    true,
    scriptKind,
  );
  return registrationSuiteSources(root, file, sourceFile, ts).map(
    (suite) => suite.file,
  );
}

function testTitles(root, file, source) {
  const cacheKey = `${root}\0${file}\0${source}`;
  if (sourceTitleCache.has(cacheKey)) {
    return sourceTitleCache.get(cacheKey);
  }
  const requireFromWeb = createRequire(path.join(root, "apps", "web", "package.json"));
  const ts = requireFromWeb("typescript");
  const scriptKind = file.endsWith("x") ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
  const sourceFile = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true, scriptKind);
  const bindings = new Map();
  function collectBindings(node) {
    if (
      ts.isVariableDeclaration(node) &&
      ts.isIdentifier(node.name) &&
      node.initializer
    ) {
      bindings.set(node.name.text, node.initializer);
    }
    ts.forEachChild(node, collectBindings);
  }
  collectBindings(sourceFile);
  const matches = new Map();
  let testIdentity = 0;

  function record(candidate, identity) {
    if (!matches.has(candidate)) matches.set(candidate, new Set());
    matches.get(candidate).add(identity);
  }

  function visit(node, ancestors) {
    if (ts.isCallExpression(node)) {
      const name = calleeName(node.expression, ts);
      const base = name.split(".")[0];
      const segments = name.split(".");
      const title = node.arguments[0]
        ? staticString(node.arguments[0], ts, bindings)
        : null;
      if (segments.some((segment) => ["describe", "suite"].includes(segment)) && title !== null) {
        const callback = [...node.arguments].reverse().find(
          (argument) => ts.isArrowFunction(argument) || ts.isFunctionExpression(argument),
        );
        if (callback) {
          visit(callback.body, [...ancestors, title]);
          return;
        }
      }
      if (
        ["it", "test"].includes(base) &&
        !segments.some((segment) => ["describe", "suite"].includes(segment)) &&
        title !== null
      ) {
        const expanded = expandedEachTitles(node.expression, title, ts, bindings) ?? [title];
        for (const expandedTitle of expanded) {
          const fullTitle = [...ancestors, expandedTitle].join(" ");
          const identity = testIdentity;
          testIdentity += 1;
          record(expandedTitle, identity);
          record(fullTitle, identity);
        }
        return;
      }
    }
    ts.forEachChild(node, (child) => visit(child, ancestors));
  }

  visit(sourceFile, []);
  const counts = new Map(
    [...matches].map(([candidate, identities]) => [candidate, identities.size]),
  );
  for (const suite of registrationSuiteSources(root, file, sourceFile, ts)) {
    for (const [title, count] of testTitles(root, suite.file, suite.source)) {
      counts.set(title, (counts.get(title) ?? 0) + count);
    }
  }
  sourceTitleCache.set(cacheKey, counts);
  return counts;
}

export function collectSourceTestTitleCounts({ root, file, approvedRoots }) {
  const { source } = containedFile(root, file, approvedRoots, `${file}.title_discovery`);
  return new Map(testTitles(root, file, source));
}

function packageDirectory(root, packagePath, approvedRoots, label) {
  if (!packagePath.startsWith("./") || /[*?{}[\]\\]/u.test(packagePath)) {
    throw new Error(`${label} must be an exact repository package`);
  }
  const relative = packagePath.slice(2);
  if (path.posix.normalize(relative) !== relative || relative.split("/").includes("..")) {
    throw new Error(`${label} contains traversal or normalization drift`);
  }
  const lexical = path.resolve(root, relative);
  if (!existsSync(lexical)) {
    throw new Error(`${label} does not exist: ${packagePath}`);
  }
  assertNoSymlinkComponents(root, relative, label);
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

function goRunnerBuildContext(root) {
  const cacheKey = realpathSync(root);
  if (goBuildContextCache.has(cacheKey)) return goBuildContextCache.get(cacheKey);
  const pins = JSON.parse(readFileSync(path.join(root, "tools/toolchain_pins.json"), "utf8"));
  const version = /^(\d+)\.(\d+)(?:\.(\d+))?$/u.exec(
    String(pins.go_version ?? ""),
  );
  if (!version || version[1] !== "1") {
    throw new Error(
      "tools/toolchain_pins.json.go_version must be a Go 1.x or Go 1.x.patch release",
    );
  }
  const releaseMinor = Number(version[2]);
  const tags = new Set(["amd64", "amd64.v1", "gc", "linux", "unix"]);
  for (let minor = 1; minor <= releaseMinor; minor += 1) tags.add(`go1.${minor}`);
  const context = Object.freeze({
    goarch: "amd64",
    goos: "linux",
    label: `linux/amd64 Go ${pins.go_version} with no custom build tags`,
    tags,
  });
  goBuildContextCache.set(cacheKey, context);
  return context;
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

function goSymbols(root, directory) {
  const context = goRunnerBuildContext(root);
  const activeCounts = new Map();
  const excludedFiles = new Map();
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    if (!entry.isFile() || !entry.name.endsWith("_test.go")) {
      continue;
    }
    const source = readFileSync(path.join(directory, entry.name), "utf8");
    const active = fileNameMatchesGoBuildContext(entry.name, context) &&
      sourceMatchesGoBuildContext(source, context, entry.name);
    for (const match of source.matchAll(/^func\s+(Test[A-Za-z0-9_]+)\s*\(/gmu)) {
      if (active) activeCounts.set(match[1], (activeCounts.get(match[1]) ?? 0) + 1);
      else {
        const files = excludedFiles.get(match[1]) ?? [];
        files.push(entry.name);
        excludedFiles.set(match[1], files);
      }
    }
  }
  return { activeCounts, context, excludedFiles };
}

export function resolveRowSelector({ root, row, runner, taskSurfaceCommandIDs }) {
  const label = `${row.row_id}.selector`;
  if (row.runner === "go") {
    const directory = packageDirectory(root, row.selector.package, runner.approved_roots, `${label}.package`);
    const symbols = goSymbols(root, directory);
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
    const { source } = containedFile(root, row.selector.file, runner.approved_roots, `${label}.file`);
    const titles = testTitles(root, row.selector.file, source);
    for (const title of row.selector.titles) {
      if ((titles.get(title) ?? 0) !== 1) {
        throw new Error(`${label}.titles ${JSON.stringify(title)} must resolve exactly once`);
      }
    }
    if (row.runner === "playwright") {
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
      );
    }
    return row.selector.titles.map((title) => `vitest:${row.selector.file}:${title}`);
  }
  if (row.runner === "shell") {
    if (!taskSurfaceCommandIDs.has(row.selector.command_id)) {
      throw new Error(`${label}.command_id is not registered`);
    }
    return [`shell:${row.selector.command_id}`];
  }
  throw new Error(`${row.row_id}.runner is unsupported`);
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
