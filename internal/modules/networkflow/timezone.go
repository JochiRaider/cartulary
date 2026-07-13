package networkflow

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/gen/networkflowtz"
)

var pinnedTimezoneCache sync.Map

func loadPinnedTimezone(name string) (*time.Location, error) {
	if networkflowtz.RulesetID != timestampRulesetID {
		return nil, fmt.Errorf("embedded Network Flow timezone ruleset is %q", networkflowtz.RulesetID)
	}
	if cached, ok := pinnedTimezoneCache.Load(name); ok {
		return cached.(*time.Location), nil
	}
	reader, err := zip.NewReader(strings.NewReader(networkflowtz.ZoneinfoZip), int64(len(networkflowtz.ZoneinfoZip)))
	if err != nil {
		return nil, fmt.Errorf("open embedded Network Flow timezone bundle: %w", err)
	}
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}
		opened, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(io.LimitReader(opened, 1<<20))
		closeErr := opened.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		location, err := time.LoadLocationFromTZData(name, data)
		if err != nil {
			return nil, fmt.Errorf("load %s from embedded Network Flow timezone bundle: %w", name, err)
		}
		actual, _ := pinnedTimezoneCache.LoadOrStore(name, location)
		return actual.(*time.Location), nil
	}
	return nil, fmt.Errorf("timezone %q is absent from embedded Network Flow ruleset", name)
}
