//go:build !linux

package performancefixturelifecycle

import "errors"

func checkObjectSnapshotCapacity(_ string, requiredBytes int64) error {
	if requiredBytes == 0 {
		return nil
	}
	return errors.New("performance fixture object storage capacity requires the supported Linux runtime")
}
