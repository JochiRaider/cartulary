import { catalogDDLSource } from "./catalog-ddl-source.mjs";

const identifier = "[a-z_][a-z0-9_$]*";
const relation = `${identifier}(?:\\.${identifier})?`;
const qualify = (name) => name.includes(".") ? name : `public.${name}`;

// This is the bounded key-bearing subset of authored DDL, not a SQL executor.
// Unknown structural syntax fails closed. PostgreSQL independently witnesses
// the resulting catalog; generation never treats these source facts as live facts.
export function collectForeignKeyCoverage(migrations) {
  const foreignKeys = new Map();
  const indexes = new Map();
  for (const { filename, source } of migrations) {
    const up = source.split(/^-- \+goose Down\s*$/mu, 1)[0];
    const ddl = catalogDDLSource(up).toLowerCase();
    for (const raw of ddl.split(";")) {
      const statement = raw.trim().replace(/\s+/gu, " ");
      if (!statement) continue;
      const fail = (detail) => { throw new Error(`${filename}: unsupported key-bearing DDL: ${detail}`); };
      let match;
      if ((match = new RegExp(`^create table (${relation})\\s*\\(`, "u").exec(statement))) {
        const table = qualify(match[1]);
        const [body, tail] = parenthesized(statement.slice(match[0].length - 1), fail);
        if (tail.trim()) fail(`table ${table} suffix`);
        for (const element of splitKeys(body, fail)) collectConstraints(table, element, fail);
      } else if (new RegExp(`^create table ${relation} as (?:select|with)\\b`, "u").test(statement)) {
        // CTAS copies values, not source constraints or indexes.
      } else if ((match = new RegExp(`^alter table (?:only )?(${relation})\\s+([\\s\\S]+)$`, "u").exec(statement))) {
        const table = qualify(match[1]);
        for (const action of splitKeys(match[2], fail)) {
          const drop = new RegExp(`^drop constraint (${identifier})$`, "u").exec(action);
          if (drop) {
            const name = `${table}.${drop[1]}`;
            foreignKeys.delete(name);
            for (const [key, index] of indexes) if (index.constraint === name) indexes.delete(key);
          } else if (/^add\s+(?:constraint|column)\b/u.test(action)) {
            collectConstraints(table, action.replace(/^add\s+(?:column\s+)?/u, ""), fail);
          } else if (/^(?:alter column|enable (?:always |replica )?trigger|disable trigger)\b/u.test(action)) {
            // These operations do not change column identity or index key order.
          } else if ((match = new RegExp(`^rename column (${identifier}) to (${identifier})$`, "u").exec(action))) {
            for (const key of [...foreignKeys.values(), ...indexes.values()]) {
              if (key.table !== table) continue;
              const member = key.columns ? "columns" : "keys";
              key[member] = key[member].map((column) => column === match[1] ? match[2] : column);
              if (key.included) key.included = key.included.map((column) => column === match[1] ? match[2] : column);
            }
          } else if ((match = new RegExp(`^drop column (${identifier})(?: restrict)?$`, "u").exec(action))) {
            const column = match[1];
            // PostgreSQL also drops indexes that reference this column only in
            // INCLUDE or an expression. Opaque expressions cannot prove absence
            // of that dependency; require explicit index removal first.
            if ([...foreignKeys.values(), ...indexes.values()].some((key) => key.table === table &&
                ((key.columns ?? key.keys).some((value) => value === column || value === null) ||
                  key.included?.includes(column)))) fail(`drop referenced or unresolved key column ${table}.${column}`);
          } else fail(`${table}: ${action}`);
        }
      } else if ((match = new RegExp(`^create (?:unique )?index (${relation}) on (?:only )?(${relation})(?: using (${identifier}))?\\s*\\(`, "u").exec(statement))) {
        const name = qualify(match[1]);
        const [body, tail] = parenthesized(statement.slice(match[0].length - 1), fail);
        let suffix = tail.trim();
        let included = [];
        if (/^include\s*\(/u.test(suffix)) {
          const [columns, rest] = parenthesized(suffix.slice(suffix.indexOf("(")), fail);
          included = columnList(columns, fail);
          suffix = rest.trim();
        }
        suffix = suffix.replace(/^nulls (?:not )?distinct\b/u, "").trim();
        if (suffix && !/^where\s+\S/u.test(suffix)) fail(`index ${name} suffix`);
        const keys = splitKeys(body, fail).map((key) => {
          const plain = new RegExp(`^(${identifier})(?: (?:asc|desc))?(?: nulls (?:first|last))?$`, "u").exec(key);
          // Expressions and non-default operator/collation semantics are never
          // evidence for the required plain-column prefix.
          return plain?.[1] ?? null;
        });
        indexes.set(name, { name, table: qualify(match[2]), method: match[3] ?? "btree", keys, included,
          partial: Boolean(suffix), constraint: null });
      } else if ((match = /^drop index\s+([\s\S]+)$/u.exec(statement))) {
        for (const name of splitKeys(match[1], fail)) {
          if (!new RegExp(`^${relation}$`, "u").test(name)) fail(`drop index ${name}`);
          indexes.delete(qualify(name));
        }
      } else if ((match = /^drop table\s+([\s\S]+)$/u.exec(statement))) {
        for (const name of splitKeys(match[1], fail)) {
          if (!new RegExp(`^${relation}$`, "u").test(name)) fail(`drop table ${name}`);
          const table = qualify(name);
          for (const [key, fk] of foreignKeys) if (fk.table === table) foreignKeys.delete(key);
          for (const [key, index] of indexes) if (index.table === table) indexes.delete(key);
        }
      } else if (/^(?:create\s+(?:unique\s+)?(?:index|table)|alter\s+(?:table|index)|drop\s+(?:table|index))\b/u.test(statement)) {
        fail(statement);
      }
    }
  }
  return [...foreignKeys.values()].sort(byName).map((fk) => {
    const candidates = [...indexes.values()].filter((index) => index.table === fk.table).sort(byName);
    const coveringIndexes = candidates.filter((index) => index.method === "btree" && !index.partial &&
      fk.columns.every((column, position) => index.keys[position] === column)).map((index) => index.name);
    return { ...fk, candidates, coveringIndexes };
  });

  function collectConstraints(table, element, fail) {
    const matches = [...element.matchAll(new RegExp(`\\bconstraint (${identifier})\\s+`, "gu"))];
    if (matches.length === 0 && /\b(?:references|foreign key|primary key|unique)\b/u.test(element)) {
      fail(`unnamed key on ${table}`);
    }
    for (const [position, match] of matches.entries()) {
      const definition = element.slice(match.index + match[0].length, matches[position + 1]?.index).trim();
      const name = `${table}.${match[1]}`;
      const column = element.slice(0, match.index).trim().match(new RegExp(`^(${identifier})\\s+`, "u"))?.[1];
      if (/^(?:foreign key|references)\b/u.test(definition)) {
        let columns;
        let reference = definition;
        if (/^foreign key\s*\(/u.test(definition)) {
          const [body, tail] = parenthesized(definition.slice(definition.indexOf("(")), fail);
          columns = columnList(body, fail);
          reference = tail.trim();
        } else if (column) columns = [column];
        else fail(`foreign key ${name}`);
        const target = new RegExp(`^references (${relation})\\s*\\(`, "u").exec(reference);
        if (!target) fail(`foreign key reference ${name}`);
        const [targetColumns] = parenthesized(reference.slice(target[0].length - 1), fail);
        if (columnList(targetColumns, fail).length !== columns.length) fail(`foreign key arity ${name}`);
        foreignKeys.set(name, { name, table, columns });
      } else if (/^(?:primary key|unique)\b/u.test(definition)) {
        let keys;
        const key = /^(?:primary key|unique)(?: nulls (?:not )?distinct)?\s*/u.exec(definition);
        const remainder = definition.slice(key[0].length);
        if (remainder.startsWith("(")) {
          const [body, tail] = parenthesized(remainder, fail);
          if (tail.trim()) fail(`backing index ${name} suffix`);
          keys = columnList(body, fail);
        }
        else if (column && !remainder) keys = [column];
        else fail(`backing index ${name}`);
        indexes.set(qualify(match[1]), { name: qualify(match[1]), table, keys,
          method: "btree", partial: false, constraint: name });
      } else if (!/^check\s*\(/u.test(definition)) fail(`constraint ${name}`);
    }
  }
}

function byName(left, right) { return left.name < right.name ? -1 : left.name > right.name ? 1 : 0; }

function parenthesized(source, fail) {
  let depth = 0;
  for (let index = 0; index < source.length; index++) {
    if (source[index] === "(") depth++;
    if (source[index] === ")" && --depth === 0) return [source.slice(1, index), source.slice(index + 1)];
  }
  fail("unbalanced key declaration");
}

function splitKeys(source, fail) {
  const parts = [];
  let depth = 0;
  let start = 0;
  for (let index = 0; index < source.length; index++) {
    if (source[index] === "(") depth++;
    if (source[index] === ")" && --depth < 0) fail("unbalanced declaration");
    if (source[index] === "," && depth === 0) { parts.push(source.slice(start, index).trim()); start = index + 1; }
  }
  if (depth !== 0) fail("unbalanced declaration");
  parts.push(source.slice(start).trim());
  if (parts.some((part) => !part)) fail("empty key declaration");
  return parts;
}

function columnList(source, fail) {
  const columns = splitKeys(source, fail);
  if (columns.some((column) => !new RegExp(`^${identifier}$`, "u").test(column)) || new Set(columns).size !== columns.length) {
    fail("expected distinct plain key columns");
  }
  return columns;
}

export function requireForeignKeyCoverage(coverage, entries) {
  const foreignKeys = new Map(coverage.map((fk) => [fk.name, fk]));
  for (const fk of coverage) {
    if (fk.coveringIndexes.length === 0) {
      throw new Error(`Uncovered foreign key ${fk.name} (${fk.columns.join(", ")}); candidates=${JSON.stringify(fk.candidates)}`);
    }
  }
  if (!entries) return;
  for (const entry of entries) {
    const fk = entry.object_kind === "constraint" ? foreignKeys.get(entry.qualified_name) : undefined;
    const expected = fk ? "covered" : "not_applicable";
    if (entry.foreign_key_index_status !== expected) throw new Error(`FK coverage status mismatch: ${entry.qualified_name}; expected ${expected}`);
    if (fk) foreignKeys.delete(entry.qualified_name);
  }
  if (foreignKeys.size) throw new Error(`Missing foreign keys in manifest: ${[...foreignKeys.keys()].join(", ")}`);
}
