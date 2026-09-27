import {
  type RefCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { WorkbookCandidateAuthorityContext } from "../hooks/useWorkbookCandidateDiscovery";
import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";
import type {
  WorkbookAuthoringReadPort,
  WorkbookAuthoringSelection,
} from "../ports/WorkbookAuthoringReadPort";
import { useSelectedReferenceRemovalFocus } from "./useSelectedReferenceRemovalFocus";
import { WorkbookAuthoringReferencePicker } from "./WorkbookAuthoringReferencePicker";
import { menuStyle } from "./workbookGridControlStyles";

type Props = Readonly<{
  compact?: boolean | undefined;
  focusTargetRef?:
    | RefCallback<
        | HTMLInputElement
        | HTMLSelectElement
        | HTMLTextAreaElement
        | HTMLButtonElement
      >
    | undefined;
  label: string;
  targetKey: string;
  maximum: number;
  errorId?: string | undefined;
  required?: boolean | undefined;
  testId: string;
  views: readonly string[];
  multiple: boolean;
  captureRowVersion?: boolean;
  selected: readonly WorkbookAuthoringSelection[];
  reader: Pick<WorkbookAuthoringReadPort, "page" | "availableViews">;
  revision: number;
  disabled: boolean;
  onApply: (selected: readonly WorkbookAuthoringSelection[]) => void;
}>;
/** Neutral staged identity picker. Page membership never owns retained selection. */
export function WorkbookAuthoringReferenceControl(props: Props) {
  const [openTarget, setOpenTarget] = useState<string | null>(null);
  const open = openTarget === props.targetKey;
  if (openTarget !== null && openTarget !== props.targetKey)
    setOpenTarget(null);
  const setOpen = (value: boolean) =>
    setOpenTarget(value ? props.targetKey : null);
  useEffect(() => {
    if (props.disabled) setOpenTarget(null);
  }, [props.disabled]);
  const trigger = useRef<HTMLButtonElement>(null);
  const group = useRef<HTMLFieldSetElement>(null);
  const popover = useRef<HTMLDivElement>(null);
  const authority = useContext(WorkbookCandidateAuthorityContext);
  const removalFocus = useSelectedReferenceRemovalFocus({
    ids: props.selected.map((item) => item.recordId),
    scopeKey: `${props.targetKey}:${authority.identity}:${props.compact}`,
    disabled: props.disabled || !authority.canRead,
    fallback: () => trigger.current,
    groupRef: group,
  });
  useLayoutEffect(() => {
    if (!open || !props.compact || !popover.current || !trigger.current) return;
    const panel = popover.current;
    const anchor = trigger.current;
    Object.assign(panel.style, {
      ...menuStyle,
      position: "fixed",
      inset: "auto",
      margin: "0",
      overflow: "auto",
      color: "var(--ct-colors-ink)",
    });
    panel.showPopover?.();
    const position = () => {
      const viewport = window.visualViewport;
      const left = viewport?.offsetLeft ?? 0;
      const top = viewport?.offsetTop ?? 0;
      const width = viewport?.width ?? window.innerWidth;
      const height = viewport?.height ?? window.innerHeight;
      const bounds = anchor.getBoundingClientRect();
      const cssWidth = Number.parseFloat(getComputedStyle(panel).width);
      const scale =
        cssWidth > 0 ? panel.getBoundingClientRect().width / cssWidth : 1;
      panel.style.maxWidth = `${width / scale}px`;
      panel.style.maxHeight = `${height / scale}px`;
      const size = panel.getBoundingClientRect();
      panel.style.left = `${Math.max(left, Math.min(bounds.left, left + width - size.width)) / scale}px`;
      panel.style.top = `${Math.max(top, Math.min(bounds.bottom, top + height - size.height)) / scale}px`;
    };
    position();
    const observer = new ResizeObserver(position);
    observer.observe(panel);
    window.addEventListener("resize", position);
    window.addEventListener("scroll", position, true);
    window.visualViewport?.addEventListener("resize", position);
    return () => {
      observer.disconnect();
      window.removeEventListener("resize", position);
      window.removeEventListener("scroll", position, true);
      window.visualViewport?.removeEventListener("resize", position);
    };
  }, [open, props.compact]);
  const close = () => {
    setOpen(false);
    trigger.current?.focus({ preventScroll: true });
  };
  return (
    <fieldset
      ref={group}
      aria-label={`${props.label} selected references`}
      data-reference-focus-fallback
      tabIndex={-1}
      style={groupStyle}
    >
      {!props.compact ? <span>{props.label}</span> : null}
      {props.compact ? null : props.selected.length ? (
        <ul style={{ margin: 0, paddingInlineStart: "var(--ct-spacing-lg)" }}>
          {props.selected.map((item) => (
            <li key={item.recordId} style={{ overflowWrap: "anywhere" }}>
              {item.displayText || "Selected reference"}{" "}
              <Button
                ref={removalFocus.buttonRef(item.recordId)}
                tone="secondary"
                type="button"
                disabled={props.disabled || !authority.canRead}
                aria-label={`Remove ${props.label} ${item.displayText || "reference"}${props.selected.filter((value) => value.displayText === item.displayText).length > 1 ? ` (${item.recordId})` : ""}`}
                onClick={(event) =>
                  removalFocus.remove(item.recordId, event.currentTarget, () =>
                    props.onApply(
                      props.selected.filter(
                        (i) => i.recordId !== item.recordId,
                      ),
                    ),
                  )
                }
              >
                Remove
              </Button>
            </li>
          ))}
        </ul>
      ) : (
        <span>No {props.label.toLowerCase()} selected.</span>
      )}
      <Button
        ref={(element) => {
          trigger.current = element;
          props.focusTargetRef?.(element);
        }}
        style={
          props.compact
            ? {
                maxWidth: "6rem",
                whiteSpace: "nowrap",
                overflow: "hidden",
                textOverflow: "ellipsis",
              }
            : undefined
        }
        title={
          props.compact
            ? props.selected
                .map((item) => item.displayText || item.recordId)
                .join(", ")
            : undefined
        }
        aria-label={`Choose ${props.label.toLowerCase()}`}
        aria-expanded={open}
        data-create-required={props.required}
        aria-invalid={!!props.errorId}
        aria-describedby={props.errorId}
        tone="secondary"
        type="button"
        disabled={props.disabled || !authority.canRead}
        onClick={() => setOpen(true)}
      >
        {props.compact && props.selected.length
          ? props.selected
              .map((item) => item.displayText || "Selected reference")
              .join(", ")
          : `Choose ${props.label.toLowerCase()}`}
      </Button>
      {open && !props.disabled ? (
        <div
          ref={popover}
          popover={props.compact ? "auto" : undefined}
          onToggle={(event) => {
            if (props.compact && event.newState === "closed") setOpen(false);
          }}
        >
          <WorkbookAuthoringReferencePicker
            key={props.targetKey}
            {...props}
            onCancel={close}
            onApply={(items) => {
              props.onApply(items);
              close();
            }}
          />
        </div>
      ) : null}
    </fieldset>
  );
}
const groupStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
  border: 0,
  margin: 0,
  padding: 0,
} as const;
