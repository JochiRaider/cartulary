//go:build linux

package performancefixturelifecycle

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// Capacity is an admission lower bound, not a reservation or a guarantee that
// later writes will succeed. Metadata and concurrent writers consume more space.
func checkObjectSnapshotCapacity(directory string, requiredBytes int64) error {
	if requiredBytes == 0 {
		return nil
	}
	root, err := os.Open(directory) // #nosec G304 -- exact parent derived from the admitted private suite runtime.
	if err != nil {
		return errors.New("performance fixture storage capacity is unavailable")
	}
	defer root.Close()
	var stats unix.Statfs_t
	if err := unix.Fstatfs(int(root.Fd()), &stats); err != nil {
		return errors.New("performance fixture storage capacity is unavailable")
	}
	return validateObjectSnapshotCapacity(requiredBytes, stats.Bsize, stats.Bavail)
}

func validateObjectSnapshotCapacity(requiredBytes, blockBytes int64, availableBlocks uint64) error {
	if requiredBytes < 0 || blockBytes <= 0 {
		return errors.New("performance fixture storage capacity is invalid")
	}
	// Round required bytes up without multiplying potentially overflowing
	// filesystem counters, including on filesystems with very large capacities.
	requiredBlocks := uint64(requiredBytes / blockBytes)
	if requiredBytes%blockBytes != 0 {
		requiredBlocks++
	}
	if requiredBlocks > availableBlocks {
		return errors.New("performance fixture object population exceeds available scratch storage; configure CARTULARY_HARNESS_SCRATCH_ROOT on a sufficiently sized external filesystem")
	}
	return nil
}
