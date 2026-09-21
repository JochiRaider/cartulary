import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import { HistoryPageLookup } from "./HistoryPageLookup";
import {
  buildRecordRollbackTargetFromHistoryAction,
  historyTargetEqual,
  type WorkbookRecordHistoryPendingAction,
} from "./workbookHistoryItem";

const unavailable: WorkbookOperationFailure = {
  kind: "stale_target",
  message: "This action is no longer available in current history.",
};

/** Action proof stays separate from read navigation; both use the same page scanner. */
export class HistoryActionLookup extends HistoryPageLookup {
  constructor(
    options: Omit<
      ConstructorParameters<typeof HistoryPageLookup>[0],
      | "evaluate"
      | "unavailable"
      | "onPage"
      | "maxRetainedPages"
      | "retainResultPage"
    > & {
      readonly pending: WorkbookRecordHistoryPendingAction;
    },
  ) {
    super({
      ...options,
      unavailable,
      evaluate: (page) => {
        const pending = options.pending;
        if (pending.kind === "destructive") {
          const legal = page.deleted === (pending.operation === "restore");
          return {
            phase: legal ? "matched" : "unavailable",
            ...(legal ? {} : { failure: unavailable }),
          };
        }
        const item = page.items.find(
          (item) => item.history_item_ref === pending.historyItemRef,
        );
        if (!item) return null;
        const target = buildRecordRollbackTargetFromHistoryAction(
          item,
          pending.action,
        );
        const legal =
          !page.deleted &&
          item.reversible &&
          target !== null &&
          historyTargetEqual(target, pending.target);
        return {
          phase: legal ? "matched" : "unavailable",
          ...(legal ? {} : { failure: unavailable }),
        };
      },
    });
  }
}
