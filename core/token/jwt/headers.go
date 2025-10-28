package jwt

type Headers struct {
	fields map[string]any
}

func (h Headers) Add(key string, value any) {
	if h.fields == nil {
		h.fields = make(map[string]any)
	}
	h.fields[key] = value
}

func (h Headers) Get(key string) any {
	return h.fields[key]
}
