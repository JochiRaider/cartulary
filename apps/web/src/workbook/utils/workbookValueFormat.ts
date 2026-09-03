export function stringifyGridValue(value: unknown): string {
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "boolean") {
    return value ? "true" : "false";
  }
  if (typeof value === "number") {
    return String(value);
  }
  if (
    value &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    "items" in value &&
    Array.isArray(value.items)
  ) {
    return value.items
      .map((item) => {
        if (!item || typeof item !== "object" || Array.isArray(item)) {
          return null;
        }
        return "display_text" in item && typeof item.display_text === "string"
          ? item.display_text
          : "raw_text" in item && typeof item.raw_text === "string"
            ? item.raw_text
            : null;
      })
      .filter((item): item is string => item !== null)
      .join(", ");
  }
  return "";
}
