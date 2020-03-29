package todolist

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ToDoFilter struct {
	Todos []*Todo
}

func NewToDoFilter(todos []*Todo) *ToDoFilter {
	return &ToDoFilter{Todos: todos}
}

func (f *ToDoFilter) Filter(filters []string) []*Todo {

	numTodos := len(f.Todos)
	//fmt.Println("filters before IDs: ", filters)
	f.Todos, filters = f.filterIDs(filters)
	//fmt.Println("filters after IDs: ", filters)
	//If matched specific id numbers, ignore the waiting filter. Presumably user intended to operate on the specific todo(s)
	if len(f.Todos) == numTodos {
		f.Todos, filters = NewDateFilter(f.Todos).FilterWaiting(filters)
	}
	//fmt.Println("filters after wait: ", filters)
	f.Todos, filters = f.filterArchived(filters) //includes filter for completed OR filter for archived
	//fmt.Println("filters after archive: ", filters)
	f.Todos, filters = NewDateFilter(f.Todos).FilterDoneDate(filters) //filter by completed date
	//fmt.Println("filters after done/completed: ", filters)
	f.Todos, filters = NewDateFilter(f.Todos).FilterModDate(filters) //filter by completed date
	//fmt.Println("filters after modified: ", filters)
	f.Todos, filters = NewDateFilter(f.Todos).FilterAge(filters) //filter by create date
	//fmt.Println("filters after age: ", filters)
	f.Todos, filters = NewDateFilter(f.Todos).FilterDueDate(filters) //filter by due date
	//fmt.Println("filters after due: ", filters)
	f.Todos, filters = f.filterPrioritized(filters) //Not useful until we redefine how priority works (number or ordinal values, rather than bool)
	//fmt.Println("filters after priority: ", filters)
	f.Todos, filters = f.filterProjects(filters)
	//fmt.Println("filters after project: ", filters)
	f.Todos, filters = f.filterContexts(filters)
	//fmt.Println("filters after Context: ", filters)
	f.Todos, filters = f.filterHasNotes(filters)
	//fmt.Println("filters after HasNotes: ", filters)
	f.Todos, filters = f.filterTopN(filters)
	//fmt.Println("filters after TopN: ", filters)
	f.Todos = f.filterSubject(filters)
	//fmt.Println("filters after Subject: ", filters)
	return f.Todos
}

func (f *ToDoFilter) filterHasNotes(filters []string) ([]*Todo, []string) {
	var ret []*Todo
	var filter string
	index := -1
	var exclude bool
	for i, part := range filters {
		exclude = false
		if strings.HasPrefix(part, "-") {
			exclude = true
			filter = part[1:]
		} else {
			filter = part
		}
		if strings.HasPrefix(filter, "notes:") {
			index = i
			for _, todo := range f.Todos {
				if exclude {
					if len(todo.Notes) == 0 {
						//fmt.Println("Adding todo to the list: ", todo.Id)
						ret = AddTodoIfNotThere(ret, todo)
					}
				} else {
					if len(todo.Notes) > 0 {
						//fmt.Println("Adding todo to the list: ", todo.Id)
						ret = AddTodoIfNotThere(ret, todo)
					}
				}
			}
			break
		}
	}
	if index > -1 {
		filters = append(filters[0:index], filters[index+1:]...)
	} else {
		ret = f.Todos
	}
	return ret, filters
}

func (f *ToDoFilter) filterIDs(filters []string) ([]*Todo, []string) {

	//filter by specific id or range of ids
	//if group 2, then call getIDs
	//else if group 1 only, then getID
	var ret []*Todo
	var ids []int
	var filter string
	index := -1
	var exclude bool

	re, _ := regexp.Compile("^(((\\d+)|(\\d+-\\d+)),*)+")
	for i, part := range filters {
		exclude = false
		if strings.HasPrefix(part, "-") {
			exclude = true
			filter = part[1:]
		} else {
			filter = part
		}
		if re.MatchString(filter) {
			index = i
			if matches := re.FindStringSubmatch(part); len(matches) > 0 {
				if len(matches) > 1 {
					//fmt.Println("Found a range or list of ids")
					ids = f.getIds(part)
				} else if len(matches) == 1 {
					//fmt.Println("Found a single id")
					ids = []int{f.getId(part)}
				}
			}
		}
	}

	if len(ids) == 0 {
		return f.Todos, filters
	}
	for _, todo := range f.Todos {
		for _, id := range ids {
			//fmt.Println("Copmare ID: ", id, " To Todo.Id: ", todo.Id)
			if exclude {
				if todo.Id != id {
					//fmt.Println("Adding todo to the list: ", todo.Id)
					ret = AddTodoIfNotThere(ret, todo)
				}
			} else {
				if todo.Id == id {
					//fmt.Println("Adding todo to the list: ", todo.Id)
					ret = AddTodoIfNotThere(ret, todo)
				}
			}
		}
	}
	if index > -1 {
		filters = append(filters[0:index], filters[index+1:]...)
	}
	return ret, filters

}

func (f *ToDoFilter) getId(input string) int {
	re, _ := regexp.Compile("\\d+")
	if re.MatchString(input) {
		id, _ := strconv.Atoi(re.FindString(input))
		return id
	}

	fmt.Println("Invalid id.")
	return -1
}

func (f *ToDoFilter) getIds(input string) (ids []int) {

	idGroups := strings.Split(input, ",")
	for _, idGroup := range idGroups {
		if rangedIds, err := f.parseRangedIds(idGroup); len(rangedIds) > 0 || err != nil {
			if err != nil {
				fmt.Printf("Invalid id group: %s.\n", input)
				continue
			}
			ids = append(ids, rangedIds...)
		} else if id := f.getId(idGroup); id != -1 {
			ids = append(ids, id)
		} else {
			fmt.Printf("Invalid id: %s.\n", idGroup)
		}
	}
	return ids
}

func (f *ToDoFilter) parseRangedIds(input string) (ids []int, err error) {
	rangeNumberRE, _ := regexp.Compile("(\\d+)-(\\d+)")
	if matches := rangeNumberRE.FindStringSubmatch(input); len(matches) > 0 {
		lowerID, _ := strconv.Atoi(matches[1])
		upperID, _ := strconv.Atoi(matches[2])
		if lowerID >= upperID {
			return ids, fmt.Errorf("Invalid id group: %s.\n", input)
		}
		for id := lowerID; id <= upperID; id++ {
			ids = append(ids, id)
		}
	}
	return ids, err
}

func (f *ToDoFilter) filterArchived(filters []string) ([]*Todo, []string) {
	exclude := false
	var filter string
	var todos []*Todo
	index := -1
	for i, part := range filters {
		exclude = false
		if strings.HasPrefix(part, "-") {
			exclude = true
			filter = part[1:]
		} else {
			filter = part
		}
		filter = strings.ToLower(filter)
		// do not filter archived if want completed items
		if filter == "completed" {
			index = i
			if exclude {
				todos = f.getIncomplete()
			} else {
				todos = f.getCompleted()
			}
		} else if filter == "archived" {
			index = i
			if exclude {
				todos = f.getUnarchived()
			} else {
				todos = f.getArchived()
			}
		} else if filter == "unarchived" {
			index = i
			if exclude {
				todos = f.getArchived()
			} else {
				todos = f.getUnarchived()
			}
		}
	}
	if index > -1 {
		filters = append(filters[0:index], filters[index+1:]...)
	} else {
		todos = f.Todos
	}
	return todos, filters
}

func (f *ToDoFilter) filterPrioritized(filters []string) ([]*Todo, []string) {
	exclude := false
	var filter string
	index := -1
	todos := []*Todo{}

	for i, part := range filters {
		if strings.HasPrefix(part, "-") {
			exclude = true
			filter = part[1:]
		} else {
			filter = part
		}
		filter := strings.ToLower(filter)
		if strings.HasPrefix(filter, "pri:") {
			index = i
			tmp := filter[4:] //e.g. pri:H,L or -pri:M
			p := strings.Split(tmp, ",")
			for _, pri := range p {
				todos = append(todos, f.getTodosByPriority(pri, exclude)...)
				//println("Length with p=", pri, " is ", len(todos))
			}
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

func (f *ToDoFilter) getTodosByPriority(p string, exclude bool) []*Todo {
	ret := []*Todo{}
	for _, todo := range f.Todos {
		pri := strings.ToLower(todo.Priority)
		if !exclude && pri == p {
			ret = append(ret, todo)
		} else if exclude && pri != p {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *ToDoFilter) filterTopN(filters []string) ([]*Todo, []string) {
	//Filter to top N tasks (ideally sort first, then filter to top N)
	//Top 1 per project -- "top:pro:1"
	//Top 2 per context -- "top:ctx:2"
	var filter string
	index := -1
	todos := []*Todo{}

	for i, part := range filters {

		filter = strings.ToLower(part)
		if strings.HasPrefix(filter, "top:pro:") {
			index = i
			max, _ := strconv.Atoi(strings.TrimPrefix(filter, "top:pro:"))
			//loop thru todos and keep only topN for each project
			pmap := map[string]int{}
			count := 0
			for _, todo := range f.Todos {
				for _, proj := range todo.Projects {
					count = pmap[proj] + 1
					pmap[proj] = count
					if count <= max {
						todos = append(todos, todo)
					}
				}
			}
		} else if strings.HasPrefix(filter, "top:ctx:") {
			index = i
			max, _ := strconv.Atoi(strings.TrimPrefix(filter, "top:ctx:"))
			//loop thru todos and keep only topN for each project
			pmap := map[string]int{}
			count := 0
			for _, todo := range f.Todos {
				for _, ctx := range todo.Contexts {
					count = pmap[ctx] + 1
					pmap[ctx] = count
					if count <= max {
						todos = append(todos, todo)
					}
				}
			}
		}
	}
	if index > -1 {
		filters = append(filters[0:index], filters[index+1:]...)
	} else {
		todos = f.Todos
	}
	return todos, filters
}

func (f *ToDoFilter) filterProjects(filters []string) ([]*Todo, []string) {

	srcTodoList := f.Todos
	notExcluded := f.Todos
	var included []*Todo

	var doInclude bool
	var ret []*Todo
	var project string
	indexes := []int{}
	for i, filter := range filters {

		if strings.Contains(filter, "+") {

			if strings.HasPrefix(filter, "-") {
				filter = filter[1:]

				srcTodoList = notExcluded
				notExcluded = []*Todo{}

				project = strings.ToLower(filter[1:])
				indexes = append(indexes, i)

				for _, todo := range srcTodoList {

					doInclude = true
					for _, todoProject := range todo.Projects {
						if project == strings.ToLower(todoProject) {
							doInclude = false
							break
						}
					}
					if doInclude {
						notExcluded = AddTodoIfNotThere(notExcluded, todo)

					}
				}

			} else {
				srcTodoList = f.Todos
				project = strings.ToLower(filter[1:])
				indexes = append(indexes, i)
				for _, todo := range srcTodoList {
					doInclude = false
					for _, todoProject := range todo.Projects {
						if project == strings.ToLower(todoProject) {
							doInclude = true
						}
					}
					if doInclude {
						included = AddTodoIfNotThere(included, todo)

					}
				}
			}
		}
	}

	if len(indexes) > 0 {
		for i, index := range indexes {
			index = index - i
			filters = append(filters[0:index], filters[index+1:]...)
		}
		ret = f.union(included, notExcluded)
	} else {
		ret = f.Todos
	}
	return ret, filters
}

func (f *ToDoFilter) filterContexts(filters []string) ([]*Todo, []string) {

	srcTodoList := f.Todos
	notExcluded := f.Todos
	var included []*Todo
	var doInclude bool
	var ret []*Todo
	var context string
	indexes := []int{}
	for i, filter := range filters {

		if strings.Contains(filter, "@") {

			if strings.HasPrefix(filter, "-") {
				filter = filter[1:]

				srcTodoList = notExcluded
				notExcluded = []*Todo{}

				context = strings.ToLower(filter[1:])
				indexes = append(indexes, i)
				for _, todo := range srcTodoList {
					doInclude = true
					for _, todoContext := range todo.Contexts {
						if context == strings.ToLower(todoContext) {
							doInclude = false
							break
						}
					}
					if doInclude {
						notExcluded = AddTodoIfNotThere(notExcluded, todo)

					}
				}

			} else {
				srcTodoList = f.Todos
				context = strings.ToLower(filter[1:])
				indexes = append(indexes, i)
				for _, todo := range srcTodoList {
					doInclude = false
					for _, todoContext := range todo.Contexts {
						if context == strings.ToLower(todoContext) {
							doInclude = true
						}
					}
					if doInclude {
						included = AddTodoIfNotThere(included, todo)

					}
				}
			}
		}
	}

	if len(indexes) > 0 {
		for i, index := range indexes {
			index = index - i
			filters = append(filters[0:index], filters[index+1:]...)
		}
		ret = f.union(included, notExcluded)
	} else {
		ret = f.Todos
	}
	return ret, filters
}

func (f *ToDoFilter) union(included []*Todo, notExcluded []*Todo) []*Todo {
	var ret []*Todo
	if len(included) > 0 && len(notExcluded) > 0 {
		sliceOne := included
		sliceTwo := notExcluded
		if len(included) > len(notExcluded) {
			sliceOne = notExcluded
			sliceTwo = included
		}
		for _, todo2 := range sliceTwo {
			for _, todo1 := range sliceOne {
				if todo2.Id == todo1.Id {
					ret = AddTodoIfNotThere(ret, todo2)
					break
				}
			}
		}
	} else if len(included) > 0 {
		ret = included
	} else {
		ret = notExcluded
	}
	return ret
}

func (f *ToDoFilter) filterSubject(filters []string) []*Todo {

	idMatcher, _ := regexp.Compile("(((\\d+)|(\\d+-\\d+)),*)+")
	subj := []string{}
	exclude := false
	var toFind string
	for i, part := range filters {
		//if !(strings.HasPrefix(part, "+") || strings.HasPrefix(part, "@") || strings.HasPrefix(part, "-") || strings.HasPrefix(part, "due:") || part == "archived" || part == "unarchived" || part == "p" || re.MatchString(part)) {
		if !(strings.HasPrefix(part, "+") || strings.HasPrefix(part, "@") || strings.HasPrefix(part, "due:") || part == "archived" || part == "unarchived" || part == "completed" || part == "p" || idMatcher.MatchString(part)) {
			subj = append(subj, filters[i])
		}
	}
	if len(subj) > 0 {
		toFind = strings.TrimSpace(strings.ToLower(strings.Join(subj, " ")))
	}

	//exclude subject if indicates an exclusion filter
	if strings.HasPrefix(toFind, "-") {
		exclude = true
		toFind = toFind[1:]
	}

	var todosubj string
	var ret []*Todo
	for _, todo := range f.Todos {
		todosubj = strings.TrimSpace(strings.ToLower(todo.Subject))
		if exclude {
			if !strings.Contains(todosubj, toFind) {
				ret = append(ret, todo)
			}
		} else {
			if strings.Contains(todosubj, toFind) {
				ret = append(ret, todo)
			}
		}
	}
	return ret
}

func (f *ToDoFilter) getArchived() []*Todo {
	var ret []*Todo
	for _, todo := range f.Todos {
		if todo.Status == "Archived" {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *ToDoFilter) getUnarchived() []*Todo {
	var ret []*Todo
	for _, todo := range f.Todos {
		if todo.Status != "Archived" {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *ToDoFilter) getCompleted() []*Todo {
	var ret []*Todo
	for _, todo := range f.Todos {
		if todo.Completed && todo.Status != "Archived" {
			ret = append(ret, todo)
		}
	}
	return ret
}

func (f *ToDoFilter) getIncomplete() []*Todo {
	var ret []*Todo
	for _, todo := range f.Todos {
		if !todo.Completed && todo.Status != "Archived" {
			ret = append(ret, todo)
		}
	}
	return ret
}
