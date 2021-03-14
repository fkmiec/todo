package todolist

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"sync"
)

type FileStore struct {
	PendingFileLocation  string
	ArchivedFileLocation string
	BacklogFileLocation  string
	PendingLoaded        bool
	ArchivedLoaded       bool
}

func NewFileStore() *FileStore {
	return &FileStore{PendingFileLocation: "", ArchivedFileLocation: "", BacklogFileLocation: "", PendingLoaded: false, ArchivedLoaded: false}
}

func (f *FileStore) Initialize() {
	if f.PendingFileLocation == "" {
		f.PendingFileLocation = ".todos.json"
	}
	if f.ArchivedFileLocation == "" {
		f.ArchivedFileLocation = ".todos_archive.json"
	}
	if f.BacklogFileLocation == "" {
		f.BacklogFileLocation = ".todos_backlog.json"
	}

	filePendingCreated := false
	fileArchivedCreated := false
	fileBacklogCreated := false
	_, err := ioutil.ReadFile(f.PendingFileLocation)
	if err != nil {
		if err := ioutil.WriteFile(f.PendingFileLocation, []byte("[]"), 0644); err != nil {
			fmt.Println("Error writing json file", err)
			os.Exit(1)
		}
		filePendingCreated = true
	}
	_, err = ioutil.ReadFile(f.ArchivedFileLocation)
	if err != nil {
		if err := ioutil.WriteFile(f.ArchivedFileLocation, []byte("[]"), 0644); err != nil {
			fmt.Println("Error writing json file", err)
			os.Exit(1)
		}
		fileArchivedCreated = true
	}

	_, err = ioutil.ReadFile(f.BacklogFileLocation)
	if err != nil {
		if err := ioutil.WriteFile(f.BacklogFileLocation, []byte(""), 0644); err != nil {
			fmt.Println("Error writing json file", err)
			os.Exit(1)
		}
		fileBacklogCreated = true
	}

	if !filePendingCreated && !fileArchivedCreated && !fileBacklogCreated {
		fmt.Println("It looks like a .todos.json file already exists!  Doing nothing.")
		os.Exit(0)
	}
	fmt.Println("Todo repo initialized.")
}

func (f *FileStore) LoadPending() ([]*Todo, error) {
	if f.PendingFileLocation == "" {
		f.PendingFileLocation = getPendingLocation()
	}
	if f.ArchivedFileLocation == "" {
		f.ArchivedFileLocation = getArchivedLocation()
	}
	if f.BacklogFileLocation == "" {
		f.BacklogFileLocation = getBacklogLocation()
	}
	todos, _ := f.load(f.PendingFileLocation)
	f.PendingLoaded = true

	return todos, nil
}

func (f *FileStore) LoadArchived() ([]*Todo, error) {
	if f.ArchivedFileLocation == "" {
		f.ArchivedFileLocation = getArchivedLocation()
	}
	if f.PendingFileLocation == "" {
		f.PendingFileLocation = getPendingLocation()
	}
	if f.BacklogFileLocation == "" {
		f.BacklogFileLocation = getBacklogLocation()
	}
	todos, _ := f.load(f.ArchivedFileLocation)
	f.ArchivedLoaded = true
	return todos, nil
}

func (f *FileStore) Import(filepath string) ([]*Todo, error) {
	return f.load(filepath)
}

func (f *FileStore) Export(filepath string, todos []*Todo) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, _ := json.Marshal(todos)
		f.saveTodos(data, filepath)
	}()
	wg.Wait()
}

func (f *FileStore) load(filepath string) ([]*Todo, error) {

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("No todo file found!")
		fmt.Println("Initialize a new todo repo by running 'todolist init'")
		os.Exit(0)
		return nil, err
	}

	var todos []*Todo
	jerr := json.Unmarshal(data, &todos)
	if jerr != nil {
		fmt.Println("Error reading json data", jerr)
		os.Exit(1)
		return nil, jerr
	}
	return todos, nil
}

func (f *FileStore) Save(todos []*Todo) {
	//Separate archived and pending and save separately
	archivedTodos := []*Todo{}
	pendingTodos := []*Todo{}
	modifiedTodos := []*Todo{}

	for _, todo := range todos {
		if todo.Status == "Pending" {
			pendingTodos = append(pendingTodos, todo)
		} else if todo.Status == "Archived" {
			archivedTodos = append(archivedTodos, todo)
		}
		//If status == deleted, it will simply be dropped from pending and archived,
		//but will be written to backlog per the modified slice below.
		if todo.IsModified {
			modifiedTodos = append(modifiedTodos, todo)
		}
	}
	var wg sync.WaitGroup

	//Save archived
	if f.ArchivedLoaded {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, _ := json.Marshal(archivedTodos)
			f.saveTodos(data, f.ArchivedFileLocation)
		}()
	}
	//save pending
	if f.PendingLoaded {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, _ := json.Marshal(pendingTodos)
			f.saveTodos(data, f.PendingFileLocation)
		}()
	}

	//save modified to backlog
	if len(modifiedTodos) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.AppendBacklog(f.BacklogFileLocation, modifiedTodos)
		}()
	}

	wg.Wait()
}

func (f *FileStore) saveTodos(data []byte, filepath string) {
	if err := ioutil.WriteFile(filepath, []byte(data), 0644); err != nil {
		fmt.Println("Error writing json file: ", filepath, ". Error: ", err)
	}
}

func (f *FileStore) AppendBacklog(filepath string, todos []*Todo) {
	fd, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		println("Error opening file ", filepath, " : ", err.Error)
		panic(err)
	}
	defer fd.Close()

	for _, todo := range todos {
		data, _ := json.Marshal(todo)
		if _, err = fd.Write(data); err != nil {
			fmt.Println("Error appending to backlog json file: ", filepath, ". Error: ", err)
			panic(err)
		}
		if _, err = fd.WriteString("\n"); err != nil {
			fmt.Println("Error appending newline to backlog json file: ", filepath, ". Error: ", err)
			panic(err)
		}
	}
	fd.Sync()
}

func (f *FileStore) LoadBacklog(filepath string) ([]*Todo, error) {

	//Read in the backlog file line by line
	inputFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	var results []string
	for scanner.Scan() {
		results = append(results, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	//Unmarshal each line of backlog file into a Todo struct
	todos := []*Todo{}
	for _, tmp := range results {
		//println("JSON:'",tmp,"'")
		var todo *Todo
		jerr := json.Unmarshal([]byte(tmp), &todo)
		if jerr != nil {
			fmt.Println("Error reading json data", jerr)
			return nil, jerr
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func (f *FileStore) DeleteBacklog(filepath string) {
	var err = os.Remove(filepath)
	if err != nil {
		return
	}
}

func getPendingLocation() string {
	localrepo := ".todos.json"
	usr, _ := user.Current()
	homerepo := fmt.Sprintf("%s/.todos.json", usr.HomeDir)
	_, ferr := os.Stat(localrepo)

	if ferr == nil {
		return localrepo
	} else {
		return homerepo
	}
}

func getArchivedLocation() string {
	localrepo := ".todos_archive.json"
	usr, _ := user.Current()
	homerepo := fmt.Sprintf("%s/.todos_archive.json", usr.HomeDir)
	_, ferr := os.Stat(localrepo)

	if ferr == nil {
		return localrepo
	} else {
		return homerepo
	}
}

func (f *FileStore) GetBacklogFilepath() string {
	return f.BacklogFileLocation
}

func getBacklogLocation() string {
	localrepo := ".todos_backlog.json"
	usr, _ := user.Current()
	homerepo := fmt.Sprintf("%s/.todos_backlog.json", usr.HomeDir)
	_, ferr := os.Stat(localrepo)

	if ferr == nil {
		return localrepo
	} else {
		return homerepo
	}
}
