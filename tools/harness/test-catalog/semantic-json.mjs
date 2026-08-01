import { createHash } from "node:crypto";
import { parseStrictJSON } from "../contract/index.mjs";

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

export { parseStrictJSON };

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
