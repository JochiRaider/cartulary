package packformat

import "github.com/JochiRaider/cartulary/referencepacks/format/envelope"

func DecodeBootstrap(data []byte) (map[string]TrustSnapshot, error) {
	v, e := envelope.DecodeBootstrap(data)
	return v, formatError(e)
}
func RootVersion(data []byte) (int64, error) {
	v, e := envelope.RootVersion(data)
	return v, formatError(e)
}
