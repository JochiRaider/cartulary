import { useContext, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { NoteCreateContext, noteSheetAttachment } from "./NoteCreateContext";
import { NoteSourceControl } from "./NoteSourceControl";

const noSubscribe = () => () => {};
const noSnapshot = () => null;
export function NoteSheetAuthoring() {
  const context = useContext(NoteCreateContext);
  const state = useSyncExternalStore(
    context?.owner.subscribe ?? noSubscribe,
    context?.owner.getSnapshot ?? noSnapshot,
  );
  if (!context || !state?.authority) return null;
  const { owner } = context;
  const reader = owner.getReader();
  return (
    <div style={{ display: "grid", gap: "0.5rem" }}>
      {reader ? (
        <NoteSourceControl
          targetKey={`note:${state.draft?.id ?? "new"}`}
          source={state.draft?.source ?? null}
          reader={reader}
          revision={state.candidateRevision}
          disabled={owner.busy || !owner.canSubmit()}
          onChange={(source) => {
            if (!owner.getSnapshot().draft)
              owner.beginSheet(context.sheetRef, noteSheetAttachment);
            owner.changeSource(source);
          }}
        />
      ) : null}
      {Object.entries(state.errors).map(([key, error]) => (
        <p key={key} role="alert">
          {error}
        </p>
      ))}
      {state.message ? <p role="status">{state.message}</p> : null}
      {state.needsReview && state.draft ? (
        <Button
          tone="secondary"
          type="button"
          disabled={owner.busy || !owner.canSubmit()}
          onClick={() => void owner.review()}
        >
          Review source and access
        </Button>
      ) : null}
      {state.draft ? (
        <Button
          tone="secondary"
          type="button"
          disabled={owner.busy}
          onClick={() => owner.discard()}
        >
          Discard Note draft
        </Button>
      ) : null}
    </div>
  );
}
