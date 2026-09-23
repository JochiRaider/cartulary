// SQL relation classification for repository ownership checks. This is a lexical
// query walker, not a query executor: unsupported relation syntax fails closed.
const queryStarts = new Set(['select', 'with', 'insert', 'update', 'delete', 'values', 'table']);
const reserved = new Set(('where group order having limit offset returning union except intersect window for join left right full inner outer cross natural on using set values conflict do into from as tablesample repeatable').split(' '));

function sqlSegments(source) {
  if (!/^\s*package\s+\w+/mu.test(source)) return [source];
  // Read Go string expressions, including constant composition. Scanning each
  // literal independently loses both balanced query structure and relation names.
  const lexemes = [...source.matchAll(/\/\/[^\n]*|\/\*[\s\S]*?\*\/|`[^`]*`|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|[a-z_][a-z_0-9]*|[^\s]/giu)]
    .map(([text]) => text).filter((text) => !text.startsWith('//') && !text.startsWith('/*'));
  const isLiteral = (text) => text?.startsWith('`') || text?.startsWith('"');
  const decode = (text) => text.startsWith('`') ? text.slice(1, -1)
    : text.slice(1, -1).replace(/\\(U[0-9a-f]{8}|u[0-9a-f]{4}|x[0-9a-f]{2}|[0-7]{3}|.)/gu, (_, escape) => {
      const simple = { a: '\x07', b: '\b', f: '\f', n: '\n', r: '\r', t: '\t', v: '\v', '\\': '\\', '"': '"', "'": "'" };
      if (Object.hasOwn(simple, escape)) return simple[escape];
      if (/^[Uux]/u.test(escape)) return String.fromCodePoint(Number.parseInt(escape.slice(1), 16));
      if (/^[0-7]{3}$/u.test(escape)) return String.fromCodePoint(Number.parseInt(escape, 8));
      throw new Error('unsupported Go string escape');
    });
  const definitions = new Map();
  for (let i = 0; i < lexemes.length; i++) {
    if (lexemes[i] === 'const' && lexemes[i + 2] === '=' && isLiteral(lexemes[i + 3])) {
      definitions.set(lexemes[i + 1], i + 3);
    }
  }
  function expression(start, seen = new Set()) {
    let text = ''; let end = start;
    do {
      const part = lexemes[end++];
      if (isLiteral(part)) text += decode(part);
      else if (definitions.has(part) && !seen.has(part)) {
        text += expression(definitions.get(part), new Set([...seen, part])).text;
      } else {
        text += '%s';
        while (lexemes[end] === '.' || lexemes[end] === '(' || lexemes[end] === '[') {
          if (lexemes[end] === '.') { end += 2; continue; }
          const opening = lexemes[end++]; const closing = opening === '(' ? ')' : ']';
          let depth = 1;
          while (end < lexemes.length && depth) {
            if (lexemes[end] === opening) depth++;
            if (lexemes[end] === closing) depth--;
            end++;
          }
          if (depth) throw new Error('unclosed Go SQL expression');
        }
      }
      if (lexemes[end] !== '+') break;
      end++;
    } while (end < lexemes.length);
    return { text, end };
  }
  const result = [];
  for (let i = 0; i < lexemes.length; i++) {
    if (!isLiteral(lexemes[i])) continue;
    const parsed = expression(i); result.push(parsed.text); i = parsed.end - 1;
  }
  return result.filter((text) => /^\s*(?:(?:--[^\n]*\n|\/\*[\s\S]*?\*\/)\s*)*(?:with\b|select\b|merge\s+into\b|table\s+|insert\s+into\b|update\s+\S+(?:\s+\S+)?\s+set\b|delete\s+from\b|from\b|join\b|left\s+join\b|right\s+join\b|cross\s+join\b|lock\s+table\s+|truncate\s+)/iu.test(text));
}

function tokenize(source) {
  const tokens = [];
  let i = 0;
  while (i < source.length) {
    const rest = source.slice(i);
    if (/^\s/u.test(rest)) { i++; continue; }
    if (rest.startsWith('--')) { const end = source.indexOf('\n', i); i = end < 0 ? source.length : end; continue; }
    if (rest.startsWith('/*')) {
      let depth = 1; i += 2;
      while (i < source.length && depth) {
        if (source.startsWith('/*', i)) { depth++; i += 2; }
        else if (source.startsWith('*/', i)) { depth--; i += 2; }
        else i++;
      }
      if (depth) throw new Error('unterminated SQL comment');
      continue;
    }
    const dollar = rest.match(/^\$(?:[a-z_][a-z_0-9]*)?\$/iu)?.[0];
    if (dollar) {
      const end = source.indexOf(dollar, i + dollar.length);
      if (end < 0) throw new Error('unterminated SQL dollar string');
      tokens.push({ kind: 'literal', value: '' }); i = end + dollar.length; continue;
    }
    if (source[i] === "'" || source[i] === '"') {
      const quote = source[i++]; let value = ''; let closed = false;
      while (i < source.length) {
        if (source[i] === quote) {
          if (source[i + 1] === quote) { value += quote; i += 2; continue; }
          i++; closed = true; break;
        }
        if (quote === "'" && source[i] === '\\') { i += 2; continue; }
        value += source[i++];
      }
      if (!closed) throw new Error('unterminated SQL quoted token');
      tokens.push({ kind: quote === '"' ? 'identifier' : 'literal', value }); continue;
    }
    const word = rest.match(/^[a-z_][a-z_0-9$]*/iu)?.[0];
    if (word) { tokens.push({ kind: 'word', value: word.toLowerCase() }); i += word.length; continue; }
    tokens.push({ kind: 'symbol', value: source[i++] });
  }
  return tokens;
}

function classify(source) {
  const tokens = tokenize(source);
  const pairs = new Map(); const stack = [];
  for (let i = 0; i < tokens.length; i++) {
    if (tokens[i].value === '(') stack.push(i);
    if (tokens[i].value === ')') {
      const start = stack.pop(); if (start === undefined) throw new Error('unmatched SQL parenthesis');
      pairs.set(start, i);
    }
  }
  if (stack.length) throw new Error('unmatched SQL parenthesis');
  const accesses = [];
  const is = (i, word) => tokens[i]?.kind === 'word' && tokens[i].value === word;
  const identifier = (i) => ['word', 'identifier'].includes(tokens[i]?.kind);
  const value = (i) => tokens[i]?.value;
  const keyword = (i) => tokens[i]?.kind === 'word' && reserved.has(value(i));

  function nested(start, end, ctes) {
    if (queryStarts.has(value(start)) && tokens[start]?.kind === 'word') {
      query(start, end, ctes); return;
    }
    for (let i = start; i < end; i++) {
      if (value(i) === '(') { const close = pairs.get(i); nested(i + 1, close, ctes); i = close; }
    }
  }

  function query(start, end, inherited) {
    const ctes = new Set(inherited);
    let i = start;
    if (is(i, 'with')) {
      i++; const recursive = is(i, 'recursive'); if (recursive) i++;
      const definitions = [];
      while (i < end) {
        if (!identifier(i)) throw new Error('unclassified CTE name');
        const name = value(i++);
        if (value(i) === '(') i = pairs.get(i) + 1;
        if (!is(i++, 'as')) throw new Error('unclassified CTE binding');
        if (is(i, 'not')) i++;
        if (is(i, 'materialized')) i++;
        if (value(i) !== '(') throw new Error('unclassified CTE body');
        const close = pairs.get(i); definitions.push({ name, start: i + 1, end: close }); i = close + 1;
        if (value(i) !== ',') break;
        i++;
      }
      if (recursive) for (const def of definitions) ctes.add(def.name);
      for (const def of definitions) { query(def.start, def.end, ctes); ctes.add(def.name); }
    }
    const relations = [];
    const locks = [];
    let inFrom = false;
    const command = value(i);
    if (i >= end) return;
    if (!['select', 'insert', 'update', 'delete', 'values', 'table', 'from', 'join', 'left', 'right', 'cross', 'lock', 'truncate'].includes(command)) {
      throw new Error(`unsupported SQL statement ${command}`);
    }

    function relation(at, operation) {
      let j = at;
      if (is(j, 'lateral')) j++;
      if (is(j, 'only')) j++;
      let tables = [];
      let defaultAlias = null;
      if (value(j) === '%' && value(j + 1) === 's') {
        accesses.push({ table: '<dynamic>', operation, dynamic: true });
        tables = ['<dynamic>']; j += 2;
      } else if (value(j) === '(') {
        if (!queryStarts.has(value(j + 1))) throw new Error('unsupported parenthesized SQL relation');
        const close = pairs.get(j); const before = accesses.length;
        nested(j + 1, close, ctes);
        tables = accesses.slice(before).filter((a) => a.operation === 'read').map((a) => a.table);
        j = close + 1;
      } else {
        if (!identifier(j) || keyword(j)) throw new Error(`unclassified SQL relation after ${value(at - 1)} (${tokens.slice(j, j + 5).map((token) => token.value).join(" ")})`);
        const parts = [value(j++)];
        while (value(j) === '.') {
          if (!identifier(j + 1)) throw new Error('unclassified qualified SQL relation');
          parts.push(value(j + 1)); j += 2;
        }
        defaultAlias = parts.at(-1);
        if (value(j) === '(' && operation === 'read') {
          const close = pairs.get(j); nested(j + 1, close, ctes); j = close + 1;
        } else if (!(parts.length === 1 && ctes.has(parts[0]) && operation === 'read')) {
          const table = parts.length === 2 && parts[0] === 'public' ? parts[1] : parts.join('.');
          accesses.push({ table, operation }); tables = [table];
        }
      }
      if (value(j) === '*') j++;
      if (is(j, 'as')) j++;
      const alias = identifier(j) && !keyword(j) ? value(j++) : defaultAlias;
      relations.push({ tables, alias });
      return j;
    }

    for (; i < end; i++) {
      if (value(i) === ';') { query(i + 1, end, inherited); break; }
      if (value(i) === '(') { const close = pairs.get(i); nested(i + 1, close, ctes); i = close; continue; }
      if (is(i, 'for')) {
        let j = i + 1;
        if (is(j, 'no')) j++;
        if (is(j, 'key')) j++;
        if (is(j, 'update') || is(j, 'share')) {
          j++; const aliases = [];
          if (is(j, 'of')) {
            j++;
            while (identifier(j) && !is(j, 'nowait') && !is(j, 'skip')) {
              aliases.push(value(j++)); if (value(j) !== ',') break; j++;
            }
            if (!aliases.length) throw new Error('unclassified SQL lock target');
          }
          locks.push(aliases); i = j - 1; inFrom = false; continue;
        }
      }
      if ((i === start && is(i, 'table')) || (command === 'select' && is(i, 'into'))) { i = relation(i + 1, command === 'select' ? 'write' : 'read') - 1; continue; }
      if (is(i, 'insert') && is(i + 1, 'into')) { i = relation(i + 2, 'write') - 1; continue; }
      if (is(i, 'delete') && is(i + 1, 'from')) { i = relation(i + 2, 'write') - 1; continue; }
      if (is(i, 'update') && !is(i - 1, 'do')) { i = relation(i + 1, 'write') - 1; continue; }
      if (is(i, 'lock') && is(i + 1, 'table')) { i = relation(i + 2, 'lock') - 1; inFrom = true; continue; }
      if (is(i, 'truncate')) { i = relation(i + (is(i + 1, 'table') ? 2 : 1), 'write') - 1; inFrom = true; continue; }
      if ((is(i, 'from') && !is(i - 1, 'distinct')) || is(i, 'join') || (command === 'delete' && is(i, 'using'))) {
        inFrom = true; i = relation(i + 1, 'read') - 1; continue;
      }
      if (tokens[i]?.kind === 'word' && ['where','group','order','having','limit','offset','returning','union','except','intersect','window','set','values'].includes(value(i))) inFrom = false;
      if (inFrom && value(i) === ',') i = relation(i + 1, command === 'truncate' ? 'write' : command === 'lock' ? 'lock' : 'read') - 1;
    }
    for (const aliases of locks) {
      for (const alias of aliases) if (!relations.some((r) => r.alias === alias)) throw new Error('unresolved SQL lock alias');
      for (const r of relations.filter((r) => !aliases.length || aliases.includes(r.alias))) {
        for (const table of r.tables) accesses.push({ table, operation: 'lock' });
      }
    }
  }
  query(0, tokens.length, new Set());
  return accesses;
}

export function sqlRelationAccesses(source) {
  return sqlSegments(source).flatMap((segment) => {
    try { return classify(segment); }
    catch (error) { throw new Error(`${error.message}; SQL: ${segment.slice(0, 500)}`, { cause: error }); }
  });
}
