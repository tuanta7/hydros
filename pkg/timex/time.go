package timex

import "time"

func NowUTC() time.Time {
	return time.Now().UTC()
}

func IsExpired(created time.Time, expiresInSeconds int64) bool {
	expiredAt := created.Add(time.Duration(expiresInSeconds) * time.Second)
	return NowUTC().After(expiredAt)
}
