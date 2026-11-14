import path from "node:path";
// Names are derived from the admitted allocation. This parser provides input
// validation only; the native journal is the authority for physical cleanup.
export const packageSourceProjectPattern = /^cartulary-qualification-[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}-source$/u;
export function packageDestinationProject(source) {
 if (typeof source !== "string" || !packageSourceProjectPattern.test(source)) throw new Error("invalid package source project");
 return source.replace(/-source$/u, "-destination");
}
export function assertPackageQualificationSource(work, project, environment = process.env) {
 const destination = packageDestinationProject(project);
 if (!path.isAbsolute(work || "") || work !== environment.CARTULARY_PACKAGE_WORK_DIR ||
     project !== environment.CARTULARY_PACKAGE_SOURCE_PROJECT || destination !== environment.CARTULARY_PACKAGE_DESTINATION_PROJECT) {
  throw new Error("package qualification binding mismatch");
 }
 return destination;
}
