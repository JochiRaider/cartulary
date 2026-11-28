package seeded

import "crypto/md5"

// Digest uses a weak hash, which G401 reports.
func Digest(data []byte) []byte {
	sum := md5.Sum(data)
	return sum[:]
}

// Suppressed uses the same hash under a suppression.
func Suppressed(data []byte) []byte {
	sum := md5.Sum(data) // #nosec G401 -- fixture suppression
	return sum[:]
}
