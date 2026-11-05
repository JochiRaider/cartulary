package postgresresult

// WithResultLimitForTest exercises the real SQL guard using a small envelope.
func (reader Reader) WithResultLimitForTest(limit int) *Reader {
	reader.maxResultBytes = limit
	return &reader
}
