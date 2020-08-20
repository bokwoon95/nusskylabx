// Package timeutil provides various time related utilities
package timeutil

import (
	"database/sql"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Timestatus struct {
	Now             time.Time
	Start           sql.NullTime
	End             sql.NullTime
	IsOpen          bool
	NotYetOpen      bool
	AlreadyClosed   bool
	InvalidStartEnd bool
}

func Funcs(funcs template.FuncMap) template.FuncMap {
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs["TimeutilResolveTimestatus"] = ResolveTimestatus
	return funcs
}

func ResolveTimestatus(start, end sql.NullTime) (status Timestatus) {
	status.Now = time.Now()
	status.Start.Valid = start.Valid
	status.Start.Time = start.Time
	status.End.Valid = end.Valid
	status.End.Time = end.Time
	switch {
	case status.Start.Valid && status.End.Valid:
		switch {
		case status.Start.Time.After(status.End.Time):
			status.InvalidStartEnd = true
			status.Start.Valid = false
			status.End.Valid = false
		case status.Start.Time.Before(status.Now) && status.Now.Before(status.End.Time):
			status.IsOpen = true
		case status.Now.Before(status.Start.Time):
			status.NotYetOpen = true
		case status.Now.After(status.End.Time):
			status.AlreadyClosed = true
		}
	case status.Start.Valid:
		switch {
		case status.Start.Time.Before(status.Now):
			status.IsOpen = true
		default:
			status.NotYetOpen = true
		}
	case status.End.Valid:
		switch {
		case status.Now.Before(status.End.Time):
			status.IsOpen = true
		default:
			status.AlreadyClosed = true
		}
	default:
		// Do nothing, up to receiver to decide whether both start and end == null means opened or closed
	}
	return status
}

func ParseDateTimeString(datestring, timestring string) (datetime sql.NullTime) {
	if datestring == "" {
		return datetime
	}
	t, err := time.Parse("2006-01-02", datestring)
	if err != nil {
		return datetime
	}
	// The date and time inputs in HTML are oblivious to timezone (implicitly
	// assuming UTC+00 time). Which means whatever date/time strings we get from
	// the front end, we must compensate for Singapore's UTC+08 timezone.
	singapore, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		panic(err) // should never happen as "Asia/Singapore" is a valid postgres location string
	}
	t = t.In(singapore).Add(-8 * time.Hour)
	datetime.Valid = true
	datetime.Time = t
	matches := regexp.MustCompile(`^(\d{2}):(\d{2})$`).FindStringSubmatch(timestring)
	if len(matches) <= 0 {
		return datetime
	}
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	datetime.Time = t.Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute)
	return datetime
}

// Graft a year (in string form) onto a target time.
//
// E.g. if the target time is 01-Jan-2020 and you graft "2021" onto it, result
// is 01-Jan-2021. If the provided year is not a valid integer, nothing will be
// done.
func GraftYear(target sql.NullTime, year string) sql.NullTime {
	if target.Valid {
		yr, err := strconv.Atoi(strings.TrimSpace(year))
		if err != nil {
			return target
		}
		target.Time = time.Date(
			yr,
			target.Time.Month(),
			target.Time.Day(),
			target.Time.Hour(),
			target.Time.Minute(),
			target.Time.Second(),
			target.Time.Nanosecond(),
			target.Time.Location(),
		)
	}
	return target
}
