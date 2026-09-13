/** Retained reviews, requests and receipts never share mutable editor aliases. */
export function freezeWorkbookValue<T>(value: T): T {
  if (value && typeof value === "object" && !Object.isFrozen(value)) {
    Object.freeze(value);
    for (const child of Object.values(value)) freezeWorkbookValue(child);
  }
  return value;
}
