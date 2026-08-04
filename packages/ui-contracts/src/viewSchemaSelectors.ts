import { listViewSchemaRegistryEntries } from "@cartulary/protocol-ts/view-schemas";
import type { StableTestId } from "./selectorCore";
import {
  encodeSelectorSegment,
  requireClosedToken,
  requireNonEmptySelectorValue,
  stableTestId,
} from "./selectorCore";

export type SystemViewSwitcherGroupToken =
  | "coordination"
  | "optional-artifact-surfaces"
  | "review-learning"
  | "scope-indicators";

export const systemViewSwitcherGroupTokens = [
  "scope-indicators",
  "coordination",
  "review-learning",
  "optional-artifact-surfaces",
] as const satisfies readonly SystemViewSwitcherGroupToken[];

const registeredViewSchemaIds = Object.freeze(
  new Set<string>(
    listViewSchemaRegistryEntries().map((entry) => entry.view_schema_id),
  ),
);

export function gridShellTestId(viewSchemaId: string): string {
  return viewFirstTestId(viewSchemaId, "grid-shell");
}

export function surfaceTabTestId(viewSchemaId: string): string {
  return viewScopedTestId("surface-tab", viewSchemaId);
}

export function systemViewSwitcherTriggerTestId(): StableTestId {
  return stableTestId("system-view-selector");
}

export function systemViewSwitcherMenuTestId(): StableTestId {
  return stableTestId("system-view-switcher-menu");
}

export function systemViewSwitcherGroupTestId(
  groupToken: SystemViewSwitcherGroupToken,
): StableTestId {
  return stableTestId(
    `system-view-switcher-group-${requireSystemViewSwitcherGroupToken(groupToken)}`,
  );
}

export function systemViewSwitcherOptionTestId(
  groupToken: SystemViewSwitcherGroupToken,
  viewSchemaId: string,
): StableTestId {
  return stableTestId(
    `system-view-switcher-option-${requireSystemViewSwitcherGroupToken(
      groupToken,
    )}-${requireViewSchemaId(viewSchemaId)}`,
  );
}

function requireViewSchemaId(value: string): string {
  const token = requireNonEmptySelectorValue(value, "view_schema_id");
  if (
    !/^cartulary\.view\.[a-z][a-z0-9_]*(?:\.[a-z0-9_]+)*\.v[1-9][0-9]*$/u.test(
      token,
    )
  ) {
    throw new Error(`Invalid view_schema_id selector token: ${value}`);
  }
  if (!registeredViewSchemaIds.has(token)) {
    throw new Error(`Unknown view_schema_id selector token: ${value}`);
  }
  const encoded = encodeSelectorSegment(token, "view_schema_id");
  return encoded;
}

export function viewScopedTestId(prefix: string, viewSchemaId: string): string {
  return `${prefix}-${requireViewSchemaId(viewSchemaId)}`;
}

export function viewFirstTestId(viewSchemaId: string, suffix: string): string {
  return `${requireViewSchemaId(viewSchemaId)}-${suffix}`;
}

function requireSystemViewSwitcherGroupToken(
  groupToken: SystemViewSwitcherGroupToken,
): SystemViewSwitcherGroupToken {
  return requireClosedToken(
    systemViewSwitcherGroupTokens,
    groupToken,
    "system view switcher group",
  );
}
