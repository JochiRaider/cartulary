import { errorRegistry } from "../generated/error-registry.js";

export { errorRegistry };

export type PublicError = (typeof errorRegistry)["errors"][number];
