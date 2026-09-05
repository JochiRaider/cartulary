import {
  type CartularyStatusSecondaryKind,
  cartularyDesignPresentation,
} from "@cartulary/ui-contracts";
import { type SheetRef, sheetRefsEqual } from "../../shared/sheetRef";

export type WorkbookStatusScope =
  | { readonly kind: "workbook" }
  | { readonly kind: "surface"; readonly sheetRef: SheetRef };

export type WorkbookStatusAction =
  | {
      readonly kind: "transaction_recovery" | "terminal_failure";
      readonly unitId: string;
    }
  | { readonly kind: "overflow" }
  | { readonly kind: "same_field_resolver"; readonly conflictKey: string }
  | { readonly kind: "session_recovery" };

export type WorkbookStatusSecondaryCandidate = {
  readonly kind: CartularyStatusSecondaryKind;
  readonly scope: WorkbookStatusScope;
  readonly count: number;
  readonly message: string;
  readonly action: WorkbookStatusAction | null;
};

export function selectWorkbookStatusSecondary(
  candidates: readonly WorkbookStatusSecondaryCandidate[],
  activeSheetRef?: SheetRef,
): WorkbookStatusSecondaryCandidate | null {
  const eligible = candidates.filter(
    (candidate) =>
      candidate.scope.kind === "workbook" ||
      (activeSheetRef !== undefined &&
        sheetRefsEqual(candidate.scope.sheetRef, activeSheetRef)),
  );
  for (const kind of cartularyDesignPresentation.statusSecondaryPriority) {
    const selected = eligible.find((candidate) => candidate.kind === kind);
    if (selected !== undefined) return selected;
  }
  return null;
}
