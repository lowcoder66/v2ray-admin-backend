package util

import "time"

func TruncMonth(t time.Time) time.Time {
	m := t.AddDate(0, 0, -t.Day()+1)
	return time.Date(m.Year(), m.Month(), m.Day(), 0, 0, 0, 0, m.Location())
}

func CeilMonth(t time.Time) time.Time {
	m := t.AddDate(0, 1, -t.Day()+1)
	m = time.Date(m.Year(), m.Month(), m.Day(), 0, 0, 0, 0, m.Location())
	return m.Add(1 - time.Microsecond)
}
