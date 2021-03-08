package todolist

import (
	"sort"
	"strings"
)

type lessFunc func(p1, p2 *Todo) int

// multiSorter implements the Sort interface, sorting the changes within.
type TodoSorter struct {
	todos       []*Todo
	less        []lessFunc
	SortColumns []string
}

func NewTodoSorter(sortCols ...string) *TodoSorter {
	sorter := &TodoSorter{}
	sorter.SortColumns = sortCols
	asc := true
	sorters := []lessFunc{}
	for _, col := range sortCols {
		col = strings.ToLower(col)
		if strings.HasPrefix(col, "-") {
			asc = false
			col = col[1:]
		} else if strings.HasPrefix(col, "+") {
			asc = true
			col = col[1:]
		} else {
			asc = true
		}
		switch col {
		case "project":
			sorters = append(sorters, Project(asc))
		case "context":
			sorters = append(sorters, Context(asc))
		case "due":
			sorters = append(sorters, Due(asc))
		case "priority":
			sorters = append(sorters, PrioritySorter(asc))
		case "id":
			sorters = append(sorters, Id(asc))
		case "notes":
			sorters = append(sorters, Notes(asc))
		case "age":
			sorters = append(sorters, Age(asc))
		case "idle":
			sorters = append(sorters, Modified(asc))
		case "effort":
			sorters = append(sorters, Effort(asc))
		case "exec":
			sorters = append(sorters, ExecOrder(asc))
		case "ord:all":
			sorters = append(sorters, OrdinalAll(asc))
		case "ord:pro":
			sorters = append(sorters, OrdinalProject(asc))
		case "ord:ctx":
			sorters = append(sorters, OrdinalContext(asc))
		case "created":
			sorters = append(sorters, Created(asc))
		case "modified":
			sorters = append(sorters, Modified(asc))
		case "subject":
			sorters = append(sorters, Subject(asc))
		}
	}
	sorter.less = sorters
	return sorter
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (s *TodoSorter) Sort(todos []*Todo) {
	calcAllExecOrder(todos)
	s.todos = todos
	sort.Sort(s)
}

func PrioritySorter(asc bool) lessFunc {
	priorityMap := Priority
	priority := func(t1, t2 *Todo) int {
		ret := 0
		var p1 int
		var p2 int
		var ok bool
		if p1, ok = priorityMap[t1.Priority]; !ok {
			p1 = 999999999 //sort unknown priority values to last
		}
		if p2, ok = priorityMap[t2.Priority]; !ok {
			p2 = 999999999 //sort unknown priority values to last
		}

		if p1 < p2 {
			ret = -1
		} else if p1 > p2 {
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
	return priority
}

func Id(asc bool) lessFunc {
	id := func(t1, t2 *Todo) int {
		ret := 0
		if t1.Id < t2.Id {
			ret = -1
		} else if t1.Id > t2.Id {
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
	return id
}

func Notes(asc bool) lessFunc {
	notes := func(t1, t2 *Todo) int {
		len1 := len(t1.Notes)
		len2 := len(t2.Notes)
		ret := 0
		if len1 < len2 {
			ret = -1
		} else if len1 > len2 {
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
	return notes
}

func OrdinalAll(asc bool) lessFunc {
	ord := func(t1, t2 *Todo) int {
		ord1 := t1.Ordinals["all"]
		ord2 := t2.Ordinals["all"]
		ret := 0
		if ord1 < ord2 {
			ret = -1
		} else if ord1 > ord2 {
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
	return ord
}

func OrdinalProject(asc bool) lessFunc {
	ord := func(t1, t2 *Todo) int {
		ord1 := -1
		if len(t1.Projects) > 0 {
			ord1 = t1.Ordinals["+"+t1.Projects[0]]
		}
		ord2 := -1
		if len(t2.Projects) > 0 {
			ord2 = t2.Ordinals["+"+t2.Projects[0]]
		}
		ret := 0
		if ord1 < ord2 {
			ret = -1
		} else if ord1 > ord2 {
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
	return ord
}

func OrdinalContext(asc bool) lessFunc {
	ord := func(t1, t2 *Todo) int {
		ord1 := -1
		if len(t1.Contexts) > 0 {
			ord1 = t1.Ordinals["@"+t1.Contexts[0]]
		}
		ord2 := -1
		if len(t2.Contexts) > 0 {
			ord2 = t2.Ordinals["@"+t2.Contexts[0]]
		}
		ret := 0
		if ord1 < ord2 {
			ret = -1
		} else if ord1 > ord2 {
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
	return ord
}

func Subject(asc bool) lessFunc {
	subject := func(t1, t2 *Todo) int {
		t1p := strings.ToLower(t1.Subject)
		t2p := strings.ToLower(t2.Subject)
		ret := 0
		if t1p < t2p {
			ret = -1
		} else if t1p > t2p {
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
	return subject
}

func Age(asc bool) lessFunc {
	age := func(t1, t2 *Todo) int {
		ret := 0
		if t1.CreatedDate < t2.CreatedDate {
			ret = -1
		} else if t1.CreatedDate > t2.CreatedDate {
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
	return age
}

func Effort(asc bool) lessFunc {
	order := func(t1, t2 *Todo) int {
		ret := 0
		if t1.EffortDays < t2.EffortDays {
			ret = -1
		} else if t1.EffortDays > t2.EffortDays {
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
	return order
}

//execution order = priority / (days til due / days of effort)
func ExecOrder(asc bool) lessFunc {
	order := func(t1, t2 *Todo) int {
		ret := 0
		if t1.ExecOrder < t2.ExecOrder {
			ret = -1
		} else if t1.ExecOrder > t2.ExecOrder {
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
	return order
}

func Due(asc bool) lessFunc {
	due := func(t1, t2 *Todo) int {
		ret := 0
		if t1.Due < t2.Due {
			ret = -1
		} else if t1.Due > t2.Due {
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
	return due
}

func Created(asc bool) lessFunc {
	d := func(t1, t2 *Todo) int {
		ret := 0
		if t1.CreatedDate < t2.CreatedDate {
			ret = -1
		} else if t1.CreatedDate > t2.CreatedDate {
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

func Modified(asc bool) lessFunc {
	d := func(t1, t2 *Todo) int {
		ret := 0
		if t1.ModifiedDate < t2.ModifiedDate {
			ret = -1
		} else if t1.ModifiedDate > t2.ModifiedDate {
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

func Project(asc bool) lessFunc {
	project := func(t1, t2 *Todo) int {
		t1p := ""
		t2p := ""
		ret := 0
		if len(t1.Projects) > 0 {
			t1p = strings.Join(t1.Projects, "")
		}
		if len(t2.Projects) > 0 {
			t2p = strings.Join(t2.Projects, "")
		}
		if t1p < t2p {
			ret = -1
		} else if t1p > t2p {
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
	return project
}

func Context(asc bool) lessFunc {
	context := func(t1, t2 *Todo) int {
		t1c := ""
		t2c := ""
		ret := 0
		if len(t1.Contexts) > 0 {
			t1c = strings.Join(t1.Contexts, "")
		}
		if len(t2.Contexts) > 0 {
			t2c = strings.Join(t2.Contexts, "")
		}
		if t1c < t2c {
			ret = -1
		} else if t1c > t2c {
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
	return context
}

// Len is part of sort.Interface.
func (s *TodoSorter) Len() int {
	return len(s.todos)
}

// Swap is part of sort.Interface.
func (s *TodoSorter) Swap(i, j int) {
	s.todos[i], s.todos[j] = s.todos[j], s.todos[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that is either Less or
// !Less. Note that it can call the less functions twice per call. We
// could change the functions to return -1, 0, 1 and reduce the
// number of calls for greater efficiency: an exercise for the reader.
func (s *TodoSorter) Less(i, j int) bool {
	p, q := s.todos[i], s.todos[j]
	// Try all but the last comparison.
	var k int
	res := 0
	for k = 0; k < len(s.less); k++ {
		less := s.less[k]
		res = less(p, q)
		switch res {
		case -1:
			// p < q, so we have a decision.
			return true
		case 1:
			// p > q, so we have a decision.
			return false
		}
		// case 0: //p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return false
}
