package projections

import (
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/projectionprovider"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionprovider"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/projectionprovider"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func contractQuerySurfacesForTest() map[string]genericSurface {
	contracts := make([]providercontract.QuerySurface, 0)
	contracts = append(contracts, timelineprojection.QuerySurfaces()...)
	contracts = append(contracts, assessmentprojection.QuerySurfaces()...)
	contracts = append(contracts, artifactprojection.QuerySurfaces()...)
	contracts = append(contracts, evidenceprojection.QuerySurfaces()...)
	contracts = append(contracts, indicatorprojection.QuerySurfaces()...)
	contracts = append(contracts, partyprojection.QuerySurfaces()...)
	contracts = append(contracts, taskdecisionprojection.TaskRequestQuerySurfaces()...)
	contracts = append(contracts, taskdecisionprojection.DecisionQuerySurfaces()...)
	surfaces := make(map[string]genericSurface, len(contracts))
	for _, contract := range contracts {
		surface, err := genericSurfaceFromContract(contract)
		if err != nil {
			panic(err)
		}
		if _, exists := surfaces[surface.viewSchemaID]; exists {
			panic("duplicate query surface " + surface.viewSchemaID)
		}
		surfaces[surface.viewSchemaID] = surface
	}
	return surfaces
}
