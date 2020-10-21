package util

import "time"

func ParseTime(value string) (time.Time, error) {
	loc := time.FixedZone("UTC+8", +8*60*60)

	return time.ParseInLocation("2006-01-02 15:04:05", value, loc)
}

func ParseDate(value string) (time.Time, error) {
	loc := time.FixedZone("UTC+8", +8*60*60)

	return time.ParseInLocation("2006-01-02", value, loc)
}

func TimeFormat(t time.Time) string {
	loc := time.FixedZone("UTC+8", +8*60*60)
	return t.In(loc).Format("2006-01-02 15:04:05")
}

type MyTime time.Time

func (myT MyTime) MarshalText() (data []byte, err error) {
	t := time.Time(myT)
	data = []byte(TimeFormat(t))
	return
}

func (myT *MyTime) UnmarshalText(text []byte) (err error) {
	t := (*time.Time)(myT)
	*t, err = ParseTime(string(text))
	return
}
