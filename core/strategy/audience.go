package strategy

import (
	"github.com/tuanta7/hydros/core"
)

type AudienceStrategyProvider interface {
	GetAudienceStrategy() AudienceStrategy
}

type AudienceStrategy func(haystack []string, needle []string) error

func ExactAudienceStrategy(haystack []string, needle []string) error {
	if len(needle) == 0 {
		return nil
	}

	for _, n := range needle {
		var found bool
		for _, h := range haystack {
			if n == h {
				found = true
			}
		}

		if !found {
			return core.ErrInvalidRequest.WithHint(`Requested audience "%s" has not been whitelisted by the OAuth 2.0 Client.`, n)
		}
	}

	return nil
}
