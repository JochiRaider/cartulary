package inner

import "crypto/md5"

// Digest is in a nested module, which loads from its own root.
func Digest(data []byte) [16]byte { return md5.Sum(data) }
