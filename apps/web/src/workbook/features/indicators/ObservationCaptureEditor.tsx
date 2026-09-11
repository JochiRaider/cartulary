import { indicatorObservationTestId } from "@cartulary/ui-contracts";
import { useEffect, useId, useRef, useState } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type { ObservationDraftStore } from "./ObservationDraftStore";
import { ObservationTargetPicker } from "./ObservationTargetPicker";
import {
  type ObservationDraft,
  type ObservationSource,
  observationSelection,
  observationTextMap,
  observationTypes,
  sameObservationSource,
} from "./observationModel";
import type {
  ObservationIntent,
  ObservationReadPort,
} from "./observationOperation";
import {
  observationField,
  observationInput,
  observationStack,
  observationText,
} from "./observationStyles";

export function ObservationCaptureEditor({
  source,
  ready,
  draft,
  drafts,
  reader,
  generation,
  disabled,
  onSubmit,
}: {
  source: ObservationSource;
  ready: boolean;
  draft: ObservationDraft;
  drafts: ObservationDraftStore;
  reader: ObservationReadPort;
  generation: number;
  disabled: boolean;
  onSubmit: (intent: ObservationIntent, draft: ObservationDraft) => void;
}) {
  const textarea = useRef<HTMLTextAreaElement>(null),
    help = useId(),
    sourceId = useId(),
    typeId = useId();
  const [error, setError] = useState<string | null>(null);
  const display = observationTextMap(source.text).display;
  useEffect(() => {
    const control = textarea.current;
    // Native readonly textareas suppress caret navigation in Chromium. Retain
    // the native selection editor, expose readonly semantics, and reject every
    // text-changing input before it can change this immutable source preview.
    const rejectInput = (event: Event) => event.preventDefault();
    control?.addEventListener("beforeinput", rejectInput);
    return () => control?.removeEventListener("beforeinput", rejectInput);
  }, []);
  const priorSource = useRef({ source, ready });
  useEffect(() => {
    if (
      !sameObservationSource(priorSource.current.source, source) ||
      priorSource.current.ready !== ready
    ) {
      textarea.current?.setSelectionRange(0, 0);
      priorSource.current = { source, ready };
    }
  }, [source, ready]);
  const current = sameObservationSource(source, draft.source) && ready;
  useEffect(() => {
    if (
      (!ready || !sameObservationSource(source, draft.source)) &&
      draft.selection
    )
      drafts.update(draft.key, { source: null, selection: null });
  }, [source, ready, draft, drafts]);
  const selection = current ? draft.selection : null;
  return (
    <form
      style={observationStack}
      onSubmit={(event) => {
        event.preventDefault();
        if (!selection || disabled || !ready) return;
        onSubmit(
          {
            action: "create",
            source,
            selection,
            ...(draft.parsedType ? { parsedType: draft.parsedType } : {}),
            ...(draft.target ? { targetId: draft.target.recordId } : {}),
          },
          draft,
        );
      }}
    >
      <div style={observationField}>
        <label htmlFor={sourceId}>Saved source text</label>
        <textarea
          id={sourceId}
          data-testid={indicatorObservationTestId("source")}
          ref={textarea}
          aria-readonly="true"
          rows={5}
          style={observationInput}
          value={display}
          aria-describedby={help}
          onChange={(event) => {
            event.currentTarget.value = display;
          }}
          onPaste={(event) => event.preventDefault()}
          onCut={(event) => event.preventDefault()}
          onDrop={(event) => event.preventDefault()}
        />
      </div>
      <p id={help} style={observationText}>
        {ready
          ? "Select the saved text with the pointer or Shift and arrow keys, then use the selection below."
          : "Finish saving source edits, then select the saved text again. Your edits are unchanged."}
      </p>
      <WorkbookInspectorActionButton
        disabled={disabled || !ready}
        onClick={() => {
          const control = textarea.current;
          const selected = control
            ? observationSelection(
                source.text,
                control.selectionStart,
                control.selectionEnd,
              )
            : null;
          setError(
            selected
              ? null
              : "Select a nonempty span containing complete characters.",
          );
          if (selected)
            drafts.update(draft.key, { source, selection: selected });
        }}
      >
        Use selected text
      </WorkbookInspectorActionButton>
      {error ? <p role="alert">{error}</p> : null}
      <section
        data-testid={indicatorObservationTestId("preview")}
        aria-label="Selected text preview"
        style={observationStack}
      >
        <strong>Selected text</strong>
        <p style={observationText}>
          {selection?.text ?? "Select text from the current saved source."}
        </p>
      </section>
      <div style={observationField}>
        <label htmlFor={typeId}>Parsed type (optional)</label>
        <select
          id={typeId}
          style={observationInput}
          disabled={disabled}
          value={draft.parsedType}
          onChange={(event) => {
            const type =
              observationTypes.find((value) => value === event.target.value) ??
              "";
            drafts.update(draft.key, { parsedType: type });
          }}
        >
          <option value="">Infer from selected text</option>
          {observationTypes.map((type) => (
            <option key={type} value={type}>
              {type}
            </option>
          ))}
        </select>
      </div>
      <ObservationTargetPicker
        reader={reader}
        generation={generation}
        disabled={disabled}
        selected={draft.target}
        onChange={(target) => drafts.update(draft.key, { target })}
      />
      <WorkbookInspectorActionButton
        type="submit"
        disabled={disabled || !selection || !ready}
      >
        Create observation
      </WorkbookInspectorActionButton>
    </form>
  );
}
