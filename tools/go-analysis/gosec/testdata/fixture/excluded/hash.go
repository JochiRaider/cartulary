//go:build ignore

package excluded

import "crypto/md5"

// Digest is excluded from every build, so no package is analyzed here.
func Digest(data []byte) [16]byte { return md5.Sum(data) }
