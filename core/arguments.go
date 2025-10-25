package core

import "strings"

type Arguments []string

func (r Arguments) ExactOne(name string) bool {
	return len(r) == 1 && r[0] == name
}

func (r Arguments) IncludeAll(items ...string) bool {
	for _, item := range items {
		if !r.Include(item) {
			return false
		}
	}

	return true
}

func (r Arguments) IncludeOne(items ...string) bool {
	for _, item := range items {
		if r.Include(item) {
			return true
		}
	}

	return false
}

func (r Arguments) Include(needle string) bool {
	for _, b := range r {
		if strings.ToLower(b) == strings.ToLower(needle) {
			return true
		}
	}
	return false
}
