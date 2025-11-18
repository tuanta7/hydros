package strategy

import (
	"strings"

	"github.com/tuanta7/hydros/core"
)

type ScopeStrategyProvider interface {
	GetScopeStrategy() ScopeStrategy
}

type ScopeStrategy func(haystack []string, needles []string) error

func ExactScopeStrategy(haystack []string, needles []string) error {
	if len(needles) == 0 {
		return nil
	}

	for _, needle := range needles {
		var found bool
		for _, h := range haystack {
			if needle == h {
				found = true
			}
		}

		if !found {
			return core.ErrInvalidScope.WithHint("The request scope '%s' has not been granted or is not allowed to be requested.", needle)
		}
	}

	return nil
}

func PrefixScopeStrategy(haystack []string, needles []string) error {
	if len(needles) == 0 {
		return nil
	}

	for _, needle := range needles {
		var found bool
		needleParts := strings.Split(needle, ".")

		for _, candidate := range haystack {
			// exact match
			if candidate == needle {
				found = true
				break
			}

			candidateParts := strings.Split(candidate, ".")

			// if a candidate is more specific (has more parts) than the requested broad scope,
			// it cannot include the requested broader scope (e.g., candidate "a.b" vs. request "a")
			if len(candidateParts) > len(needleParts) {
				continue
			}

			// check whether a candidate is a prefix of the requested scope
			isPrefix := true
			for i := 0; i < len(candidateParts); i++ {
				if candidateParts[i] != needleParts[i] {
					isPrefix = false
					break
				}
			}

			if isPrefix {
				found = true
				break
			}
		}

		if !found {
			return core.ErrInvalidScope.WithHint("The request scope '%s' has not been granted or is not allowed to be requested.", needle)
		}
	}

	return nil
}
