import {
  type ComponentPropsWithoutRef,
  type CSSProperties,
  forwardRef,
} from "react";

export const RecoverySurface = forwardRef<
  HTMLElement,
  ComponentPropsWithoutRef<"aside">
>(function RecoverySurface({ children, style, ...props }, ref) {
  return (
    <aside ref={ref} style={{ ...recoverySurfaceStyle, ...style }} {...props}>
      {children}
    </aside>
  );
});

const recoverySurfaceStyle = {
  position: "absolute",
  insetBlockStart:
    "calc(var(--ct-layout-viewBarHeight) + var(--ct-spacing-sm))",
  insetInlineEnd: "var(--ct-spacing-sm)",
  zIndex: 8,
  inlineSize: "min(46rem, calc(100% - var(--ct-spacing-xl)))",
  minInlineSize: 0,
  maxInlineSize: "calc(100% - var(--ct-spacing-xl))",
  blockSize: "max-content",
  maxBlockSize:
    "calc(100% - var(--ct-layout-viewBarHeight) - var(--ct-layout-statusStripHeight) - (2 * var(--ct-spacing-sm)))",
  minBlockSize: 0,
  overflowX: "hidden",
  overflowY: "auto",
  overscrollBehavior: "contain",
  boxSizing: "border-box",
  border: "var(--ct-border-strong)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
  color: "var(--ct-colors-ink)",
  padding: "var(--ct-spacing-lg)",
  display: "grid",
  alignContent: "start",
  gap: "var(--ct-spacing-md)",
} satisfies CSSProperties;
