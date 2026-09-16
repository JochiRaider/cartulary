import {
  partiesViewSchemaId,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { useEffect, useRef, useState } from "react";
import { WorkbookAuthoringReferencePicker } from "../../components/WorkbookAuthoringReferencePicker";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";

export function RelatedEvidencePartyControl({
  disabled,
  targetKey,
  field,
  value,
  labels,
  reader,
  revision,
  errorId,
  onChange,
}: {
  readonly disabled: boolean;
  readonly targetKey: string;
  readonly field: ViewFieldContract;
  readonly value: string;
  readonly labels: Readonly<Record<string, string>>;
  readonly reader: WorkbookAuthoringReadPort;
  readonly revision: number;
  readonly errorId?: string | undefined;
  readonly onChange: (
    value: string,
    labels: Readonly<Record<string, string>>,
  ) => void;
}) {
  const [openTarget, setOpenTarget] = useState<string | null>(null);
  const open = openTarget === targetKey;
  if (openTarget !== null && openTarget !== targetKey) setOpenTarget(null);
  const setOpen = (value: boolean) => setOpenTarget(value ? targetKey : null);
  useEffect(() => {
    if (disabled) setOpenTarget(null);
  }, [disabled]);
  const trigger = useRef<HTMLButtonElement>(null);
  const close = () => {
    setOpen(false);
    trigger.current?.focus({ preventScroll: true });
  };
  return (
    <div style={{ display: "grid", gap: "var(--ct-spacing-xs)" }}>
      <span>
        {field.label}:{" "}
        {value
          ? (labels[value] ?? "Selected Party (availability needs review)")
          : "None selected"}
      </span>
      {value ? (
        <Button type="button" tone="secondary" onClick={() => onChange("", {})}>
          Remove {field.label}
        </Button>
      ) : null}
      <Button
        ref={trigger}
        aria-describedby={errorId}
        type="button"
        tone="secondary"
        disabled={disabled}
        onClick={() => setOpen(true)}
      >
        Choose {field.label}
      </Button>
      {open && !disabled ? (
        <PartyPicker
          key={targetKey}
          targetKey={targetKey}
          label={field.label}
          fieldKey={field.fieldKey}
          value={value}
          labels={labels}
          reader={reader}
          revision={revision}
          onCancel={close}
          onApply={(id, names) => {
            onChange(id, names);
            close();
          }}
        />
      ) : null}
    </div>
  );
}
function PartyPicker({
  targetKey,
  label,
  fieldKey,
  value,
  labels,
  reader,
  revision,
  onApply,
  onCancel,
}: {
  readonly targetKey: string;
  readonly label: string;
  readonly fieldKey: string;
  readonly value: string;
  readonly labels: Readonly<Record<string, string>>;
  readonly reader: WorkbookAuthoringReadPort;
  readonly revision: number;
  readonly onApply: (
    id: string,
    labels: Readonly<Record<string, string>>,
  ) => void;
  readonly onCancel: () => void;
}) {
  return (
    <WorkbookAuthoringReferencePicker
      regionLabel={`Choose ${label}`}
      targetKey={`${targetKey}:${fieldKey}`}
      label={label}
      testId={`related-evidence-reference-${fieldKey}`}
      views={[partiesViewSchemaId]}
      multiple={false}
      maximum={1}
      selected={
        value
          ? [
              {
                recordId: value,
                displayText: labels[value] ?? value,
                viewSchemaId: partiesViewSchemaId,
              },
            ]
          : []
      }
      reader={reader}
      revision={revision}
      disabled={false}
      onCancel={onCancel}
      applyLabel="Apply Party"
      cancelLabel="Cancel Party selection"
      onApply={(items) =>
        onApply(
          items[0]?.recordId ?? "",
          Object.fromEntries(
            items.map((item) => [item.recordId, item.displayText]),
          ),
        )
      }
    />
  );
}
