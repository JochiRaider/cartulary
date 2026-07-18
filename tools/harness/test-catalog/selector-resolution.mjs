import { createRequire } from "node:module";
import { existsSync, lstatSync, readFileSync, readdirSync, realpathSync } from "node:fs";
import path from "node:path";

const sourceTitleCache = new Map();

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

function staticString(node, ts) {
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
  if (
    ts.isBinaryExpression(node) &&
    node.operatorToken.kind === ts.SyntaxKind.PlusToken
  ) {
    const left = staticString(node.left, ts);
    const right = staticString(node.right, ts);
    return left === null || right === null ? null : left + right;
  }
  return null;
}

function staticValue(node, ts) {
  while (
    ts.isAsExpression(node) ||
    ts.isTypeAssertionExpression(node) ||
    ts.isParenthesizedExpression(node)
  ) {
    node = node.expression;
  }
  const string = staticString(node, ts);
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
    return node.elements.map((element) => staticValue(element, ts));
  }
  if (ts.isObjectLiteralExpression(node)) {
    const result = {};
    for (const property of node.properties) {
      if (!ts.isPropertyAssignment(property)) return undefined;
      const key = property.name && ts.isIdentifier(property.name)
        ? property.name.text
        : property.name && staticString(property.name, ts);
      if (key === null) return undefined;
      const value = staticValue(property.initializer, ts);
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

function expandedEachTitles(expression, template, ts) {
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
  const rows = dataset.elements.map((element) => staticValue(element, ts));
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

function testTitles(root, file, source) {
  const cacheKey = `${root}\0${file}\0${source}`;
  if (sourceTitleCache.has(cacheKey)) {
    return sourceTitleCache.get(cacheKey);
  }
  const requireFromWeb = createRequire(path.join(root, "apps", "web", "package.json"));
  const ts = requireFromWeb("typescript");
  const scriptKind = file.endsWith("x") ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
  const sourceFile = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true, scriptKind);
  const counts = new Map();

  function visit(node, ancestors) {
    if (ts.isCallExpression(node)) {
      const name = calleeName(node.expression, ts);
      const base = name.split(".")[0];
      const title = node.arguments[0] ? staticString(node.arguments[0], ts) : null;
      if (["describe", "suite"].includes(base) && title !== null) {
        const callback = [...node.arguments].reverse().find(
          (argument) => ts.isArrowFunction(argument) || ts.isFunctionExpression(argument),
        );
        if (callback) {
          visit(callback.body, [...ancestors, title]);
          return;
        }
      }
      if (["it", "test"].includes(base) && title !== null) {
        const expanded = expandedEachTitles(node.expression, title, ts) ?? [title];
        for (const expandedTitle of expanded) {
          const fullTitle = [...ancestors, expandedTitle].join(" ");
          counts.set(fullTitle, (counts.get(fullTitle) ?? 0) + 1);
        }
        return;
      }
    }
    ts.forEachChild(node, (child) => visit(child, ancestors));
  }

  visit(sourceFile, []);
  sourceTitleCache.set(cacheKey, counts);
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

function goSymbols(directory) {
  const counts = new Map();
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    if (!entry.isFile() || !entry.name.endsWith("_test.go")) {
      continue;
    }
    const source = readFileSync(path.join(directory, entry.name), "utf8");
    for (const match of source.matchAll(/^func\s+(Test[A-Za-z0-9_]+)\s*\(/gmu)) {
      counts.set(match[1], (counts.get(match[1]) ?? 0) + 1);
    }
  }
  return counts;
}

export function resolveRowSelector({ root, row, runner, taskSurfaceCommandIDs }) {
  const label = `${row.row_id}.selector`;
  if (row.runner === "go") {
    const directory = packageDirectory(root, row.selector.package, runner.approved_roots, `${label}.package`);
    const symbols = goSymbols(directory);
    for (const symbol of row.selector.tests) {
      if ((symbols.get(symbol) ?? 0) !== 1) {
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
