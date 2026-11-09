package envelope

import (
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
	"path"
)

type shape = schema.Shape

func compileProjection(name string) *shape {
	resources := map[string][]byte{}
	entries, err := schemas.ReadDir("schemas")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if path.Ext(entry.Name()) == ".json" {
			resources[entry.Name()] = projection(entry.Name())
		}
	}
	return schema.Compile(resources, fixedShapeLimits, name)
}
