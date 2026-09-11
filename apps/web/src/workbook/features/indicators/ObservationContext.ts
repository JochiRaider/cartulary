import { createContext } from "react";
import type { ObservationOwnerPort } from "./observationOperation";
export const ObservationContext = createContext<ObservationOwnerPort | null>(
  null,
);
