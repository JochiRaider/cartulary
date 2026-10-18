import path from "node:path";
import { buildSourceSnapshot } from "./source-snapshot.mjs";
import { loadTestCatalog } from "./test-catalog.mjs";

function immutable(value, seen = new Map()) {
  if (value === null || typeof value !== "object") return value;
  if (seen.has(value)) return seen.get(value);
  if (value instanceof Map || value instanceof Set) {
    const copy = value instanceof Map ? new Map() : new Set();
    const proxy = new Proxy(copy, {
      get(target, property) {
        if (["set", "add", "delete", "clear"].includes(property)) return () => { throw new TypeError("validated facts are immutable"); };
        if (property === "valueOf") return () => proxy;
        if (property === "forEach") return (callback, receiver) => target.forEach((entry, key) => callback.call(receiver, entry, key, proxy));
        const member = Reflect.get(target, property, target);
        return ["get", "has", "keys", "values", "entries", Symbol.iterator].includes(property) ? member.bind(target) : member;
      },
      set() { throw new TypeError("validated facts are immutable"); },
      defineProperty() { throw new TypeError("validated facts are immutable"); },
      deleteProperty() { throw new TypeError("validated facts are immutable"); },
    });
    seen.set(value, proxy);
    for (const entry of value) {
      if (value instanceof Map) copy.set(immutable(entry[0], seen), immutable(entry[1], seen));
      else copy.add(immutable(entry, seen));
    }
    return proxy;
  }
  const copy = Array.isArray(value) ? [] : {};
  seen.set(value, copy);
  for (const key of Reflect.ownKeys(value)) copy[key] = immutable(value[key], seen);
  return Object.freeze(copy);
}

// An in-memory capability, never a serializable receipt or a cross-run cache.
// Every consumer proves freshness; failures and changing inputs are never facts.
export function createValidatedContext(root, ports = {}) {
  root = path.resolve(root);
  const snapshot = ports.snapshot ?? buildSourceSnapshot;
  const loadCatalog = ports.loadCatalog ?? loadTestCatalog;
  const source = immutable(snapshot(root));
  const facts = new Map();
  let invalidated = false;
  const assertCurrent = () => {
    if (invalidated || snapshot(root).digest !== source.digest) {
      invalidated = true;
      facts.clear();
      throw new Error("validated context is stale; create a new context for current executable inputs");
    }
  };
  const context = Object.freeze({
    root,
    source,
    assertCurrent,
    invalidate() { invalidated = true; facts.clear(); },
    validate(validator) {
      assertCurrent();
      if (!facts.has(validator)) {
        const result = validator(root, context);
        if (result instanceof Promise) throw new Error("validation facts must be synchronous");
        assertCurrent();
        facts.set(validator, immutable(result));
      }
      return facts.get(validator);
    },
    catalog() { return context.validate(loadCatalog); },
  });
  return context;
}
