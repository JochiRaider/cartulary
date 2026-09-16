import { networkAnalysisSheetRef } from "../extensions/extensionWorkspaceIdentities";
import type { WorkbookRecoveryItem } from "../shared/workbookRecoveryNavigation";
import type { IndicatorLinkSnapshot } from "./NetworkFlowIndicatorLinkController";
export function networkFlowIndicatorRecoveryItems(
  state: IndicatorLinkSnapshot,
): readonly WorkbookRecoveryItem[] {
  if (state.hidden) return [];
  const items = new Map<string, WorkbookRecoveryItem>();
  if (state.draft) {
    const draft = state.draft;
    items.set(String(draft.workId), {
      id: String(draft.workId),
      label: "Indicator link draft",
      origin: draft.candidate.label,
      sheetRef: networkAnalysisSheetRef(),
      summary:
        draft.feedback || !draft.applicable
          ? "Review original source and target"
          : "Draft retained",
      attention: draft.feedback || !draft.applicable ? "attention" : "draft",
      order: draft.workId,
    });
  }
  if (state.attempt) {
    const attempt = state.attempt;
    const complete =
      state.settlement?.kind === "confirmed" ||
      state.settlement?.kind === "reused";
    const progress =
      state.settlement?.kind === "pending" ||
      state.settlement?.kind === "queued";
    items.set(String(attempt.workId), {
      id: String(attempt.workId),
      label: "Indicator link",
      origin: attempt.candidate.label,
      sheetRef: networkAnalysisSheetRef(),
      summary: complete
        ? "Link saved"
        : progress
          ? "Linking indicator"
          : "Review captured link request",
      attention: complete ? "completed" : progress ? "progress" : "attention",
      order: attempt.workId,
    });
  }
  return [...items.values()];
}
