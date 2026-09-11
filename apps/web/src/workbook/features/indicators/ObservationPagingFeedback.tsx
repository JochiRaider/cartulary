import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "../../inspector/presentation/WorkbookInspectorFeedback";
import { workbookInspectorErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import type { ObservationCollection } from "./ObservationCollection";

export function ObservationPagingFeedback<T>({
  pages,
  state,
  label,
}: {
  pages: ObservationCollection<T>;
  state: ReturnType<ObservationCollection<T>["getSnapshot"]>;
  label: string;
}) {
  return (
    <>
      {["initial_loading", "loading_more", "refreshing"].includes(
        state.phase,
      ) ? (
        <p role="status">
          {state.phase === "loading_more"
            ? "Loading more"
            : state.phase === "refreshing"
              ? "Refreshing"
              : "Loading"}{" "}
          {label}…
        </p>
      ) : null}
      {state.failure ? (
        <div>
          <WorkbookInspectorPublicError
            error={workbookInspectorErrorPresentation(state.failure)}
          />
          {state.items.length ? (
            <p>
              Previously loaded results remain visible and may be incomplete or
              stale.
            </p>
          ) : null}
          <WorkbookInspectorActionButton onClick={() => void pages.retry()}>
            {state.restartRequired ? "Restart" : "Retry"} {label}
          </WorkbookInspectorActionButton>
        </div>
      ) : null}
      {state.hasMore && state.phase === "ready" ? (
        <WorkbookInspectorActionButton onClick={() => void pages.more()}>
          Load more {label}
        </WorkbookInspectorActionButton>
      ) : null}
    </>
  );
}
