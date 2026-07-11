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

// QuerySurface is an internal PostgreSQL persistence descriptor shared by
// source-owner projection providers and the projections query engine. Its SQL
// members are compiled repository constants, never request-supplied text. It
// does not define public workbook query semantics.
type QuerySurface struct {
	ViewSchemaID string
	FromSQL      string
	RecordExpr   string
	IncidentExpr string
	WhereSQL     string
	Fields       []QueryField
}
