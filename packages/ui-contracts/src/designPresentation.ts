import {
  type CartularyErrorFamily,
  type CartularyErrorPresentation,
  type CartularyGridDataState,
  type CartularyGridDataStatePresentation,
  type CartularyGridInteractionMode,
  type CartularyGridInteractionModePresentation,
  cartularyDesignPresentation,
} from "./generated/design-presentation";

export {
  type CartularyErrorFamily,
  type CartularyErrorPresentation,
  type CartularyGridDataState,
  type CartularyGridDataStatePresentation,
  type CartularyGridInteractionMode,
  type CartularyGridInteractionModePresentation,
  type CartularyStatusSecondaryKind,
  cartularyDesignPresentation,
} from "./generated/design-presentation";

export function cartularyErrorPresentation(
  family: CartularyErrorFamily,
): CartularyErrorPresentation {
  const presentation = cartularyDesignPresentation.errorPresentations.find(
    (candidate) => candidate.family === family,
  );
  if (presentation === undefined) {
    throw new Error(`Missing Cartulary error presentation for ${family}`);
  }
  return presentation;
}

export function cartularyGridDataStatePresentation(
  state: CartularyGridDataState,
): CartularyGridDataStatePresentation {
  const presentation =
    cartularyDesignPresentation.gridDataStatePresentations.find(
      (candidate) => candidate.state === state,
    );
  if (presentation === undefined) {
    throw new Error(
      `Missing Cartulary grid data-state presentation for ${state}`,
    );
  }
  return presentation;
}

export function cartularyGridInteractionModePresentation(
  mode: CartularyGridInteractionMode,
): CartularyGridInteractionModePresentation {
  const presentation =
    cartularyDesignPresentation.gridInteractionModePresentations.find(
      (candidate) => candidate.mode === mode,
    );
  if (presentation === undefined) {
    throw new Error(
      `Missing Cartulary grid interaction presentation for ${mode}`,
    );
  }
  return presentation;
}
