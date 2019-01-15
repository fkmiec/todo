package todolist

import (
	"sort"
	"strings"
	"time"
)

type sortFunc func(p1, p2 *TodoStat) int

type StatSorter struct {
	stats []*TodoStat
	less  sortFunc
}

func DateSort(asc bool) sortFunc {
	d := func(t1, t2 *TodoStat) int {
		ret := 0
		if t1.PeriodStartDate.Before(t2.PeriodStartDate) {
			ret = -1
		} else if t1.PeriodStartDate.After(t2.PeriodStartDate) {
			ret = 1
		} else {
			ret = 0
		}
		if asc {
			return ret
		} else {
			return -1 * ret
		}
	}
	return d
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ss *StatSorter) Sort(stats []*TodoStat) {
	ss.stats = stats
	sort.Sort(ss)
}

// Len is part of sort.Interface.
func (ss *StatSorter) Len() int {
	return len(ss.stats)
}

// Swap is part of sort.Interface.
func (ss *StatSorter) Swap(i, j int) {
	ss.stats[i], ss.stats[j] = ss.stats[j], ss.stats[i]
}

func (ss *StatSorter) Less(i, j int) bool {
	p, q := ss.stats[i], ss.stats[j]
	// Try all but the last comparison.
	res := ss.less(p, q)
	switch res {
	case -1:
		// p < q, so we have a decision.
		return true
	case 1:
		// p > q, so we have a decision.
		return false
	}
	// case 0: //p == q; try the next comparison.
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return false
}

type TodoStat struct {
	PeriodStartDate time.Time
	Pending         int
	Unpending       int
	Added           int
	Modified        int
	Completed       int
	Archived        int
}

type StatsGroup struct {
	Group    string
	PrevStat *TodoStat //Use to track pending todos across dates
	Stats    []*TodoStat
}

type StatsData struct {
	Groups map[string]*StatsGroup
}

func (s *StatsData) GetSortedGroups() []*StatsGroup {
	sortedGroups := []*StatsGroup{}
	for k, _ := range s.Groups {
		sortedGroups = append(sortedGroups, s.SortedGroup(k))
	}
	return sortedGroups
}

func (s *StatsData) SortedGroup(group string) *StatsGroup {
	statsGroup := s.Groups[group]
	lessFn := DateSort(true)
	sorter := &StatSorter{}
	sorter.less = lessFn
	sorter.stats = statsGroup.Stats
	sort.Sort(sorter)
	return statsGroup
}

func (s *StatsData) CalcStats(todos []*Todo, groupBy string, sum int, rangeTimes []time.Time) {
	doGroups := (groupBy != "" && !strings.HasPrefix(strings.ToLower(groupBy), "a"))
	for _, todo := range todos {
		if doGroups {
			if strings.HasPrefix(strings.ToLower(groupBy), "p") {
				projects := strings.Join(todo.Projects, ",")
				s.CalcStatsForTodoAndGroup(todo, projects, sum)
			} else {
				contexts := strings.Join(todo.Contexts, ",")
				s.CalcStatsForTodoAndGroup(todo, contexts, sum)
			}
		} else {
			s.CalcStatsForTodoAndGroup(todo, "all", sum)
		}
	}

	groups := s.GetSortedGroups()
	for _, sg := range groups {
		for _, stat := range sg.Stats {
			if sg.PrevStat != nil {
				stat.Pending = sg.PrevStat.Pending
			}
			sg.PrevStat = stat
			stat.Pending += stat.Added
			stat.Pending -= stat.Unpending
		}
	}

	if len(rangeTimes) > 0 {
		startDate := rangeTimes[0]
		var endDate time.Time
		if len(rangeTimes) > 1 {
			endDate = rangeTimes[1]
		} else {
			endDate = Now
		}

		var rangeStats []*TodoStat
		for _, sg := range groups {
			for _, stat := range sg.Stats {
				if stat.PeriodStartDate.Before(startDate) {
					continue
				}
				if stat.PeriodStartDate.After(endDate) {
					continue
				}
				rangeStats = append(rangeStats, stat)
			}
			sg.Stats = rangeStats
		}
	}
}

func (s *StatsData) CalcStatsForTodoAndGroup(todo *Todo, group string, sumBy int) {
	addDate, _ := time.Parse(time.RFC3339, todo.CreatedDate)
	modDate, _ := time.Parse(time.RFC3339, todo.ModifiedDate)
	compDate, _ := time.Parse(time.RFC3339, todo.CompletedDate)
	sg, ok := s.Groups[group]
	if !ok {
		sg = &StatsGroup{Group: group, Stats: []*TodoStat{}}
		s.Groups[group] = sg
	}
	var stat *TodoStat
	var pending = true
	if sumBy == 1 { //weekly
		startDateFunc := bow
		stat = sg.getStatsForDate(startDateFunc(addDate))
		stat.Added++
		sg.getStatsForDate(startDateFunc(modDate)).Modified++
		if todo.Completed {
			stat = sg.getStatsForDate(startDateFunc(compDate))
			stat.Completed++
			if pending {
				stat.Unpending++
				pending = false
			}
		}
		if todo.Status == "Archived" {
			stat = sg.getStatsForDate(startDateFunc(modDate))
			stat.Archived++
			if pending {
				stat.Unpending++
				pending = false
			}
		}
	} else if sumBy == 2 { //monthly
		startDateFunc := bom
		stat = sg.getStatsForDate(startDateFunc(addDate))
		stat.Added++
		sg.getStatsForDate(startDateFunc(modDate)).Modified++
		if todo.Completed {
			stat = sg.getStatsForDate(startDateFunc(compDate))
			stat.Completed++
			if pending {
				stat.Unpending++
				pending = false
			}
		}
		if todo.Status == "Archived" {
			stat = sg.getStatsForDate(startDateFunc(modDate))
			stat.Archived++
			if pending {
				stat.Unpending++
				pending = false
			}
		}
	} else { //default to daily
		startDateFunc := bod
		stat = sg.getStatsForDate(startDateFunc(addDate))
		stat.Added++
		sg.getStatsForDate(startDateFunc(modDate)).Modified++
		if todo.Completed {
			stat = sg.getStatsForDate(startDateFunc(compDate))
			stat.Completed++
			if pending {
				stat.Unpending++
				pending = false
			}
		}
		if todo.Status == "Archived" {
			stat = sg.getStatsForDate(startDateFunc(modDate))
			stat.Archived++
			if pending {
				stat.Unpending++
				pending = false
			}
		}
	}
}

func (sg *StatsGroup) getStatsForDate(date time.Time) *TodoStat {
	var stat *TodoStat
	for _, stat = range sg.Stats {
		if stat.PeriodStartDate.Equal(date) {
			return stat
		}
	}
	stat = &TodoStat{PeriodStartDate: date}
	sg.Stats = append(sg.Stats, stat)
	return stat
}
