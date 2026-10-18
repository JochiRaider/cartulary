//go:build linux

package harnessfixtures

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

var safeName = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)
var component = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

func identifier() string {
	var v [16]byte
	if _, err := rand.Read(v[:]); err != nil {
		panic(err)
	}
	v[6] = (v[6] & 15) | 64
	v[8] = (v[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", v[:4], v[4:6], v[6:8], v[8:10], v[10:])
}
func validateJSON(schema string, value any) ([]byte, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	v, err := harnessschema.Decode(bytes)
	if err == nil {
		err = harnessschema.Validate(schema, v)
	}
	if err != nil {
		err = fmt.Errorf("%s: %w", schema, err)
	}
	return bytes, err
}
func privateJSON(name, schema string, value any, replace bool) error {
	data, err := validateJSON(schema, value)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if replace {
		return harnesslocal.ReplaceFile(name, data)
	}
	return harnesslocal.PublishNew(name, data)
}
