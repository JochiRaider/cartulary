package viewschema

type Filter struct {
	FieldKey string         `json:"field_key"`
	Op       string         `json:"op"`
	Arg      map[string]any `json:"arg"`
}

type QueryMeta struct {
	Filters []Filter    `json:"filters"`
	Sort    []SortEntry `json:"sort"`
	GroupBy *string     `json:"group_by,omitempty"`
}

func (s Schema) DefaultQueryMeta() QueryMeta {
	return QueryMeta{
		Filters: []Filter{},
		Sort:    s.DefaultSort(),
	}
}
