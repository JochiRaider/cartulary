package workbook_test

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"net/http"
	"testing"
)

func TestWorkbookStartupConstructionFailure(t *testing.T) {
	constructionErr := errors.New("startup construction failed")
	registrar := workbook.RegisterRoutes(workbook.RouteDependencies{
		StartupStoreFactory: func(httpapi.DependencySet) (*workbookstartup.Store, error) {
			return nil, constructionErr
		},
	})
	err := registrar(http.NewServeMux(), httpapi.DependencySet{})
	if !errors.Is(err, constructionErr) {
		t.Fatalf("route composition did not propagate startup construction error: %v", err)
	}
}
