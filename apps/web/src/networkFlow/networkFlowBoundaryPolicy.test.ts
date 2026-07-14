import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const networkFlowDirectory = path.dirname(fileURLToPath(import.meta.url));
const webSourceDirectory = path.resolve(networkFlowDirectory, "..");

function sourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      return sourceFiles(target);
    }
    return /\.(?:ts|tsx)$/u.test(entry.name) && !entry.name.includes(".test.")
      ? [target]
      : [];
  });
}

describe("Network Flow frontend boundary policy", () => {
  it("composes each capability through a dedicated workspace controller", () => {
    const workspace = readFileSync(
      path.join(networkFlowDirectory, "NetworkAnalysisWorkspace.tsx"),
      "utf8",
    );
    for (const controller of [
      "useNetworkFlowTableController",
      "useNetworkFlowRowsController",
      "useNetworkFlowRejectedRowsController",
      "useNetworkFlowGraphController",
      "useNetworkFlowImportController",
      "useNetworkFlowIndicatorLinkController",
      "useNetworkFlowCollaborationController",
    ]) {
      expect(workspace).toContain(`${controller}({`);
    }
  });

  it("keeps runtime schema compilation and projection adapter input out of the browser", () => {
    for (const file of sourceFiles(webSourceDirectory)) {
      const source = readFileSync(file, "utf8");
      expect(source, file).not.toContain("$defs");
      expect(source, file).not.toContain("GraphProjectionEphemeralInput");
      expect(source, file).not.toContain("graph_projection_ephemeral_input");
    }
  });

  it("keeps Network Flow wire decoding behind the generated contract adapter", () => {
    const client = readFileSync(
      path.join(networkFlowDirectory, "networkFlowClient.ts"),
      "utf8",
    );
    expect(client).toContain("../services/networkFlowContractAdapter");
    expect(client).not.toContain("@cartulary/protocol-ts");
  });
});
