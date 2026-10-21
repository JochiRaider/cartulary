package harnessdiag

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"os"
)

var openDirectory = harnesslocal.OpenDirectory
var readLocal = harnesslocal.ReadFile
var publishNew = harnesslocal.PublishNew

func readObject(name, schema, role string, optional, mutable bool) (map[string]any, error) {
	data, err := readLocal(name, 16*1024*1024, mutable)
	if optional && errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		var rejection *harnesslocal.ReadRejection
		if errors.As(err, &rejection) {
			return nil, rejected(role, rejection.Stage(), rejection.Reason())
		}
		return nil, rejected(role, "opening", "file_unavailable")
	}
	value, err := Decode(data)
	if err != nil {
		return nil, rejected(role, "decoding", "invalid_document")
	}
	if schema != "" {
		if err := Validate(schema, value); err != nil {
			return nil, rejected(role, "schema", "invalid_document")
		}
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, rejected(role, "decoding", "invalid_document")
	}
	return object, nil
}
