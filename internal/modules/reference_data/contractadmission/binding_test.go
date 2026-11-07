package contractadmission

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestReferenceBindingLoadFailures(t *testing.T) {
	id := "rpset_" + strings.Repeat("a", 64)
	binding := []byte(`{"pack_set_id":"` + id + `","pack_set_sha256":"` + strings.Repeat("a", 64) + `","provenance":[]}`)
	for _, cause := range []error{errors.New("database unavailable"), context.Canceled, context.DeadlineExceeded} {
		t.Run(cause.Error(), func(t *testing.T) {
			called := false
			err := validateBinding(binding, func(requested string) ([]byte, error) {
				called = true
				if requested != id {
					t.Fatal("lookup changed retained identity", requested)
				}
				return nil, cause
			})
			if !called || !errors.Is(err, cause) || IsIncompatible(err) {
				t.Fatal("binding lookup lost operational failure", err)
			}
		})
	}
	for _, raw := range [][]byte{nil, []byte(`{"schema_id":"unsupported"}`)} {
		if err := validateBinding(binding, func(string) ([]byte, error) { return raw, nil }); !IsIncompatible(err) {
			t.Fatal("missing or incompatible retained set was not rejected", err)
		}
	}
}
