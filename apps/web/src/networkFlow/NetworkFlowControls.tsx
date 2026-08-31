import {
  type CartularyDesignTokenVarName,
  cartularyDesignTokenReference as token,
} from "@cartulary/ui-contracts";
import {
  type ButtonHTMLAttributes,
  forwardRef,
  type HTMLAttributes,
  type InputHTMLAttributes,
  type ReactNode,
  type SelectHTMLAttributes,
} from "react";

export const networkFlowChromeRootClassName = "network-flow-chrome";

export type NetworkFlowButtonVariant =
  | "primary"
  | "secondary"
  | "danger"
  | "ghost"
  | "mode";

export type NetworkFlowButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  readonly pending?: boolean;
  readonly selected?: boolean;
  readonly variant?: NetworkFlowButtonVariant;
};

export const NetworkFlowButton = forwardRef<
  HTMLButtonElement,
  NetworkFlowButtonProps
>(function NetworkFlowButton(
  {
    children,
    className,
    disabled,
    pending = false,
    selected,
    type = "button",
    variant = "secondary",
    ...props
  },
  ref,
) {
  return (
    <button
      {...props}
      aria-busy={pending || undefined}
      aria-pressed={
        selected === undefined ||
        props["aria-pressed"] !== undefined ||
        props.role === "tab"
          ? props["aria-pressed"]
          : selected
      }
      className={joinClassNames(
        "network-flow-control network-flow-button",
        `network-flow-button--${variant}`,
        selected && "is-selected",
        className,
      )}
      data-network-flow-control="button"
      data-network-flow-variant={variant}
      disabled={disabled || pending}
      ref={ref}
      type={type}
    >
      {pending ? (
        <span aria-hidden="true" className="network-flow-pending-mark" />
      ) : null}
      {children}
    </button>
  );
});

export type NetworkFlowIconButtonProps = Omit<
  NetworkFlowButtonProps,
  "aria-label"
> & {
  readonly "aria-label": string;
};

export const NetworkFlowIconButton = forwardRef<
  HTMLButtonElement,
  NetworkFlowIconButtonProps
>(function NetworkFlowIconButton({ className, ...props }, ref) {
  return (
    <NetworkFlowButton
      {...props}
      className={joinClassNames("network-flow-icon-button", className)}
      ref={ref}
    />
  );
});

export const NetworkFlowTextInput = forwardRef<
  HTMLInputElement,
  InputHTMLAttributes<HTMLInputElement>
>(function NetworkFlowTextInput({ className, type = "text", ...props }, ref) {
  return (
    <input
      {...props}
      className={joinClassNames(
        "network-flow-control network-flow-input",
        className,
      )}
      data-network-flow-control="input"
      ref={ref}
      type={type}
    />
  );
});

export const NetworkFlowNumberInput = forwardRef<
  HTMLInputElement,
  Omit<InputHTMLAttributes<HTMLInputElement>, "type">
>(function NetworkFlowNumberInput(
  { className, inputMode = "numeric", ...props },
  ref,
) {
  return (
    <NetworkFlowTextInput
      {...props}
      className={className}
      inputMode={inputMode}
      ref={ref}
      type="number"
    />
  );
});

export const NetworkFlowSelect = forwardRef<
  HTMLSelectElement,
  SelectHTMLAttributes<HTMLSelectElement>
>(function NetworkFlowSelect({ className, ...props }, ref) {
  return (
    <select
      {...props}
      className={joinClassNames(
        "network-flow-control network-flow-select",
        className,
      )}
      data-network-flow-control="select"
      ref={ref}
    />
  );
});

export const NetworkFlowChoice = forwardRef<
  HTMLInputElement,
  InputHTMLAttributes<HTMLInputElement>
>(function NetworkFlowChoice({ className, type = "checkbox", ...props }, ref) {
  return (
    <input
      {...props}
      className={joinClassNames(
        "network-flow-control network-flow-choice",
        className,
      )}
      data-network-flow-control="choice"
      ref={ref}
      type={type}
    />
  );
});

export function NetworkFlowField({
  children,
  className,
  error,
  errorId,
  help,
  helpId,
  htmlFor,
  label,
}: {
  readonly children: ReactNode;
  readonly className?: string;
  readonly error?: string | number | undefined;
  readonly errorId?: string;
  readonly help?: string | number | undefined;
  readonly helpId?: string;
  readonly htmlFor: string;
  readonly label: ReactNode;
}) {
  return (
    <div className={joinClassNames("network-flow-field", className)}>
      <label className="network-flow-field__label" htmlFor={htmlFor}>
        {label}
      </label>
      {children}
      {help ? (
        <span className="network-flow-field__help" id={helpId}>
          {help}
        </span>
      ) : null}
      {error ? (
        <span className="network-flow-field__error" id={errorId}>
          <span aria-hidden="true">!</span> {error}
        </span>
      ) : null}
    </div>
  );
}

export function NetworkFlowActionGroup({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      {...props}
      className={joinClassNames("network-flow-action-group", className)}
    />
  );
}

export function NetworkFlowToolbar({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      {...props}
      className={joinClassNames("network-flow-toolbar", className)}
    />
  );
}

export function NetworkFlowChromeStyles() {
  return <style>{networkFlowChromeCssText}</style>;
}

function cssToken(name: CartularyDesignTokenVarName): string {
  return token(name);
}

function joinClassNames(
  ...names: ReadonlyArray<string | false | null | undefined>
): string {
  return names.filter(Boolean).join(" ");
}

export const networkFlowChromeCssText = `
.${networkFlowChromeRootClassName} {
  color-scheme: dark;
  color: ${cssToken("--ct-colors-ink")};
  font-family: ${cssToken("--ct-typography-ui-fontFamily")};
}

.${networkFlowChromeRootClassName} *,
.${networkFlowChromeRootClassName} *::before,
.${networkFlowChromeRootClassName} *::after {
  box-sizing: border-box;
}

.${networkFlowChromeRootClassName} .network-flow-control {
  font: inherit;
}

.${networkFlowChromeRootClassName} .network-flow-button {
  min-block-size: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: ${cssToken("--ct-spacing-xs")};
  border: ${cssToken("--ct-component-button-secondary-border")};
  border-radius: ${cssToken("--ct-component-button-secondary-rounded")};
  padding: ${cssToken("--ct-component-button-secondary-padding")};
  background: ${cssToken("--ct-component-button-secondary-backgroundColor")};
  color: ${cssToken("--ct-component-button-secondary-textColor")};
  font-family: ${cssToken("--ct-typography-button-fontFamily")};
  font-size: ${cssToken("--ct-typography-button-fontSize")};
  font-weight: ${cssToken("--ct-typography-button-fontWeight")};
  line-height: ${cssToken("--ct-typography-button-lineHeight")};
  letter-spacing: ${cssToken("--ct-typography-button-letterSpacing")};
  white-space: nowrap;
  cursor: pointer;
}

.${networkFlowChromeRootClassName} .network-flow-button:hover:not(:disabled) {
  border-color: ${cssToken("--ct-colors-hairline-strong")};
  background: ${cssToken("--ct-colors-surface-3")};
}

.${networkFlowChromeRootClassName} .network-flow-button:focus-visible,
.${networkFlowChromeRootClassName} .network-flow-input:focus-visible,
.${networkFlowChromeRootClassName} .network-flow-select:focus-visible,
.${networkFlowChromeRootClassName} .network-flow-choice:focus-visible,
.${networkFlowChromeRootClassName} .network-flow-list-action:focus-visible {
  outline: ${cssToken("--ct-component-focus-ring-border")};
  outline-offset: ${cssToken("--ct-component-focus-ring-offset")};
}

.${networkFlowChromeRootClassName} .network-flow-button:disabled,
.${networkFlowChromeRootClassName} .network-flow-input:disabled,
.${networkFlowChromeRootClassName} .network-flow-select:disabled {
  border-color: ${cssToken("--ct-colors-hairline")};
  background: ${cssToken("--ct-colors-surface-2")};
  color: ${cssToken("--ct-colors-ink-tertiary")};
  cursor: not-allowed;
}

.${networkFlowChromeRootClassName} .network-flow-button--primary {
  border-color: ${cssToken("--ct-component-button-primary-backgroundColor")};
  border-radius: ${cssToken("--ct-component-button-primary-rounded")};
  padding: ${cssToken("--ct-component-button-primary-padding")};
  background: ${cssToken("--ct-component-button-primary-backgroundColor")};
  color: ${cssToken("--ct-component-button-primary-textColor")};
}

.${networkFlowChromeRootClassName} .network-flow-button--primary:hover:not(:disabled) {
  border-color: ${cssToken("--ct-colors-accent-hover")};
  background: ${cssToken("--ct-colors-accent-hover")};
}

.${networkFlowChromeRootClassName} .network-flow-button--danger {
  border-color: ${cssToken("--ct-colors-semantic-destructive")};
  border-radius: ${cssToken("--ct-component-button-danger-rounded")};
  padding: ${cssToken("--ct-component-button-danger-padding")};
  background: ${cssToken("--ct-component-button-danger-backgroundColor")};
  color: ${cssToken("--ct-component-button-danger-textColor")};
}

.${networkFlowChromeRootClassName} .network-flow-button--danger:hover:not(:disabled) {
  background: ${cssToken("--ct-colors-surface-3")};
}

.${networkFlowChromeRootClassName} .network-flow-button--ghost,
.${networkFlowChromeRootClassName} .network-flow-button--mode {
  background: transparent;
}

.${networkFlowChromeRootClassName} .network-flow-button--mode {
  position: relative;
  border-color: transparent;
  border-radius: 0;
}

.${networkFlowChromeRootClassName} .network-flow-button--mode.is-selected,
.${networkFlowChromeRootClassName} .network-flow-button--mode[aria-selected="true"],
.${networkFlowChromeRootClassName} .network-flow-button--mode[aria-pressed="true"] {
  border-color: ${cssToken("--ct-colors-hairline-strong")};
  background: ${cssToken("--ct-colors-surface-2")};
  font-weight: 700;
  box-shadow: inset 0 -3px ${cssToken("--ct-colors-accent")};
}

.${networkFlowChromeRootClassName} .network-flow-icon-button {
  min-inline-size: 32px;
  padding-inline: ${cssToken("--ct-spacing-sm")};
}

.${networkFlowChromeRootClassName} .network-flow-pending-mark {
  inline-size: 0.55em;
  block-size: 0.55em;
  border: ${cssToken("--ct-border-strong")};
  border-color: currentColor;
  transform: rotate(45deg);
}

.${networkFlowChromeRootClassName} .network-flow-input,
.${networkFlowChromeRootClassName} .network-flow-select {
  min-block-size: 32px;
  min-inline-size: 0;
  max-inline-size: 100%;
  border: ${cssToken("--ct-component-text-input-border")};
  border-radius: ${cssToken("--ct-component-text-input-rounded")};
  padding: ${cssToken("--ct-component-text-input-padding")};
  background: ${cssToken("--ct-component-text-input-backgroundColor")};
  color: ${cssToken("--ct-component-text-input-textColor")};
  caret-color: ${cssToken("--ct-colors-accent")};
}

.${networkFlowChromeRootClassName} .network-flow-input::placeholder {
  color: ${cssToken("--ct-colors-ink-tertiary")};
}

.${networkFlowChromeRootClassName} .network-flow-input[aria-invalid="true"],
.${networkFlowChromeRootClassName} .network-flow-select[aria-invalid="true"] {
  border-color: ${cssToken("--ct-colors-semantic-destructive")};
  box-shadow: inset 3px 0 ${cssToken("--ct-colors-semantic-destructive")};
}

.${networkFlowChromeRootClassName} .network-flow-input:-webkit-autofill,
.${networkFlowChromeRootClassName} .network-flow-input:-webkit-autofill:hover,
.${networkFlowChromeRootClassName} .network-flow-input:-webkit-autofill:focus {
  -webkit-text-fill-color: ${cssToken("--ct-component-text-input-textColor")};
  caret-color: ${cssToken("--ct-colors-accent")};
  box-shadow: 0 0 0 1000px ${cssToken("--ct-component-text-input-backgroundColor")} inset;
}

.${networkFlowChromeRootClassName} .network-flow-choice {
  inline-size: 16px;
  block-size: 16px;
  margin: 0;
  accent-color: ${cssToken("--ct-colors-accent")};
}

.${networkFlowChromeRootClassName} .network-flow-choice:disabled {
  filter: grayscale(1);
  cursor: not-allowed;
}

.${networkFlowChromeRootClassName} .network-flow-field {
  min-inline-size: 0;
  display: grid;
  align-content: start;
  gap: ${cssToken("--ct-spacing-xs")};
}

.${networkFlowChromeRootClassName} .network-flow-field__label {
  color: ${cssToken("--ct-colors-ink-muted")};
  font-size: ${cssToken("--ct-typography-compact-metadata-fontSize")};
  font-weight: ${cssToken("--ct-typography-compact-metadata-fontWeight")};
  line-height: ${cssToken("--ct-typography-compact-metadata-lineHeight")};
}

.${networkFlowChromeRootClassName} .network-flow-field__help {
  color: ${cssToken("--ct-colors-ink-subtle")};
  overflow-wrap: anywhere;
}

.${networkFlowChromeRootClassName} .network-flow-field__error {
  color: ${cssToken("--ct-colors-semantic-destructive")};
  font-weight: 600;
  overflow-wrap: anywhere;
}

.${networkFlowChromeRootClassName} .network-flow-action-group,
.${networkFlowChromeRootClassName} .network-flow-toolbar,
.${networkFlowChromeRootClassName} .network-flow-pagination {
  min-inline-size: 0;
  display: flex;
  align-items: center;
  gap: ${cssToken("--ct-spacing-sm")};
  flex-wrap: wrap;
}

.${networkFlowChromeRootClassName} .network-flow-action-group {
  justify-content: flex-end;
}

.${networkFlowChromeRootClassName} .network-flow-query-band {
  min-inline-size: 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(13rem, 100%), 1fr));
  align-items: end;
  gap: ${cssToken("--ct-spacing-sm")};
  border-block-end: ${cssToken("--ct-border-hairline")};
  padding: ${cssToken("--ct-spacing-sm")} ${cssToken("--ct-spacing-md")};
  background: ${cssToken("--ct-colors-surface-2")};
}

.${networkFlowChromeRootClassName} .network-flow-query-band > .network-flow-action-group {
  align-self: end;
}

.${networkFlowChromeRootClassName} .network-flow-inline-fields {
  min-inline-size: 0;
  display: grid;
  grid-template-columns: minmax(7rem, auto) minmax(8rem, 1fr);
  gap: ${cssToken("--ct-spacing-xs")};
}

.${networkFlowChromeRootClassName} .network-flow-advanced {
  position: relative;
  min-inline-size: 0;
}

.${networkFlowChromeRootClassName} .network-flow-advanced > summary {
  min-block-size: 32px;
  display: flex;
  align-items: center;
  border: ${cssToken("--ct-component-button-secondary-border")};
  border-radius: ${cssToken("--ct-component-button-secondary-rounded")};
  padding: ${cssToken("--ct-component-button-secondary-padding")};
  background: ${cssToken("--ct-component-button-secondary-backgroundColor")};
  color: ${cssToken("--ct-component-button-secondary-textColor")};
  cursor: pointer;
}

.${networkFlowChromeRootClassName} .network-flow-advanced > summary:focus-visible {
  outline: ${cssToken("--ct-component-focus-ring-border")};
  outline-offset: ${cssToken("--ct-component-focus-ring-offset")};
}

.${networkFlowChromeRootClassName} .network-flow-advanced__editor {
  position: fixed;
  inset-block-start: auto;
  inset-block-end: ${cssToken("--ct-spacing-lg")};
  inset-inline: max(${cssToken("--ct-spacing-lg")}, calc((100vw - 50rem) / 2));
  z-index: 4;
  inline-size: auto;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(11rem, 100%), 1fr));
  align-items: end;
  gap: ${cssToken("--ct-spacing-sm")};
}

.${networkFlowChromeRootClassName} .network-flow-filter-chip {
  min-inline-size: 0;
  overflow-wrap: anywhere;
  white-space: normal;
}

.${networkFlowChromeRootClassName} .network-flow-popover {
  max-inline-size: min(50rem, calc(100vw - 2 * ${cssToken("--ct-spacing-lg")}));
  max-block-size: min(34rem, calc(100vh - 2 * ${cssToken("--ct-spacing-lg")}));
  overflow: auto;
  border: ${cssToken("--ct-border-strong")};
  border-radius: ${cssToken("--ct-rounded-md")};
  padding: ${cssToken("--ct-spacing-md")};
  background: ${cssToken("--ct-colors-surface-2")};
  color: ${cssToken("--ct-colors-ink")};
  box-shadow: ${cssToken("--ct-elevation-popover")};
}

.${networkFlowChromeRootClassName} .network-flow-dialog-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: grid;
  place-items: center;
  overflow: auto;
  padding: ${cssToken("--ct-spacing-lg")};
  background: ${cssToken("--ct-colors-overlay-scrim")};
}

.${networkFlowChromeRootClassName} .network-flow-dialog {
  inline-size: min(48rem, 100%);
  max-block-size: calc(100vh - 2 * ${cssToken("--ct-spacing-lg")});
  overflow: auto;
  display: grid;
  gap: ${cssToken("--ct-spacing-md")};
  border: ${cssToken("--ct-border-strong")};
  border-radius: ${cssToken("--ct-rounded-lg")};
  padding: ${cssToken("--ct-spacing-lg")};
  background: ${cssToken("--ct-colors-surface-1")};
  color: ${cssToken("--ct-colors-ink")};
  box-shadow: ${cssToken("--ct-elevation-popover")};
}

.${networkFlowChromeRootClassName} .network-flow-dialog-form {
  min-inline-size: 0;
  display: grid;
  gap: ${cssToken("--ct-spacing-md")};
}

.${networkFlowChromeRootClassName} .network-flow-saved-workspace {
  min-inline-size: 0;
  display: grid;
  grid-template-columns: minmax(12rem, 15rem) minmax(0, 1fr);
  gap: ${cssToken("--ct-spacing-md")};
}

.${networkFlowChromeRootClassName} .network-flow-saved-list,
.${networkFlowChromeRootClassName} .network-flow-saved-result {
  min-inline-size: 0;
  border: ${cssToken("--ct-border-hairline")};
  border-radius: ${cssToken("--ct-rounded-md")};
  padding: ${cssToken("--ct-spacing-md")};
  background: ${cssToken("--ct-colors-surface-1")};
}

.${networkFlowChromeRootClassName} .network-flow-object-list .network-flow-button {
  inline-size: 100%;
  min-inline-size: 0;
  justify-content: flex-start;
  overflow-wrap: anywhere;
  white-space: normal;
  text-align: start;
}

.${networkFlowChromeRootClassName} .network-flow-mapping-row {
  min-inline-size: 0;
}

.${networkFlowChromeRootClassName} .network-flow-list-action {
  inline-size: 100%;
  min-inline-size: 0;
  display: grid;
  gap: ${cssToken("--ct-spacing-xs")};
  border: ${cssToken("--ct-border-hairline")};
  border-inline-start-width: 3px;
  border-radius: ${cssToken("--ct-rounded-sm")};
  padding: ${cssToken("--ct-spacing-sm")};
  background: ${cssToken("--ct-colors-surface-1")};
  color: ${cssToken("--ct-colors-ink")};
  text-align: start;
  cursor: pointer;
}

.${networkFlowChromeRootClassName} .network-flow-list-action[aria-current="true"] {
  border-inline-start-color: ${cssToken("--ct-colors-accent")};
  background: ${cssToken("--ct-colors-surface-2")};
  font-weight: 700;
}

.${networkFlowChromeRootClassName} .network-flow-status {
  min-inline-size: 0;
  border-inline-start: 3px solid ${cssToken("--ct-colors-semantic-info")};
  padding-inline-start: ${cssToken("--ct-spacing-sm")};
  overflow-wrap: anywhere;
}

.${networkFlowChromeRootClassName} .network-flow-status[data-tone="error"] {
  border-inline-start-color: ${cssToken("--ct-colors-semantic-destructive")};
}

.${networkFlowChromeRootClassName} .network-flow-status[data-tone="stale"] {
  border-inline-start-color: ${cssToken("--ct-colors-semantic-caution")};
}

.${networkFlowChromeRootClassName} .network-flow-mono {
  font-family: ${cssToken("--ct-typography-mono-fontFamily")};
  overflow-wrap: anywhere;
}

.${networkFlowChromeRootClassName} .network-flow-truncate {
  min-inline-size: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1024px) {
  .${networkFlowChromeRootClassName} .network-flow-toolbar {
    align-items: flex-start;
  }

  .${networkFlowChromeRootClassName} .network-flow-field {
    flex: 1 1 12rem;
  }

  .${networkFlowChromeRootClassName} .network-flow-advanced__editor {
    position: fixed;
    inset-block-start: auto;
    inset-block-end: ${cssToken("--ct-spacing-lg")};
    inset-inline: ${cssToken("--ct-spacing-lg")};
    inline-size: auto;
  }

  .${networkFlowChromeRootClassName} .network-flow-saved-workspace {
    grid-template-columns: minmax(10rem, 13rem) minmax(0, 1fr);
  }

  .${networkFlowChromeRootClassName} .network-flow-mapping-row {
    grid-template-columns: minmax(0, 1fr) minmax(12rem, 0.8fr) !important;
  }
}

@media (max-width: 768px) {
  .${networkFlowChromeRootClassName} .network-flow-action-group {
    justify-content: flex-start;
  }

  .${networkFlowChromeRootClassName} .network-flow-dialog-backdrop {
    place-items: start center;
    padding: ${cssToken("--ct-spacing-sm")};
  }

  .${networkFlowChromeRootClassName} .network-flow-dialog {
    max-block-size: calc(100vh - 2 * ${cssToken("--ct-spacing-sm")});
    padding: ${cssToken("--ct-spacing-md")};
  }

  .${networkFlowChromeRootClassName} .network-flow-saved-workspace {
    grid-template-columns: minmax(0, 1fr);
  }

  .${networkFlowChromeRootClassName} .network-flow-mapping-row {
    grid-template-columns: minmax(0, 1fr) !important;
  }

  .${networkFlowChromeRootClassName} .network-flow-advanced__editor {
    inset-block-end: ${cssToken("--ct-spacing-sm")};
    inset-inline: ${cssToken("--ct-spacing-sm")};
  }
}
`;
