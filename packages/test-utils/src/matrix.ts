export function pasteMatrixText(
  matrix: readonly (readonly string[])[],
): string {
  return matrix.map((row) => row.join("\t")).join("\n");
}
