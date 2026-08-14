import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const layoutDirectory = path.dirname(fileURLToPath(import.meta.url));
const workbookDirectory = path.resolve(layoutDirectory, "..");

const concreteSurfacePaths = [
  "components/AssessmentWorkbookSurface.tsx",
  "components/EntityWorkbookSurface.tsx",
  "components/GenericWorkbookSurface.tsx",
  "timeline/presentation/TimelineWorkbookView.tsx",
] as const;

function productionTypeScriptPaths(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const absolutePath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      return productionTypeScriptPaths(absolutePath);
    }
    if (
      !entry.isFile() ||
      (!entry.name.endsWith(".ts") && !entry.name.endsWith(".tsx")) ||
      entry.name.endsWith(".test.ts") ||
      entry.name.endsWith(".test.tsx")
    ) {
      return [];
    }
    return [absolutePath];
  });
}

describe("workbook layout policy", () => {
  it("keeps viewport and row-count geometry behind semantic layout rules", () => {
    for (const file of productionTypeScriptPaths(workbookDirectory)) {
      const source = readFileSync(file, "utf8");
      expect(source, file).not.toMatch(/calc\(\s*100d?vh\s*-/u);
      expect(source, file).not.toMatch(
        /(?:height|minHeight|blockSize|minBlockSize)\s*:\s*[^,\n]*(?:rowCount|rows?\.length)/u,
      );
      expect(source, file).not.toMatch(
        /(?:synthetic|filler|placeholder)(?:Grid)?Rows?\b/u,
      );
    }
  });

  it("keeps surface geometry in the shared layout owner", () => {
    for (const relativePath of concreteSurfacePaths) {
      const source = readFileSync(
        path.join(workbookDirectory, relativePath),
        "utf8",
      );
      expect(source, relativePath).toContain("WorkbookSurfaceLayout");
      expect(source, relativePath).not.toMatch(
        /\b(?:minHeight|minBlockSize)\s*:/u,
      );
    }
  });
});
