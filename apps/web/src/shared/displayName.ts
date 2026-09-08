/** Shared client projection of display_name_line_v1; server normalization is authoritative. */
export function validateDisplayName(input: string): {
  value: string;
  error: string | null;
} {
  const value = input
    .normalize("NFC")
    .replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");
  const scalars = Array.from(value);
  const invalid = scalars.some((scalar) => {
    const point = scalar.codePointAt(0) ?? 0;
    return (
      point <= 31 ||
      (point >= 127 && point <= 159) ||
      (point >= 0xd800 && point <= 0xdfff)
    );
  });
  return {
    value,
    error:
      value === ""
        ? "Enter a display name."
        : invalid
          ? "Remove control characters from the display name."
          : scalars.length > 256
            ? "Use 256 characters or fewer."
            : null,
  };
}
