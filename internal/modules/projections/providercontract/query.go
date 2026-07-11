package providercontract

const DescriptorSchemaVersion = "projection_provider_descriptor.v3"

type FieldKind string

const (
	FieldKindText       FieldKind = "text"
	FieldKindTimestamp  FieldKind = "timestamp"
	FieldKindDate       FieldKind = "date"
	FieldKindBool       FieldKind = "bool"
	FieldKindNumber     FieldKind = "number"
	FieldKindCollection FieldKind = "collection"
)

type QueryField struct {
	Key      string
	Expr     string
	SortExpr string
	Kind     FieldKind
	Ordered  bool
}

type QuerySurface struct {
	ViewSchemaID string
	FromSQL      string
	RecordExpr   string
	IncidentExpr string
	WhereSQL     string
	Fields       []QueryField
}
