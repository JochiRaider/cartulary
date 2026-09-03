import type { StableTestId } from "./selectorCore";
import {
  dataTestIdPrefixSelector,
  encodeSelectorSegment,
  stableTestId,
} from "./selectorCore";

import { viewScopedTestId } from "./viewSchemaSelectors";

export function savedViewFamilySelector(): string {
  return dataTestIdPrefixSelector("saved-view-");
}

export function savedViewSelectorTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-selector", viewSchemaId));
}

export function savedViewOptionTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-option", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewNameInputTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-name", viewSchemaId));
}

export function savedViewScopeSelectTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-scope", viewSchemaId));
}

export function savedViewActionMenuTriggerTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(
    viewScopedTestId("saved-view-action-menu-trigger", viewSchemaId),
  );
}

export function savedViewActionMenuTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-action-menu", viewSchemaId));
}

export function savedViewCreateButtonTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-create", viewSchemaId));
}

export function savedViewDuplicateButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-duplicate", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewUpdateButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-update", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewDeleteButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-delete", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewSetHomeButtonTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-set-home", viewSchemaId));
}

export function savedViewSetDefaultButtonTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-set-default", viewSchemaId));
}

export function savedViewModifiedTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-modified", viewSchemaId));
}

export function savedViewResetButtonTestId(
  viewSchemaId: string,
  savedViewId: string,
): StableTestId {
  return stableTestId(
    `${viewScopedTestId("saved-view-reset", viewSchemaId)}-${encodeSelectorSegment(
      savedViewId,
      "saved_view_id",
    )}`,
  );
}

export function savedViewStatusTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewScopedTestId("saved-view-status", viewSchemaId));
}
