import { createHash } from "node:crypto";

function rejectUnpairedSurrogates(value, label) {
  for (let index = 0; index < value.length; index += 1) {
    const unit = value.charCodeAt(index);
    if (unit >= 0xd800 && unit <= 0xdbff) {
      const next = value.charCodeAt(index + 1);
      if (!(next >= 0xdc00 && next <= 0xdfff)) {
        throw new Error(`${label} contains an unpaired high surrogate`);
      }
      index += 1;
      continue;
    }
    if (unit >= 0xdc00 && unit <= 0xdfff) {
      throw new Error(`${label} contains an unpaired low surrogate`);
    }
  }
}

export function parseStrictJSON(source, label = "JSON") {
  let offset = 0;

  function fail(message) {
    throw new Error(`${label}:${offset + 1}: ${message}`);
  }

  function skipWhitespace() {
    while (/[\u0009\u000a\u000d\u0020]/u.test(source[offset] ?? "") && offset < source.length) {
      offset += 1;
    }
  }

  function parseString() {
    const start = offset;
    offset += 1;
    let escaped = false;
    while (offset < source.length) {
      const character = source[offset];
      if (!escaped && character === '"') {
        offset += 1;
        const value = JSON.parse(source.slice(start, offset));
        rejectUnpairedSurrogates(value, label);
        return value;
      }
      if (!escaped && character === "\\") {
        escaped = true;
      } else {
        escaped = false;
      }
      offset += 1;
    }
    fail("unterminated string");
  }

  function parseNumber() {
    const match = source.slice(offset).match(/^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?/u);
    if (!match) {
      fail("invalid number");
    }
    offset += match[0].length;
    const value = Number(match[0]);
    if (!Number.isFinite(value)) {
      fail("number is not finite");
    }
    if (Object.is(value, -0)) {
      fail("negative zero is not I-JSON");
    }
    if (Number.isInteger(value) && !Number.isSafeInteger(value)) {
      fail("integer is outside the I-JSON interoperable range");
    }
    return value;
  }

  function parseArray() {
    const result = [];
    offset += 1;
    skipWhitespace();
    if (source[offset] === "]") {
      offset += 1;
      return result;
    }
    while (offset < source.length) {
      result.push(parseValue());
      skipWhitespace();
      if (source[offset] === "]") {
        offset += 1;
        return result;
      }
      if (source[offset] !== ",") {
        fail("array requires ',' or ']'");
      }
      offset += 1;
      skipWhitespace();
    }
    fail("unterminated array");
  }

  function parseObject() {
    const result = {};
    const keys = new Set();
    offset += 1;
    skipWhitespace();
    if (source[offset] === "}") {
      offset += 1;
      return result;
    }
    while (offset < source.length) {
      if (source[offset] !== '"') {
        fail("object key must be a string");
      }
      const key = parseString();
      if (keys.has(key)) {
        fail(`duplicate object member ${JSON.stringify(key)}`);
      }
      keys.add(key);
      skipWhitespace();
      if (source[offset] !== ":") {
        fail("object key requires ':'");
      }
      offset += 1;
      result[key] = parseValue();
      skipWhitespace();
      if (source[offset] === "}") {
        offset += 1;
        return result;
      }
      if (source[offset] !== ",") {
        fail("object requires ',' or '}'");
      }
      offset += 1;
      skipWhitespace();
    }
    fail("unterminated object");
  }

  function parseValue() {
    skipWhitespace();
    const character = source[offset];
    if (character === '"') {
      return parseString();
    }
    if (character === "{") {
      return parseObject();
    }
    if (character === "[") {
      return parseArray();
    }
    for (const [token, value] of [["true", true], ["false", false], ["null", null]]) {
      if (source.startsWith(token, offset)) {
        offset += token.length;
        return value;
      }
    }
    return parseNumber();
  }

  const value = parseValue();
  skipWhitespace();
  if (offset !== source.length) {
    fail("unexpected trailing content");
  }
  return value;
}

export function canonicalJSONString(value) {
  if (value === null || typeof value === "boolean" || typeof value === "string") {
    if (typeof value === "string") {
      rejectUnpairedSurrogates(value, "semantic JSON string");
    }
    return JSON.stringify(value);
  }
  if (typeof value === "number") {
    if (
      !Number.isFinite(value) ||
      Object.is(value, -0) ||
      (Number.isInteger(value) && !Number.isSafeInteger(value))
    ) {
      throw new Error("semantic JSON contains a non-I-JSON number");
    }
    return JSON.stringify(value);
  }
  if (Array.isArray(value)) {
    return `[${value.map((entry) => canonicalJSONString(entry)).join(",")}]`;
  }
  if (!value || typeof value !== "object") {
    throw new Error(`semantic JSON contains unsupported ${typeof value}`);
  }
  return `{${Object.keys(value)
    .sort()
    .map((key) => `${canonicalJSONString(key)}:${canonicalJSONString(value[key])}`)
    .join(",")}}`;
}

export function semanticJSONSHA256(value) {
  return createHash("sha256").update(canonicalJSONString(value)).digest("hex");
}

export function semanticJSONDigest(value) {
  return `sha256:${semanticJSONSHA256(value)}`;
}
