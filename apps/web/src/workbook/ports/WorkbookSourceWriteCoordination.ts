/** Settlement of local source work; the source owner must still verify current records and authority. */
export type WorkbookSourceWriteSettlement =
  | { readonly kind: "settled"; readonly minimumRowVersion: number }
  | {
      readonly kind: "blocked";
      readonly reason: "pending_recovery" | "uncertain_source";
    }
  | { readonly kind: "cancelled" };
