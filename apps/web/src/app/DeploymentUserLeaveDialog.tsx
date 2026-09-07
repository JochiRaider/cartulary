import { useRef, useSyncExternalStore } from "react";
import { useRegisteredOverlayNavigation } from "../shared/useRegisteredOverlayNavigation";
import type { DeploymentUsersController } from "./deploymentUsersModel";
import { primaryButtonStyle, secondaryButtonStyle } from "./landingAdminStyles";

const keys = ["stay", "discard", "save"] as const;
export function DeploymentUserLeaveDialog({
  controller,
}: {
  controller: DeploymentUsersController;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const triggerRef = useRef<HTMLElement | null>(null);
  const previouslyOpen = useRef(false);
  if (state.leavePrompt && !previouslyOpen.current)
    triggerRef.current =
      document.activeElement instanceof HTMLElement
        ? document.activeElement
        : null;
  previouslyOpen.current = state.leavePrompt;
  const navigation = useRegisteredOverlayNavigation({
    isOpen: state.leavePrompt,
    initialItemKey: "stay",
    itemKeys: keys,
    triggerRef,
    subjectKey: state.selected?.user_id ?? "",
    trapTab: true,
    reconcileItems: true,
    restoreFocusOnSubjectChange: false,
    onRequestClose: () => {
      void controller.resolveLeave("stay");
    },
  });
  if (!state.leavePrompt) return null;
  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 1200,
        display: "grid",
        placeItems: "center",
        padding: "var(--ct-spacing-md)",
        background: "var(--ct-colors-overlay-backdrop, rgba(0,0,0,0.45))",
        overflow: "auto",
      }}
    >
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="deployment-user-leave-title"
        aria-describedby="deployment-user-leave-description"
        onBlur={navigation.onOverlayBlur}
        style={{
          width: "min(100%, 32rem)",
          maxHeight: "calc(100dvh - 2rem)",
          overflow: "auto",
          borderRadius: "var(--ct-rounded-md)",
          padding: "var(--ct-spacing-lg)",
          background: "var(--ct-colors-surface-1)",
          color: "var(--ct-colors-ink)",
          boxShadow: "var(--ct-elevation-panel)",
        }}
      >
        <h2 id="deployment-user-leave-title">Unsaved user changes</h2>
        <p id="deployment-user-leave-description">
          Save or discard the current user draft before leaving. Discard does
          not cancel a pending or uncertain action.
        </p>
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "var(--ct-spacing-sm)",
          }}
        >
          {keys.map((key) => (
            <button
              key={key}
              ref={navigation.registerItem(key)}
              type="button"
              tabIndex={navigation.tabIndexFor(key)}
              disabled={key === "save" && !controller.canSave()}
              onFocus={() => navigation.onItemFocus(key)}
              onKeyDown={(event) => navigation.onItemKeyDown(event, key)}
              style={key === "save" ? primaryButtonStyle : secondaryButtonStyle}
              onClick={() => {
                if (key === "stay")
                  navigation.close({ restoreTriggerFocus: true });
                else void controller.resolveLeave(key);
              }}
            >
              {key === "stay"
                ? "Stay"
                : key === "discard"
                  ? "Discard and leave"
                  : "Save and leave"}
            </button>
          ))}
        </div>
        {state.operation.kind === "pending" ? (
          <p role="status">
            Saving the captured changes. Newer edits will keep this dialog open.
          </p>
        ) : null}
        {state.operation.kind === "rejected" ||
        state.operation.kind === "uncertain" ? (
          <p role="status">
            The save is not confirmed. Stay to review and recover the action.
          </p>
        ) : null}
      </section>
    </div>
  );
}
