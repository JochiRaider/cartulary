// tools/harness_redaction_manifest.json is the single machine owner of the
// retained marker, the structured secret-key tokens, the secret CLI flag
// names, the value rules and the retained-secret scan's vocabulary
// (TH-HARNESS-REQ-604, TH-HARNESS-REQ-617). This module interprets that data
// once for both the generated native policy and the script redactor. It is
// pure: callers supply the parsed manifest.

export const redactionManifestSchemaID = "cartulary.harness_redaction_manifest.v2";

const caseInsensitivePrefix = "(?i)";

// ASCII whitespace as RE2 and the retained-secret scan define it.
export const asciiSpace = String.raw`[\t\n\f\r ]`;

function invalid(message) {
  throw new Error(`invalid redaction manifest: ${message}`);
}

// Escape a literal identically for ECMAScript and RE2.
export function regexpLiteral(text) {
  return text.replace(/[\\^$.*+?()[\]{}|]/gu, "\\$&");
}

function uniqueStrings(values, label, shape) {
  if (!Array.isArray(values) || values.length === 0) invalid(`${label} must be a non-empty array`);
  const seen = new Set();
  for (const value of values) {
    if (typeof value !== "string" || !shape.test(value) || seen.has(value)) {
      invalid(`${label} has an invalid or duplicate entry`);
    }
    seen.add(value);
  }
  return [...values];
}

function scanTerms(entries, label, field, shape, excludedShape) {
  if (!Array.isArray(entries) || entries.length === 0) invalid(`${label} must be a non-empty array`);
  const seen = new Set();
  return entries.map((entry) => {
    if (!entry || typeof entry !== "object" || Array.isArray(entry) ||
        Object.keys(entry).some((key) => key !== field && key !== "not_followed_by")) {
      invalid(`${label} entries declare only ${field} and not_followed_by`);
    }
    const text = entry[field];
    const excluded = entry.not_followed_by ?? null;
    if (typeof text !== "string" || !shape.test(text) || seen.has(text) ||
        (excluded !== null && (typeof excluded !== "string" || !excludedShape.test(excluded)))) {
      invalid(`${label} has an invalid or duplicate entry`);
    }
    seen.add(text);
    return { text, excluded };
  });
}

// Shared patterns use explicit ASCII classes (TH-HARNESS-REQ-617). Shorthand
// classes, word boundaries and the dot differ between ECMAScript and RE2.
function assertEngineNeutral(pattern, label) {
  let inClass = false;
  for (let index = 0; index < pattern.length; index += 1) {
    const character = pattern[index];
    if (character === "\\") {
      const escaped = pattern[index + 1] ?? "";
      if ("sSdDwWbB".includes(escaped) && escaped !== "") {
        invalid(`${label} uses the engine-dependent \\${escaped}; use an explicit ASCII class`);
      }
      index += 1;
    } else if (inClass) {
      inClass = character !== "]";
    } else if (character === "[") {
      inClass = true;
      // A leading "]" is literal in RE2; patterns never use an empty class.
      if (pattern[index + 1] === "^") index += 1;
      if (pattern[index + 1] === "]") index += 1;
    } else if (character === ".") {
      invalid(`${label} uses the engine-dependent "."; use an explicit ASCII class`);
    }
  }
  return pattern;
}

const metacharacters = new Set([..."\\^$.|?*+()[]{}"]);
const quantifiers = new Set([..."*+?{"]);

// Index just past the group opened at `open`, honouring escapes and classes.
function groupEnd(pattern, open) {
  let depth = 0;
  let inClass = false;
  for (let index = open; index < pattern.length; index += 1) {
    const character = pattern[index];
    if (character === "\\") {
      index += 1;
    } else if (inClass) {
      inClass = character !== "]";
    } else if (character === "[") {
      inClass = true;
      if (pattern[index + 1] === "^") index += 1;
      if (pattern[index + 1] === "]") index += 1;
    } else if (character === "(") {
      depth += 1;
    } else if (character === ")") {
      depth -= 1;
      if (depth === 0) return index + 1;
    }
  }
  return -1;
}

// Every top-level alternative of pattern[start, end) begins literally.
function alternativesBeginLiterally(pattern, start, end) {
  let depth = 0;
  let inClass = false;
  let alternative = start;
  for (let index = start; index < end; index += 1) {
    const character = pattern[index];
    if (character === "\\") {
      index += 1;
    } else if (inClass) {
      inClass = character !== "]";
    } else if (character === "[") {
      inClass = true;
      if (pattern[index + 1] === "^") index += 1;
      if (pattern[index + 1] === "]") index += 1;
    } else if (character === "(") {
      depth += 1;
    } else if (character === ")") {
      depth -= 1;
    } else if (character === "|" && depth === 0) {
      if (!beginsLiterally(pattern, alternative, index)) return false;
      alternative = index + 1;
    }
  }
  return beginsLiterally(pattern, alternative, end);
}

// pattern[start, end) begins with an unquantified literal character, possibly
// inside groups whose every alternative does too.
function beginsLiterally(pattern, start, end) {
  if (start >= end) return false;
  const character = pattern[start];
  if (character === "(") {
    const close = groupEnd(pattern, start);
    if (close < 0 || close > end || quantifiers.has(pattern[close])) return false;
    const inner = pattern.startsWith("(?:", start) ? start + 3 : pattern[start + 1] === "?" ? -1 : start + 1;
    return inner >= 0 && alternativesBeginLiterally(pattern, inner, close - 1);
  }
  let next = start + 1;
  if (character === "\\") {
    if (/[A-Za-z0-9]/u.test(pattern[start + 1] ?? "0")) return false;
    next = start + 2;
  } else if (metacharacters.has(character)) {
    return false;
  }
  return !quantifiers.has(pattern[next]);
}

// A consumed ASCII boundary group "(^|[^...])" opening a pattern.
const boundaryGroup = /^\(\^\|\[\^(?:\\[^A-Za-z0-9]|\\[tnfr]|[\x20-\x5b\x5e-\x7e])+\]\)/u;

// TH-HARNESS-REQ-617: a shared value pattern begins with a literal, or with a
// consumed ASCII boundary group that its replacement re-emits as $1, so no
// match can start inside a token it cannot complete and a backtracking engine
// stays linear.
function assertBoundedStart(rule) {
  if (boundaryGroup.test(rule.body)) {
    if (!rule.replacement.startsWith("$1")) invalid(`value rule ${rule.name} must re-emit its consumed boundary as $1`);
  } else if (!beginsLiterally(rule.body, 0, rule.body.length)) {
    invalid(`value rule ${rule.name} must begin with a literal or a consumed boundary group (^|[^...])`);
  }
}

function splitCaseInsensitive(pattern) {
  return pattern.startsWith(caseInsensitivePrefix)
    ? { body: pattern.slice(caseInsensitivePrefix.length), caseInsensitive: true }
    : { body: pattern, caseInsensitive: false };
}

function valueRules(entries, marker) {
  if (!Array.isArray(entries) || entries.length === 0) invalid("value_patterns must be a non-empty array");
  const names = new Set();
  return entries.map((rule) => {
    if (!rule || typeof rule.name !== "string" || !rule.name || names.has(rule.name) ||
        typeof rule.pattern !== "string" || !rule.pattern ||
        typeof rule.replacement !== "string" || !rule.replacement) {
      invalid("value_patterns has an invalid or duplicate rule");
    }
    // Every rule emits the single marker and no second or typed form.
    if (!rule.replacement.includes(marker) || rule.replacement.includes(`${marker}:`) ||
        /redact/iu.test(rule.replacement.split(marker).join(""))) {
      invalid(`value rule ${rule.name} must emit only the marker ${marker}`);
    }
    names.add(rule.name);
    const { body, caseInsensitive } = splitCaseInsensitive(rule.pattern);
    new RegExp(body, caseInsensitive ? "giu" : "gu");
    return { name: rule.name, body, caseInsensitive, replacement: rule.replacement };
  });
}

// The native scan's JSON-member prefix in RE2 (automaton, so linear): a
// quoted key containing a term, followed by a colon and the opening quote of
// a string value. A term with an exclusion is expanded without look-ahead.
// Redactors never use it; they decide key membership in code.
function jsonMemberBody(terms) {
  const plain = terms.filter((term) => term.excluded === null).map((term) => regexpLiteral(term.text));
  const alternatives = plain.length ? [`[^"]*(?:${plain.join("|")})[^"]*`] : [];
  for (const term of terms.filter((entry) => entry.excluded !== null)) {
    alternatives.push(`[^"]*${regexpLiteral(term.text)}(?:[^${term.excluded}"][^"]*)?`);
  }
  return `"(?:${alternatives.join("|")})"${asciiSpace}*:${asciiSpace}*"`;
}

/** Validates the manifest and returns its interpreted vocabulary. */
export function redactionVocabulary(manifest) {
  if (manifest?.schema_id !== redactionManifestSchemaID) invalid(`schema_id must be ${redactionManifestSchemaID}`);
  const marker = manifest.replacement;
  if (typeof marker !== "string" || marker === "") invalid("replacement must be the non-empty marker");
  const tokens = uniqueStrings(manifest.structured_secret_key_tokens, "structured_secret_key_tokens", /^[A-Z0-9]+(?:_[A-Z0-9]+)*$/u);
  const cliFlagNames = uniqueStrings(manifest.secret_cli_flag_names, "secret_cli_flag_names", /^[a-z0-9]+(?:-[a-z0-9]+)*$/u);
  const wholeStringAssignment = manifest.whole_string_assignment;
  if (typeof wholeStringAssignment !== "string" || !wholeStringAssignment.startsWith("^") ||
      new RegExp(`${wholeStringAssignment}|`, "u").exec("").length !== 2) {
    invalid("whole_string_assignment must be an anchored pattern with one capture group");
  }
  const scan = manifest.retained_scan;
  if (!scan || typeof scan !== "object") invalid("retained_scan is required");
  if (!Array.isArray(scan.forbidden_names) || scan.forbidden_names.length === 0) {
    invalid("retained_scan.forbidden_names must be a non-empty array");
  }
  const forbiddenNames = scan.forbidden_names.map((entry) => {
    if (!entry || typeof entry.pattern !== "string" || !entry.pattern || !["u", "iu"].includes(entry.flags) ||
        Object.keys(entry).some((key) => key !== "pattern" && key !== "flags")) {
      invalid("retained_scan.forbidden_names entries declare a pattern and the flags u or iu");
    }
    // Forbidden names are native scan syntax inside the frozen identity, not
    // shared redaction patterns, so they keep their original notation.
    new RegExp(entry.pattern, entry.flags);
    return { pattern: entry.pattern, flags: entry.flags };
  });
  const jsonMemberTerms = scanTerms(scan.json_member_terms, "retained_scan.json_member_terms", "term", /^[a-z][a-z0-9_]*$/u, /^[a-z]$/u);
  const assignmentNames = scanTerms(scan.assignment_names, "retained_scan.assignment_names", "name", /^[A-Z][A-Z0-9_]*$/u, /^[A-Z]$/u);
  const urlUserinfoSchemes = uniqueStrings(scan.url_userinfo_schemes, "retained_scan.url_userinfo_schemes", /^[a-z][a-z0-9+.-]*$/u);
  const memberBody = jsonMemberBody(jsonMemberTerms);
  return {
    marker,
    tokens,
    cliFlagNames,
    // A name's hyphen matches a hyphen, an underscore or nothing.
    cliFlagPattern: `^--(?:${cliFlagNames.map((name) => name.split("-").map(regexpLiteral).join("[_-]?")).join("|")})$`,
    wholeStringAssignment,
    jsonMemberBody: memberBody,
    // Authored rules apply first; the scan closure for JSON members embedded
    // in raw text then runs in code (redactJSONMembers).
    valueRules: valueRules(manifest.value_patterns, marker),
    retainedScan: { forbiddenNames, jsonMemberTerms, assignmentNames, urlUserinfoSchemes },
  };
}

/**
 * Generation rejects a vocabulary whose shared patterns are engine-dependent
 * or can start inside a token they cannot complete, so the native and script
 * redactors produce identical output in linear time (TH-HARNESS-REQ-617).
 */
export function assertSharedPatterns(vocabulary) {
  for (const rule of vocabulary.valueRules) {
    assertEngineNeutral(rule.body, `value rule ${rule.name}`);
    assertBoundedStart(rule);
  }
  assertEngineNeutral(vocabulary.wholeStringAssignment, "whole_string_assignment");
  assertEngineNeutral(vocabulary.cliFlagPattern, "secret CLI flag pattern");
  return vocabulary;
}

/** Compiles the vocabulary for the script redactor. Predicates are not global. */
export function compileRedactionVocabulary(vocabulary) {
  return {
    replacement: vocabulary.marker,
    tokens: vocabulary.tokens,
    valuePatterns: vocabulary.valueRules.map((rule) => ({
      name: rule.name,
      regex: new RegExp(rule.body, rule.caseInsensitive ? "giu" : "gu"),
      replacement: rule.replacement,
    })),
    cliFlag: new RegExp(vocabulary.cliFlagPattern, "iu"),
    scanMemberTerms: vocabulary.retainedScan.jsonMemberTerms,
    wholeStringAssignment: new RegExp(vocabulary.wholeStringAssignment, "u"),
  };
}

/** The raw-text redactor: the value rules, then the JSON-member closure. */
export function applyValueRules(compiled, value) {
  let text = String(value);
  for (const rule of compiled.valuePatterns) text = text.replace(rule.regex, rule.replacement);
  return redactJSONMembers(compiled.scanMemberTerms, compiled.replacement, text);
}

// Simple case folding as the scan's case-insensitive matching applies it to
// ASCII terms: ASCII letters, the long s (U+017F) and the Kelvin sign (U+212A).
function foldScanCharacter(character) {
  if (character >= "A" && character <= "Z") return character.toLowerCase();
  if (character === "\u017f") return "s";
  if (character === "\u212a") return "k";
  return character;
}

/**
 * Whether a whole key contains a retained-scan JSON-member term, decided in
 * code in linear time (TH-HARNESS-REQ-604/617). A term with an exclusion does
 * not match when the next character folds to the excluded letter.
 */
export function scanKeyMatches(terms, key) {
  const folded = Array.from(String(key), foldScanCharacter);
  for (const { text, excluded } of terms) {
    search: for (let index = 0; index + text.length <= folded.length; index += 1) {
      for (let offset = 0; offset < text.length; offset += 1) {
        if (folded[index + offset] !== text[offset]) continue search;
      }
      if (excluded === null || folded[index + text.length] !== excluded) return true;
    }
  }
  return false;
}

const asciiSpaceCharacters = new Set(["\t", "\n", "\f", "\r", " "]);

/**
 * The scan closure for JSON members embedded in raw text (TH-HARNESS-REQ-604):
 * a quoted key, ASCII whitespace, a colon, ASCII whitespace and a non-empty
 * quoted value has its value replaced when the key contains a scan term.
 * Every quote is a candidate key start, as in the scan, and each candidate
 * inspects only its own key, separator and value, so the pass is linear.
 */
export function redactJSONMembers(terms, marker, text) {
  const quotes = [];
  for (let index = text.indexOf('"'); index !== -1; index = text.indexOf('"', index + 1)) quotes.push(index);
  let output = "";
  let previous = 0;
  for (let quote = 0; quote + 3 < quotes.length; quote += 1) {
    let cursor = quotes[quote + 1] + 1;
    while (asciiSpaceCharacters.has(text[cursor])) cursor += 1;
    if (text[cursor] !== ":") continue;
    cursor += 1;
    while (asciiSpaceCharacters.has(text[cursor])) cursor += 1;
    if (cursor !== quotes[quote + 2] || quotes[quote + 3] === cursor + 1 ||
        !scanKeyMatches(terms, text.slice(quotes[quote] + 1, quotes[quote + 1]))) continue;
    output += text.slice(previous, cursor + 1) + marker;
    previous = quotes[quote + 3];
  }
  return previous === 0 ? text : output + text.slice(previous);
}

// TH-HARNESS-REQ-604: split camelCase, uppercase ASCII, and collapse every
// other run into one underscore. Non-ASCII characters are separators.
export function canonicalStructuredKey(value) {
  return String(value ?? "")
    .replace(/([a-z0-9])(?=[A-Z])/gu, "$1_")
    .replace(/[a-z]+/gu, (letters) => letters.toUpperCase())
    .replace(/[^A-Z0-9]+/gu, "_")
    .replace(/^_+|_+$/gu, "");
}

export function isStructuredSecretKey(tokens, key) {
  const canonical = canonicalStructuredKey(key);
  if (!canonical) return false;
  return tokens.some((token) => canonical === token || canonical.startsWith(`${token}_`) || canonical.endsWith(`_${token}`));
}
