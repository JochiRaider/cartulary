import { prepareSuiteAdminState } from "./helpers";

export default async function globalSetup() {
  await prepareSuiteAdminState();
}
