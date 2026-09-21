import type { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import type { ReactNode } from "react";
import type { WorkbookInspectorNotice } from "../workbookInspectorErrorModel";
import { WorkbookInspectorNoticeView } from "./WorkbookInspectorFeedback";

type State<
  T extends (typeof cartularyDesignPresentation.inspector.dataStates)[number],
> = T;
export type WorkbookInspectorPanelData =
  | { readonly state: State<"initial_loading">; readonly message?: string }
  | {
      readonly state: State<"ready" | "refreshing">;
      readonly content: WorkbookInspectorPanelContentModel;
    }
  | {
      readonly state: State<"stale_failure">;
      readonly content: WorkbookInspectorPanelContentModel;
      readonly message: string;
    }
  | {
      readonly state: State<"unavailable">;
      readonly cause: (typeof cartularyDesignPresentation.inspector.unavailableCauses)[number];
      readonly message: string;
    };
export type WorkbookInspectorPanelContentModel =
  | { readonly kind: "empty"; readonly message: string }
  | { readonly kind: "populated"; readonly content: ReactNode };

/** Adapts an owner's existing read observation without storing or fetching data. */
export function inspectorReadData(input: {
  readonly requested: boolean;
  readonly pending: boolean;
  readonly accepted: WorkbookInspectorPanelContentModel | null;
  readonly failure: string | null;
  readonly notLoadedMessage: string;
}): WorkbookInspectorPanelData {
  if (input.accepted)
    return input.failure
      ? {
          state: "stale_failure",
          content: input.accepted,
          message: input.failure,
        }
      : {
          state: input.pending ? "refreshing" : "ready",
          content: input.accepted,
        };
  if (input.pending && input.requested) return { state: "initial_loading" };
  return {
    state: "unavailable",
    cause: input.failure ? "load_failed" : "not_requested",
    message: input.failure ?? input.notLoadedMessage,
  };
}
export type WorkbookInspectorRegionModel =
  | { readonly access: "concealed" }
  | {
      readonly access: "readable";
      readonly data: WorkbookInspectorPanelData;
      // Authoring is independent of accepted-data emptiness and read freshness.
      readonly authoring?: ReactNode;
      readonly commands?: ReactNode;
      readonly commandsPlacement?: "before_content";
      readonly messageId?: string;
      // An existing operation/status owner may already announce this transition.
      readonly announcement?: "owner";
      readonly notice?: {
        readonly value: WorkbookInspectorNotice;
        readonly consume: (notice: WorkbookInspectorNotice) => boolean;
      };
    };

export type PresentInspectorRegion = (
  model: WorkbookInspectorRegionModel,
) => ReactNode;
type PresentRegion = PresentInspectorRegion;
export type WorkbookInspectorRegion = {
  readonly id: string;
} & (
  | { readonly kind: "snapshot"; readonly model: WorkbookInspectorRegionModel }
  | {
      readonly kind: "owner";
      // The subscribed owner delivers its model directly to the renderer. This
      // slot preserves its existing component lifetime; it is not a state cache
      // or a ready wrapper around an independently stateful child.
      readonly render: (present: PresentRegion) => ReactNode;
    }
);

export type WorkbookInspectorPanelModel =
  | { readonly access: "concealed" }
  | {
      readonly access: "readable";
      readonly regions: readonly [
        WorkbookInspectorRegion,
        ...WorkbookInspectorRegion[],
      ];
      readonly authoring?: ReactNode;
    };

export function inspectorPanel(
  ...regions: readonly [WorkbookInspectorRegion, ...WorkbookInspectorRegion[]]
): Extract<WorkbookInspectorPanelModel, { access: "readable" }> {
  return { access: "readable", regions };
}

export function savedInspectorRegion(
  id: string,
  content: WorkbookInspectorPanelContentModel,
): WorkbookInspectorRegion {
  return {
    id,
    kind: "snapshot",
    model: { access: "readable", data: { state: "ready", content } },
  };
}

export function ownedInspectorRegion(
  id: string,
  render: (present: PresentRegion) => ReactNode,
): WorkbookInspectorRegion {
  return { id, kind: "owner", render };
}

export function WorkbookInspectorPanelContent({
  model,
}: {
  readonly model: WorkbookInspectorPanelModel;
}) {
  if (model.access === "concealed") return null;
  if (model.regions.length === 0)
    throw new Error("Readable inspector panel requires an owner region");
  const ids = new Set<string>();
  for (const region of model.regions) {
    if (!region.id || ids.has(region.id))
      throw new Error(`Invalid inspector region identity: ${region.id}`);
    ids.add(region.id);
  }
  return (
    <>
      {model.regions.map((region) => (
        <div key={region.id} data-inspector-region={region.id}>
          {region.kind === "snapshot" ? (
            <WorkbookInspectorRegionContent model={region.model} />
          ) : (
            region.render((value) => (
              <WorkbookInspectorRegionContent model={value} />
            ))
          )}
        </div>
      ))}
      {model.authoring}
    </>
  );
}

// Owners determine data and access independently. This component never fetches,
// inspects child output, or equates permission loss with an empty result.
export function WorkbookInspectorRegionContent({
  model,
}: {
  readonly model: WorkbookInspectorRegionModel;
}) {
  if (model.access === "concealed") return null;
  const data = model.data;
  return (
    <div
      data-inspector-data-state={data.state}
      aria-busy={
        data.state === "initial_loading" || data.state === "refreshing"
      }
    >
      {data.state === "initial_loading" ? (
        <p
          id={model.messageId}
          data-testid={model.messageId}
          role={
            model.notice || model.announcement === "owner"
              ? undefined
              : "status"
          }
        >
          {data.message ?? "Loading…"}
        </p>
      ) : null}
      {data.state === "refreshing" ? (
        <p
          role={
            model.notice || model.announcement === "owner"
              ? undefined
              : "status"
          }
        >
          Refreshing…
        </p>
      ) : null}
      {data.state === "unavailable" || data.state === "stale_failure" ? (
        <p
          id={model.messageId}
          data-testid={model.messageId}
          role={
            !model.notice &&
            model.announcement !== "owner" &&
            (data.state === "stale_failure" || data.cause === "load_failed")
              ? "status"
              : undefined
          }
        >
          {data.message}
        </p>
      ) : null}
      {data.state === "stale_failure" && data.content.kind === "empty" ? (
        <p>Last loaded observation:</p>
      ) : null}
      {model.commandsPlacement === "before_content" ? model.commands : null}
      {"content" in data ? (
        data.content.kind === "empty" ? (
          <p
            id={data.state === "stale_failure" ? undefined : model.messageId}
            data-testid={
              data.state === "stale_failure" ? undefined : model.messageId
            }
          >
            {data.content.message}
          </p>
        ) : (
          data.content.content
        )
      ) : null}
      {model.authoring}
      {model.commandsPlacement === "before_content" ? null : model.commands}
      {model.notice ? (
        <WorkbookInspectorNoticeView
          notice={model.notice.value}
          consume={model.notice.consume}
          visible={false}
        />
      ) : null}
    </div>
  );
}
