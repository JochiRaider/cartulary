import { describe, expect, it } from "vitest";

import {
  assessmentsViewSchemaId,
  buildWorkbookSurfaceRegistry,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  listBuiltInWorkbookSurfaceRegistryEntries,
  listSystemWorkbookSurfaceGroups,
  listSystemWorkbookSurfaceRegistryEntries,
  listWorkbookSurfaceRegistryEntries,
  optionalStandardizedWorkbookSurfaceIds,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  requiredSystemWorkbookSurfaceIds,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
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
        .filter(
          (entry) =>
            entry.surfaceStatus !== "standardized_optional_workbook_surface",
        )
        .map((entry) => entry.viewSchemaId),
    ).toEqual(requiredIds);

    const shuffled = buildWorkbookSurfaceRegistry(
      [...entries].reverse().map((entry) => entry.contract),
    );
    expect(shuffled.map((entry) => entry.viewSchemaId)).toEqual(ids);
  });

  it("FE-U-P2-02 groups System views by stable group tokens and registry-backed IDs", () => {
    const groups = listSystemWorkbookSurfaceGroups();

    expect(groups.map((group) => group.token)).toEqual([
      "scope-assessment",
      "coordination",
      "review-learning",
      "optional-artifact-surfaces",
    ]);
    expect(
      groups.map((group) => group.entries.map((entry) => entry.viewSchemaId)),
    ).toEqual([
      [indicatorsViewSchemaId, assessmentsViewSchemaId, partiesViewSchemaId],
      [
        taskRequestsViewSchemaId,
        decisionsViewSchemaId,
        commLogViewSchemaId,
        handoffViewSchemaId,
      ],
      [statusReviewViewSchemaId, lessonViewSchemaId],
      [
        findingsViewSchemaId,
        investigativeQueriesViewSchemaId,
        forensicKeywordsViewSchemaId,
      ],
    ]);
    expect(
      groups.flatMap((group) =>
        group.entries.map((entry) => entry.contract.viewSchemaId),
      ),
    ).toEqual([
      indicatorsViewSchemaId,
      assessmentsViewSchemaId,
      partiesViewSchemaId,
      taskRequestsViewSchemaId,
      decisionsViewSchemaId,
      commLogViewSchemaId,
      handoffViewSchemaId,
      statusReviewViewSchemaId,
      lessonViewSchemaId,
      ...optionalStandardizedWorkbookSurfaceIds,
    ]);
  });

  it("FE-U-P2-02 remains keyed by stable IDs when registry labels are relabeled", () => {
    const entries = listWorkbookSurfaceRegistryEntries();
    const relabeledContracts = entries.map((entry) => ({
      ...entry.contract,
      title: `Surface ${entry.viewSchemaId}`,
    }));

    const relabeled = buildWorkbookSurfaceRegistry(relabeledContracts);

    expect(relabeled.map((entry) => entry.viewSchemaId)).toEqual(
      entries.map((entry) => entry.viewSchemaId),
    );
    expect(relabeled.map((entry) => entry.contract.title)).toEqual(
      relabeled.map((entry) => `Surface ${entry.viewSchemaId}`),
    );
    expect(relabeled.map((entry) => entry.surfaceStatus)).toEqual(
      entries.map((entry) => entry.surfaceStatus),
    );
  });

  it("FE-U-P2-02 tolerates absent optional standardized surfaces while requiring required surfaces", () => {
    const entries = listWorkbookSurfaceRegistryEntries();
    const requiredIds = [
      ...requiredBuiltInWorkbookSurfaceIds,
      ...requiredSystemWorkbookSurfaceIds,
    ];
    const requiredIdSet = new Set<string>(requiredIds);
    const optionalIdSet = new Set<string>(
      optionalStandardizedWorkbookSurfaceIds,
    );
    const requiredContracts = entries
      .filter((entry) => requiredIdSet.has(entry.viewSchemaId))
      .map((entry) => entry.contract);

    const requiredOnly = buildWorkbookSurfaceRegistry(requiredContracts);

    expect(requiredOnly.map((entry) => entry.viewSchemaId)).toEqual(
      requiredIds,
    );
    expect(
      requiredOnly.some((entry) => optionalIdSet.has(entry.viewSchemaId)),
    ).toBe(false);
    expect(() =>
      buildWorkbookSurfaceRegistry(
        requiredContracts.filter(
          (contract) =>
            contract.viewSchemaId !== requiredBuiltInWorkbookSurfaceIds[0],
        ),
      ),
    ).toThrow(/Missing workbook surface contract/);
  });

  it("FE-U-P2-02 exposes required reference-pack keys from view contracts", () => {
    const entries = listWorkbookSurfaceRegistryEntries();
    const packBoundContracts = entries.map((entry) =>
      entry.viewSchemaId === findingsViewSchemaId
        ? {
            ...entry.contract,
            requiredReferencePackKeys: ["mitre_attack_enterprise"],
          }
        : entry.contract,
    );

    expect(entries.map((entry) => entry.requiredReferencePackKeys)).toEqual(
      entries.map(() => []),
    );
    expect(
      buildWorkbookSurfaceRegistry(packBoundContracts).find(
        (entry) => entry.viewSchemaId === findingsViewSchemaId,
      )?.requiredReferencePackKeys,
    ).toEqual(["mitre_attack_enterprise"]);
  });
});
