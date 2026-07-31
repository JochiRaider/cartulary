import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const webSourceDirectory = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);

function runtimeSourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      return runtimeSourceFiles(target);
    }
    return /\.(?:ts|tsx)$/u.test(entry.name) &&
      !entry.name.includes(".test.") &&
      !entry.name.endsWith("TestSupport.ts") &&
      !entry.name.endsWith("TestSupport.tsx")
      ? [target]
      : [];
  });
}

describe("frontend transport boundary policy", () => {
  it("allows raw fetch only in the shared transport and presigned upload adapter", () => {
    const callers = runtimeSourceFiles(webSourceDirectory).flatMap((file) => {
      const source = readFileSync(file, "utf8");
      const count = source.match(/\bfetch\s*\(/gu)?.length ?? 0;
      return count === 0 ? [] : [{ file, count }];
    });
    expect(
      callers.map(({ file, count }) => ({
        file: path.relative(webSourceDirectory, file),
        count,
      })),
    ).toEqual([
      { file: "services/httpTransport.ts", count: 2 },
      { file: "services/workbookEvidence.ts", count: 1 },
    ]);
  });

  it("keeps the raw upload call constrained to the server-issued target", () => {
    const source = readFileSync(
      path.join(webSourceDirectory, "services/workbookEvidence.ts"),
      "utf8",
    );
    expect(source).toContain("uploadTarget.href");
    expect(source).toContain('method: uploadTarget.method ?? "PUT"');
    expect(source).toContain('credentials: "omit"');
    expect(source).toContain("Object.entries(uploadTarget.headers ?? {})");
    expect(source).toContain("headers.set(key, value);");
  });
});
