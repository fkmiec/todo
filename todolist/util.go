package todolist

import (
	"strings"
	"time"
)

func AddIfNotThere(arr []string, items []string) []string {
	for _, item := range items {
		there := false
		for _, arrItem := range arr {
			if item == arrItem {
				there = true
			}
		}
		if !there {
			arr = append(arr, item)
		}
	}
	return arr
}

func AddTodoIfNotThere(arr []*Todo, item *Todo) []*Todo {
	there := false
	for _, arrItem := range arr {
		if item.Id == arrItem.Id {
			there = true
		}
	}
	if !there {
		arr = append(arr, item)
	}
	return arr
}

func bod(t time.Time) time.Time {
	year, month, day := t.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func timestamp(t time.Time) time.Time {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	return time.Date(year, month, day, hour, min, sec, 0, t.Location())
}

func bom(t time.Time) time.Time {
	for {
		if t.Day() != 1 {
			t = t.AddDate(0, 0, -1)
		} else {
			return bod(t)
		}
	}
}

func bow(t time.Time) time.Time {
	for {
		if t.Weekday() != time.Sunday {
			t = t.AddDate(0, 0, -1)
		} else {
			return bod(t)
		}
	}
}

func getNearestMonday(t time.Time) time.Time {
	for {
		if t.Weekday() != time.Monday {
			t = t.AddDate(0, 0, -1)
		} else {
			return t
		}
	}
}

func pluralize(count int, singular, plural string) string {
	if count > 1 {
		return plural
	}
	return singular
}

func isToday(t time.Time) bool {
	nowYear, nowMonth, nowDay := time.Now().Date()
	timeYear, timeMonth, timeDay := t.Date()
	return nowYear == timeYear &&
		nowMonth == timeMonth &&
		nowDay == timeDay
}

func isTomorrow(t time.Time) bool {
	nowYear, nowMonth, nowDay := time.Now().AddDate(0, 0, 1).Date()
	timeYear, timeMonth, timeDay := t.Date()
	return nowYear == timeYear &&
		nowMonth == timeMonth &&
		nowDay == timeDay
}

func isPastDue(t time.Time) bool {
	return time.Now().After(t)
}

func translateToDates(t time.Time, vals ...string) []time.Time {
	times := []time.Time{}
	p := Parser{}
	for i, val := range vals {

		//Interpret blank values to support filter for due after and due before
		if val == "" {
			if i == 0 {
				//Treat blank begin date as an indefinite past date (-100 years)
				times = append(times, bod(t).AddDate(-100, 0, 0))
				continue
			} else if i == 1 {
				//Treat blank end date as an indefinite future date (+100 years)
				times = append(times, bod(t).AddDate(100, 0, 0))
				continue
			}
		}

		switch {
		case strings.HasPrefix(val, "this_week"):
			begin := bow(t)
			end := begin.AddDate(0, 0, 7)
			times = append(times, begin, end)
			break
		case strings.HasPrefix(val, "next_week"):
			begin := bow(t).AddDate(0, 0, 7)
			end := begin.AddDate(0, 0, 7)
			times = append(times, begin, end)
			break
		case strings.HasPrefix(val, "last_week"):
			begin := bow(t).AddDate(0, 0, -7)
			end := begin.AddDate(0, 0, 7)
			times = append(times, begin, end)
			break
		default:
			//If not blank or one of the range terms, parse for day of week or relative references
			t2 := p.ParseDateTime(val, t)
			times = append(times, t2)
		}

	}
	return times
}
