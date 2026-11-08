package producer

import (
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

// compileProducerSchema resolves only explicitly supplied resources. It has no
// network or filesystem loaders and is private to offline producer contracts.
func compileProducerSchema(b []byte, fragment string) (*jsonschema.Schema, error) {
	v, err := canonicaljson.DecodeStrict(b)
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	c.UseLoader(jsonschema.SchemeURLLoader{})
	c.DefaultDraft(jsonschema.Draft2020)
	c.AssertFormat()
	const location = "https://cartulary.invalid/proposals/input"
	if err = c.AddResource(location, v); err != nil {
		return nil, err
	}
	return c.Compile(location + fragment)
}
