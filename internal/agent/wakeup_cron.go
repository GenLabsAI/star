package agent

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Cron schedule support for wakeups: standard five-field cron
// (minute hour day-of-month month day-of-week) plus @hourly,
// @daily, @weekly, and @every <duration>. It covers the requested
// shapes — every N minutes, every morning — without a new
// dependency. Field syntax: *, */step, a-b, a-b/step, lists, and
// English month/weekday names.

// maxCronLookahead bounds the minute-by-minute search for the next
// occurrence to about a year.
const maxCronLookahead = 366 * 24 * 60

// NextCronOccurrence returns the next fire strictly after from for
// a cron spec.
func NextCronOccurrence(spec string, from time.Time) (time.Time, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return time.Time{}, fmt.Errorf("Cron schedule is empty")
	}
	if rest, ok := strings.CutPrefix(spec, "@every "); ok {
		d, err := time.ParseDuration(strings.TrimSpace(rest))
		if err != nil || d <= 0 {
			return time.Time{}, fmt.Errorf("Invalid @every cron schedule %q: use a positive duration", spec)
		}
		return from.Add(d), nil
	}
	switch spec {
	case "@hourly":
		return from.Truncate(time.Hour).Add(time.Hour), nil
	case "@daily":
		return nextDaily(from, 0, 0), nil
	case "@weekly":
		return nextWeekly(from, time.Sunday, 0, 0), nil
	}
	fields := strings.Fields(spec)
	if len(fields) != 5 {
		return time.Time{}, fmt.Errorf("Invalid cron schedule %q: want five fields (minute hour day month weekday)", spec)
	}
	minutes, err := parseCronField(fields[0], 0, 59, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("Invalid cron minute in %q: %w", spec, err)
	}
	hours, err := parseCronField(fields[1], 0, 23, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("Invalid cron hour in %q: %w", spec, err)
	}
	dom, err := parseCronField(fields[2], 1, 31, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("Invalid cron day-of-month in %q: %w", spec, err)
	}
	months, err := parseCronField(fields[3], 1, 12, monthNames)
	if err != nil {
		return time.Time{}, fmt.Errorf("Invalid cron month in %q: %w", spec, err)
	}
	dow, err := parseCronField(fields[4], 0, 7, weekdayNames)
	if err != nil {
		return time.Time{}, fmt.Errorf("Invalid cron weekday in %q: %w", spec, err)
	}
	domStar := fields[2] == "*"
	dowStar := fields[4] == "*"

	candidate := from.Truncate(time.Minute).Add(time.Minute)
	for i := 0; i < maxCronLookahead; i++ {
		if minutes[candidate.Minute()] &&
			hours[candidate.Hour()] &&
			months[int(candidate.Month())] &&
			cronDayMatch(dom, dow, domStar, dowStar, candidate) {
			return candidate, nil
		}
		candidate = candidate.Add(time.Minute)
	}
	return time.Time{}, fmt.Errorf("Cron schedule %q has no occurrence within a year", spec)
}

// cronDayMatch applies the standard rule: when both day-of-month
// and day-of-week are restricted, either may match; when one is *,
// the other decides.
func cronDayMatch(dom, dow map[int]bool, domStar, dowStar bool, t time.Time) bool {
	domMatch := dom[t.Day()]
	dowMatch := dow[int(t.Weekday())] || dow[7] && t.Weekday() == time.Sunday
	switch {
	case domStar && dowStar:
		return true
	case domStar:
		return dowMatch
	case dowStar:
		return domMatch
	default:
		return domMatch || dowMatch
	}
}

var monthNames = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

var weekdayNames = map[string]int{
	"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
}

func parseCronField(field string, min, max int, names map[string]int) (map[int]bool, error) {
	out := make(map[int]bool)
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty list element")
		}
		step := 1
		base := part
		if slash := strings.Index(part, "/"); slash >= 0 {
			base = part[:slash]
			var err error
			step, err = strconv.Atoi(part[slash+1:])
			if err != nil || step <= 0 {
				return nil, fmt.Errorf("bad step in %q", part)
			}
		}
		lo, hi := min, max
		switch {
		case base == "*" || base == "":
			// Whole range with the step.
		default:
			bounds := strings.SplitN(base, "-", 2)
			start, err := parseCronValue(bounds[0], min, max, names)
			if err != nil {
				return nil, err
			}
			lo, hi = start, start
			if len(bounds) == 2 {
				end, err := parseCronValue(bounds[1], min, max, names)
				if err != nil {
					return nil, err
				}
				hi = end
			}
			if lo > hi {
				return nil, fmt.Errorf("reversed range in %q", part)
			}
		}
		for v := lo; v <= hi; v += step {
			out[v] = true
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("field %q matches nothing", field)
	}
	return out, nil
}

func parseCronValue(token string, min, max int, names map[string]int) (int, error) {
	token = strings.ToLower(strings.TrimSpace(token))
	if v, ok := names[token]; ok {
		return v, nil
	}
	v, err := strconv.Atoi(token)
	if err != nil || v < min || v > max {
		return 0, fmt.Errorf("%q is out of range %d-%d", token, min, max)
	}
	return v, nil
}

// NextDailyAt returns the next local occurrence of a wall-clock
// HH:MM time, for "every morning" style schedules.
func NextDailyAt(dailyAt string, from time.Time) (time.Time, error) {
	hour, minute, err := parseDailyAt(dailyAt)
	if err != nil {
		return time.Time{}, err
	}
	candidate := time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, from.Location())
	if !candidate.After(from) {
		candidate = candidate.Add(24 * time.Hour)
	}
	return candidate, nil
}

func parseDailyAt(dailyAt string) (int, int, error) {
	parts := strings.Split(strings.TrimSpace(dailyAt), ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("Invalid daily time %q: want HH:MM", dailyAt)
	}
	hour, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("Invalid daily time %q: want HH:MM", dailyAt)
	}
	minute, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("Invalid daily time %q: want HH:MM", dailyAt)
	}
	return hour, minute, nil
}

// DailyAtCron renders a daily-at time as a cron spec so daily
// schedules reuse the cron advancement path.
func DailyAtCron(dailyAt string) (string, error) {
	hour, minute, err := parseDailyAt(dailyAt)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d %d * * *", minute, hour), nil
}

func nextDaily(from time.Time, hour, minute int) time.Time {
	candidate := time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, from.Location())
	if !candidate.After(from) {
		candidate = candidate.Add(24 * time.Hour)
	}
	return candidate
}

func nextWeekly(from time.Time, day time.Weekday, hour, minute int) time.Time {
	candidate := nextDaily(from, hour, minute)
	for candidate.Weekday() != day {
		candidate = candidate.Add(24 * time.Hour)
	}
	return candidate
}
