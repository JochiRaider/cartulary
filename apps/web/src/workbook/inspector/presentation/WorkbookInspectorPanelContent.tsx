import type { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import type { ReactNode } from "react";

type State<
  T extends (typeof cartularyDesignPresentation.inspector.dataStates)[number],
> = T;
export type WorkbookInspectorPanelData =
  | { readonly state: State<"initial_loading"> }
  | {
      readonly state: State<"ready" | "refreshing">;
      readonly content: WorkbookInspectorPanelContentModel;
    }
  | {
      readonly state: State<"stale_failure">;
      readonly content: WorkbookInspectorPanelContentModel;
      readonly message: string;
    }
  | { readonly state: State<"unavailable">; readonly message: string };
export type WorkbookInspectorPanelContentModel =
  | { readonly kind: "empty"; readonly message: string }
  | { readonly kind: "populated"; readonly content: ReactNode };
export type WorkbookInspectorPanelModel =
  | { readonly access: "concealed" }
  | {
      readonly access: "readable";
      readonly data: WorkbookInspectorPanelData;
      readonly commands?: ReactNode;
    };

export function inspectorPanel(
  content: ReactNode,
): WorkbookInspectorPanelModel {
  return {
    access: "readable",
    data: { state: "ready", content: { kind: "populated", content } },
  };
}
export function emptyInspectorPanel(
  message: string,
): WorkbookInspectorPanelModel {
  return {
    access: "readable",
    data: { state: "ready", content: { kind: "empty", message } },
  };
}

// Owners determine data and access independently. This component never fetches,
// inspects child output, or equates permission loss with an empty result.
export function WorkbookInspectorPanelContent({
  model,
}: {
  readonly model: WorkbookInspectorPanelModel;
}) {
  if (model.access === "concealed") return null;
  const data = model.data;
  return (
    <>
      {data.state === "initial_loading" ? <p role="status">Loading…</p> : null}
      {data.state === "refreshing" ? <p role="status">Refreshing…</p> : null}
      {data.state === "unavailable" || data.state === "stale_failure" ? (
        <p role="alert">{data.message}</p>
      ) : null}
      {"content" in data ? (
        data.content.kind === "empty" ? (
          <p>{data.content.message}</p>
        ) : (
          data.content.content
        )
      ) : null}
      {model.commands}
    </>
  );
}
