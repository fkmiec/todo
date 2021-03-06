package todolist

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type DateFilter struct {
	Todos  []*Todo
	Now    time.Time
	Parser Parser
}

func NewDateFilter(todos []*Todo) *DateFilter {
	now := Now
	return &DateFilter{Todos: todos, Now: now}
}

func filterOnDue(todo *Todo) time.Time {
	return stringToTime(todo.Due)
}

func filterOnCompletedDate(todo *Todo) time.Time {
	return stringToTime(todo.CompletedDate)
}

func filterOnModifiedDate(todo *Todo) time.Time {
	return stringToTime(todo.ModifiedDate)
}

func (f *DateFilter) FilterExpired() []*Todo {
	//dateInfo will be a simple date format (2018-12-01) or a day of week (sun-sat) or relative day reference (ie. today, tomorrow, yesterday)
	var ret []*Todo
	for _, todo := range f.Todos {
		if todo.Until != "" {
			todoTime := stringToTime(todo.Until) // time.ParseInLocation(time.RFC3339, todo.Due, f.Location)
			if todoTime.Before(f.Now) {
				ret = append(ret, todo)
			}
		}
	}
	return ret
}

func (f *DateFilter) FilterWaiting(filters []string) ([]*Todo, []string) {
	index := -1
	var todos []*Todo
	for i, filter := range filters {
		if strings.HasPrefix(filter, "wait") {
			index = i
			todos = f.FilterIncludeWaiting()
			break
		}
	}
	if index > -1 {
		filters = append(filters[0:index], filters[index+1:]...)
		return todos, filters
	}
	return f.FilterExcludeWaiting(), filters
}

func (f *DateFilter) FilterExcludeWaiting() []*Todo {
	var ret []*Todo
	for _, todo := range f.Todos {
		if !stringToTime(todo.Wait).After(f.Now) {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *DateFilter) FilterIncludeWaiting() []*Todo {
	var ret []*Todo
	for _, todo := range f.Todos {
		if stringToTime(todo.Wait).After(f.Now) {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *DateFilter) FilterAge(filters []string) ([]*Todo, []string) {
	//e.g. age:gt:7d or age:lt:7d or age:eq:3d
	index := -1
	var todos []*Todo
	min := -1
	max := -1
	re, _ := regexp.Compile("age:(\\d+)(-(\\d+))?\\w*")
	for i, filter := range filters {
		if re.MatchString(filter) {
			index = i
			matches := re.FindStringSubmatch(filter)
			min, _ = strconv.Atoi(matches[1])
			if matches[3] != "" {
				max, _ = strconv.Atoi(matches[3])
			}
			break
		}
	}
	//If no max specified, match min age exactly
	if max == -1 {
		max = min
		//If max specified with another value less than min (e.g. 1-0)
		//return all todos older than min
	} else if max < min {
		max = 999999999
	}
	for _, todo := range f.Todos {
		days := f.ageInDays(todo)
		if days >= min && days <= max {
			todos = append(todos, todo)
		}
	}

	if index > -1 {
		filters = append(filters[0:index], filters[index+1:]...)
		return todos, filters
	}
	return f.Todos, filters
}

func (f *DateFilter) ageInDays(todo *Todo) int {
	days := 0
	if len(todo.CreatedDate) > 0 {
		tmpTime, err := time.Parse(time.RFC3339, todo.CreatedDate)
		if err == nil {
			createTime := tmpTime.Unix()
			now := Now.Unix()
			diff := now - createTime
			days = (int)(diff / (60 * 60 * 24))
		}
	}
	return days
}

func (f *DateFilter) FilterDueDate(filters []string) ([]*Todo, []string) {
	r, _ := regexp.Compile(`due:([^:]+)?(:(.*))?`)
	return f.FilterDateRange(filters, r, filterOnDue)
}

func (f *DateFilter) FilterDoneDate(filters []string) ([]*Todo, []string) {
	r, _ := regexp.Compile(`done:([^:]+)?(:(.*))?`)
	return f.FilterDateRange(filters, r, filterOnCompletedDate)
}

func (f *DateFilter) FilterModDate(filters []string) ([]*Todo, []string) {
	r, _ := regexp.Compile(`mod:([^:]+)?(:(.*))?`)
	return f.FilterDateRange(filters, r, filterOnModifiedDate)
}

func (f *DateFilter) FilterDateRange(filters []string, regex *regexp.Regexp, dateConvFunc func(*Todo) time.Time) ([]*Todo, []string) {

	var todos []*Todo
	index := -1

loop:
	for i, filter := range filters {
		// filter due items
		var d1 string
		var d2 string
		var times []time.Time
		//r, _ := regexp.Compile(`due:([^:]+)?(:(.*))?`)
		matches := regex.FindStringSubmatch(filter)
		if len(matches) > 0 {
			index = i
			d1 = strings.ToLower(matches[1])
			//Handle special values not mapping to dates
			switch {
			case strings.HasPrefix(d1, "any"):
				todos = f.filterAnyDueDate()
				break loop
			case strings.HasPrefix(d1, "non"):
				todos = f.filterNoDueDate()
				break loop
			case strings.HasPrefix(d1, "overdue"):
				todos = f.filterOverdue(bod(f.Now))
				break loop
			}

			//Handle if there is a date range
			if strings.HasPrefix(matches[2], ":") {
				d2 = strings.ToLower(matches[3])
				times = translateToDates(f.Now, d1, d2)
			} else {
				times = translateToDates(f.Now, d1)
			}
			len := len(times)
			switch len {
			case 1: //single date to match
				todos = f.filterToExactDate(times[0], dateConvFunc)
			case 2: //date range to match
				todos = f.filterBetweenDatesInclusive(times[0], times[1], dateConvFunc)
			case 0: //this should not happen, but if no times provided, ignore due filter and return all todos
				//println("Received due: ", d1, " and time parsing returned nothing. Applying no due filter.")
				todos = f.Todos
			}
		}
		if index > -1 {
			break
		}
	}

	if index > -1 {
		filters = append(filters[0:index], filters[index+1:]...)
	} else {
		todos = f.Todos
	}
	return todos, filters
}

func (f *DateFilter) filterAnyDueDate() []*Todo {
	var ret []*Todo
	for _, todo := range f.Todos {
		if len(todo.Due) > 6 {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *DateFilter) filterNoDueDate() []*Todo {
	var ret []*Todo
	for _, todo := range f.Todos {
		if len(todo.Due) < 6 {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *DateFilter) equalSimpleDates(t1 time.Time, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay()
}

func (f *DateFilter) filterToExactDate(pivot time.Time, filterOn func(*Todo) time.Time) []*Todo {
	var ret []*Todo
	var todoTime time.Time
	for _, todo := range f.Todos {
		todoTime = filterOn(todo)
		if f.equalSimpleDates(todoTime, pivot) {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *DateFilter) filterBetweenDatesInclusive(begin, end time.Time, filterOn func(*Todo) time.Time) []*Todo {
	var ret []*Todo

	for _, todo := range f.Todos {
		todoTime := filterOn(todo)
		if (begin.Before(todoTime) || begin.Equal(todoTime)) && (end.After(todoTime) || end.Equal(todoTime)) {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *DateFilter) filterCompletedToday(pivot time.Time) []*Todo {
	return f.filterToExactDate(pivot, filterOnCompletedDate)
}

func (f *DateFilter) filterCompletedThisWeek(pivot time.Time) []*Todo {

	begin := bod(mostRecentSunday(pivot))
	end := begin.AddDate(0, 0, 7)

	return f.filterBetweenDatesInclusive(begin, end, filterOnCompletedDate)
}

func (f *DateFilter) filterOverdue(pivot time.Time) []*Todo {
	var ret []*Todo

	for _, todo := range f.Todos {
		if todo.Due == "" {
			continue
		}
		todoTime := stringToTime(todo.Due) //time.ParseInLocation(time.RFC3339, todo.Due, f.Location)
		if todoTime.Before(pivot) {        //&& pivot.pivotDate != todo.Due {
			ret = append(ret, todo)
		}
	}
	return ret
}
