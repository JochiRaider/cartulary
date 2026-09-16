import { type CSSProperties, forwardRef } from "react";
import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";

export const WorkbookQueueOverflowNotice = forwardRef<
  HTMLElement,
  {
    readonly message: string;
    readonly onClose: () => void;
  }
>(function WorkbookQueueOverflowNotice({ message, onClose }, ref) {
  return (
    <section
      aria-label="Workbook queued edit overflow"
      ref={ref}
      tabIndex={-1}
      onKeyDown={(event) => {
        if (event.key !== "Escape") return;
        event.preventDefault();
        event.stopPropagation();
        onClose();
      }}
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
      <Button type="button" tone="secondary" onClick={onClose}>
        Close queued edit notice
      </Button>
    </section>
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
