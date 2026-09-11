import { createContext } from "react";
import type { DecisionSupersessionOwnerPort } from "./decisionSupersessionOperation";

export const DecisionSupersessionContext =
  createContext<DecisionSupersessionOwnerPort | null>(null);
