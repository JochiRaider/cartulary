import type { CSSProperties } from "react";
import { useId, useRef } from "react";
import {
  workbookFormButtonStyle,
  workbookGridActionButtonStyle,
  workbookTypography,
} from "../../components/workbookFormStyles";

/** Picker, drop and paste share one entry; the source owner rechecks admission. */
export function EvidenceAttachmentEntry({
  title,
  testId,
  disabledReason,
  busy,
  onAttach,
  compact = false,
  regionTestId,
}: {
  readonly title: string;
  readonly testId: string;
  readonly disabledReason: string | null;
  readonly busy: boolean;
  readonly onAttach: (files: readonly File[]) => void;
  readonly compact?: boolean;
  readonly regionTestId?: string | undefined;
}) {
  const input = useRef<HTMLInputElement>(null);
  const description = useId();
  const disabled = disabledReason !== null || busy;
  const attach = (files: FileList | readonly File[]) => {
    if (!disabled && files.length) onAttach(Array.from(files));
  };
  const controls = (
    <>
      <button
        type="button"
        aria-label={`Attach file to ${title}`}
        aria-describedby={compact ? undefined : description}
        disabled={disabled}
        aria-busy={busy || undefined}
        style={
          compact ? workbookGridActionButtonStyle : workbookFormButtonStyle
        }
        title={disabledReason ?? undefined}
        onClick={() => input.current?.click()}
      >
        {compact ? "Attach" : "Attach file"}
      </button>
      <input
        ref={input}
        hidden
        type="file"
        data-testid={testId}
        disabled={disabled}
        aria-label={`Attach file to ${title}`}
        accept="image/*,.txt,.pdf,text/plain,application/pdf"
        onChange={(event) => {
          attach(event.currentTarget.files ?? []);
          event.currentTarget.value = "";
        }}
      />
    </>
  );
  if (compact) return controls;
  return (
    <section
      data-testid={regionTestId}
      aria-label={`File attachment for ${title}`}
      tabIndex={-1}
      aria-describedby={description}
      style={dropStyle}
      onDragOver={(event) => event.preventDefault()}
      onDrop={(event) => {
        event.preventDefault();
        event.stopPropagation();
        attach(event.dataTransfer.files);
      }}
      onPaste={(event) => {
        if (event.clipboardData.files.length) {
          event.preventDefault();
          event.stopPropagation();
          attach(event.clipboardData.files);
        }
      }}
      onKeyDown={(event) => {
        if (
          event.target === event.currentTarget &&
          (event.key === "Enter" || event.key === " ")
        ) {
          event.preventDefault();
          if (!disabled && !event.repeat) input.current?.click();
        }
      }}
    >
      <div>{controls}</div>
      <p id={description} style={messageStyle}>
        {disabledReason ??
          "Choose a file, drop it here, or paste a file while Attach file is focused."}
      </p>
    </section>
  );
}
const dropStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  padding: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  minInlineSize: 0,
} satisfies CSSProperties;
const messageStyle = {
  ...workbookTypography("metadata"),
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;
