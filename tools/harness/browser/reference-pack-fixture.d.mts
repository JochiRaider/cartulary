import type { Buffer } from "node:buffer";
export function referencePackBundle(options?: {
  invalidManifest?: boolean;
  preparationDate?: string;
}): {
  key: string;
  version: string;
  baseVersion: string;
  payloadSha256: string;
  preparationDate: string;
  sha256: string;
  upload: { name: string; mimeType: string; buffer: Buffer };
};
