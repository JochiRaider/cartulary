import { describe, expect, it, vi } from "vitest";
import {
  type ClientTransactionCrypto,
  createClientTransactionId,
} from "./clientTransactionId";

const uuidV4Pattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u;

describe("createClientTransactionId", () => {
  it("retains the prefix and uses the platform UUID generator", () => {
    const randomUUID = vi.fn(() => "123e4567-e89b-42d3-a456-426614174000");
    const provider = {
      getRandomValues: vi.fn(),
      randomUUID,
    } as unknown as ClientTransactionCrypto;

    expect(createClientTransactionId("timeline-client", provider)).toBe(
      "timeline-client-123e4567-e89b-42d3-a456-426614174000",
    );
    expect(randomUUID).toHaveBeenCalledOnce();
  });

  it("uses RFC 4122 v4 formatting with secure random bytes as fallback", () => {
    let seed = 0;
    const provider = {
      getRandomValues: <T extends ArrayBufferView | null>(array: T): T => {
        if (array instanceof Uint8Array) {
          for (let index = 0; index < array.length; index += 1) {
            array[index] = seed;
            seed += 1;
          }
        }
        return array;
      },
    } as ClientTransactionCrypto;

    const first = createClientTransactionId("timeline-client", provider);
    const second = createClientTransactionId("timeline-client", provider);

    expect(first.replace("timeline-client-", "")).toMatch(uuidV4Pattern);
    expect(second.replace("timeline-client-", "")).toMatch(uuidV4Pattern);
    expect(second).not.toBe(first);
  });

  it("fails locally when Web Crypto is unavailable", () => {
    const originalCrypto = globalThis.crypto;
    Object.defineProperty(globalThis, "crypto", {
      configurable: true,
      value: undefined,
    });
    try {
      expect(() => createClientTransactionId("timeline-client")).toThrow(
        "Secure random identifiers are unavailable. This edit was not submitted.",
      );
    } finally {
      Object.defineProperty(globalThis, "crypto", {
        configurable: true,
        value: originalCrypto,
      });
    }
  });
});
