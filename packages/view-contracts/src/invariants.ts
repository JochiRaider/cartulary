// All contract adapters use one owner-neutral error envelope so callers receive
// stable source and invariant context independent of internal module layout.
export function viewContractInvariant(source: string, detail: string): never {
  throw new Error(`View contract invariant failed: ${source} ${detail}`);
}

export function viewContractSourceInvariant(
  source: string,
  path: string,
  reason: string,
): never {
  throw new Error(
    `View contract source validation failed: ${source} path=${path} reason=${reason}`,
  );
}

export function viewRowInvariant(source: string, detail: string): never {
  throw new Error(`View row invariant failed: ${source} ${detail}`);
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function hasOwn(object: Record<string, unknown>, key: string): boolean {
  return Object.hasOwn(object, key);
}

export function requireContractObject(
  value: unknown,
  source: string,
  label: string,
): Record<string, unknown> {
  if (!isRecord(value)) {
    viewContractInvariant(source, `${label} must be an object`);
  }
  return value;
}

export function requireContractBoolean(
  value: unknown,
  source: string,
  label: string,
): boolean {
  if (typeof value !== "boolean") {
    viewContractInvariant(source, `${label} must be a boolean`);
  }
  return value;
}

export function requireEnumValue<T extends string>(
  value: unknown,
  allowed: readonly T[],
  source: string,
  label: string,
): T {
  if (typeof value !== "string" || !allowed.includes(value as T)) {
    viewContractInvariant(
      source,
      `${label} must be one of ${allowed.join("|")}`,
    );
  }
  return value as T;
}

export function requireStableKey(
  value: unknown,
  source: string,
  label: string,
): string {
  if (typeof value !== "string" || value.trim() === "") {
    viewContractInvariant(source, `${label} must be a non-empty string`);
  }
  return value;
}

export function stableKeySet(
  values: readonly string[],
  source: string,
  label: string,
): ReadonlySet<string> {
  const keys = new Set<string>();
  for (const [index, value] of values.entries()) {
    const fieldKey = requireStableKey(value, source, `${label}[${index + 1}]`);
    if (keys.has(fieldKey)) {
      viewContractInvariant(source, `${label} duplicate field_key ${fieldKey}`);
    }
    keys.add(fieldKey);
  }
  return keys;
}

export function stableKeyList(
  value: unknown,
  source: string,
  label: string,
): readonly string[] {
  if (value === undefined) {
    return Object.freeze([]);
  }
  if (!Array.isArray(value)) {
    viewContractInvariant(source, `${label} must be an array`);
  }
  const keys = stableKeySet(value, source, label);
  return Object.freeze([...keys]);
}

export function stableKeyMatrix(
  value: unknown,
  source: string,
  label: string,
): readonly (readonly string[])[] {
  if (!Array.isArray(value)) {
    viewContractInvariant(source, `${label} must be an array`);
  }
  return Object.freeze(
    value.map((item, index) =>
      stableKeyList(item, source, `${label}[${index + 1}]`),
    ),
  );
}

export function unionKeySet(
  ...sets: readonly ReadonlySet<string>[]
): ReadonlySet<string> {
  const keys = new Set<string>();
  for (const set of sets) {
    for (const key of set) {
      keys.add(key);
    }
  }
  return keys;
}

export function requireInspectorKey(
  value: unknown,
  source: string,
  label: string,
): string {
  const key = requireStableKey(value, source, label);
  if (!/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(key)) {
    viewContractInvariant(
      source,
      `${label} must be ASCII lower snake or dotted key`,
    );
  }
  return key;
}

export function requireFieldKey(
  value: unknown,
  source: string,
  label: string,
): string {
  const key = requireStableKey(value, source, label);
  if (!/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(key)) {
    viewContractInvariant(source, `${label} must be a stable field_key`);
  }
  return key;
}

export function truthMap(
  values: readonly string[],
): Readonly<Record<string, true>> {
  return Object.freeze(
    Object.fromEntries(values.map((value) => [value, true])) as Record<
      string,
      true
    >,
  );
}
