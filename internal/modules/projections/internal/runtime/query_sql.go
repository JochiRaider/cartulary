package runtime

import (
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func buildGenericQueryPageSQL(incidentID uuid.UUID, definition genericSurface, query viewschema.QueryMeta, window querypage.Window) (string, []any, error) {
	return queryengine.BuildQueryPageSQL(incidentID, definition.queryEngineSurface(), query, window)
}
