/** Declared writable string normalization; raw authoring never passes through this in-place. */
export function normalizeWorkbookAuthoringText(raw: string, contract: string) {
  const limits: Readonly<Record<string, number>> = {
    display_name_line_v1: 256,
    single_line_title_v1: 512,
    multiline_body_v1: 16384,
    party_text_v1: 256,
    locator_text_v1: 1024,
    tag_label_v1: 64,
    alias_text_v1: 256,
    reason_note_v1: 4096,
    email_address_v1: 320,
    timezone_name_v1: 128,
  };
  const limit = limits[contract];
  if (!limit)
    return { value: raw, error: "This text contract is unavailable." };
  const multiline =
    contract === "multiline_body_v1" || contract === "reason_note_v1";
  const text = multiline ? raw.replace(/\r\n?/gu, "\n") : raw;
  const value = text
    .normalize("NFC")
    .replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");
  const invalid = [...text].some((c) => {
    const n = c.codePointAt(0) ?? 0;
    return (
      (n < 32 && !(multiline && (n === 9 || n === 10))) ||
      (n >= 127 && n <= 159) ||
      (n >= 0xd800 && n <= 0xdfff)
    );
  });
  if (invalid)
    return { value, error: "Remove unsupported control characters." };
  if ([...value].length > limit)
    return { value, error: `Use at most ${limit} characters.` };
  if (
    value &&
    contract === "email_address_v1" &&
    (/\p{White_Space}/u.test(value) || !/^[^@]+@[^@]+$/u.test(value))
  )
    return { value, error: "Enter a valid email address." };
  return { value };
}

export function validWorkbookTimestamp(value: unknown): value is string {
  if (typeof value !== "string") return false;
  const m =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d{1,9})?(?:Z|([+-])(\d{2}):(\d{2}))$/u.exec(
      value,
    );
  if (!m) return false;
  const year = Number(m[1]),
    month = Number(m[2]),
    day = Number(m[3]);
  const leap = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  const days = [31, leap ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
  return (
    month >= 1 &&
    month <= 12 &&
    day >= 1 &&
    day <= (days[month - 1] ?? 0) &&
    Number(m[4]) < 24 &&
    Number(m[5]) < 60 &&
    Number(m[6]) < 60 &&
    Number(m[8] ?? 0) < 24 &&
    Number(m[9] ?? 0) < 60 &&
    !Number.isNaN(Date.parse(value))
  );
}
export const exactWorkbookReferenceId = (raw: string) =>
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u.test(raw);
