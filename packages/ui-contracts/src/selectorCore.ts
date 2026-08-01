declare const stableTestIdBrand: unique symbol;

export type StableTestId = string & {
  readonly [stableTestIdBrand]: "StableTestId";
};

export type WorkbookInspectorPanelId =
  | "details"
  | "relationships"
  | "evidence"
  | "history"
  | "workflow";

export function rowCellTestId(recordId: string, fieldKey: string): string {
  return recordFieldTestId("row", recordId, fieldKey);
}

export function rowInspectorFieldTestId(
  recordId: string,
  fieldKey: string,
): string {
  return recordFieldTestId("row", recordId, fieldKey, "inspector");
}

export function rowInspectButtonTestId(recordId: string): string {
  return recordTestId("row", recordId, "inspect");
}

export function dataTestIdPrefixSelector(testIdPrefix: string): string {
  return `[data-testid^="${cssAttributeValue(
    requireNonEmptySelectorValue(testIdPrefix, "data-testid prefix"),
  )}"]`;
}

export function dataTestIdSelector(testId: string): string {
  return `[data-testid="${cssAttributeValue(
    requireNonEmptySelectorValue(testId, "data-testid"),
  )}"]`;
}

export function requireFieldKey(value: string): string {
  const encoded = encodeSelectorSegment(value, "field_key");
  if (!/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(value)) {
    throw new Error(`Invalid field_key selector token: ${value}`);
  }
  return encoded;
}

export function requireFeatureGroupKey(value: string): string {
  const encoded = encodeSelectorSegment(value, "feature_group_key");
  if (!/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(value)) {
    throw new Error(`Invalid feature_group_key selector token: ${value}`);
  }
  return encoded;
}

export function requireRecordId(value: string): string {
  return encodeSelectorSegment(value, "record_id");
}

export function requireItemRef(value: string): string {
  return encodeSelectorSegment(value, "item_ref");
}

export function encodeSelectorSegment(value: string, label: string): string {
  return encodeURIComponent(requireNonEmptySelectorValue(value, label));
}

export function requireNonEmptySelectorValue(
  value: string,
  label: string,
): string {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`Invalid ${label} selector token: ${value}`);
  }
  return value;
}

function requireSelectorToken<T extends string>(
  tokens: Readonly<Record<T, string>>,
  value: T,
  label: string,
): string {
  const token = tokens[value];
  if (token === undefined) {
    throw new Error(`Invalid ${label} token: ${String(value)}`);
  }
  return token;
}

export function stableSelectorTokenTestId<T extends string>(
  tokens: Readonly<Record<T, string>>,
  value: T,
  label: string,
): StableTestId {
  return stableTestId(requireSelectorToken(tokens, value, label));
}

export function semanticSelectorTestId<T extends string>(
  tokens: Readonly<Record<T, string>>,
  value: T,
  label: string,
): StableTestId {
  return stableSelectorTokenTestId(tokens, value, label);
}

export function stableTestId(value: string): StableTestId {
  return value as StableTestId;
}

export function encodedTestId(
  prefix: string,
  value: string,
  label: string,
): string {
  return `${prefix}-${encodeSelectorSegment(value, label)}`;
}

export function userScopedTestId(prefix: string, userId: string): string {
  return encodedTestId(prefix, userId, "user_id");
}

export function stableEncodedTestId(
  prefix: string,
  value: string,
  label: string,
): StableTestId {
  return stableTestId(encodedTestId(prefix, value, label));
}

function suffixedTestId(base: string, suffix?: string): string {
  return suffix === undefined ? base : `${base}-${suffix}`;
}

export function tokenScopedTestId(
  prefix: string,
  token: string,
  suffix?: string,
): string {
  return suffixedTestId(`${prefix}-${token}`, suffix);
}

export function recordTestId(
  prefix: string,
  recordId: string,
  suffix?: string,
): string {
  return tokenScopedTestId(prefix, requireRecordId(recordId), suffix);
}

export function recordFieldTestId(
  prefix: string,
  recordId: string,
  fieldKey: string,
  suffix?: string,
): string {
  return tokenScopedTestId(
    recordTestId(prefix, recordId),
    requireFieldKey(fieldKey),
    suffix,
  );
}

export function itemRefTestId(
  prefix: string,
  itemRef: string,
  suffix?: string,
): string {
  return tokenScopedTestId(prefix, requireItemRef(itemRef), suffix);
}

export function requireClosedToken<T extends string>(
  tokens: readonly T[],
  value: T,
  label: string,
): T {
  if ((tokens as readonly string[]).includes(value)) {
    return value;
  }
  throw new Error(`Invalid ${label} token: ${String(value)}`);
}

export function cssAttributeValue(value: string): string {
  return value
    .replace(/\\/gu, "\\\\")
    .replace(/\n/gu, "\\a ")
    .replace(/\r/gu, "\\d ")
    .replace(/\f/gu, "\\c ")
    .replace(/"/gu, '\\"');
}
