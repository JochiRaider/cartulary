package publication

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func signingOutside(privatePath, publicPath string) error {
	resolve := func(p string) (string, error) {
		absolute, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		// Resolve existing ancestors as well as the final component, including
		// the creation case where the private file does not exist yet.
		parts := []string{}
		current := absolute
		for {
			resolved, err := filepath.EvalSymlinks(current)
			if err == nil {
				return filepath.Join(append([]string{resolved}, parts...)...), nil
			}
			if !errors.Is(err, os.ErrNotExist) {
				return "", err
			}
			parent := filepath.Dir(current)
			if parent == current {
				return "", err
			}
			parts = append([]string{filepath.Base(current)}, parts...)
			current = parent
		}
	}
	private, err := resolve(privatePath)
	if err != nil {
		return err
	}
	public, err := resolve(publicPath)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(public, private)
	if err != nil {
		return err
	}
	if relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("private signing context must be outside retained public roots")
	}
	return nil
}
