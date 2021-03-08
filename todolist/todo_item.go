package todolist

import (
	"fmt"
	"os"
)

// Timestamp format to include date, time with timezone support. Easy to parse
//const ISO8601_TIMESTAMP_FORMAT = "2006-01-02T15:04:05Z07:00"

type Todo struct {
	Id            int            `json:"id"`
	Uuid          string         `json:"uuid"`
	Subject       string         `json:"subject"`
	Projects      []string       `json:"projects"`
	Contexts      []string       `json:"contexts"`
	Priority      string         `json:"priority"`
	Ordinals      map[string]int `json:"ordinals"`
	CreatedDate   string         `json:"createdDate"`
	ModifiedDate  string         `json:"modifiedDate"`
	IsModified    bool           `json:"-"`
	Wait          string         `json:"wait"`
	Until         string         `json:"until"`
	Due           string         `json:"due"`
	EffortDays    float64        `json:"effortDays"`
	Completed     bool           `json:"completed"`
	CompletedDate string         `json:"completedDate"`
	Status        string         `json:"status"`
	Notes         []string       `json:"notes"`
	ExecOrder     float64
}

func NewTodo() *Todo {
	uuid, err := newUUID()
	if err != nil {
		fmt.Println("Could not create UUID for new Todo: ", err)
		os.Exit(1)
	}
	ordMap := map[string]int{}
	//fmt.Println("Creating new Todo with UUID: ", uuid)
	return &Todo{Completed: false, Status: "Pending", Uuid: uuid, Ordinals: ordMap}
}

func (t Todo) Valid() bool {
	return (t.Subject != "")
}

func (t *Todo) Complete() {
	t.Completed = true
	t.CompletedDate = timeToString(Now)
}

func (t *Todo) Uncomplete() {
	t.Completed = false
	t.CompletedDate = ""
}

func (t *Todo) Archive() {
	t.Status = "Archived"
}

func (t *Todo) Unarchive() {
	t.Status = "Pending"
	t.Until = ""
}

func (t Todo) HasProject(proj string) bool {
	for _, p := range t.Projects {
		if proj == p {
			return true
		}
	}
	return false
}

func (t Todo) HasContext(ctx string) bool {
	for _, c := range t.Contexts {
		if ctx == c {
			return true
		}
	}
	return false
}
