package envelope

import "embed"

// Only library-owned, compiled schemas are resolvable. A pack cannot supply a
// schema, URL, filesystem reference or instruction to this compiler.
//
//go:embed schemas/*.json
var schemas embed.FS

func projection(name string) []byte {
	data, err := schemas.ReadFile("schemas/" + name)
	if err != nil {
		panic("reference format: missing compiled schema")
	}
	return data
}
