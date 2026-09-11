import { indicatorObservationTestId } from "@cartulary/ui-contracts";
import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { ObservationOperationStatus } from "./ObservationOperationStatus";
import {
  type ObservationOwnerPort,
  observationIntentSource,
} from "./observationOperation";
import { observationStack, observationText } from "./observationStyles";
export function WorkbookObservationRecovery({
  owner,
}: {
  owner: ObservationOwnerPort;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const [open, setOpen] = useState(false),
    trigger = useRef<HTMLButtonElement>(null),
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
    <div style={{ position: "relative" }}>
      <WorkbookInspectorActionButton
        data-testid={indicatorObservationTestId("recovery-trigger")}
        ref={trigger}
        aria-expanded={open}
        onClick={() => {
          if (open) close();
          else {
            focus.current = true;
            setOpen(true);
          }
        }}
      >
        Indicator observations ({snapshot.entries.length})
      </WorkbookInspectorActionButton>
      {open ? (
        <section
          data-testid={indicatorObservationTestId("recovery")}
          aria-label="Indicator observation recovery"
          style={{
            ...observationStack,
            position: "absolute",
            zIndex: 30,
            insetInlineEnd: 0,
            inlineSize: "min(38rem, 90vw)",
            maxBlockSize: "75vh",
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
            Indicator observation recovery
          </h3>
          <p>
            Submitted changes remain here when the Inspector closes. Closing a
            panel does not cancel a server write.
          </p>
          <WorkbookInspectorActionButton onClick={close}>
            Close observation recovery
          </WorkbookInspectorActionButton>
          {snapshot.entries.map((entry) => (
            <article key={entry.attempt.id} style={observationStack}>
              <strong>{entry.attempt.intent.action} observation</strong>
              <p style={observationText}>
                {entry.attempt.intent.action === "create"
                  ? entry.attempt.intent.selection.text
                  : entry.attempt.intent.observation.observed_text}
              </p>
              <p style={observationText}>
                Source: {observationIntentSource(entry.attempt.intent)}
              </p>
              <ObservationOperationStatus owner={owner} entry={entry} />
            </article>
          ))}
        </section>
      ) : null}
    </div>
  );
}
