import type { ViewContract } from "@cartulary/view-contracts";

export type WorkbookSurfaceOwnerId =
  | "capture_timeline"
  | "entities_observations"
  | "assessments"
  | "evidence"
  | "artifacts"
  | "coordination";

export type WorkbookSurfaceRenderer =
  | "timeline"
  | "entity_hosts"
  | "entity_identities"
  | "assessment"
  | "contract";

export type WorkbookOwnerBinding =
  | "decision_supersede"
  | "evidence_lifecycle"
  | "linked_note_create"
  | "task_lifecycle";

export type WorkbookSurfacePolicy = {
  readonly collectionActions: Readonly<
    Record<string, "alias" | "party" | "record" | "risk" | "tag">
  >;
  readonly createDefaults: Readonly<Record<string, string>>;
  readonly currentUserDefaultFields: readonly string[];
  readonly ownerBindings: readonly WorkbookOwnerBinding[];
  readonly publicErrorPresentation: "owner_public_message";
};

export type WorkbookSurfacePolicyDefinition = {
  readonly ownerId: WorkbookSurfaceOwnerId;
  readonly policy: WorkbookSurfacePolicy;
  readonly renderer: WorkbookSurfaceRenderer;
  readonly viewSchemaId: string;
};

export type WorkbookSurfaceRegistration = WorkbookSurfacePolicyDefinition & {
  readonly contract: ViewContract;
};

export const defineWorkbookSurfacePolicy = (
  overrides: Partial<WorkbookSurfacePolicy> = {},
): WorkbookSurfacePolicy =>
  Object.freeze({
    collectionActions: Object.freeze(overrides.collectionActions ?? {}),
    createDefaults: Object.freeze(overrides.createDefaults ?? {}),
    currentUserDefaultFields: Object.freeze(
      overrides.currentUserDefaultFields ?? [],
    ),
    ownerBindings: Object.freeze(overrides.ownerBindings ?? []),
    publicErrorPresentation: "owner_public_message",
  });
