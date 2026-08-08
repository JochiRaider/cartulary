import {
  type CartularyStatusSecondaryKind,
  cartularyDesignPresentation,
} from "@cartulary/ui-contracts";

export type WorkbookStatusSecondaryCandidate = {
  readonly kind: CartularyStatusSecondaryKind;
  readonly message: string;
  readonly surfaceId: string;
};

export function selectWorkbookStatusSecondary(
  candidates: readonly WorkbookStatusSecondaryCandidate[],
  activeSurfaceId: string,
): WorkbookStatusSecondaryCandidate | null {
  const activeCandidates = candidates.filter(
    (candidate) => candidate.surfaceId === activeSurfaceId,
  );
  for (const kind of cartularyDesignPresentation.statusSecondaryPriority) {
    const selected = activeCandidates.find(
      (candidate) => candidate.kind === kind,
    );
    if (selected !== undefined) return selected;
  }
  return null;
}
