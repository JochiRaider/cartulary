import { inspectGoSQL } from "./go-import-inspection.mjs";
// SQL relation classification for repository ownership checks. This is a lexical
// query walker, not a query executor: unsupported relation syntax fails closed.
const queryStarts = new Set(['select', 'with', 'insert', 'update', 'delete', 'values', 'table']);
const reserved = new Set(('select with table insert update delete lock truncate all distinct where group order having limit offset fetch returning union except intersect window for join left right full inner outer cross natural on using set values conflict do into from as tablesample repeatable').split(' '));

// Both consumers receive the same structured Go analysis. Unknown operands are
// errors even when the visible prefix already looks like a complete statement.
export function analyzedSQLStatements(declarations) {
  return declarations.flatMap(declaration => declaration.sql.flatMap(expression => {
    if (expression.kind === "statement" && expression.text && !isSQLSegment(expression.text)) return [];
    if (expression.unresolved.length) throw new Error(`${declaration.name}:${expression.line}:${expression.column}: unresolved SQL construction: ${expression.unresolved.join(", ")}`);
    if (expression.kind === "closed" || expression.kind === "empty") return [];
    if (expression.kind === "copy") return sqlStatements(`INSERT INTO ${expression.text}`).map(statement=>({...statement,declaration:declaration.name}));
    if (expression.kind === "fragment") return fragmentStatements(expression.text).map(statement=>({...statement,declaration:declaration.name}));
    if (expression.kind === "command") {
      const statements=sqlStatements(expression.text);
      if (statements.some(statement=>statement.accesses.some(access=>access.dynamic))) throw new Error('closed commands require literal relation names');
      return statements.map(statement=>({...statement,declaration:declaration.name}));
    }
    if (!expression.text) throw new Error(`${declaration.name}: empty SQL expression`);
    if (expression.kind === "statement" && !isSQLSegment(expression.text)) return [];
    return sqlStatements(expression.text).map(statement=>({...statement,dynamic:statement.dynamic||expression.dynamic,declaration:declaration.name}));
  }));
}

// Closed fragments are separated at runtime, so tokens cannot be manufactured
// by joining partial keywords. Relation clauses must be complete within a
// template. The compiler capability is read-only; locks never inherit it.
function fragmentStatements(text) {
  const tokens = tokenize(text);
  const words = tokens.filter((token,index)=>token.kind==="word" && tokens[index-1]?.value!==".");
  // Reject operation keywords even when split across separate fragments.
  // Read templates cannot introduce SELECT INTO, COPY or indirect commands.
  if (words.some(token=>["update","insert","delete","merge","lock","truncate","share","into","copy","create","alter","drop","execute","do","call","grant","revoke","prepare","deallocate","set","reset","vacuum","analyze","refresh","reindex","cluster","comment","discard","listen","notify","unlisten","load"].includes(token.value))) throw new Error("closed query fragments cannot write, lock or execute commands");
  if (!words.some(token=>["from","join","table"].includes(token.value))) return [];
  const source = isSQLSegment(text) || (tokens[0]?.kind === "word" && tokens[0].value === "table") ? text : `SELECT ${text}`;
  const statements = sqlStatements(source);
  if (statements.some(statement=>statement.accesses.some(access=>access.dynamic))) throw new Error("closed query fragments require literal relation names");
  return statements;
}

function isSQLSegment(text) {
  // Recognize statement shapes with the same comment/identifier lexer used by
  // relation classification. Formatting between SQL keywords cannot hide access.
  if (!/^\s*(?:--|\/\*|(?:with|select|merge|table|insert|update|delete|from|join|left|right|cross|lock|truncate)\b)/iu.test(text)) return false;
  const tokens = tokenize(text, true);
  const is = (index, word) => tokens[index]?.kind === "word" && tokens[index].value === word;
  if (is(0, "update")) return tokens.some((token) => token.kind === "word" && token.value === "set");
  if (is(0, "insert") || is(0, "merge")) return is(1, "into");
  if (is(0, "delete")) return is(1, "from");
  if (is(0, "lock")) return is(1, "table") && tokens.length > 2;
  if (is(0, "table") || is(0, "truncate")) return tokens.length > 1;
  if (is(0, "left") || is(0, "right") || is(0, "cross")) return is(1, "join");
  return tokens[0]?.kind === "word" && ["with", "select", "from", "join"].includes(tokens[0].value);
}

function tokenize(source, allowCommentOnly = false) {
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
      // Go code also contains comment delimiters as ordinary string data.
      // Before the first SQL token, an unfinished comment is comment-only text.
      if (depth && allowCommentOnly && tokens.length === 0) return [];
      if (depth) throw new Error('unterminated SQL comment');
      continue;
    }
    const dollar = rest.match(/^\$(?:[a-z_][a-z_0-9]*)?\$/iu)?.[0];
    if (dollar) {
      const end = source.indexOf(dollar, i + dollar.length);
      if (end < 0) throw new Error('unterminated SQL dollar string');
      tokens.push({ kind: 'literal', value: source.slice(i, end + dollar.length) }); i = end + dollar.length; continue;
    }
    if (source[i] === "'" || source[i] === '"') {
      const start = i; const escape = source[i] === "\'" && /(?:^|[^a-z_0-9$])e$/iu.test(source.slice(0,i)); const quote = source[i++]; let value = ''; let closed = false;
      while (i < source.length) {
        if (source[i] === quote) {
          if (source[i + 1] === quote) { value += quote; i += 2; continue; }
          i++; closed = true; break;
        }
        if (escape && source[i] === '\\') { i += 2; continue; }
        value += source[i++];
      }
      if (!closed) throw new Error('unterminated SQL quoted token');
      tokens.push({ kind: quote === '"' ? 'identifier' : 'literal', value: quote === '"' ? value : source.slice(start, i) }); continue;
    }
    const word = rest.match(/^[a-z_][a-z_0-9$]*/iu)?.[0];
    if (word) { tokens.push({ kind: 'word', value: word.toLowerCase() }); i += word.length; continue; }
    const parameter = rest.match(/^\$[0-9]+/u)?.[0];
    if (parameter) { tokens.push({ kind: 'parameter', value: parameter }); i += parameter.length; continue; }
    const number = rest.match(/^(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:e[+-]?[0-9]+)?/iu)?.[0];
    if (number) { tokens.push({ kind: 'number', value: number }); i += number.length; continue; }
    const operator = rest.match(/^(?:::|->>|->|#>>|#>|<=|>=|<>|!=|\|\||&&|@>|<@|:=)/u)?.[0];
    if (operator) { tokens.push({ kind: 'symbol', value: operator }); i += operator.length; continue; }
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
  const controls = [];
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
    // Statement and parenthesis boundaries are resolved before CTE bindings or
    // operands, so neither aliases nor locks can leak into a sibling statement.
    if (start === end) return;
    for (let at=start;at<end;at++) {
      if(value(at)==='(') {at=pairs.get(at);continue;}
      if(value(at)===';') {query(start,at,inherited);query(at+1,end,inherited);return;}
    }
    if(value(start)==='(' && pairs.get(start)===end-1) {query(start+1,end-1,inherited);return;}
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
    if(i>=end) throw new Error('missing SQL query operand');
    const commandStart=i;
    const command=value(i);
    const setOperands=[];
    let operandStart=i;
    for(let at=i;at<end;at++) {
      if(value(at)==='(') {at=pairs.get(at);continue;}
      if(!['union','intersect','except'].some(word=>is(at,word))) continue;
      // INSERT has its own target and delegates its source query below.
      if(command==='insert') break;
      if(!['select','values','table'].includes(command) && command!=='(') throw new Error('unsupported SQL set operand');
      if(at===operandStart) throw new Error('missing SQL set operand');
      setOperands.push([operandStart,at]);
      at++;
      if(is(at,'all') || is(at,'distinct')) at++;
      operandStart=at;
      if(at>=end) throw new Error('missing SQL set operand');
      at--;
    }
    if(setOperands.length) {
      setOperands.push([operandStart,end]);
      for(const [begin,stop] of setOperands) {
        if(!['select','values','table'].some(word=>is(begin,word)) && value(begin)!=='(') throw new Error('unsupported SQL set operand');
        query(begin,stop,ctes);
      }
      return;
    }
    // A parenthesized query can have an outer ordering/limit clause. Its relation
    // scope is kept inside the parentheses; nested expressions still get visited.
    if(value(i)==='(') {
      const close=pairs.get(i);
      query(i+1,close,ctes);
      if(close+1<end && !['order','limit','offset','fetch'].some(word=>is(close+1,word))) throw new Error('unsupported parenthesized query suffix');
      nested(close+1,end,ctes);
      return;
    }
    const relations = [];
    const locks = [];
    let inFrom = false;
    if (is(i,'set') || is(i,'listen') || is(i,'unlisten')) {
      let stop=i; while(stop<end && value(stop)!==';') stop++;
      const shape=tokens.slice(i,stop).map(token=>token.value).join(' ');
      if(!/^(?:set (?:role [a-z_][a-z_0-9]*|local (?:row_security = (?:off|on)|session_replication_role = (?:replica|origin)))|(?:listen|unlisten) [a-z_][a-z_0-9]*)$/.test(shape)) throw new Error('unsupported SQL session command');
      controls.push(shape);
      if(stop<end) query(stop+1,end,inherited);
      return;
    }

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
      if(command==='insert' && i>commandStart && ['select','values','table','with'].some(word=>is(i,word))) {
        let stop=i;
        for(;stop<end;stop++) {
          if(value(stop)==='(') {stop=pairs.get(stop);continue;}
          if(is(stop,'returning') || is(stop,'on') && is(stop+1,'conflict')) break;
        }
        query(i,stop,ctes);i=stop-1;continue;
      }
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
      if ((i === commandStart && is(i, 'table')) || (command === 'select' && is(i, 'into'))) { i = relation(i + 1, command === 'select' ? 'write' : 'read') - 1; continue; }
      if (is(i, 'insert') && is(i + 1, 'into')) { i = relation(i + 2, 'write') - 1; continue; }
      if (is(i, 'delete') && is(i + 1, 'from')) { i = relation(i + 2, 'write') - 1; continue; }
      if (is(i, 'update') && !is(i - 1, 'do')) { i = relation(i + 1, 'write') - 1; continue; }
      if (is(i, 'lock') && is(i + 1, 'table')) { i = relation(i + 2, 'lock') - 1; inFrom = true; continue; }
      if (is(i, 'truncate')) { i = relation(i + (is(i + 1, 'table') ? 2 : 1), 'write') - 1; inFrom = true; continue; }
      if(is(i,'table')) throw new Error('unclassified TABLE query operand');
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
  return {accesses,controls};
}

// Keys preserve SQL semantics; only whitespace, comments and unquoted case fold.
// Statement separators remain tokens, so adding another statement changes the key.
export function sqlStatements(source, { goConstants = "" } = {}) {
  if (/^\s*package\s+\w+/mu.test(source)) {
    const files=[{relative:"inspection.go",content:source}];
    if (goConstants) files.push({relative:"constants.go",content:`package ${source.match(/^\s*package\s+(\w+)/mu)[1]}\n${goConstants}`});
    return analyzedSQLStatements(inspectGoSQL([files]).get("inspection.go"));
  }
  return [{text:source,dynamic:false}].map(({ text, dynamic }) => {
    try {
      const tokens = tokenize(text);
      return {
        key: JSON.stringify(tokens),
        dynamic: dynamic || tokens.some((token, index) => token.value === '%' && tokens[index + 1]?.value === 's'),
        ...classify(text),
      };
    } catch (error) { throw new Error(`${error.message}; SQL: ${text.slice(0, 500)}`, { cause: error }); }
  });
}

export function sqlRelationAccesses(source) {
  return sqlStatements(source).flatMap((statement) => statement.accesses);
}
