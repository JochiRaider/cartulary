import {
  type CartularyErrorFamily,
  type CartularyErrorPresentation,
  cartularyDesignPresentation,
} from "./generated/design-presentation";

export {
  type CartularyErrorFamily,
  type CartularyErrorPresentation,
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
