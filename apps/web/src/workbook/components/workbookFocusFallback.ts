import {
  dataTestIdSelector,
  workbookIncidentIdentityTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";

/** Last visible shell destinations in the workbook focus fallback ladder. */
export function focusWorkbookShellFallback(): boolean {
  const activeSelector = document.querySelector<HTMLElement>(
    '[aria-label="Built-in workbook surfaces"] button[aria-current="page"]',
  );
  if (
    activeSelector &&
    activeSelector.getClientRects().length > 0 &&
    !activeSelector.matches(":disabled")
  ) {
    activeSelector.focus({ preventScroll: true });
    if (document.activeElement === activeSelector) return true;
  }
  for (const id of [
    workbookSurfacesMenuTriggerTestId(),
    workbookIncidentIdentityTestId(),
  ]) {
    const container = document.querySelector<HTMLElement>(
      dataTestIdSelector(id),
    );
    const target = container?.matches("button, [tabindex]")
      ? container
      : container?.querySelector<HTMLElement>("button, [tabindex]");
    if (
      !target ||
      target.getClientRects().length === 0 ||
      target.matches(":disabled")
    )
      continue;
    target.focus({ preventScroll: true });
    if (document.activeElement === target) return true;
  }
  return false;
}
