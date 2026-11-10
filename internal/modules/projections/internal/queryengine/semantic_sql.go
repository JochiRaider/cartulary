package queryengine

import "github.com/JochiRaider/cartulary/internal/platform/sqlshape"

func recordRefItemRefSQL(recordIDExpr sqlshape.Fragment) sqlshape.Fragment {
	return sqlshape.Format("'record_ref:' || %s::text", recordIDExpr)
}
func partyRefItemRefSQL(recordIDExpr sqlshape.Fragment) sqlshape.Fragment {
	return sqlshape.Format("'party_ref:' || %s::text", recordIDExpr)
}
func recordTagItemRefSQL(recordIDExpr, recordTagIDExpr sqlshape.Fragment) sqlshape.Fragment {
	return sqlshape.Format("'record_tag:' || %s::text || ':' || %s::text", recordIDExpr, recordTagIDExpr)
}
