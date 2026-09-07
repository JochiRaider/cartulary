import { type RefObject, useSyncExternalStore } from "react";
import { AccountDialog } from "./AccountDialog";
import type { DeploymentUsersController } from "./deploymentUsersModel";
import { primaryButtonStyle, secondaryButtonStyle } from "./landingAdminStyles";

const keys = ["stay", "discard", "save"] as const;
export function DeploymentUserLeaveDialog({
  controller,
  fallbackFocusRef,
}: {
  controller: DeploymentUsersController;
  fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  if (!state.leavePrompt) return null;
  return (
    <AccountDialog
      labelledBy="deployment-user-leave-title"
      describedBy="deployment-user-leave-description"
      fallbackFocusRef={fallbackFocusRef}
      onClose={() => {
        void controller.resolveLeave("stay");
      }}
      style={{
        width: "min(100%, 32rem)",
        overflow: "auto",
        borderRadius: "var(--ct-rounded-md)",
        padding: "var(--ct-spacing-lg)",
        background: "var(--ct-colors-surface-1)",
        color: "var(--ct-colors-ink)",
        boxShadow: "var(--ct-elevation-panel)",
      }}
    >
      {(dismiss) => (
        <>
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
                type="button"
                disabled={key === "save" && !controller.canSave()}
                style={
                  key === "save" ? primaryButtonStyle : secondaryButtonStyle
                }
                onClick={() => {
                  if (key === "stay") dismiss();
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
              Saving the captured changes. Newer edits will keep this dialog
              open.
            </p>
          ) : null}
          {state.operation.kind === "rejected" ||
          state.operation.kind === "uncertain" ? (
            <p role="status">
              The save is not confirmed. Stay to review and recover the action.
            </p>
          ) : null}
        </>
      )}
    </AccountDialog>
  );
}
