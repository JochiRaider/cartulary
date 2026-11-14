export function atomicLocalFile(
  file: string,
  bytes: string | Uint8Array | Iterable<string | Uint8Array>,
  options?: { replace?: boolean; maximumReplacementBytes?: number },
): void;
