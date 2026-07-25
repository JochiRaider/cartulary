//go:build !linux

package securefile

func Read(string, int64) (Document, error) {
	return Document{}, secureFileError("read", FailureUnsupported, "platform is unsupported", ErrUnsupportedPlatform)
}
