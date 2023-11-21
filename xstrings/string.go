package xstrings

type SecretString string

func (s SecretString) MarshalJSON() ([]byte, error) {
	return []byte(`"*secret*"`), nil
}
