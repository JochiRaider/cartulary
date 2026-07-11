package projections

import "github.com/JochiRaider/cartulary/internal/modules/projections/queryengine"

func buildGenericRow(definition genericSurface, values []any) (map[string]any, error) {
	return queryengine.BuildRow(definition.queryEngineSurface(), values)
}
