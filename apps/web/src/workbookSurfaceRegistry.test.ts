import { describe, expect, it } from "vitest";

import {
  buildWorkbookSurfaceRegistry,
  listBuiltInWorkbookSurfaceRegistryEntries,
  listSystemWorkbookSurfaceRegistryEntries,
  listWorkbookSurfaceRegistryEntries,
  optionalStandardizedWorkbookSurfaceIds,
  requiredBuiltInWorkbookSurfaceIds,
  requiredSystemWorkbookSurfaceIds,
} from "./workbookSurfaceRegistry";

describe("workbook surface registry", () => {
  it("FE-U-P2-02 registers built-in and system workbook surfaces by stable IDs", () => {
    const builtIns = listBuiltInWorkbookSurfaceRegistryEntries();
    expect(builtIns.map((entry) => entry.viewSchemaId)).toEqual([
      ...requiredBuiltInWorkbookSurfaceIds,
    ]);
    expect(builtIns.map((entry) => entry.surfaceKind)).toEqual(
      requiredBuiltInWorkbookSurfaceIds.map(() => "built_in_sheet"),
    );

    const systemEntries = listSystemWorkbookSurfaceRegistryEntries().filter(
      (entry) => entry.surfaceStatus === "required_system_view",
    );
    expect(systemEntries.map((entry) => entry.viewSchemaId)).toEqual([
      ...requiredSystemWorkbookSurfaceIds,
    ]);
    expect(systemEntries.map((entry) => entry.surfaceKind)).toEqual(
      requiredSystemWorkbookSurfaceIds.map(() => "system_view"),
    );

    const allIds = listWorkbookSurfaceRegistryEntries().map(
      (entry) => entry.viewSchemaId,
    );
    expect(new Set(allIds).size).toBe(allIds.length);
    for (const entry of listWorkbookSurfaceRegistryEntries()) {
      expect(entry.contract.viewSchemaId).toBe(entry.viewSchemaId);
      expect(entry.contract.title).not.toBe(entry.viewSchemaId);
    }
  });

  it("FE-U-P2-02 keeps optional standardized surfaces additive after required surfaces", () => {
    const entries = listWorkbookSurfaceRegistryEntries();
    const ids = entries.map((entry) => entry.viewSchemaId);
    const requiredIds = [
      ...requiredBuiltInWorkbookSurfaceIds,
      ...requiredSystemWorkbookSurfaceIds,
    ];
    expect(ids.slice(0, requiredIds.length)).toEqual(requiredIds);
    expect(ids.slice(requiredIds.length)).toEqual([
      ...optionalStandardizedWorkbookSurfaceIds,
    ]);

    const optionalEntries = entries.slice(requiredIds.length);
    expect(optionalEntries.map((entry) => entry.surfaceStatus)).toEqual(
      optionalStandardizedWorkbookSurfaceIds.map(
        () => "standardized_optional_workbook_surface",
      ),
    );
    expect(
      entries
        .filter((entry) => entry.surfaceStatus !== "standardized_optional_workbook_surface")
        .map((entry) => entry.viewSchemaId),
    ).toEqual(requiredIds);

    const shuffled = buildWorkbookSurfaceRegistry(
      [...entries].reverse().map((entry) => entry.contract),
    );
    expect(shuffled.map((entry) => entry.viewSchemaId)).toEqual(ids);
  });
});
