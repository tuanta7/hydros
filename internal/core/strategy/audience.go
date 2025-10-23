package strategy

import "errors"

type AudienceMatchingStrategy func(haystack []string, needle []string) error

func ExactAudienceMatchingStrategy(haystack []string, needle []string) error {
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
			return errors.New("TODO")
		}
	}

	return nil
}
