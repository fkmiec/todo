package todolist

type Printer interface {
	//Print(*GroupedTodos, bool)
	PrintReport(*Report, []*Todo)
}
