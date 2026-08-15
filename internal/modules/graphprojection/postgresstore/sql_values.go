package postgresstore

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}
