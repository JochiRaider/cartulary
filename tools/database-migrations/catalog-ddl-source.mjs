// Catalog extraction sees top-level DDL, never comments, string values, or
// routine bodies. Preserve offsets so ordered create/drop events remain exact.
// This lexical projection does not replace the separate migration policy check.
export function catalogDDLSource(source) {
  const parts = [];
  let start = 0;
  for (let index = 0; index < source.length;) {
    let begin = index;
    if (source.startsWith("--", index)) {
      const newline = source.indexOf("\n", index + 2);
      index = newline < 0 ? source.length : newline;
    } else if (source.startsWith("/*", index)) {
      let depth = 1;
      index += 2;
      while (index < source.length && depth > 0) {
        if (source.startsWith("/*", index)) { depth++; index += 2; }
        else if (source.startsWith("*/", index)) { depth--; index += 2; }
        else index++;
      }
      if (depth !== 0) throw new Error("Unterminated SQL block comment in migration catalog input");
    } else if (source[index] === "'" || source[index] === '"') {
      const quote = source[index++];
      const escapes = quote === "'" && /[eE]/u.test(source[begin - 1] ?? "") &&
        !/[A-Za-z0-9_$]/u.test(source[begin - 2] ?? "");
      if (escapes) begin--;
      let closed = false;
      while (index < source.length) {
        if (escapes && source[index] === "\\") index += 2;
        else if (source[index] === quote) {
          index++;
          if (source[index] === quote) index++;
          else { closed = true; break; }
        } else index++;
      }
      if (!closed) throw new Error("Unterminated SQL quoted token in migration catalog input");
    } else {
      const delimiter = source[index] === "$" && !/[A-Za-z0-9_$]/u.test(source[index - 1] ?? "")
        ? /^\$(?:[A-Za-z_][A-Za-z0-9_]*)?\$/u.exec(source.slice(index))?.[0] : undefined;
      if (!delimiter) { index++; continue; }
      const end = source.indexOf(delimiter, index + delimiter.length);
      if (end < 0) throw new Error("Unterminated SQL dollar quote in migration catalog input");
      index = end + delimiter.length;
    }
    parts.push(source.slice(start, begin), source.slice(begin, index).replace(/[^\r\n]/g, " "));
    start = index;
  }
  parts.push(source.slice(start));
  return parts.join("");
}
