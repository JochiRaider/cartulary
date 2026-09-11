import { indicatorLifecycleTestId } from "@cartulary/ui-contracts";
import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { IndicatorLifecycleOperationStatus } from "./IndicatorLifecycleOperationStatus";
import type { IndicatorLifecycleOwnerPort } from "./indicatorLifecycleOperation";
import { lifecycleStack } from "./indicatorLifecycleStyles";

export function WorkbookIndicatorLifecycleRecovery({
  owner,
}: {
  owner: IndicatorLifecycleOwnerPort;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement | null>(null),
    heading = useRef<HTMLHeadingElement | null>(null);
  const focus = useRef(false);
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
  if (!snapshot.authority || snapshot.entries.length === 0) return null;
  return (
    <div style={{ position: "relative" }}>
      <WorkbookInspectorActionButton
        data-testid={indicatorLifecycleTestId("recovery-trigger")}
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
        Indicator intervals ({snapshot.entries.length})
      </WorkbookInspectorActionButton>
      {open ? (
        <section
          data-testid={indicatorLifecycleTestId("recovery")}
          aria-label="Indicator interval recovery"
          style={{
            ...lifecycleStack,
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
            Indicator interval recovery
          </h3>
          <p>
            Submitted intervals remain here when the Inspector closes. Closing a
            panel does not cancel a server write.
          </p>
          <WorkbookInspectorActionButton onClick={close}>
            Close interval recovery
          </WorkbookInspectorActionButton>
          {snapshot.entries.map((entry) => (
            <article
              key={entry.attempt.id}
              aria-label={`Interval for ${entry.attempt.draft.label}`}
            >
              <strong>{entry.attempt.draft.label}</strong>
              <p>
                {entry.attempt.values.lifecycle_state.replaceAll("_", " ")} from{" "}
                {entry.attempt.values.valid_from} (UTC)
              </p>
              <IndicatorLifecycleOperationStatus owner={owner} entry={entry} />
            </article>
          ))}
        </section>
      ) : null}
    </div>
  );
}
