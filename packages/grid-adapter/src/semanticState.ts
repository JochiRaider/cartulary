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
  const primary: GridSemanticPrimaryState = input.conflicted
    ? "conflicted"
    : invalid !== undefined
      ? "invalid"
      : input.pending
        ? "pending"
        : input.active
          ? "active"
          : input.bulkSelected
            ? "bulk-selected"
            : input.readOnlyOrDerived
              ? "read-only"
              : "saved";
  const markers: GridSemanticMarker[] = [];
  const descriptions: string[] = [];

  if (primary === "conflicted") {
    const accessibleLabel = `Conflict on ${label}`;
    markers.push({ accessibleLabel, glyph: "!", kind: "conflicted" });
    descriptions.push(accessibleLabel);
  } else if (primary === "invalid" && invalid !== undefined) {
    const accessibleLabel = `Invalid ${label}: ${invalid.message}`;
    markers.push({ accessibleLabel, glyph: "×", kind: "invalid" });
    descriptions.push(accessibleLabel);
  } else if (primary === "pending") {
    const accessibleLabel = `Pending ${label}`;
    markers.push({ accessibleLabel, glyph: "…", kind: "pending" });
    descriptions.push(accessibleLabel);
  } else if (primary === "read-only") {
    const accessibleLabel = `Read-only ${label}`;
    markers.push({ accessibleLabel, glyph: "◇", kind: "read-only" });
    descriptions.push(accessibleLabel);
  }

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
    stateIds: [
      primary,
      ...(input.active && primary !== "active" ? ["active"] : []),
      ...(input.bulkSelected && primary !== "bulk-selected"
        ? ["bulk-selected"]
        : []),
      ...(input.inspectorActive ? ["inspector-active"] : []),
      ...(input.readOnlyOrDerived && primary !== "read-only"
        ? ["read-only"]
        : []),
      ...(input.stale ? ["stale"] : []),
    ],
  };
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
