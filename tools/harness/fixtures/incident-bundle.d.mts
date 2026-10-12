import type { Buffer } from "node:buffer";

export function incidentBundleFixture(status?: "active" | "closed", root?: string): {
  incidentId: string;
  key: string;
  upload: { name: string; mimeType: string; buffer: Buffer };
};
