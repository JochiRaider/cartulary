//go:build linux

package harnessruntime

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

// The scan is required even when optional evidence is absent. Failed scans
// never write a success record and remain secondary to an existing failure.
func publishScan(root string, secrets []string) error {
	v, err := harnesssession.ScanRetained(root, secrets)
	if err != nil {
		return errors.Join(failed("artifact", "artifact_error"), err)
	}
	return writeEvidence(root, "retained-secret-scan.json", "cartulary.harness_retained_secret_scan.v1", v, false)
}
