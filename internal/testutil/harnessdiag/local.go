package harnessdiag

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"os"
)

var openDirectory = harnesslocal.OpenDirectory
var readLocal = harnesslocal.ReadFile
var publishNew = harnesslocal.PublishNew

func readObject(name, schema string, optional, mutable bool) (map[string]any, error) {
	data, err := readLocal(name, 16*1024*1024, mutable)
	if optional && errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, errArtifact
	}
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}
	if schema != "" {
		if err := Validate(schema, value); err != nil {
			return nil, err
		}
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, errArtifact
	}
	return object, nil
}
