import { createContext } from "react";
import type { IndicatorCreateOwnerPort } from "./indicatorCreateOperation";
export const IndicatorCreateContext =
  createContext<IndicatorCreateOwnerPort | null>(null);
