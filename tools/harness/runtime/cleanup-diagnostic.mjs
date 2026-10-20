import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";

const schema = "cartulary.harness_cleanup_diagnostic.v4";

export function parseCleanupDiagnostic(data) {
  if (!data || data.length > 4096) return null;
  try {
    const value = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(data));
    validateSchemaSync(schema, value);
    return value;
  } catch { return null; }
}

// Aggregate failures preserve the first physical boundary without forwarding
// arbitrary messages, private protocol proof or an unbounded exception graph.
export function findCleanupDiagnostic(error) {
  const pending = [error], seen = new Set();
  for (let visited = 0; pending.length && visited < 16; visited++) {
    const current = pending.shift();
    if (!current || seen.has(current)) continue;
    seen.add(current);
    const value = current.cleanupDiagnostic;
    if (value) {
      try { validateSchemaSync(schema, value); return value; } catch { /* reject unsafe detail */ }
    }
    if (current.cause) pending.push(current.cause);
    if (Array.isArray(current.errors)) pending.push(...current.errors.slice(0, 16));
  }
  return null;
}

export function withCleanupDiagnostic(error, stage) {
  const existing = findCleanupDiagnostic(error);
  let diagnostic = existing;
  if (!diagnostic) {
    const pending = [error], seen = new Set();
    let rejection;
    for (let visited = 0; pending.length && visited < 16; visited++) {
      const current = pending.shift();
      if (!current || seen.has(current)) continue;
      seen.add(current);
      if (current.privateDiagnostic) { rejection = current.privateDiagnostic; break; }
      if (current.cause) pending.push(current.cause);
      if (Array.isArray(current.errors)) pending.push(...current.errors.slice(0, 16));
    }
    const candidate = { schema_id: schema, stage, cause: rejection?.reason ?? "internal", errno: rejection?.errno ?? 0 };
    try { validateSchemaSync(schema, candidate); diagnostic = candidate; }
    catch { diagnostic = { schema_id: schema, stage, cause: "internal", errno: 0 }; validateSchemaSync(schema, diagnostic); }
  }
  Object.defineProperty(error, "cleanupDiagnostic", { value: diagnostic, configurable: true });
  return error;
}
