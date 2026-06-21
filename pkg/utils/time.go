package utils

import "time"

var (
	PRCLoc    = time.FixedZone("PRC", 8*3600)
	UnixEpoch = time.Unix(0, 0).UTC()
)

func Now() time.Time {
	return time.Now().In(PRCLoc)
}

func ToPRCTime(t time.Time) time.Time {
	return t.In(PRCLoc)
}

func Today() string {
	return Now().Format(time.DateOnly)
}
