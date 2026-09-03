import { useCallback, useEffect, useRef } from "react";
import type {
  ActiveSurfaceSavedViewProjection,
  SavedViewActionIdentity,
  SavedViewControlEvent,
  SavedViewEditableScope,
} from "../models/workbookSavedViewControl";
import {
  canMutateSavedView,
  type SavedViewResource,
} from "../models/workbookSavedViews";

export type SavedViewActionIntent =
  | {
      readonly kind: "create";
      readonly displayName: string;
      readonly scope: SavedViewEditableScope;
    }
  | {
      readonly kind: "update";
      readonly displayName: string;
      readonly scope: SavedViewEditableScope;
    }
  | { readonly kind: "duplicate" }
  | { readonly kind: "reset" }
  | { readonly kind: "delete" }
  | { readonly kind: "set_home" }
  | { readonly kind: "set_default" };

type SavedViewActionPorts = {
  readonly create: (input: {
    readonly displayName: string;
    readonly scope: SavedViewEditableScope;
  }) => Promise<SavedViewResource>;
  readonly delete: (savedView: SavedViewResource) => Promise<void>;
  readonly duplicate: (
    savedView: SavedViewResource,
  ) => Promise<SavedViewResource>;
  readonly reset: (savedView: SavedViewResource) => void;
  readonly setDefault: () => Promise<void>;
  readonly setHome: () => Promise<void>;
  readonly update: (
    savedView: SavedViewResource,
    input: {
      readonly displayName: string;
      readonly scope: SavedViewEditableScope;
    },
  ) => Promise<SavedViewResource>;
};

type CurrentSavedViewActionContext = {
  readonly activeViewSchemaId: string;
  readonly currentIncidentRole: string | null;
  readonly currentUserId: string | null;
  readonly isModified: boolean;
  readonly ports: SavedViewActionPorts;
  readonly projection: ActiveSurfaceSavedViewProjection;
};

export function useActiveSurfaceSavedViewActions({
  activeViewSchemaId,
  currentIncidentRole,
  currentUserId,
  dispatch,
  isModified,
  ports,
  projection,
}: CurrentSavedViewActionContext & {
  readonly dispatch: (event: SavedViewControlEvent) => void;
}) {
  const generationRef = useRef(0);
  const activeActionRef = useRef<SavedViewActionIdentity | null>(null);
  const currentRef = useRef<CurrentSavedViewActionContext>({
    activeViewSchemaId,
    currentIncidentRole,
    currentUserId,
    isModified,
    ports,
    projection,
  });
  currentRef.current = {
    activeViewSchemaId,
    currentIncidentRole,
    currentUserId,
    isModified,
    ports,
    projection,
  };
  const actionContextKey = `${activeViewSchemaId}:${projection.resourceKind}:${projection.selectedSavedView?.saved_view_id ?? "base"}:${projection.selectedSavedView?.saved_view_version ?? "none"}:${currentUserId ?? "anonymous"}:${currentIncidentRole ?? "none"}:${isModified ? "modified" : "clean"}`;

  useEffect(() => {
    if (actionContextKey === "") return;
    const activeAction = activeActionRef.current;
    if (
      activeAction === null ||
      actionIdentityIsCurrent(activeAction, currentRef.current)
    ) {
      return;
    }
    activeActionRef.current = null;
    dispatch({
      type: "invalidate_action",
      identity: activeAction,
      message:
        "Saved-view action was superseded by a newer surface or selection.",
    });
  }, [actionContextKey, dispatch]);

  const runAction = useCallback(
    (intent: SavedViewActionIntent) => {
      const current = currentRef.current;
      const existing = activeActionRef.current;
      if (existing !== null) return;
      const selected = current.projection.selectedSavedView;
      const identity: SavedViewActionIdentity = {
        surface: current.activeViewSchemaId,
        savedViewId: selected?.saved_view_id ?? null,
        savedViewVersion: selected?.saved_view_version ?? null,
        actionKind: intent.kind,
        generation: generationRef.current + 1,
      };
      generationRef.current = identity.generation;
      const unavailableReason = actionUnavailableReason(intent, current);
      if (unavailableReason !== null) {
        dispatch({
          type: "publish_notice",
          surface: current.activeViewSchemaId,
          message: unavailableReason,
        });
        return;
      }
      activeActionRef.current = identity;
      dispatch({ type: "start_action", identity });

      void executeAction(intent, current).then(
        () => {
          if (
            activeActionRef.current !== identity ||
            !actionIdentityIsCurrent(identity, currentRef.current)
          ) {
            return;
          }
          activeActionRef.current = null;
          dispatch({
            type: "complete_action",
            identity,
            feedback: {
              kind: "success",
              message: successMessage(intent.kind),
            },
          });
        },
        (error: unknown) => {
          if (
            activeActionRef.current !== identity ||
            !actionIdentityIsCurrent(identity, currentRef.current)
          ) {
            return;
          }
          activeActionRef.current = null;
          dispatch({
            type: "complete_action",
            identity,
            feedback: {
              kind: "error",
              message:
                error instanceof Error ? error.message : "Saved view failed.",
            },
          });
        },
      );
    },
    [dispatch],
  );

  return { runAction };
}

function actionIdentityIsCurrent(
  identity: SavedViewActionIdentity,
  current: CurrentSavedViewActionContext,
): boolean {
  const selected = current.projection.selectedSavedView;
  if (
    identity.surface !== current.activeViewSchemaId ||
    identity.savedViewId !== (selected?.saved_view_id ?? null) ||
    identity.savedViewVersion !== (selected?.saved_view_version ?? null)
  ) {
    return false;
  }
  return (
    actionUnavailableReason({ kind: identity.actionKind }, current) === null
  );
}

function actionUnavailableReason(
  intent: Pick<SavedViewActionIntent, "kind">,
  current: CurrentSavedViewActionContext,
): string | null {
  return (
    resourceActionUnavailableReason(current.projection.resourceKind) ??
    admittedActionUnavailableReason(intent.kind, current)
  );
}

function resourceActionUnavailableReason(
  resourceKind: ActiveSurfaceSavedViewProjection["resourceKind"],
): string | null {
  switch (resourceKind) {
    case "loading":
      return "Saved views are still loading.";
    case "unavailable":
      return "Saved-view actions are unavailable until the view list can be loaded.";
    case "invalid_selection":
      return "Select an available view or the base surface before changing saved views.";
    case "ready":
      return null;
  }
}

function admittedActionUnavailableReason(
  actionKind: SavedViewActionIntent["kind"],
  current: CurrentSavedViewActionContext,
): string | null {
  const selected = current.projection.selectedSavedView;
  switch (actionKind) {
    case "create":
    case "set_home":
      return null;
    case "set_default":
      return current.currentIncidentRole === "admin"
        ? null
        : "Only incident administrators can set the incident default.";
    case "duplicate":
      return selectedRequiredReason(selected, "duplicate");
    case "reset":
      return resetUnavailableReason(selected, current.isModified);
    case "update":
    case "delete":
      return mutationUnavailableReason(selected, current);
  }
}

function selectedRequiredReason(
  selected: SavedViewResource | null,
  action: "duplicate" | "restore",
): string | null {
  return selected === null ? `Select a saved view to ${action}.` : null;
}

function resetUnavailableReason(
  selected: SavedViewResource | null,
  isModified: boolean,
): string | null {
  return (
    selectedRequiredReason(selected, "restore") ??
    (isModified
      ? null
      : "The selected saved view already matches its saved configuration.")
  );
}

function mutationUnavailableReason(
  selected: SavedViewResource | null,
  current: CurrentSavedViewActionContext,
): string | null {
  return canMutateSavedView(
    selected,
    current.currentUserId,
    current.currentIncidentRole,
  )
    ? null
    : "Only the owner or incident administrator can change this view.";
}

async function executeAction(
  intent: SavedViewActionIntent,
  current: CurrentSavedViewActionContext,
): Promise<void> {
  const selected = current.projection.selectedSavedView;
  switch (intent.kind) {
    case "create":
      await current.ports.create({
        displayName: intent.displayName,
        scope: intent.scope,
      });
      return;
    case "update":
      if (selected === null) return;
      await current.ports.update(selected, {
        displayName: intent.displayName,
        scope: intent.scope,
      });
      return;
    case "duplicate":
      if (selected === null) return;
      await current.ports.duplicate(selected);
      return;
    case "reset":
      if (selected === null) return;
      current.ports.reset(selected);
      return;
    case "delete":
      if (selected === null) return;
      await current.ports.delete(selected);
      return;
    case "set_home":
      await current.ports.setHome();
      return;
    case "set_default":
      await current.ports.setDefault();
      return;
  }
}

function successMessage(kind: SavedViewActionIntent["kind"]): string {
  switch (kind) {
    case "create":
      return "Saved view created.";
    case "update":
      return "Saved view updated.";
    case "duplicate":
      return "Saved view duplicated.";
    case "reset":
      return "Saved configuration restored.";
    case "delete":
      return "Saved view deleted.";
    case "set_home":
      return "Home view updated.";
    case "set_default":
      return "Default view updated.";
  }
}
