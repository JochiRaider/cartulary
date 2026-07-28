import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const sourceRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);

function authoredProductionFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      return entry.name === "generated" ? [] : authoredProductionFiles(target);
    }
    return /\.(?:ts|tsx)$/u.test(entry.name) &&
      !/\.test\.(?:ts|tsx)$/u.test(entry.name)
      ? [target]
      : [];
  });
}

describe("client transaction identifier source policy", () => {
  it("keeps browser mutation IDs independent of counters, clocks, and insecure randomness", () => {
    const violations: string[] = [];
    for (const file of authoredProductionFiles(sourceRoot)) {
      const source = readFileSync(file, "utf8");
      const transactionConstruction =
        /(?:client_txn_id|clientTxnId)[\s\S]{0,180}(?:Date\.now\(|Math\.random\()/gu;
      if (
        transactionConstruction.test(source) ||
        source.includes("clientTxnRef")
      ) {
        violations.push(path.relative(sourceRoot, file));
      }
    }
    expect(violations, violations.join("\n")).toEqual([]);
  });
});
