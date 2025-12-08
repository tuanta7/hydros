package x

import "time"

func NowUTC() time.Time {
	return time.Now().UTC()
}

func SecondsFromNow(expiredAt time.Time) time.Duration {
	nanosecondsFromNow := time.Duration(expiredAt.UnixNano() - NowUTC().UnixNano())
	return time.Duration(nanosecondsFromNow.Round(time.Second).Seconds())
}
