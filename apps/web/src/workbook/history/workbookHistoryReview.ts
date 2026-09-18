import type { HistoryReadScope } from "./workbookHistoryPage";

/** A read locator. Neither receipt versions nor rollback selectors belong here. */
export type HistoryReviewLocator = {
  readonly scope: HistoryReadScope;
  readonly viewSchemaId: string;
  readonly recordId: string;
  readonly changeSetId: string;
  /** Historical context only, not a claim about current row values. */
  readonly label: string;
};

export function historyReviewAuthorized(
  locator: HistoryReviewLocator,
  scope: HistoryReadScope | null,
): boolean {
  return (
    scope !== null &&
    scope.actorId === locator.scope.actorId &&
    scope.incidentId === locator.scope.incidentId &&
    scope.sessionIdentity === locator.scope.sessionIdentity
  );
}

export const historyReviewPageLimit = 3;
