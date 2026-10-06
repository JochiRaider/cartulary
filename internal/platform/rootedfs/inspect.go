package rootedfs

import (
	"errors"
	"io/fs"
)

// InspectEmpty reports whether a root is absent or contains no entries, without
// creating directories. Unsafe roots fail closed. Any entry counts as retained
// state, including directories, links and non-regular files.
func InspectEmpty(rootPath string) (empty bool, retErr error) {
	root, err := Open(rootPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return true, nil
		}
		return false, err
	}
	defer func() { retErr = errors.Join(retErr, root.Close()) }()
	return root.isEmpty()
}
