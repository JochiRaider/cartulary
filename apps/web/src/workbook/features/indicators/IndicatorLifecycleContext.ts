import { createContext } from "react";
import type { IndicatorLifecycleOwnerPort } from "./indicatorLifecycleOperation";
export const IndicatorLifecycleContext =
  createContext<IndicatorLifecycleOwnerPort | null>(null);
