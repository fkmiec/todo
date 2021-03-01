package todolist

type Store interface {
	Initialize()
	LoadPending() ([]*Todo, error)
	LoadArchived() ([]*Todo, error)
	LoadBacklog(filepath string) ([]*Todo, error)
	GetBacklogFilepath() string
	AppendBacklog(filepath string, todos []*Todo)
	DeleteBacklog(filepath string)
	Save(todos []*Todo)
	Import(filepath string) ([]*Todo, error)
	Export(filepath string, todos []*Todo)
}
