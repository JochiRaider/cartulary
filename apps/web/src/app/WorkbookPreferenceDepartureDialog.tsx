import type { RefObject } from "react";
import type { WorkbookPreferenceController } from "../workbook/preferences/WorkbookPreferenceController";
import { useWorkbookPreferencesSnapshot } from "../workbook/preferences/WorkbookPreferencesPanel";
import { AccountDialog } from "./AccountDialog";
import { secondaryButtonStyle } from "./landingAdminStyles";
export function WorkbookPreferenceDepartureDialog({
  controller,
  fallbackFocusRef,
}: {
  controller: WorkbookPreferenceController;
  fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
}) {
  const state = useWorkbookPreferencesSnapshot(controller);
  if (!state?.departure) return null;
  return (
    <AccountDialog
      style={{
        display: "grid",
        gap: "var(--ct-spacing-md)",
        padding: "var(--ct-spacing-lg)",
        background: "var(--ct-colors-surface-1)",
        color: "var(--ct-colors-ink)",
        border: "var(--ct-border-hairline)",
        maxInlineSize: "var(--ct-layout-viewBarOverlayMaxInlineSize)",
      }}
      labelledBy="preference-departure-title"
      describedBy="preference-departure-description"
      fallbackFocusRef={fallbackFocusRef}
      onClose={() => controller.resolveDeparture("stay")}
    >
      {(dismiss) => (
        <>
          <h2 id="preference-departure-title">
            Leave workbook preference recovery?
          </h2>
          <p id="preference-departure-description">
            Leaving forgets local pending or uncertain preference work for this
            incident. It cannot cancel a server request or establish whether an
            uncertain update succeeded.
          </p>
          <button type="button" style={secondaryButtonStyle} onClick={dismiss}>
            Stay
          </button>
          <button
            type="button"
            style={secondaryButtonStyle}
            onClick={() => controller.resolveDeparture("discard")}
          >
            Leave and forget recovery
          </button>
        </>
      )}
    </AccountDialog>
  );
}
