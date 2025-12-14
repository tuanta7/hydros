package mapx

import "errors"

var (
	ErrKeyCanNotBeTypeAsserted = errors.New("key could not be type asserted")
	ErrKeyDoesNotExist         = errors.New("key is not present in map")
)

func GetStringDefault[K comparable](values map[K]any, key K, defaultValue string) string {
	if s, err := GetString(values, key); err == nil {
		return s
	}

	return defaultValue
}

func GetString[K comparable](values map[K]any, key K) (string, error) {
	if v, ok := values[key]; !ok {
		return "", ErrKeyDoesNotExist
	} else if sv, ok := v.(string); !ok {
		return "", ErrKeyCanNotBeTypeAsserted
	} else {
		return sv, nil
	}
}
