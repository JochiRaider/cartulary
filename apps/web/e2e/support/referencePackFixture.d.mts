import type { Buffer } from "node:buffer";
export function referencePackBundle(options?: { invalidManifest?: boolean }): {
  key: string;
  version: string;
  upload: { name: string; mimeType: string; buffer: Buffer };
};
