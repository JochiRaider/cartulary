// This runs after Markdown tokenization and before Vue compilation. Escaping
// braces in Markdown source alone is insufficient: Markdown removes the escape,
// and Vue would then interpret the restored braces as an expression.
export function renderLiteralText(value: string): string {
  return value.replace(/[&<>"'{}]/g, (character) => {
    switch (character) {
      case "&":
        return "&amp;";
      case "<":
        return "&lt;";
      case ">":
        return "&gt;";
      case '"':
        return "&quot;";
      case "'":
        return "&#39;";
      case "{":
        return "&#123;";
      case "}":
        return "&#125;";
      default:
        throw new Error("unreachable literal-text character");
    }
  });
}
