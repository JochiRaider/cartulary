// Compile the existing foundation schema closure into native predicates. This
// is a closed projection, not a second authored schema registry. New vocabulary
// must be implemented and tested explicitly; unsupported keywords stop generation.
const annotations = new Set(["$schema", "$id", "$comment", "$defs", "title", "description", "default", "examples", "format"]);
const assertions = new Set(["$ref", "type", "const", "enum", "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "minLength", "maxLength", "minProperties", "maxProperties", "pattern", "required", "properties", "patternProperties", "additionalProperties", "propertyNames", "items", "prefixItems", "minItems", "maxItems", "uniqueItems", "allOf", "anyOf", "oneOf", "not", "if", "then", "else"]);
const q = JSON.stringify;
// Token patterns dominate the harness schemas. Project this deliberately small
// anchored grammar to byte predicates, so every reader does not rebuild a large
// regexp automaton. Other patterns retain the validated RE2 implementation.
function tokenPredicate(pattern) {
  const match = /^\^((?:[A-Za-z0-9_:=-]|\\\.)*)(\[[A-Za-z0-9_.:-]+\])(?:(\[[A-Za-z0-9_.:-]+\]))?(\*|\+|\{\d+(?:,\d*)?\})?\$$/u.exec(pattern);
  if (!match) return null;
  const [, rawPrefix, first, repeated, quantifier = ""] = match;
  const mask = (characterClass) => {
    const expression = new RegExp(`^${characterClass}$`, "u");
    let low = 0n, high = 0n;
    for (let code = 0; code < 128; code++) if (expression.test(String.fromCharCode(code))) {
      if (code < 64) low |= 1n << BigInt(code); else high |= 1n << BigInt(code - 64);
    }
    return `0x${low.toString(16)},0x${high.toString(16)}`;
  };
  let min = 1, max = 1;
  if (quantifier === "*") { min = 0; max = -1; }
  else if (quantifier === "+") { min = 1; max = -1; }
  else if (quantifier) {
    const counts = quantifier.slice(1,-1).split(",");
    min = Number(counts[0]); max = counts.length === 1 ? min : counts[1] === "" ? -1 : Number(counts[1]);
  }
  if (repeated) { min++; if (max !== -1) max++; }
  return `schemaToken(s,${q(rawPrefix.replaceAll("\\.", "."))},${min},${max},${mask(first)},${mask(repeated ?? first)})`;
}
const relativePattern = "^(?!/)(?!.*(?:^|/)\\.\\.(?:/|$))(?!.*(?:^|/)\\.(?:/|$))(?!.*//)[A-Za-z0-9._@+,:=-]+(?:/[A-Za-z0-9._@+,:=-]+)*$";
function validatePattern(pattern) {
  if (pattern === relativePattern) return;
  if (pattern === "^(?!/)(?!.*(?:^|/)\\.\\.(?:/|$))[A-Za-z0-9._@+-]+(?:/[A-Za-z0-9._@+-]+)*$" || pattern === "^(?!/)(?!.*(?:^|/)\\.\\.(?:/|$))(?!.*[\\\\\\u0000]).+$") return;
  if (pattern === "^(?!/)(?!.*(?:^|/)\\.\\.(?:/|$)).+$") return;
  if (pattern === "^(?!/)(?!.*\\\\)(?!.*(?:^|/)\\.{1,2}(?:/|$))(?!.*//)[A-Za-z0-9._@+/-]+$") return;
  // The catalog's exact reserved-segment assertions have a native predicate.
  pattern = pattern.replaceAll("(?!phase[0-9]+\\.|fe_p[0-9]+\\.)", "").replaceAll("(?!phase[0-9]+$|fe_p[0-9]+$)", "");
  // Deliberately admit only the shared ECMAScript/RE2 vocabulary. A future
  // expression outside this subset needs an explicit projection and probes.
  let inClass = false;
  for (let index = 0; index < pattern.length; index++) {
    const c = pattern[index];
    if (c.charCodeAt(0) > 127) throw Error("native projection: unsupported Unicode regex literal");
    if (c === "\\") {
      const next = pattern[++index];
      if (!next || /[A-Za-z0-9]/u.test(next) && !/[dDwW]/u.test(next)) throw Error("native projection: unsupported regex escape");
      continue;
    }
    if (c === "[") inClass = true;
    if (c === "]") inClass = false;
    if (!inClass && (c === "." || pattern.slice(index, index + 2) === "(?" && pattern[index + 2] !== ":")) throw Error(`native projection: unsupported regex assertion or wildcard: ${pattern}`);
    if (!inClass && c === "{") {
      const repetition = /^\{(\d+)(?:,(\d*))?\}/u.exec(pattern.slice(index));
      if (!repetition || Number(repetition[1]) > 1000 || Number(repetition[2] ?? 0) > 1000) throw Error("native projection: unsupported regex repetition");
    }
  }
}
const literal = (value) => value === null ? "nil" : typeof value === "number" ? `float64(${value})` : typeof value === "string" || typeof value === "boolean" ? q(value) : Array.isArray(value) ? `[]any{${value.map(literal).join(",")}}` : `map[string]any{${Object.entries(value).map(([k,v]) => `${q(k)}:${literal(v)}`).join(",")}}`;

export function renderNativeSchemaValidators(schemas, entries, { prefix = "schema", registry = "validSchema", patterns = new Set() } = {}) {
  const functions = [], names = new Map();
  const reference = (ref, base) => {
    const [external, fragment = ""] = ref.split("#"), id = external || base;
    let schema = schemas.get(id);
    if (!schema) throw Error(`native projection: unresolved ${ref}`);
    if (fragment) {
      if (!fragment.startsWith("/")) throw Error(`native projection: unsupported fragment ${ref}`);
      for (const segment of fragment.slice(1).split("/")) schema = schema?.[segment.replaceAll("~1", "/").replaceAll("~0", "~")];
    }
    if (schema === undefined) throw Error(`native projection: unresolved ${ref}`);
    return compile(schema, id);
  };
  function compile(schema, base) {
    const key = `${base}:${JSON.stringify(schema)}`;
    if (names.has(key)) return names.get(key);
    const name = `${prefix}${names.size}`; names.set(key, name);
    const index = functions.length; functions.push("");
    const lines = [];
    if (typeof schema === "boolean") lines.push(`return ${schema}`);
    else {
      for (const key of Object.keys(schema)) if (!annotations.has(key) && !assertions.has(key)) throw Error(`native projection: unsupported keyword ${key} in ${base}`);
      const reject = (condition) => lines.push(`if ${condition} { return false }`);
      if (schema.$ref) reject(`!${reference(schema.$ref, base)}(v)`);
      if (schema.type) reject(`!(${[schema.type].flat().map((type) => `schemaType(v, ${q(type)})`).join(" || ")})`);
      if (Object.hasOwn(schema, "const")) reject(`!schemaEqual(v, ${literal(schema.const)})`);
      if (schema.enum) reject(`!(${schema.enum.map((value) => `schemaEqual(v, ${literal(value)})`).join(" || ")})`);
      for (const [key, operator] of [["minimum", "<"], ["maximum", ">"], ["exclusiveMinimum", "<="], ["exclusiveMaximum", ">="]])
        if (key in schema) lines.push(`if n, ok := v.(float64); ok && n ${operator} ${schema[key]} { return false }`);
      if (schema.minLength !== undefined || schema.maxLength !== undefined || schema.pattern) {
        lines.push("if s, ok := v.(string); ok {");
        if (schema.minLength !== undefined) reject(`schemaLength(s) < ${schema.minLength}`);
        if (schema.maxLength !== undefined) reject(`schemaLength(s) > ${schema.maxLength}`);
        if (schema.pattern) { validatePattern(schema.pattern); patterns.add(schema.pattern); reject(`!${tokenPredicate(schema.pattern) ?? `schemaPattern(${q(schema.pattern)}, s)`}`); }
        lines.push("}");
      }
      if (schema.required || schema.properties || schema.patternProperties || schema.propertyNames !== undefined || schema.additionalProperties !== undefined || schema.minProperties !== undefined || schema.maxProperties !== undefined) {
        lines.push("if o, ok := v.(map[string]any); ok {", "_ = o");
        if (schema.minProperties !== undefined) reject(`len(o) < ${schema.minProperties}`);
        if (schema.maxProperties !== undefined) reject(`len(o) > ${schema.maxProperties}`);
        for (const key of schema.required ?? []) lines.push(`if _, present := o[${q(key)}]; !present { return false }`);
        for (const [key, child] of Object.entries(schema.properties ?? {})) lines.push(`if x, present := o[${q(key)}]; present && !${compile(child, base)}(x) { return false }`);
        for (const [pattern, child] of Object.entries(schema.patternProperties ?? {})) {
          validatePattern(pattern); patterns.add(pattern);
          lines.push(`for key, x := range o { if schemaPattern(${q(pattern)}, key) && !${compile(child, base)}(x) { return false } }`);
        }
        if (schema.propertyNames !== undefined) lines.push(`for key := range o { if !${compile(schema.propertyNames, base)}(key) { return false } }`);
        if (schema.additionalProperties !== undefined && schema.additionalProperties !== true) {
          lines.push("for key, x := range o {", "_ = x", "switch key {");
          const keys = Object.keys(schema.properties ?? {});
          if (keys.length) lines.push(`case ${keys.map(q).join(",")}:`);
          lines.push("default:");
          const propertyPatterns = Object.keys(schema.patternProperties ?? {});
          if (propertyPatterns.length) lines.push(`if ${propertyPatterns.map((pattern) => `schemaPattern(${q(pattern)}, key)`).join(" || ")} { continue }`);
          if (schema.additionalProperties === false) lines.push("return false");
          else reject(`!${compile(schema.additionalProperties, base)}(x)`);
          lines.push("}", "}");
        }
        lines.push("}");
      }
      if (schema.items !== undefined || schema.prefixItems !== undefined || schema.minItems !== undefined || schema.maxItems !== undefined || schema.uniqueItems) {
        lines.push("if a, ok := v.([]any); ok {", "_ = a");
        if (schema.minItems !== undefined) reject(`len(a) < ${schema.minItems}`);
        if (schema.maxItems !== undefined) reject(`len(a) > ${schema.maxItems}`);
        if (schema.uniqueItems) reject("!schemaUnique(a)");
        for (const [index, child] of (schema.prefixItems ?? []).entries()) reject(`len(a) > ${index} && !${compile(child,base)}(a[${index}])`);
        if (schema.items !== undefined) lines.push(`for i := ${schema.prefixItems?.length ?? 0}; i < len(a); i++ { if !${compile(schema.items, base)}(a[i]) { return false } }`);
        lines.push("}");
      }
      for (const key of ["allOf", "anyOf"]) if (schema[key]) reject(`!(${schema[key].map((child) => `${compile(child, base)}(v)`).join(key === "allOf" ? " && " : " || ")})`);
      if (schema.oneOf) { lines.push("matches := 0"); for (const child of schema.oneOf) lines.push(`if ${compile(child, base)}(v) { matches++ }`); reject("matches != 1"); }
      if (schema.not !== undefined) reject(`${compile(schema.not, base)}(v)`);
      if (schema.if !== undefined) {
        lines.push(`if ${compile(schema.if, base)}(v) {`);
        if (schema.then !== undefined) reject(`!${compile(schema.then, base)}(v)`);
        lines.push("} else {");
        if (schema.else !== undefined) reject(`!${compile(schema.else, base)}(v)`);
        lines.push("}");
      }
      lines.push("return true");
    }
    functions[index] = `func ${name}(v any) bool {\n${lines.join("\n")}\n}`;
    return name;
  }
  const cases = entries.map((entry) => `case ${q(entry.schema_id)}: return ${reference(entry.schema_id, entry.schema_id)}(value)`);
  return `// Code generated by tools/harness/generated-artifacts/generate-foundation-schema-validators.mjs; DO NOT EDIT.\npackage harnessschema\n\nfunc ${registry}(id string, value any) bool {\nswitch id {\n${cases.join("\n")}\ndefault: return false\n}\n}\n\n${functions.join("\n\n")}\n`;
}
