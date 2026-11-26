package goenv

import (
	"strings"
	"testing"
)

// The pinned Go environment is the declared settings with the pinned
// toolchain and cryptographic module; the repository's pins declare it with
// cgo disabled.
func TestPinnedEnvironment(t *testing.T) {
	environment, err := Read("../../..")
	if err != nil {
		t.Fatal(err)
	}
	if environment["CGO_ENABLED"] != "0" || environment["GOOS"] != "linux" || environment["GOENV"] != "off" || environment["GOFLAGS"] != "" ||
		!strings.HasPrefix(environment["GOTOOLCHAIN"], "go1.") || environment["GOFIPS140"] == "" {
		t.Fatalf("pinned environment %v", environment)
	}
	if Race(environment)["CGO_ENABLED"] != "1" || environment["CGO_ENABLED"] != "0" {
		t.Fatal("a race build must enable cgo without changing the pinned environment")
	}
	entries := Entries(map[string]string{"B": "2", "A": "1"})
	if strings.Join(entries, ",") != "A=1,B=2" {
		t.Fatalf("entries %v", entries)
	}
	settings := `"GOOS":"linux","GOARCH":"amd64","GOAMD64":"v1","CGO_ENABLED":"0","GOFLAGS":"","GOENV":"off","GOEXPERIMENT":"","GOWORK":"off","GO111MODULE":"on"`
	for name, pins := range map[string]string{
		"a missing setting":  `{"go_toolchain":"go1.27.4","go_environment":{"GOOS":"linux"},"cryptographic_module":{"selector":"s"}}`,
		"an extra setting":   `{"go_toolchain":"go1.27.4","go_environment":{` + settings + `,"GOROOT":"/x"},"cryptographic_module":{"selector":"s"}}`,
		"no toolchain":       `{"go_environment":{` + settings + `},"cryptographic_module":{"selector":"s"}}`,
		"no module selector": `{"go_toolchain":"go1.27.4","go_environment":{` + settings + `}}`,
	} {
		if _, err := Parse([]byte(pins)); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}
