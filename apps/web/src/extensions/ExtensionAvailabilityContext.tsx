import { createContext, type ReactNode, useContext } from "react";
import type { ExtensionAvailabilityController } from "./extensionAvailability";

const ExtensionAvailabilityContext =
  createContext<ExtensionAvailabilityController | null>(null);

export function ExtensionAvailabilityProvider({
  children,
  controller,
}: {
  readonly children: ReactNode;
  readonly controller: ExtensionAvailabilityController;
}) {
  return (
    <ExtensionAvailabilityContext.Provider value={controller}>
      {children}
    </ExtensionAvailabilityContext.Provider>
  );
}

export function useExtensionAvailabilityController() {
  const controller = useContext(ExtensionAvailabilityContext);
  if (controller === null) {
    throw new Error("extension availability controller is unavailable");
  }
  return controller;
}
