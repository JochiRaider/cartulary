const invalidScalar = (scalar: string) => {
  const point = scalar.codePointAt(0) ?? 0;
  return (
    point <= 31 ||
    (point >= 127 && point <= 159) ||
    (point >= 0xd800 && point <= 0xdfff)
  );
};
const normalizedLine = (input: string) =>
  input.normalize("NFC").replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");

/** Core 01 email_address_v1, required for local authentication and provisioning. */
export function validateAccountEmail(input: string) {
  const value = normalizedLine(input);
  const scalars = Array.from(value);
  const parts = value.split("@");
  return {
    value,
    error:
      scalars.length > 320 ||
      scalars.some(invalidScalar) ||
      /\p{White_Space}/u.test(value) ||
      parts.length !== 2 ||
      !parts[0] ||
      !parts[1]
        ? "Enter a valid email address."
        : null,
  };
}

/** local_password_provision_v1 counts scalars and preserves the exact input. */
export function validateProvisioningPassword(value: string): string | null {
  const scalars = Array.from(value);
  return scalars.length < 12 ||
    scalars.length > 1024 ||
    scalars.some(invalidScalar) ||
    /^\p{White_Space}*$/u.test(value)
    ? "Use 12–1024 characters, without control characters or only whitespace."
    : null;
}

export { validateDisplayName as validateAccountDisplayName } from "../shared/displayName";
