package packformat

import "github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"

func RelationshipID(profile, typ, source, target string, refs any) (string, error) {
	b, err := canonicaljson.Marshal(refs)
	if err != nil {
		return "", err
	}
	return "rpr_" + Digest(append([]byte(profile+"\x00"+typ+"\x00"+source+"\x00"+target+"\x00"), b...)), nil
}
