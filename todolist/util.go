package todolist

import (
	"crypto/rand"
	"fmt"
	"io"
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

func mostRecentSunday(t time.Time) time.Time {
	for {
		if t.Weekday() != time.Sunday {
			t = t.AddDate(0, 0, -1)
		} else {
			return t
		}
	}
}

func mostRecentMonday(t time.Time) time.Time {
	for {
		if t.Weekday() != time.Monday {
			t = t.AddDate(0, 0, -1)
		} else {
			return t
		}
	}
}

func monday(day time.Time, forward bool) time.Time {
	dow := mostRecentMonday(day)
	if forward {
		return thisOrNextWeek(dow, day)
	}
	return thisOrLastWeek(dow, day)
}

func tuesday(day time.Time, forward bool) time.Time {
	dow := mostRecentMonday(day).AddDate(0, 0, 1)
	if forward {
		return thisOrNextWeek(dow, day)
	}
	return thisOrLastWeek(dow, day)
}

func wednesday(day time.Time, forward bool) time.Time {
	dow := mostRecentMonday(day).AddDate(0, 0, 2)
	if forward {
		return thisOrNextWeek(dow, day)
	}
	return thisOrLastWeek(dow, day)
}

func thursday(day time.Time, forward bool) time.Time {
	dow := mostRecentMonday(day).AddDate(0, 0, 3)
	if forward {
		return thisOrNextWeek(dow, day)
	}
	return thisOrLastWeek(dow, day)
}

func friday(day time.Time, forward bool) time.Time {
	dow := mostRecentMonday(day).AddDate(0, 0, 4)
	if forward {
		return thisOrNextWeek(dow, day)
	}
	return thisOrLastWeek(dow, day)
}

func saturday(day time.Time, forward bool) time.Time {
	dow := mostRecentMonday(day).AddDate(0, 0, 5)
	if forward {
		return thisOrNextWeek(dow, day)
	}
	return thisOrLastWeek(dow, day)
}

func sunday(day time.Time, forward bool) time.Time {
	dow := mostRecentMonday(day).AddDate(0, 0, 6)
	if forward {
		return thisOrNextWeek(dow, day)
	}
	return thisOrLastWeek(dow, day)
}

func thisOrNextWeek(day time.Time, pivotDay time.Time) time.Time {
	if day.Before(pivotDay) {
		return bod(day.AddDate(0, 0, 7))
	} else {
		return bod(day)
	}
}

func thisOrLastWeek(day time.Time, pivotDay time.Time) time.Time {
	if day.After(pivotDay) {
		return bod(day.AddDate(0, 0, -7))
	} else {
		return bod(day)
	}
}

func pluralize(count int, singular, plural string) string {
	if count > 1 {
		return plural
	}
	return singular
}

func isToday(t time.Time) bool {
	nowYear, nowMonth, nowDay := Now.Date()
	timeYear, timeMonth, timeDay := t.Date()
	return nowYear == timeYear &&
		nowMonth == timeMonth &&
		nowDay == timeDay
}

func isTomorrow(t time.Time) bool {
	nowYear, nowMonth, nowDay := Now.AddDate(0, 0, 1).Date()
	timeYear, timeMonth, timeDay := t.Date()
	return nowYear == timeYear &&
		nowMonth == timeMonth &&
		nowDay == timeDay
}

func isPastDue(t time.Time) bool {
	return Now.After(t)
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

func inSliceOneNotSliceTwo(s1, s2 []string) []string {
	// difference returns the elements in s1 that aren't in s2
	ms2 := map[string]bool{} //map of slice 2 elements
	for _, x := range s2 {
		ms2[x] = true
	}
	res := []string{} //result slice to contain s1 elements not in s2
	for _, x := range s1 {
		if _, ok := ms2[x]; !ok {
			res = append(res, x)
		}
	}
	return res
}

func getModifiedTime(todo *Todo) time.Time {
	if len(todo.ModifiedDate) > 0 {
		modTime, rerr := time.Parse(time.RFC3339, todo.ModifiedDate)
		if rerr != nil {
			createTime, _ := time.Parse(time.RFC3339, todo.CreatedDate)
			return createTime
		}
		return modTime
	}
	return Now
}

func stringToTime(val string) time.Time {
	if val != "" {
		parsedTime, _ := time.Parse(time.RFC3339, val)
		return parsedTime
	} else {
		parsedTime, _ := time.Parse(time.RFC3339, "1900-01-01T00:00:00+00:00")
		return parsedTime
	}
}

func timeToString(val time.Time) string {
	formatted := val.Format(time.RFC3339)
	return formatted
}

func timeToSimpleDateString(val time.Time) string {
	return val.Format("2006-01-02")
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
