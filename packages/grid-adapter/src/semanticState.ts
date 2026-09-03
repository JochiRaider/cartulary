import type { GridSemanticStateInput } from "./core";

export type GridSemanticPrimaryState =
  | "conflicted"
  | "invalid"
  | "pending"
  | "active"
  | "bulk-selected"
  | "read-only"
  | "saved";

export type GridSemanticMarker = {
  readonly accessibleLabel: string;
  readonly glyph: string;
  readonly kind: "conflicted" | "invalid" | "pending" | "read-only" | "stale";
};

export type GridResolvedSemanticState = {
  readonly description: string | undefined;
  readonly markers: readonly GridSemanticMarker[];
  readonly primary: GridSemanticPrimaryState;
  readonly stateIds: readonly string[];
};

export function mergeGridSemanticState(
  owner: GridSemanticStateInput | undefined,
  adapter: GridSemanticStateInput,
): GridSemanticStateInput {
  return {
    active: owner?.active === true || adapter.active === true,
    bulkSelected: owner?.bulkSelected === true || adapter.bulkSelected === true,
    conflicted: owner?.conflicted === true || adapter.conflicted === true,
    inspectorActive:
      owner?.inspectorActive === true || adapter.inspectorActive === true,
    invalid:
      owner?.invalid !== undefined && owner.invalid !== false
        ? owner.invalid
        : adapter.invalid,
    pending: owner?.pending === true || adapter.pending === true,
    readOnlyOrDerived:
      owner?.readOnlyOrDerived === true || adapter.readOnlyOrDerived === true,
    saved: owner?.saved !== false && adapter.saved !== false,
    stale: owner?.stale === true || adapter.stale === true,
  };
}

export function resolveGridSemanticState(
  input: GridSemanticStateInput,
  label: string,
): GridResolvedSemanticState {
  const invalid = input.invalid === false ? undefined : input.invalid;
  const primary = resolvePrimaryState(input, invalid !== undefined);
  const markers = primaryMarker(primary, invalid?.message, label);
  const descriptions = markers.map((marker) => marker.accessibleLabel);
  if (input.stale) {
    const accessibleLabel = `Stale ${label}; refresh required`;
    markers.push({ accessibleLabel, glyph: "↻", kind: "stale" });
    descriptions.push(accessibleLabel);
  }
  if (input.inspectorActive)
    descriptions.push(`${label} is open in the inspector`);
  if (input.bulkSelected)
    descriptions.push(`${label} is selected for bulk actions`);

  return {
    description:
      descriptions.length === 0 ? undefined : descriptions.join(". "),
    markers,
    primary,
    stateIds: resolveStateIds(input, primary),
  };
}

function resolvePrimaryState(
  input: GridSemanticStateInput,
  invalid: boolean,
): GridSemanticPrimaryState {
  const ordered: readonly [boolean | undefined, GridSemanticPrimaryState][] = [
    [input.conflicted, "conflicted"],
    [invalid, "invalid"],
    [input.pending, "pending"],
    [input.active, "active"],
    [input.bulkSelected, "bulk-selected"],
    [input.readOnlyOrDerived, "read-only"],
  ];
  return ordered.find(([enabled]) => enabled === true)?.[1] ?? "saved";
}

function primaryMarker(
  primary: GridSemanticPrimaryState,
  invalidMessage: string | undefined,
  label: string,
): GridSemanticMarker[] {
  if (primary === "conflicted") {
    return [marker(`Conflict on ${label}`, "!", "conflicted")];
  }
  if (primary === "invalid" && invalidMessage !== undefined) {
    return [marker(`Invalid ${label}: ${invalidMessage}`, "×", "invalid")];
  }
  if (primary === "pending") {
    return [marker(`Pending ${label}`, "…", "pending")];
  }
  if (primary === "read-only") {
    return [marker(`Read-only ${label}`, "◇", "read-only")];
  }
  return [];
}

function marker(
  accessibleLabel: string,
  glyph: string,
  kind: GridSemanticMarker["kind"],
): GridSemanticMarker {
  return { accessibleLabel, glyph, kind };
}

function resolveStateIds(
  input: GridSemanticStateInput,
  primary: GridSemanticPrimaryState,
): readonly string[] {
  const stateIds: string[] = [primary];
  if (input.active && primary !== "active") stateIds.push("active");
  if (input.bulkSelected && primary !== "bulk-selected") {
    stateIds.push("bulk-selected");
  }
  if (input.inspectorActive) stateIds.push("inspector-active");
  if (input.readOnlyOrDerived && primary !== "read-only") {
    stateIds.push("read-only");
  }
  if (input.stale) stateIds.push("stale");
  return stateIds;
}

export function gridSemanticStateClassNames(
  scope: "cell" | "row",
  state: GridResolvedSemanticState,
): string {
  return [
    `cartulary-grid-${scope}`,
    `cartulary-grid-${scope}-state-${state.primary}`,
    ...state.stateIds.map((stateId) => `cartulary-grid-${scope}-is-${stateId}`),
  ].join(" ");
}
