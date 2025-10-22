package core

type Arguments []string

func (r Arguments) ExactOne(name string) bool {
	return len(r) == 1 && r[0] == name
}
