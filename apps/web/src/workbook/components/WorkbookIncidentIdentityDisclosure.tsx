import { ChevronDown } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";
import {
  shellIncidentTitleStyle,
  shellTopBarValueStyle,
} from "../layout/workbookShellStyles";
import { workbookQuietCommandStyle } from "./workbookFormStyles";
import { menuStyle } from "./workbookGridControlStyles";

export function WorkbookIncidentIdentityDisclosure({
  incidentKey,
  title,
}: {
  readonly incidentKey: string;
  readonly title: string;
}) {
  const id = useId();
  const [open, setOpen] = useState(false);
  const root = useRef<HTMLFieldSetElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (!open) return;
    const outside = (event: PointerEvent) => {
      if (event.target instanceof Node && !root.current?.contains(event.target))
        setOpen(false);
    };
    document.addEventListener("pointerdown", outside);
    return () => document.removeEventListener("pointerdown", outside);
  }, [open]);
  return (
    <fieldset
      aria-label="Incident context"
      ref={root}
      style={{
        position: "relative",
        minWidth: 0,
        width: "100%",
        border: 0,
        padding: 0,
        margin: 0,
      }}
      onBlur={(event) => {
        if (
          event.relatedTarget instanceof Node &&
          !event.currentTarget.contains(event.relatedTarget)
        )
          setOpen(false);
      }}
      onKeyDown={(event) => {
        if (event.key === "Escape" && open) {
          event.preventDefault();
          event.stopPropagation();
          setOpen(false);
          trigger.current?.focus({ preventScroll: true });
        }
      }}
    >
      <button
        ref={trigger}
        type="button"
        aria-label={`Incident details: ${incidentKey}, ${title}`}
        aria-expanded={open}
        aria-controls={open ? id : undefined}
        onClick={() => setOpen(!open)}
        style={{
          ...workbookQuietCommandStyle,
          width: "100%",
          minWidth: 0,
          justifyContent: "flex-start",
          paddingInline: 0,
        }}
      >
        <strong style={shellTopBarValueStyle}>{incidentKey}</strong>
        <span style={shellIncidentTitleStyle}>{title}</span>
        <ChevronDown
          aria-hidden="true"
          size={16}
          style={{ flex: "0 0 auto" }}
        />
      </button>
      {open ? (
        <section
          id={id}
          aria-label="Incident identity"
          style={{
            ...menuStyle,
            padding: "var(--ct-spacing-md)",
            overflowWrap: "anywhere",
          }}
        >
          <strong>{incidentKey}</strong>
          <span>{title}</span>
        </section>
      ) : null}
    </fieldset>
  );
}
