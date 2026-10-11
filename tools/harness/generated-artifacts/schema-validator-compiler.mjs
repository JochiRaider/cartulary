import { readFileSync } from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import Ajv2020 from "ajv/dist/2020.js";
import standaloneCode from "ajv/dist/standalone/index.js";
import equalRuntime from "ajv/dist/runtime/equal.js";
import { _ } from "ajv/dist/compile/codegen/index.js";

export function validatorFilename(reference) {
  const name = reference.split(/[^A-Za-z0-9]+/u).filter(Boolean)
    .map((part) => `${part[0].toUpperCase()}${part.slice(1)}`).join("");
  return `validate${name}.cjs`;
}

function inlineRuntimeHelpers(source, modules) {
  const helper = `
function cartularyUcs2Length(value) {
  let length = 0;
  for (let index = 0; index < value.length; index += 1) {
    length += 1;
    const code = value.charCodeAt(index);
    if (code >= 0xd800 && code <= 0xdbff && index + 1 < value.length) {
      const next = value.charCodeAt(index + 1);
      if (next >= 0xdc00 && next <= 0xdfff) index += 1;
    }
  }
  return length;
}
`;
  let rewritten = source.replace(
    /const (func\d+) = require\("ajv\/dist\/runtime\/ucs2length"\)\.default;/gu,
    "const $1 = cartularyUcs2Length;",
  );
  if (rewritten.includes('require("ajv/dist/runtime/equal").default')) {
    const runtimeRequire = createRequire(import.meta.resolve("ajv/dist/runtime/equal.js"));
    const license = readFileSync(path.join(path.dirname(runtimeRequire.resolve("fast-deep-equal")), "LICENSE"), "utf8");
    // Embed the pinned validator's own equality function, preserving its exact
    // semantics and license without imposing an installed-package dependency.
    rewritten = `/* ${license.replaceAll("*/", "* /")} */\nconst cartularyEqual = (${equalRuntime.default.toString()});\n${rewritten.replaceAll('require("ajv/dist/runtime/equal").default', "cartularyEqual")}`;
  }
  // Only the generated sibling closure may remain. Bootstrap does not need AJV
  // or installed packages, and arbitrary compiler runtime imports fail closed.
  const checked = rewritten.replace(/\brequire\(("(?:[^"\\]|\\.)*")\)/gu, (call, argument) =>
    modules.has(JSON.parse(argument)) ? "generatedValidator" : call);
  if (/\brequire\s*\(/u.test(checked) || /^\s*import\s/mu.test(checked)) {
    throw new Error(`foundation validator generation produced a runtime dependency: ${checked.match(/require\([^)]*\)/u)?.[0] ?? "import"}`);
  }
  return `${helper}\n${rewritten}`;
}

// This adapter is deliberately confined to the pinned generation-time compiler.
// AJV's standalone emitter accepts code for its external scope values. Link only
// exact compiled-function identities; never infer equivalence from schema text
// or rewrite validation bodies. Anonymous/local references retain AJV's closure.
function linkCompiledReferences(validators) {
  const modules = new Map([...validators].map(([reference, validate]) => [validate, validatorFilename(reference)]));
  const seen = new Set(), linked = new Set();
  function visit(validate) {
    if (seen.has(validate)) return;
    seen.add(validate);
    if (!validate?.source?.scopeValues) throw new Error("unsupported standalone validator source");
    for (const names of Object.values(validate.source.scopeValues)) for (const name of names) {
      const ref = name.value?.ref;
      const wrapper = name.prefix === "root" || name.prefix === "wrapper";
      const dependency = name.prefix === "validate" ? ref : wrapper ? ref?.validate : null;
      if (typeof dependency !== "function") continue;
      const file = modules.get(dependency);
      if (file && !linked.has(name)) {
        if (name.value.code) throw new Error("unexpected compiled reference code");
        const requireCode = _`require(${`./${file}`})`;
        name.value.code = wrapper ? _`{validate: ${requireCode}}` : requireCode;
        linked.add(name);
      }
      visit(dependency);
    }
  }
  validators.forEach(visit);
}

// Registered roots form a lazy generated module graph. Small definition
// fragments stay in their compiled closure: separate modules for those cost
// more than their duplicated code. The ordinary module loader owns sharing.
export function compileStandaloneValidators(schemas) {
  const ajv = new Ajv2020({
    // Runtime validation needs a rejection witness, not every possible error.
    // Stop once rejection is established, including within failed union arms.
    allErrors: false,
    code: { esm: false, source: true },
    inlineRefs: false,
    strict: false,
    validateFormats: false,
    validateSchema: true,
  });
  for (const schema of schemas) ajv.addSchema(schema, schema.$id);
  const references = new Set(schemas.map((schema) => schema.$id));
  const filenames = new Set();
  const validators = new Map([...references].map((reference) => {
    const file = validatorFilename(reference);
    if (filenames.has(file)) throw new Error(`colliding generated validator filename: ${file}`);
    filenames.add(file);
    const validate = ajv.getSchema(reference);
    if (!validate) throw new Error(`unresolved generated validator: ${reference}`);
    return [reference, validate];
  }));
  linkCompiledReferences(validators);
  const modules = new Set([...filenames].map((file) => `./${file}`));
  return new Map([...validators].map(([reference, validate]) => [
    reference, inlineRuntimeHelpers(standaloneCode(ajv, validate), modules),
  ]));
}
