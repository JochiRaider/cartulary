import { useEffect, useRef, useState } from "react";
import { WorkbookAuthoringReferencePicker } from "../../components/WorkbookAuthoringReferencePicker";
import { workbookFormFieldsStyle } from "../../components/workbookFormStyles";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import {
  type NoteCreateReader,
  type NoteSource,
  noteSourceViews,
} from "./noteCreateModel";

export function NoteSourceControl({
  targetKey,
  source,
  reader,
  revision,
  disabled,
  onChange,
}: {
  readonly targetKey: string;
  readonly source: NoteSource | null;
  readonly reader: NoteCreateReader;
  readonly revision: number;
  readonly disabled: boolean;
  readonly onChange: (source: NoteSource | null) => void;
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
    <div style={workbookFormFieldsStyle}>
      <span style={{ overflowWrap: "anywhere" }}>
        Source:{" "}
        {source ? source.label || source.recordId : "None (unlinked Note)"}
      </span>
      <Button
        type="button"
        tone="secondary"
        ref={trigger}
        disabled={disabled}
        onClick={() => setOpen(true)}
      >
        Choose source
      </Button>
      {source ? (
        <Button
          type="button"
          tone="secondary"
          disabled={disabled}
          onClick={() => onChange(null)}
        >
          Clear source
        </Button>
      ) : null}
      {open && !disabled ? (
        <SourcePicker
          key={targetKey}
          targetKey={targetKey}
          source={source}
          reader={reader}
          revision={revision}
          onCancel={close}
          onApply={(value) => {
            onChange(value);
            close();
          }}
        />
      ) : null}
    </div>
  );
}
function SourcePicker({
  targetKey,
  source,
  reader,
  revision,
  onCancel,
  onApply,
}: {
  readonly targetKey: string;
  readonly source: NoteSource | null;
  readonly reader: NoteCreateReader;
  readonly revision: number;
  readonly onCancel: () => void;
  readonly onApply: (source: NoteSource | null) => void;
}) {
  return (
    <WorkbookAuthoringReferencePicker
      regionLabel="Choose Note source"
      captureRowVersion
      targetKey={`${targetKey}:source`}
      label="Note source"
      testId="note-source-record"
      views={noteSourceViews}
      multiple={false}
      maximum={1}
      surfaceLabel="Source sheet"
      applyLabel="Apply source"
      cancelLabel="Cancel source"
      selected={
        source
          ? [
              {
                recordId: source.recordId,
                displayText: source.label,
                viewSchemaId: source.viewSchemaId,
                rowVersion: source.rowVersion,
              },
            ]
          : []
      }
      reader={reader}
      revision={revision}
      disabled={false}
      onCancel={onCancel}
      onApply={(items) => {
        const item = items[0];
        if (!item) onApply(null);
        else if (item.rowVersion !== undefined)
          onApply({
            recordId: item.recordId,
            viewSchemaId: item.viewSchemaId,
            rowVersion: item.rowVersion,
            label: item.displayText,
          });
      }}
    />
  );
}
