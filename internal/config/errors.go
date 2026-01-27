package config

type UnknownKeyError struct {
	Key string
}

func (e *UnknownKeyError) Error() string {
	return "unknown key: " + e.Key
}
