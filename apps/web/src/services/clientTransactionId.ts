export type ClientTransactionCrypto = Pick<Crypto, "getRandomValues"> &
  Partial<Pick<Crypto, "randomUUID">>;

function secureCrypto(): ClientTransactionCrypto {
  const provider = globalThis.crypto;
  if (
    provider === undefined ||
    typeof provider.getRandomValues !== "function"
  ) {
    throw new Error(
      "Secure random identifiers are unavailable. This edit was not submitted.",
    );
  }
  return provider;
}

function uuidV4FromSecureRandom(provider: ClientTransactionCrypto): string {
  const bytes = provider.getRandomValues(new Uint8Array(16));
  bytes[6] = ((bytes[6] ?? 0) & 0x0f) | 0x40;
  bytes[8] = ((bytes[8] ?? 0) & 0x3f) | 0x80;
  const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0"));
  return [
    hex.slice(0, 4).join(""),
    hex.slice(4, 6).join(""),
    hex.slice(6, 8).join(""),
    hex.slice(8, 10).join(""),
    hex.slice(10, 16).join(""),
  ].join("-");
}

export function createClientTransactionId(
  prefix: string,
  provider: ClientTransactionCrypto = secureCrypto(),
): string {
  const normalizedPrefix = prefix.trim();
  if (normalizedPrefix === "") {
    throw new Error("A client transaction identifier prefix is required.");
  }
  const uuid =
    typeof provider.randomUUID === "function"
      ? provider.randomUUID()
      : uuidV4FromSecureRandom(provider);
  return `${normalizedPrefix}-${uuid}`;
}
