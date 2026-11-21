import { createHash } from "node:crypto";

import {
  applyValueRules,
  asciiSpace,
  assertSharedPatterns,
  compileRedactionVocabulary,
  redactionVocabulary,
  regexpLiteral,
} from "../contract/redaction-vocabulary.mjs";

// The retained-secret scan's policy identity is frozen (TH-HARNESS-REQ-617).
// Generation renders the identity from the manifest and must reproduce it.
export const frozenRetainedScanPolicyDigest =
  "sha256:eaa74efae69ea73a6e0f18823a94ccc70a19b242e2cb2880e446611c3084ee11";

// The identity keeps the scan's original ECMAScript notation byte for byte:
// the forbidden-name lines, then the JSON-member, assignment and URL lines.
export function renderRetainedScanIdentity(vocabulary) {
  const { marker, retainedScan: scan } = vocabulary;
  const names = scan.forbiddenNames.map(({ pattern, flags }) => `/${pattern}/${flags}`).join("\n");
  const lookahead = ({ text, excluded }) => (excluded === null ? text : `${text}(?!${excluded})`);
  const json = String.raw`/"[^"]*(?:${scan.jsonMemberTerms.map(lookahead).join("|")})[^"]*"\s*:\s*"(?!${marker})[^"]+"/iu`;
  const assignment = String.raw`/(?:${scan.assignmentNames.map(lookahead).join("|")})=(?!${marker})[^\s]+/u`;
  const url = String.raw`/(?:${scan.urlUserinfoSchemes.map(regexpLiteral).join("|")}):\/\/[^\s/:]+:(?!${marker})[^\s/@]+@/iu`;
  return `${names}${json}\n${assignment}\n${url}`;
}

// RE2 contexts of the streaming scan: a prefix, the bytes a value excludes and
// its terminator. The native scan expands the marker exemption from these.
function retainedScanContexts(vocabulary) {
  const scan = vocabulary.retainedScan;
  // RE2's \s, written as the explicit ASCII class the shared patterns use.
  const space = asciiSpace.slice(1, -1);
  // An assignment name is followed by "=", which no exclusion letter matches.
  const assignment = `(?:${scan.assignmentNames.map(({ text }) => regexpLiteral(text)).join("|")})=`;
  return [
    { prefix: `(?i:${vocabulary.jsonMemberBody})`, excluded: String.raw`"`, terminator: String.raw`"` },
    { prefix: assignment, excluded: space, terminator: `(?:${asciiSpace}|$)` },
    { prefix: `(?i:(?:${scan.urlUserinfoSchemes.map(regexpLiteral).join("|")})://[^${space}/:]+:)`, excluded: `${space}/@`, terminator: "@" },
  ];
}

function forbiddenNamesProgram(vocabulary) {
  return vocabulary.retainedScan.forbiddenNames
    .map(({ pattern, flags }) => (flags === "iu" ? `(?i:${pattern})` : pattern))
    .join("|");
}

// The authored rules must redact every name, scheme and flag in the scan and
// CLI vocabularies, in the form the scan checks, to the bare marker.
function assertVocabularyCoverage(vocabulary) {
  const compiled = compileRedactionVocabulary(vocabulary);
  const { marker } = vocabulary;
  const expect = (input, want) => {
    if (applyValueRules(compiled, input) !== want) {
      throw new Error(`redaction rules do not cover ${JSON.stringify(input)}`);
    }
  };
  for (const { text } of vocabulary.retainedScan.assignmentNames) {
    expect(`${text}=fixture`, `${text}=${marker}`);
    expect(`PREFIX_${text}=fixture`, `PREFIX_${text}=${marker}`);
  }
  for (const scheme of vocabulary.retainedScan.urlUserinfoSchemes) {
    expect(`${scheme}://user:fix?ture@host`, `${scheme}://user:${marker}@host`);
  }
  for (const { text } of vocabulary.retainedScan.jsonMemberTerms) {
    expect(`{"prefix_${text}":"fixture"}`, `{"prefix_${text}":"${marker}"}`);
  }
  for (const name of vocabulary.cliFlagNames) {
    for (const spelling of new Set([name, name.replaceAll("-", "_"), name.replaceAll("-", "")])) {
      if (!compiled.cliFlag.test(`--${spelling}`)) throw new Error(`secret CLI flag pattern misses --${spelling}`);
      expect(`--${spelling}=fixture`, `--${spelling}=${marker}`);
      expect(`--${spelling} fixture`, `--${spelling} ${marker}`);
    }
  }
}

const goString = (value) => JSON.stringify(value);
const goPattern = (rule) => goString(`${rule.caseInsensitive ? "(?i)" : ""}${rule.body}`);

// The native redactor and scan carry the complete authored policy in their
// executable. They must never substitute a smaller runtime fallback.
export function renderNativeRedactionPolicy(manifest) {
  let vocabulary;
  try {
    vocabulary = assertSharedPatterns(redactionVocabulary(manifest));
  } catch (error) {
    throw new Error(`invalid native redaction policy: ${error.message}`);
  }
  const identity = renderRetainedScanIdentity(vocabulary);
  const digest = `sha256:${createHash("sha256").update(identity).digest("hex")}`;
  if (digest !== frozenRetainedScanPolicyDigest) {
    throw new Error(`retained-secret scan identity ${digest} does not reproduce the frozen policy digest ${frozenRetainedScanPolicyDigest}`);
  }
  assertVocabularyCoverage(vocabulary);
  const rules = vocabulary.valueRules.map((rule) => `\t\t{pattern: regexp.MustCompile(${goPattern(rule)}), replacement: ${goString(rule.replacement)}},`);
  const tokens = vocabulary.tokens.map((token) => `\t${goString(token)},`);
  const terms = vocabulary.retainedScan.jsonMemberTerms.map(({ text, excluded }) =>
    `\t{text: ${goString(text)}${excluded === null ? "" : `, excluded: '${excluded}'`}},`);
  const contexts = retainedScanContexts(vocabulary).map(({ prefix, excluded, terminator }) =>
    `\t{Prefix: ${goString(prefix)}, Excluded: ${goString(excluded)}, Terminator: ${goString(terminator)}},`);
  return `// Code generated by tools/harness/generated-artifacts/generate-foundation-schema-validators.mjs; DO NOT EDIT.
package harnessredact

import "regexp"

// Marker is the single retained redaction marker owned by
// tools/harness_redaction_manifest.json (TH-HARNESS-REQ-617).
const Marker = ${goString(vocabulary.marker)}

// structuredSecretKeyTokens is the closed TH-HARNESS-REQ-604 token set.
var structuredSecretKeyTokens = []string{
${tokens.join("\n")}
}

var sensitiveCLIFlagPattern = regexp.MustCompile(${goString(`(?i)${vocabulary.cliFlagPattern}`)})

var wholeStringAssignmentPattern = regexp.MustCompile(${goString(vocabulary.wholeStringAssignment)})

// scanMemberTerms are the retained-secret scan's JSON-member terms. The
// TH-HARNESS-REQ-604 closure decides key membership in code from them.
var scanMemberTerms = []scanTerm{
${terms.join("\n")}
}

var rules = compiledManifest{
\treplacement: Marker,
\tvalueRules: []compiledRule{
${rules.join("\n")}
\t},
}

// RetainedScanPolicyIdentity is the retained-secret scan's frozen identity
// text. Its SHA-256 digest is ${frozenRetainedScanPolicyDigest}.
const RetainedScanPolicyIdentity = ${goString(identity)}

// RetainedScanForbiddenNames matches retained file names that are never safe.
const RetainedScanForbiddenNames = ${goString(forbiddenNamesProgram(vocabulary))}

// RetainedScanContexts are the syntactic contexts the retained-secret scan
// checks; a value in each is safe only when it begins with Marker.
var RetainedScanContexts = []RetainedScanContext{
${contexts.join("\n")}
}
`;
}
