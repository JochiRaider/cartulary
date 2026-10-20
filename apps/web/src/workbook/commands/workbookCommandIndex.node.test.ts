import { describe, expect, it } from "vitest";
import {
  searchWorkbookCommands,
  type WorkbookCommandDescriptor,
} from "./workbookCommandIndex";

const command = (
  id: string,
  label: string,
  terms: readonly string[] = [],
  family: WorkbookCommandDescriptor["family"] = "View",
): WorkbookCommandDescriptor => ({
  id,
  label,
  terms,
  family,
  targetKind: "shell",
  availability: () => null,
  invoke: () => true,
});
describe("command metadata search", () => {
  it("normalizes NFC and Unicode whitespace with deterministic match tiers and family ties", () => {
    const commands = [
      command("z", "Source reader", ["open evidence"]),
      command("b", "Evidence open"),
      command("a", "Open evidence"),
      command("c", "Open evidence", [], "Inspect"),
    ];
    expect(
      searchWorkbookCommands(commands, "OPEN\u2003evidence").map(
        (item) => item.id,
      ),
    ).toEqual(["c", "a", "b", "z"]);
    expect(
      searchWorkbookCommands(
        [command("accent", "Résumé")],
        "re\u0301sume\u0301",
      ).map((item) => item.id),
    ).toEqual(["accent"]);
    expect(searchWorkbookCommands(commands, "open absent")).toEqual([]);
  });
  it("breaks equal labels by stable identity and never invokes actions during matching", () => {
    const commands = [command("b", "Inspect"), command("a", "Inspect")];
    expect(searchWorkbookCommands(commands, "").map((item) => item.id)).toEqual(
      ["a", "b"],
    );
  });
});
