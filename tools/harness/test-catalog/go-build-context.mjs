import { readExecutableRepositoryJSON } from "./restricted-input-boundary.mjs";
import { independentGoModules } from "./go-modules.mjs";

// The pinned Go environment (tools/toolchain_pins.json go_environment, the
// pinned toolchain and cryptographic module) that every build and test runs
// under; internal/testutil/goenv reads the same table. Static discovery and
// execution share this projection, and runtime profiles never add build tags.
const goSettings = ["CGO_ENABLED", "GO111MODULE", "GOAMD64", "GOARCH", "GOENV", "GOEXPERIMENT", "GOFLAGS", "GOOS", "GOWORK"];

export function goRunnerBuildContext(root) {
  const pins = readExecutableRepositoryJSON(root, "tools/toolchain_pins.json");
  const version = /^(1)\.(\d+)(?:\.(\d+))?$/u.exec(String(pins.go_version ?? ""));
  if (!version) throw new Error("tools/toolchain_pins.json.go_version must be a Go 1.x or Go 1.x.patch release");
  const declared = pins.go_environment ?? {};
  if (Object.keys(declared).sort().join(",") !== goSettings.join(",") || Object.values(declared).some((value) => typeof value !== "string")) {
    throw new Error(`tools/toolchain_pins.json.go_environment must declare exactly ${goSettings.join(", ")}`);
  }
  if (pins.go_toolchain !== `go${pins.go_version}` || typeof pins.cryptographic_module?.selector !== "string") {
    throw new Error("tools/toolchain_pins.json must pin go_toolchain for go_version and the cryptographic module selector");
  }
  const environment = Object.freeze({ ...declared, GOTOOLCHAIN: pins.go_toolchain, GOFIPS140: pins.cryptographic_module.selector });
  const tags = new Set([environment.GOARCH, `${environment.GOARCH}.${environment.GOAMD64}`, "gc", environment.GOOS, "unix"]);
  for (let minor = 1; minor <= Number(version[2]); minor += 1) tags.add(`go1.${minor}`);
  return Object.freeze({
    modules: independentGoModules(root),
    goarch: environment.GOARCH,
    goos: environment.GOOS,
    label: `${environment.GOOS}/${environment.GOARCH}.${environment.GOAMD64} Go ${pins.go_version}, gc, cgo ${environment.CGO_ENABLED === "0" ? "disabled" : "enabled"}, no custom build tags`,
    tags,
    environment,
  });
}
