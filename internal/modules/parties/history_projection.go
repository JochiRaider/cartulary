package parties

import "github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"

func projectRecordHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("party", "text", "display_name party_kind organization_name role_title primary_email timezone_name external_ref notes")
	return historycontract.Row(facts, fields)
}
