import { createContext, useContext } from "react";

/** Immutable source identity only; this port owns no import workflow state. */
export const CommittedSourceContext = createContext<{
  readonly open: (sourceId: string, trigger: HTMLElement) => void;
  readonly openRecord: (recordId: string, trigger: HTMLElement) => void;
} | null>(null);

export function CommittedSourceButton({
  sourceId,
}: {
  readonly sourceId: string;
}) {
  const inspection = useContext(CommittedSourceContext);
  return inspection ? (
    <button
      type="button"
      onClick={(event) => inspection.open(sourceId, event.currentTarget)}
    >
      Inspect original source
    </button>
  ) : null;
}

export function RecordImportSourcesDetails({
  recordId,
}: {
  readonly recordId: string;
}) {
  const inspection = useContext(CommittedSourceContext);
  return inspection ? (
    <div>
      <dt>Import history</dt>
      <dd>
        <button
          type="button"
          onClick={(event) =>
            inspection.openRecord(recordId, event.currentTarget)
          }
        >
          Find original import sources
        </button>
      </dd>
    </div>
  ) : null;
}
