import { type CSSProperties, forwardRef } from "react";
import { RecoverySurface } from "./RecoverySurface";

export const WorkbookQueueOverflowNotice = forwardRef<
  HTMLElement,
  {
    readonly message: string;
    readonly onFocusWithinChange: (focused: boolean) => void;
  }
>(function WorkbookQueueOverflowNotice({ message, onFocusWithinChange }, ref) {
  return (
    <RecoverySurface
      aria-label="Workbook queued edit overflow"
      ref={ref}
      tabIndex={-1}
      onBlurCapture={(event) => {
        const relatedTarget = event.relatedTarget;
        if (
          !(relatedTarget instanceof Node) ||
          !event.currentTarget.contains(relatedTarget)
        ) {
          onFocusWithinChange(false);
        }
      }}
      onFocusCapture={() => onFocusWithinChange(true)}
    >
      <div>
        <p style={eyebrowStyle}>Local edit needs attention</p>
        <h2 style={titleStyle}>Queued edits</h2>
      </div>
      <p
        aria-atomic="true"
        aria-live="assertive"
        role="status"
        style={bodyStyle}
      >
        {message}
      </p>
    </RecoverySurface>
  );
});

const eyebrowStyle = {
  margin: 0,
  color: "var(--ct-colors-semantic-conflict)",
  fontSize: "0.78rem",
  fontWeight: 800,
  letterSpacing: "0.08em",
  textTransform: "uppercase",
} satisfies CSSProperties;

const titleStyle = {
  margin: "0.2rem 0 0",
  overflowWrap: "anywhere",
} satisfies CSSProperties;

const bodyStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
