package todolist

import (
	"sort"
	"strings"
)

type TodoList struct {
	Data []*Todo
}

func (t *TodoList) Load(todos []*Todo) {
	if t.Data != nil {
		t.Data = append(t.Data, todos...)
	} else {
		t.Data = todos
	}
}

func (t *TodoList) Add(todo *Todo) int {
	todo.Id = t.NextId()
	todo.ModifiedDate = timeToString(Now)
	todo.CreatedDate = todo.ModifiedDate
	todo.IsModified = true
	t.Data = append(t.Data, todo)
	return todo.Id
}

func (t *TodoList) getSetSortedByOrdinal(set string) []*Todo {
	//Get all the todos in the same group into a separate slice
	todos := []*Todo{}
	if "all" == set {
		todos = append(todos, t.Data...)
	} else {
		orderByProject := false
		val := set[1:]
		if strings.HasPrefix(set, "+") {
			orderByProject = true
		}
		for _, todo := range t.Data {
			if orderByProject {
				if todo.HasProject(val) {
					todos = append(todos, todo)
				}
			} else {
				if todo.HasContext(val) {
					todos = append(todos, todo)
				}
			}
		}
	}
	//Sort them by existing ordinal values
	less1 := func(i, j int) bool {
		ordA := todos[i].Ordinals[set]
		ordB := todos[j].Ordinals[set]

		if ordA == -1 {
			todos[i].Ordinals[set] = i
			return false
		}
		if ordB == -1 {
			todos[i].Ordinals[set] = j
			return false
		}
		if ordA < ordB {
			return true
		} else {
			return false
		}
	}
	sort.Slice(todos, less1)
	return todos
}

func (t *TodoList) getIndexForId(todos []*Todo, id int) int {
	for p, todo := range todos {
		if todo.Id == id {
			return p
		}
	}
	return 0
}

func (t *TodoList) getMaxOrdinal(set string) int {
	maxOrd := -1
	tmpOrd := 0
	setType := 0 // 0 = all, 1 = project, 2 = context
	val := set[1:]
	if strings.HasPrefix(set, "+") {
		setType = 1
	} else if strings.HasPrefix(set, "@") {
		setType = 2
	}
	for _, todo := range t.Data {
		if setType == 0 {
			tmpOrd = todo.Ordinals[set]
			if tmpOrd > maxOrd {
				maxOrd = tmpOrd
			}
		} else if setType == 1 {
			if todo.HasProject(val) {
				tmpOrd = todo.Ordinals[set]
				if tmpOrd > maxOrd {
					maxOrd = tmpOrd
				}
			}
		} else {
			if todo.HasContext(val) {
				tmpOrd = todo.Ordinals[set]
				if tmpOrd > maxOrd {
					maxOrd = tmpOrd
				}
			}
		}
	}
	return maxOrd
}

func (t *TodoList) AddProject(p string, todo *Todo) {
	todo.Projects = append(todo.Projects, p)
	t.AddOrdinal("+"+p, todo)
}

func (t *TodoList) AddContext(c string, todo *Todo) {
	todo.Contexts = append(todo.Contexts, c)
	t.AddOrdinal("@"+c, todo)
}

func (t *TodoList) RemoveProject(p string, todo *Todo) {
	for i, project := range todo.Projects {
		if project == p {
			todo.Projects = append(todo.Projects[:i], todo.Projects[i+1:]...)
			t.RemoveOrdinal("+"+p, todo)
			break
		}
	}
}

func (t *TodoList) RemoveContext(c string, todo *Todo) {
	for i, context := range todo.Contexts {
		if context == c {
			todo.Contexts = append(todo.Contexts[:i], todo.Contexts[i+1:]...)
			t.RemoveOrdinal("@"+c, todo)
			break
		}
	}
}

func (t *TodoList) AddOrdinal(set string, todo *Todo) {
	//Set ordinal to last in the set
	todo.Ordinals[set] = (t.getMaxOrdinal(set) + 1)
}

func (t *TodoList) RemoveOrdinal(set string, todo *Todo) {
	//Remove set from the map of ordinals
	delete(todo.Ordinals, set)
}

func (t *TodoList) UpdateOrdinals(set string, ids []int) {
	//Get todos for the set ordered by current ordinal values
	todos := t.getSetSortedByOrdinal(set)

	//Figure out current ordinal from todos for the first id in the slice of ids
	//If user supplied ID 0 at the start of his list, insert ids at top of the list
	insertAt := 0
	if ids[0] < 1 {
		ids = ids[1:]
	} else {
		insertAt = t.getIndexForId(todos, ids[0])
	}

	idsContains := func(slice []int, id int) bool {
		for _, i := range slice {
			if id == i {
				return true
			}
		}
		return false
	}

	//Copy existing ids up to the current ordinal if not in the ids list
	//Insert the values from the ids array starting at current ordinal
	//Copy remaining ids in to the end (making sure to drop any that are found in the slice of ids
	res := []*Todo{}
	resIdx := 0
	for i, todo := range todos {
		if i == insertAt {
			for _, tmpId := range ids {
			inner:
				for _, tmpTodo := range todos {
					if tmpTodo.Id == tmpId {
						res = append(res, tmpTodo)
						resIdx++
						break inner
					}
				}
			}
			//Account for the todo at current index in todos when supplanting it with inserted from ids
			if idsContains(ids, todo.Id) {
				continue
			} else {
				res = append(res, todo)
				resIdx++
			}
		} else {
			if idsContains(ids, todo.Id) {
				continue
			} else {
				res = append(res, todo)
				resIdx++
			}
		}
	}

	//Assign new ordinal values
	for i, todo := range res {
		todo.Ordinals[set] = i
		todo.ModifiedDate = timeToString(Now)
		todo.IsModified = true
	}
}

func (t *TodoList) Edit(mods []string, todos ...*Todo) bool {
	parser := &Parser{}
	isEdited := false
	for _, todo := range todos {
		if parser.ParseEditTodo(todo, mods, t) {
			todo.ModifiedDate = timeToString(Now)
			todo.IsModified = true
			t.remove(todo)
			t.Data = append(t.Data, todo)
			isEdited = true
		}
	}
	return isEdited
}

func (t *TodoList) Delete(todos ...*Todo) {
	for _, td := range todos {
		for _, todo := range t.Data {
			if todo.Id == td.Id {
				todo.ModifiedDate = timeToString(Now)
				todo.IsModified = true
				todo.Status = "Deleted"
				t.remove(todo)
				t.Data = append(t.Data, todo)
			}
		}
	}
}

func (t *TodoList) remove(todos ...*Todo) {
	for _, td := range todos {
		i := -1
		for index, todo := range t.Data {
			if todo.Id == td.Id {
				i = index
				break
			}
		}
		t.Data = append(t.Data[:i], t.Data[i+1:]...)
	}
}

func (t *TodoList) Complete(todos ...*Todo) {
	for _, td := range todos {
		td.Complete()
		td.ModifiedDate = timeToString(Now)
		td.IsModified = true
		t.remove(td)
		t.Data = append(t.Data, td)
	}
}

func (t *TodoList) Uncomplete(todos ...*Todo) {
	for _, td := range todos {
		td.Uncomplete()
		td.ModifiedDate = timeToString(Now)
		td.IsModified = true
		t.remove(td)
		t.Data = append(t.Data, td)
	}
}

func (t *TodoList) Archive(todos ...*Todo) {
	for _, td := range todos {
		td.Archive()
		td.ModifiedDate = timeToString(Now)
		td.IsModified = true
		t.remove(td)
		t.Data = append(t.Data, td)
	}
}

func (t *TodoList) Unarchive(todos ...*Todo) {
	for _, td := range todos {
		td.Unarchive()
		td.ModifiedDate = timeToString(Now)
		td.IsModified = true
		t.remove(td)
		t.Data = append(t.Data, td)
	}
}

func (t *TodoList) IndexOf(todoToFind *Todo) int {
	for i, todo := range t.Data {
		if todo.Id == todoToFind.Id {
			return i
		}
	}
	return -1
}

type ByDate []*Todo

func (a ByDate) Len() int      { return len(a) }
func (a ByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool {
	t1Due := stringToTime(a[i].Due) //CalculateDueTime()
	t2Due := stringToTime(a[j].Due) //CalculateDueTime()
	return t1Due.Before(t2Due)
}

func (t *TodoList) Todos() []*Todo {
	sort.Sort(ByDate(t.Data))
	return t.Data
}

func (t *TodoList) MaxId() int {
	maxId := 0
	for _, todo := range t.Data {
		if todo.Id > maxId {
			maxId = todo.Id
		}
	}
	return maxId
}

func (t *TodoList) NextId() int {
	var found bool
	maxID := t.MaxId()
	for i := 1; i <= maxID; i++ {
		found = false
		for _, todo := range t.Data {
			if todo.Id == i {
				found = true
				break
			}
		}
		if !found {
			return i
		}
	}
	return maxID + 1
}

func (t *TodoList) FindById(id int) *Todo {
	for _, todo := range t.Data {
		if todo.Id == id {
			return todo
		}
	}
	return nil
}

type ByUuid []*Todo

func (a ByUuid) Len() int      { return len(a) }
func (a ByUuid) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByUuid) Less(i, j int) bool {
	return a[i].Uuid < a[j].Uuid
}

func (t *TodoList) ReassignAllIds() []*Todo {
	sort.Sort(ByUuid(t.Data))
	for i, todo := range t.Data {
		todo.Id = (i + 1)
	}
	return t.Data
}

func (t *TodoList) ExpireTodos() bool {
	dateFilter := NewDateFilter(t.Data)
	expired := dateFilter.FilterExpired()
	if len(expired) > 0 {
		t.Archive(expired...)
		return true
	}
	return false
}

func (t *TodoList) GarbageCollect() {
	var toDelete []*Todo
	for _, todo := range t.Data {
		if todo.Status == "Archived" {
			toDelete = append(toDelete, todo)
		}
	}
	for _, todo := range toDelete {
		t.Delete(todo)
	}
}
