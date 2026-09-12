import { indicatorCreateTestId } from "@cartulary/ui-contracts";
import { Fingerprint } from "lucide-react";
import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { IndicatorCreateOperationStatus } from "./IndicatorCreateOperationStatus";
import type { IndicatorCreateOwnerPort } from "./indicatorCreateOperation";
import { observationStack, observationText } from "./observationStyles";

export function WorkbookIndicatorCreateRecovery({
  owner,
}: {
  owner: IndicatorCreateOwnerPort;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot),
    [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null),
    heading = useRef<HTMLHeadingElement>(null),
    focus = useRef(false);
  useLayoutEffect(() => {
    if (open && focus.current) {
      heading.current?.focus();
      focus.current = false;
    }
  }, [open]);
  const close = () => {
    setOpen(false);
    trigger.current?.focus();
  };
  if (!snapshot.authority || !snapshot.entries.length) return null;
  return (
    <div style={{ position: "relative", flex: "0 0 auto" }}>
      <WorkbookInspectorActionButton
        ref={trigger}
        data-testid={indicatorCreateTestId("recovery-trigger")}
        aria-label={`Canonical Indicator recovery, ${snapshot.entries.length} results`}
        title="Canonical Indicator recovery"
        style={{
          display: "inline-flex",
          alignItems: "center",
          gap: "var(--ct-spacing-xs)",
          minInlineSize: "2.5rem",
          whiteSpace: "nowrap",
        }}
        aria-expanded={open}
        onClick={() => {
          if (open) close();
          else {
            focus.current = true;
            setOpen(true);
          }
        }}
      >
        <Fingerprint size={16} aria-hidden="true" />
        <span aria-hidden="true">{snapshot.entries.length}</span>
      </WorkbookInspectorActionButton>
      {open ? (
        <div
          style={{
            position: "fixed",
            inset: "var(--ct-spacing-md)",
            zIndex: 30,
            display: "flex",
            alignItems: "flex-start",
            justifyContent: "flex-end",
            pointerEvents: "none",
          }}
        >
          <section
            data-testid={indicatorCreateTestId("recovery")}
            aria-label="Canonical Indicator recovery"
            style={{
              ...observationStack,
              inlineSize: "min(38rem, 100%)",
              maxBlockSize: "100%",
              pointerEvents: "auto",
              boxSizing: "border-box",
              overflow: "auto",
              padding: "var(--ct-spacing-md)",
              background: "var(--ct-colors-surface-1)",
              border: "var(--ct-border-hairline)",
              boxShadow: "var(--ct-elevation-popover)",
            }}
            onKeyDown={(event) => {
              if (event.key === "Escape") {
                event.stopPropagation();
                close();
              }
            }}
          >
            <h3 ref={heading} tabIndex={-1}>
              Canonical Indicator recovery
            </h3>
            <p>
              Canonical creation and observation resolution have separate
              receipts. Review linking from the source record’s Relationships
              panel.
            </p>
            <WorkbookInspectorActionButton onClick={close}>
              Close canonical recovery
            </WorkbookInspectorActionButton>
            {snapshot.entries.map((entry) => (
              <article key={entry.attempt.id} style={observationStack}>
                <p style={observationText}>
                  Observed text: {entry.attempt.observation.observed_text}
                </p>
                <p style={observationText}>
                  Source: {entry.attempt.observation.source_record_id}
                </p>
                <IndicatorCreateOperationStatus owner={owner} entry={entry} />
              </article>
            ))}
          </section>
        </div>
      ) : null}
    </div>
  );
}
