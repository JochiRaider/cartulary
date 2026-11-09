// Package native owns this producer's source-model obligations. Cartulary does
// not import these schemas or interpret these native fields at runtime.
package native

import (
	"embed"
	"encoding/json"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
	"io/fs"
	"strings"
)

//go:embed schemas/profiles/*.json profile.json limits.json
var inputs embed.FS
var definition = func() struct {
	Key          string            `json:"pack_key"`
	ID           string            `json:"content_profile_id"`
	Shape        string            `json:"payload_shape"`
	Schemas      map[string]string `json:"schemas"`
	MaxLineBytes int               `json:"max_line_bytes"`
} {
	var d struct {
		Key          string            `json:"pack_key"`
		ID           string            `json:"content_profile_id"`
		Shape        string            `json:"payload_shape"`
		Schemas      map[string]string `json:"schemas"`
		MaxLineBytes int               `json:"max_line_bytes"`
	}
	b, err := inputs.ReadFile("profile.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(b, &d); err != nil {
		panic(err)
	}
	return d
}()
var resources = func() map[string][]byte {
	result := map[string][]byte{}
	err := fs.WalkDir(inputs, "schemas", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := inputs.ReadFile(name)
		if err != nil {
			return err
		}
		result[strings.TrimPrefix(name, "schemas/")] = b
		return nil
	})
	if err != nil {
		panic(err)
	}
	return result
}()
var contentShapes = func() map[string]*schema.Shape {
	var limits map[string]int
	b, err := inputs.ReadFile("limits.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(b, &limits); err != nil {
		panic(err)
	}
	result := map[string]*schema.Shape{}
	for kind, name := range definition.Schemas {
		result[definition.Key+"/"+kind] = schema.Compile(resources, limits, name)
	}
	return result
}()

func ProjectionDigest(name string) string {
	b, ok := resources[name]
	if !ok {
		panic("unknown producer schema")
	}
	canonical, err := canonicaljson.Canonicalize(b)
	if err != nil {
		panic(err)
	}
	return envelope.Digest(canonical)
}
