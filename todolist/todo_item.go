package todolist

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"time"
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
	Completed     bool           `json:"completed"`
	CompletedDate string         `json:"completedDate"`
	Status        string         `json:"status"`
	Notes         []string       `json:"notes"`
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
	t.CompletedDate = timestamp(time.Now()).Format(time.RFC3339)
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

func (t Todo) ToTime(val string) time.Time {
	if val != "" {
		parsedTime, _ := time.Parse(time.RFC3339, val)
		return parsedTime
	} else {
		parsedTime, _ := time.Parse(time.RFC3339, "1900-01-01T00:00:00+00:00")
		return parsedTime
	}
}

func (t Todo) ToSimpleDate(val time.Time) string {
	formatted := val.Format("2006-01-02")
	return formatted
}

func (t Todo) ToDateTimeString(val time.Time) string {
	formatted := val.Format(time.RFC3339)
	return formatted
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
