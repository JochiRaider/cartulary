package harnessdiag

import "github.com/JochiRaider/cartulary/internal/testutil/harnessschema"

var errArtifact = harnessschema.ErrInvalid

func Decode(data []byte) (any, error)          { return harnessschema.Decode(data) }
func Validate(id string, value any) error      { return harnessschema.Validate(id, value) }
func schemaPattern(pattern, value string) bool { return harnessschema.MatchPattern(pattern, value) }
