package repository

func jsonRawOrNil(raw []byte) interface{} {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
