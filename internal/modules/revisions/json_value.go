package revisions

import "encoding/json"

func jsonOrNil(value any) any {
	if value == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return payload
}
