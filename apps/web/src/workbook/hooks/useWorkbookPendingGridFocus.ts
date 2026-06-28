import { dataTestIdSelector, gridShellTestId } from "@cartulary/ui-contracts";
import { useEffect } from "react";

export function useWorkbookPendingGridFocus({
  pendingGridFocusSurface,
  setPendingGridFocusSurface,
  surface,
}: {
  readonly pendingGridFocusSurface: string | null;
  readonly setPendingGridFocusSurface: (
    update: (current: string | null) => string | null,
  ) => void;
  readonly surface: string;
}) {
  useEffect(() => {
    if (
      pendingGridFocusSurface === null ||
      pendingGridFocusSurface !== surface
    ) {
      return;
    }

    let cancelled = false;
    let timer: number | null = null;
    let attempt = 0;
    const focusFirstTarget = () => {
      if (cancelled) {
        return;
      }
      const gridShell = document.querySelector<HTMLElement>(
        dataTestIdSelector(gridShellTestId(pendingGridFocusSurface)),
      );
      const focusTarget = gridShell?.querySelector<HTMLElement>(
        '[role="row"][data-grid-record-id] [role="gridcell"] [data-testid][tabindex="0"], [role="row"][data-grid-record-id] [role="gridcell"] button:not([disabled]), [role="row"][data-grid-record-id] [role="gridcell"] input:not([disabled]), [role="row"][data-grid-record-id] [role="gridcell"] select:not([disabled]), [role="row"][data-grid-record-id] [role="gridcell"] textarea:not([disabled]), [role="row"][data-grid-record-id] [role="gridcell"] a[href]',
      );
      if (focusTarget) {
        focusTarget.focus({ preventScroll: true });
        setPendingGridFocusSurface((current) =>
          current === pendingGridFocusSurface ? null : current,
        );
        return;
      }
      attempt += 1;
      if (attempt < 30) {
        timer = window.setTimeout(focusFirstTarget, 50);
      }
    };

    timer = window.setTimeout(focusFirstTarget, 0);
    return () => {
      cancelled = true;
      if (timer !== null) {
        window.clearTimeout(timer);
      }
    };
  }, [pendingGridFocusSurface, setPendingGridFocusSurface, surface]);
}
