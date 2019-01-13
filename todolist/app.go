package todolist

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	//"os/user"
	"github.com/skratchdot/open-golang/open"
)

type App struct {
	TodoStore  Store
	Cfg        *Config
	Printer    Printer
	TodoList   *TodoList
	CommandMap map[string]Command
}

func NewApp() *App {
	app := &App{
		TodoList:   &TodoList{},
		Printer:    NewScreenPrinter(),
		TodoStore:  NewFileStore(),
		CommandMap: map[string]Command{},
	}
	app.loadConfig()
	app.mapCommands()

	return app
}

func (a *App) loadConfig() {
	cfgStore := NewConfigStore()
	config, err := cfgStore.Load()
	if err != nil {
		fmt.Println("Error reading .todorc config file: ", err)
	}
	a.Cfg = config
	//Iterate over alias and create commands
	for key, val := range config.Aliases {
		a.AddAliasCommand(key, val)
	}
	//Iterate over reports and create commands
	for key, _ := range config.Reports {
		r, ok := a.Cfg.GetReport(key)
		if ok {
			a.AddReportCommand(key, r)
		}
	}
}

func (a *App) LoadPending() error {
	todos, err := a.TodoStore.LoadPending()
	if err != nil {
		return err
	}
	a.TodoList.Load(todos)
	return nil
}

func (a *App) LoadArchived() error {
	todos, err := a.TodoStore.LoadArchived()
	if err != nil {
		return err
	}
	a.TodoList.Load(todos)
	return nil
}

func (a *App) Save() {
	a.TodoStore.Save(a.TodoList.Data)
}

func (a *App) ProcessCmdLine(input string) Command {

	/*
		Parse the input into [filter] [command] [modifications] [miscellaneous args]
		- Split input on space
		- Find first instance of any of the command values
		- If command is first and does not accept modifications or args, treat anything to right as filter
		- If command is second or later, treat anything to left as filter
		- If command supports modifications (cannot also support args), treat anything to right of command as modification
		- If command supports args (cannot also support modifications), treat anything to the right as an arg
		- Add command should be treated as if it has "modifications" since subject, due, project, etc. are all handled same if add or modify.
	*/
	cmdMap := a.CommandMap

	parts := strings.Split(input, " ")
	filters := []string{}
	mods := []string{}
	args := []string{}
	var command Command

	for _, part := range parts {
		//try to find command
		if command == nil {
			command = cmdMap[part]
			if command == nil { //Filters to left of command
				filters = append(filters, part)
			}
		} else {
			if command.AcceptsMods() { //mods to right
				mods = append(mods, part)
			} else if command.AcceptsArgs() { //args to right
				args = append(args, part)
			} else { //filter to right
				filters = append(filters, part)
			}
		}
	}

	if command == nil {
		command = cmdMap["list"]
	}

	command.SetFilters(filters)
	command.SetMods(mods)
	command.SetArgs(args)
	/*
		fmt.Println("Command: ", command.GetCmd())
		fmt.Println("Filters: ", command.GetFilters())
		fmt.Println("Mods: ", command.GetMods())
		fmt.Println("Args: ", command.GetArgs())
	*/
	return command
}

func (a *App) AddAliasCommand(alias string, command string) {
	aliasCmd := NewCommand(command, true, false, a.ExecAlias)
	a.CommandMap[alias] = aliasCmd
}

func (a *App) AddReportCommand(key string, val *Report) *ReportCmd {
	reportCmd := &ReportCmd{
		CommandImpl{
			Cmd:          "list",
			IsAcceptMods: false,
			IsAcceptArgs: true,
		},
		nil,
	}
	reportCmd.SavedReport = val
	a.CommandMap[key] = reportCmd
	return reportCmd
}

//Functions that implement command logic

func (a *App) GarbageCollect(c *CommandImpl) {
	a.LoadPending()
	a.TodoList.GarbageCollect()
	a.Save()
	fmt.Println("Garbage collection complete.")
}

func (a *App) InitializeRepo(c *CommandImpl) {
	CreateDefaultConfig()
	a.TodoStore.Initialize()
}

func (a *App) SetView(c *CommandImpl) {
	cfgStore := NewConfigStore()
	if len(c.Mods) > 0 {
		err := cfgStore.SetConfigValue("view.current", c.Mods[0])
		if err != nil {
			fmt.Println("Error setting view: ", err)
		} else {
			fmt.Println("View set to: ", c.Mods[0])
		}
	}
}

func (a *App) AddTodo(c *CommandImpl) {
	a.LoadPending()
	parser := &Parser{}
	todo := parser.ParseNewTodo(c.Mods, a.TodoList)
	if todo == nil {
		fmt.Println("I need more information. Try something like 'todo a chat with bob @Bob due:tom'")
		return
	}

	id := a.TodoList.Add(todo)
	a.Save()
	fmt.Printf("Todo %d added.\n", id)
}

// AddDoneTodo Adds a todo and immediately completed it.
func (a *App) AddDoneTodo(c *CommandImpl) {
	a.LoadPending()
	parser := &Parser{}
	todo := parser.ParseNewTodo(c.Mods, a.TodoList)
	if todo == nil {
		fmt.Println("I need more information. Try something like 'todo done chating with bob'")
		return
	}

	id := a.TodoList.Add(todo)
	a.TodoList.Complete(todo)
	a.Save()
	fmt.Printf("Completed Todo %d added.\n", id)
}

func (a *App) DeleteTodo(c *CommandImpl) {
	a.LoadPending()
	filtered := NewToDoFilter(a.TodoList.Todos()).Filter(c.Filters)
	if len(filtered) == 0 {
		return
	}

	a.TodoList.Delete(filtered...)
	a.Save()
	fmt.Printf("%s deleted.\n", pluralize(len(filtered), "Todo", "Todos"))
}

func (a *App) CompleteTodo(c *CommandImpl) {
	a.LoadPending()
	filtered := NewToDoFilter(a.TodoList.Todos()).Filter(c.Filters)
	if len(filtered) == 0 {
		return
	}
	a.TodoList.Complete(filtered...)
	a.Save()
	fmt.Printf("%s completed.\n", pluralize(len(filtered), "Todo", "Todos"))
}

func (a *App) UncompleteTodo(c *CommandImpl) {
	a.LoadPending()
	filtered := NewToDoFilter(a.TodoList.Todos()).Filter(c.Filters)
	if len(filtered) == 0 {
		return
	}
	a.TodoList.Uncomplete(filtered...)
	a.Save()
	fmt.Printf("%s uncompleted.\n", pluralize(len(filtered), "Todo", "Todos"))
}

func (a *App) ArchiveTodo(c *CommandImpl) {
	a.LoadPending()
	filtered := NewToDoFilter(a.TodoList.Todos()).Filter(c.Filters)
	if len(filtered) == 0 {
		return
	}
	a.TodoList.Archive(filtered...)
	//load the archived todos from file so a.Save() call will save them all to the same file
	a.LoadArchived() //only do this when operating on archived
	a.Save()
	fmt.Printf("%s archived.\n", pluralize(len(filtered), "Todo", "Todos"))
}

func (a *App) UnarchiveTodo(c *CommandImpl) {
	a.LoadArchived() //only do this when operating on archived
	filtered := NewToDoFilter(a.TodoList.Todos()).Filter(c.Filters)
	if len(filtered) == 0 {
		println("UnarchiveTodo: filtered list is 0 for filter: ", c.Filters[0])
		return
	}
	a.TodoList.Unarchive(filtered...)
	a.LoadPending() //load in complete set of unarchived so that they get saved together with newly unarchived
	a.Save()
	fmt.Printf("%s unarchived.\n", pluralize(len(filtered), "Todo", "Todos"))
}

func (a *App) EditTodo(c *CommandImpl) {
	a.LoadPending()
	filtered := NewToDoFilter(a.TodoList.Todos()).Filter(c.Filters)
	isEdited := a.TodoList.Edit(c.Mods, filtered...)
	if isEdited {
		a.Save()
		fmt.Printf("%s edited.\n", pluralize(len(filtered), "Todo", "Todos"))
	}
}

func (a *App) AddNote(c *CommandImpl) {
	a.LoadPending()
	id, _ := strconv.Atoi(c.Filters[0])
	if id == -1 {
		return
	}
	todo := a.TodoList.FindById(id)
	if todo == nil {
		fmt.Println("No such id.")
		return
	}
	parser := &Parser{}

	if parser.ParseAddNote(todo, c.Mods) {
		todo.ModifiedDate = time.Now().Format(time.RFC3339)
		todo.IsModified = true
		fmt.Println("Note added.")
	}
	a.Save()
}

func (a *App) EditNote(c *CommandImpl) {
	a.LoadPending()
	id, _ := strconv.Atoi(c.Filters[0])
	if id == -1 {
		return
	}
	todo := a.TodoList.FindById(id)
	if todo == nil {
		fmt.Println("No such id.")
		return
	}
	parser := &Parser{}

	if parser.ParseEditNote(todo, c.Mods) {
		todo.ModifiedDate = time.Now().Format(time.RFC3339)
		todo.IsModified = true
		fmt.Println("Note edited.")
	}
	a.Save()
}

func (a *App) DeleteNote(c *CommandImpl) {
	a.LoadPending()
	id, _ := strconv.Atoi(c.Filters[0])
	if id == -1 {
		return
	}
	todo := a.TodoList.FindById(id)
	if todo == nil {
		fmt.Println("No such id.")
		return
	}
	parser := &Parser{}

	if parser.ParseDeleteNote(todo, c.Mods) {
		todo.ModifiedDate = time.Now().Format(time.RFC3339)
		todo.IsModified = true
		fmt.Println("Note deleted.")
	}
	a.Save()
}

func (a *App) ArchiveCompleted(c *CommandImpl) {
	a.LoadPending()
	filtered := NewToDoFilter(a.TodoList.Todos()).Filter([]string{"Completed"})
	a.TodoList.Archive(filtered...)
	//load the archived todos from file so a.Save() call will save them all to the same file
	a.LoadArchived() //only do this when operating on archived
	a.Save()
	fmt.Println("All completed todos have been archived.")
}

func (a *App) OrderTodos(c *CommandImpl) {
	a.LoadPending()
	//td ord all:3,4,5,6,7 OR td ord +BigProject:3,2,5,6 OR td ord @home:5,3,1
	if len(c.Mods) < 1 || !strings.Contains(c.Mods[0], ":") {
		println("Invalid input. Expected ord(er) <set>:<comma-separated ids>")
		return
	}
	tmp := strings.Split(c.Mods[0], ":")
	set := tmp[0]
	tmp2 := strings.Split(tmp[1], ",")
	ids := []int{}
	for _, val := range tmp2 {
		id, err := strconv.Atoi(val)
		if err != nil {
			println("Invalid input. Unable to parse id: ", val)
			return
		}
		ids = append(ids, id)
	}
	a.TodoList.UpdateOrdinals(set, ids)
	a.Save()
	println("Ordered Todos.")
}

func (a *App) ListProjects(c *CommandImpl) {
	a.LoadPending()
	todos := a.TodoList.Data
	m := map[string]int{}
	var set []string
	for _, todo := range todos {
		set = todo.Projects
		for _, name := range set {
			//if map[name] == nil {
			//	map[name] = 0
			//}
			m[name]++
		}
	}
	p := NewScreenPrinter()
	p.PrintSetCounts("Projects", m)
}

func (a *App) ListContexts(c *CommandImpl) {
	a.LoadPending()
	todos := a.TodoList.Data
	m := map[string]int{}
	var set []string
	for _, todo := range todos {
		set = todo.Contexts
		for _, name := range set {
			//if map[name] == nil {
			//	map[name] = 0
			//}
			m[name]++
		}
	}
	p := NewScreenPrinter()
	p.PrintSetCounts("Contexts", m)
}

func (a *App) Stats(c *CommandImpl) {
	/*
		pending, added, modified, completed, archived, deleted
		per all, project or context
		filtered by everything we allow to filter

		td <filters> stats [by:all|pro|ctx] [cols:p,a,m,c,ar,d] [sum:all|daily|weekly|monthly]
		output is tabular.
		If sum and by:pro/ctx then group stats and row per day/week/month
		If no sum (or sum:all) then no grouping and row per pro/ctx
	*/

	a.LoadPending()
	a.LoadArchived()
	cols := []string{"p", "a", "m", "c", "ar"}
	var groupBy string
	var sumBy string
	var chart bool
	for _, m := range c.Mods {
		if strings.HasPrefix(m, "cols:") {
			cols = strings.Split(m[5:], ",")
		} else if strings.HasPrefix(m, "by:") {
			groupBy = m[3:]
		} else if strings.HasPrefix(m, "sum:") {
			sumBy = m[4:]
		} else if strings.HasPrefix(m, "chart:") {
			chart = (m[6:] == "true")
		}
	}
	filtered := NewToDoFilter(a.TodoList.Todos()).Filter(c.Filters)
	if len(filtered) == 0 {
		return
	}
	p := NewScreenPrinter()
	p.PrintStats(filtered, groupBy, sumBy, cols, chart)
}

func (a *App) PrintTodoDetail(c *CommandImpl) {
	a.LoadPending()
	filtered := NewToDoFilter(a.TodoList.Todos()).Filter(c.Filters)
	if len(filtered) == 0 {
		return
	}
	p := NewScreenPrinter()
	p.PrintTodoDetail(filtered)
}

func (a *App) Sync(c *CommandImpl) {
	s := NewTodoSync(a.Cfg, a.TodoStore)
	verbose := false
	if len(c.Mods) > 0 {
		if c.Mods[0] == "verbose" {
			verbose = true
		}
	}
	err := s.Sync(verbose)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		os.Exit(1)
	}
}

func (a *App) NewWebApp(c *CommandImpl) {
	if err := a.LoadPending(); err != nil {
		os.Exit(1)
	} else {
		web := NewWebapp()
		fmt.Println("Now serving todolist web.\nHead to http://localhost:7890 to see your todo list!")
		open.Start("http://localhost:7890")
		web.Run()
	}
}

type OpenTask struct {
	Todo      *Todo
	NoteIndex int
	UriType   string
	Uri       string
	Cmd       string
}

func NewOpenTask(todo *Todo, idx int, uriType string, uri string, cmd string) *OpenTask {
	openTask := &OpenTask{
		Todo:      todo,
		NoteIndex: idx,
		UriType:   uriType,
		Uri:       uri,
		Cmd:       cmd,
	}
	return openTask
}

func (a *App) Open(c *CommandImpl) {
	if err := a.LoadPending(); err != nil {
		os.Exit(1)
	} else {
		todos := NewToDoFilter(a.TodoList.Data).Filter(c.Filters)
		if len(todos) == 0 {
			fmt.Println("No todos matching filter criteria.")
			return
		}
		openNotesRegex, err := regexp.Compile(a.Cfg.OpenNotesRegex)
		if err != nil {
			fmt.Println("Error compiling notes regex: " + a.Cfg.OpenNotesRegex)
			os.Exit(1)
		}
		openCustomRegex := map[string]*regexp.Regexp{}
		for k, v := range a.Cfg.OpenCustomRegex {
			re, err := regexp.Compile(v)
			if err != nil {
				fmt.Println("Error compiling regex. ", k, "=", v)
				os.Exit(1)
			}
			openCustomRegex[k] = re
		}
		openCustomCmd := map[string]string{}
		for k, v := range a.Cfg.OpenCustomCmd {
			openCustomCmd[k] = v
		}

		chosenNoteIndex := -1
		verbose := false
		if len(c.Args) > 0 {
			for _, arg := range c.Args {
				if arg == "verbose" {
					verbose = true
				} else {
					if idx, err := strconv.Atoi(arg); err == nil {
						chosenNoteIndex = idx
					}
				}
			}
		}

		openTasks := []*OpenTask{}

		var notes []string
		for _, todo := range todos {
			notes = todo.Notes
			for noteIndex, note := range notes {
				if chosenNoteIndex > -1 && noteIndex != chosenNoteIndex {
					continue
				}
				if verbose {
					println("Todo Uuid: " + todo.Uuid)
					println("Todo Note: '" + note + "'")
					println("Open file type: notes")
					println("Open regex: '" + a.Cfg.OpenNotesRegex + "'")
				}
				if openNotesRegex.MatchString(strings.TrimSpace(note)) {
					if verbose {
						if a.Cfg.OpenNotesCmd != "" {
							println("Matched notes regex. Opening notes with command '" + a.Cfg.OpenNotesCmd + "'")
						} else {
							println("Matched notes regex. Opening notes with default command")
						}
					}
					//a.openNotes(todo.Uuid, a.Cfg.OpenNotesCmd)
					openTasks = append(openTasks, NewOpenTask(todo, noteIndex, "note", "notes", a.Cfg.OpenNotesCmd))
				} else {
					/*
						NOTE - Browser URL was sort of a bitch to get right.
							   Would prefer to match things like github.com,
							   but requries www.github.com or http://github.com
							   to avoid over-matching.
						# Regular expression that identifies annotations openable by BROWSER.
						# URL must start with www|http
						open.browser.regex=((((https?://)?(www.))|(https?://))\S+)

						# Regular expression to identify files in annotations
						((\/|\.\/|~\/|\w:\/)\S+)
					*/

					//Preference browser match vs a file
					var task *OpenTask

					for k, v := range openCustomRegex {
						if verbose {
							println("Open regex not matched.")
							println("Open file type: " + k)
							println("Open regex: '" + a.Cfg.OpenCustomRegex[k] + "'")
						}
						if matches := v.FindStringSubmatch(note); len(matches) > 0 {
							if cmd, ok := openCustomCmd[k]; ok {
								if verbose {
									println("Opening file " + matches[1] + " with command '" + openCustomCmd[k] + "'")
								}
								//a.openUriWithCmd(matches[1], cmd)
								//openTasks = append(openTasks, NewOpenTask(todo, noteIndex, k, matches[1], cmd))
								task = NewOpenTask(todo, noteIndex, k, matches[1], cmd)
							} else {
								if verbose {
									println("Opening file: " + matches[1] + " with default command")
								}
								//a.openUri(matches[1])
								//openTasks = append(openTasks, NewOpenTask(todo, noteIndex, k, matches[1], ""))
								task = NewOpenTask(todo, noteIndex, k, matches[1], "")
							}
							//If matched browser type, quit so not overwritten by less specific file type matching
							if k == "browser" {
								break
							}
						} else {
							if verbose {
								println("Open regex not matched.")
							}
						}
					}
					if task != nil {
						openTasks = append(openTasks, task)
					}
				}
			}
		}

		//If more than one open task was found, let user choose
		var task *OpenTask
		if len(openTasks) > 1 {
			task = selectOpenTaskInput(openTasks)
			//Else just open it
		} else if len(openTasks) > 0 {
			task = openTasks[0]
		}
		//Open the note
		if task != nil {
			if task.Uri == "notes" {
				a.openNotes(task.Todo.Uuid, task.Cmd)
			} else {
				if task.Cmd == "" {
					a.openUri(task.Uri)
				} else {
					a.openUriWithCmd(task.Uri, task.Cmd)
				}
			}
		}
	}
}

func selectOpenTaskInput(openTasks []*OpenTask) *OpenTask {
	fmt.Println("Options: ")
	for i, task := range openTasks {
		fmt.Println("\t" + strconv.Itoa(i) + ") Todo " + strconv.Itoa(task.Todo.Id) + " (" + task.UriType + "): " + task.Uri)
	}
	fmt.Print("Select note to open: ")
	var i int
	_, err := fmt.Scanf("%d", &i)
	if err == nil {
		return openTasks[i]
	} else {
		return nil
	}

	//byteVal, _ := terminal.Read(int(syscall.Stdin))
	//if err == nil {
	//	fmt.Println("\nPassword typed: " + string(bytePassword))
	//}
	//val := string(byteVal)
	//fmt.Println("")
	//return strings.TrimSpace(password)
}

func (a *App) openUri(uri string) {
	err := open.Start(uri)
	if err != nil {
		println("Error opening uri " + uri + ": " + err.Error())
	}
}

func (a *App) openUriWithCmd(uri string, cmd string) {
	//err := open.StartWith(uri, cmd)
	err := open.RunWith(uri, cmd)
	if err != nil {
		println("Error opening uri " + uri + ": " + err.Error())
	}
}

func (a *App) openNotes(uuid string, cmd string) {
	notesDir := a.Cfg.OpenNotesFolder
	notes := fmt.Sprintf("%s/%s_notes"+a.Cfg.OpenNotesExt, notesDir, uuid)
	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		fmt.Println("Notes directory does not exist. Creating.")
		err2 := os.MkdirAll(notesDir, 0755)
		if err2 != nil {
			fmt.Println("Error creating notes directory: ", err2.Error())
		}
	}
	if _, err := os.Stat(notes); os.IsNotExist(err) {
		fmt.Println("File does not exist. Creating.")
		fd, err2 := os.OpenFile(notes, os.O_CREATE, 0644)
		if err2 != nil {
			fmt.Println("Error creating notes file: ", err2.Error())
		}
		fd.Close()
	}
	if cmd != "" {
		a.openUriWithCmd(notes, cmd)
	} else {
		a.openUri(notes)
	}
}

func (a *App) PrintHelp(c *CommandImpl) {
	p := NewScreenPrinter()
	if len(c.Args) == 0 {
		p.PrintOverallHelp()
	} else {
		for _, arg := range c.Args {
			switch arg {
			case "add", "a":
				p.PrintAddHelp()
			case "list", "l":
				p.PrintListHelp()
			case "edit", "e":
				p.PrintEditHelp()
			case "delete", "d":
				p.PrintDeleteHelp()
			case "done":
				p.PrintDoneHelp()
			case "complete", "c":
				p.PrintCompleteHelp()
			case "uncomplete", "uc":
				p.PrintUncompleteHelp()
			case "archive", "ar":
				p.PrintArchiveHelp()
			case "unarchive", "uar":
				p.PrintUnarchiveHelp()
			case "order", "ord", "reorder":
				p.PrintOrderTodosHelp()
			case "an":
				p.PrintAddNoteHelp()
			case "en":
				p.PrintEditNoteHelp()
			case "dn":
				p.PrintDeleteNoteHelp()
			case "gc":
				p.PrintGarbageCollectHelp()
			case "sync":
				p.PrintSyncHelp()
			case "ac":
				p.PrintArchiveCompletedHelp()
			case "projects":
				p.PrintProjectsHelp()
			case "contexts":
				p.PrintContextsHelp()
			case "print":
				p.PrintPrintTodoDetailHelp()
			case "view":
				p.PrintViewHelp()
			case "init":
				p.PrintInitHelp()
			case "config":
				p.PrintConfigHelp()
			case "open":
				p.PrintOpenHelp()
			case "stats":
				p.PrintStatsHelp()
			}
		}
	}
}

func (a *App) ExecAlias(c *CommandImpl) {
	//Parse the command line into command plus mods
	//TODO - Pull command line parsing into function in todolist package. Call from todo.go and from here to process
	//the alias with the full logic of parsing a command line. Then combine original contents from command line and
	//the result of parsing the alias command. Currently, assumes command is first "part" in cmdline. Also assumes
	//filters and args are not possible.

	//At time this is called, actual command line has already been parsed. Filters, Mods and Args will have been set.
	//Process the config value (c.Cmd) as a new command line
	//Then add the filters, mods and args from this AliasCmd back to it before calling Exec()

	origCmd := a.ProcessCmdLine(c.Cmd)

	//Parsing of Alias will assume command accepts mods. Now that we know the original command, if it doesn't accept mods,
	//those mods should be treated as filters.
	if !origCmd.AcceptsMods() {
		if c.Filters != nil {
			c.Filters = append(c.Filters, c.Mods...)
		} else {
			c.Filters = c.Mods
		}
		c.Mods = []string{}
	}

	if origCmd.GetMods() != nil {
		origCmd.SetMods(append(origCmd.GetMods(), c.Mods...))
	} else {
		origCmd.SetMods(c.Mods)
	}

	if origCmd.GetFilters() != nil {
		origCmd.SetFilters(append(origCmd.GetFilters(), c.Filters...))
	} else {
		origCmd.SetFilters(c.Filters)
	}

	if origCmd.GetArgs() != nil {
		origCmd.SetArgs(append(origCmd.GetArgs(), c.Args...))
	} else {
		origCmd.SetArgs(c.Args)
	}

	/*
		fmt.Println("Command: ", origCmd.GetCmd())
		fmt.Println("Filters: ", origCmd.GetFilters())
		fmt.Println("Mods: ", origCmd.GetMods())
		fmt.Println("Args: ", origCmd.GetArgs())
	*/

	//Execute the original command. Only one command executed at a time, so no worries about over-writing origCmd mods, filters, etc.
	origCmd.Exec(a)
}

type Command interface {
	Exec(a *App)
	GetCmd() string
	SetCmd(cmd string)
	GetFilters() []string
	SetFilters(filters []string)
	GetMods() []string
	SetMods(mods []string)
	GetArgs() []string
	SetArgs(args []string)
	AcceptsMods() bool
	SetAcceptsMods(acceptsMods bool)
	AcceptsArgs() bool
	SetAcceptsArgs(acceptsArgs bool)
}

func NewCommand(cmd string, iam bool, iaa bool, ef func(c *CommandImpl)) *CommandImpl {
	c := CommandImpl{
		Cmd:          cmd,
		IsAcceptMods: iam,
		IsAcceptArgs: iaa,
		ExecFunc:     ef,
	}
	return &c
}

//Define command

type CommandImpl struct {
	Cmd          string
	Filters      []string
	Mods         []string
	Args         []string
	IsAcceptMods bool
	IsAcceptArgs bool
	ExecFunc     func(c *CommandImpl)
}

func (c *CommandImpl) Exec(a *App) {
	c.ExecFunc(c)
}

func (c *CommandImpl) GetCmd() string {
	return c.Cmd
}

func (c *CommandImpl) SetCmd(cmd string) {
	c.Cmd = cmd
}

func (c *CommandImpl) GetFilters() []string {
	return c.Filters
}
func (c *CommandImpl) SetFilters(filters []string) {
	c.Filters = filters
}
func (c *CommandImpl) GetMods() []string {
	return c.Mods
}
func (c *CommandImpl) SetMods(mods []string) {
	c.Mods = mods
}
func (c *CommandImpl) GetArgs() []string {
	return c.Args
}
func (c *CommandImpl) SetArgs(args []string) {
	c.Args = args
}
func (c *CommandImpl) AcceptsMods() bool {
	return c.IsAcceptMods
}
func (c *CommandImpl) SetAcceptsMods(acceptsMods bool) {
	c.IsAcceptMods = acceptsMods
}
func (c *CommandImpl) AcceptsArgs() bool {
	return c.IsAcceptArgs
}
func (c *CommandImpl) SetAcceptsArgs(acceptsArgs bool) {
	c.IsAcceptArgs = acceptsArgs
}

//Special report command

type ReportCmd struct {
	CommandImpl
	SavedReport *Report
}

func (c *ReportCmd) Exec(a *App) {
	c.SavedReport.Filters = append(c.SavedReport.Filters, c.Filters...)
	filterArchived := false
	for _, f := range c.SavedReport.Filters {
		if f == "archived" {
			filterArchived = true
		}
	}
	if filterArchived {
		a.LoadArchived()
	} else {
		a.LoadPending()
		if a.TodoList.ExpireTodos() {
			a.LoadArchived()
			a.Save()
			a.TodoList.Data = []*Todo{}
			a.LoadPending()
		}
	}

	//Process args
	// notes:<bool> - Show \ Hide the notes
	// sort:<replace sorting> - Modify sorting
	// filter:<replace filters>
	// group:<replace group>
	groupBy := ""
	for _, arg := range c.Args {
		if strings.HasPrefix(arg, "notes:") {
			c.SavedReport.PrintNotes, _ = strconv.ParseBool(arg[6:])
		} else if strings.HasPrefix(arg, "sort:") {
			sorts := strings.Split(arg[5:], ",")
			c.SavedReport.Sorter = NewTodoSorter(sorts...)
		} else if strings.HasPrefix(arg, "filter:") {
			c.SavedReport.Filters = strings.Split(arg[7:], ",")
		} else if strings.HasPrefix(arg, "group:") {
			groupBy = strings.TrimSpace(arg[6:])
		}
	}

	//Ensure do grouping last, since it modifies the sort that may have been modified above
	if groupBy != "" {
		c.SavedReport.Group = groupBy
		if c.SavedReport.Group != "" {
			sorts := c.SavedReport.Sorter.SortColumns
			if !strings.Contains(sorts[0], c.SavedReport.Group) {
				sorts = append([]string{c.SavedReport.Group}, sorts...)
				c.SavedReport.Sorter = NewTodoSorter(sorts...)
			}
		}
	}
	//pass report and slice of todos to printer to print the columns and headers
	a.Printer.PrintReport(c.SavedReport, a.TodoList.Todos())
}

//Create command instances, map command text to required app function

func (a *App) mapCommands() {
	//Apply default report
	var report *Report
	var exists bool
	report, exists = a.Cfg.GetReport("default")
	if !exists {
		report = &Report{
			Description: "Default report of pending todos",
			Filters:     []string{},
			Columns:     []string{"id", "completed", "due", "context", "project", "subject"},
			Headers:     []string{"Id", "Status", "Due", "Context", "Project", "Subject"},
			Sorter:      NewTodoSorter("project", "due"),
		}
	}

	listCmd := a.AddReportCommand("list", report)
	a.CommandMap["l"] = listCmd
	a.CommandMap["list"] = listCmd

	addCmd := NewCommand("add", true, false, a.AddTodo)
	a.CommandMap["a"] = addCmd
	a.CommandMap["add"] = addCmd

	doneCmd := NewCommand("done", true, false, a.AddDoneTodo)
	a.CommandMap["done"] = doneCmd

	deleteCmd := NewCommand("delete", false, false, a.DeleteTodo)
	a.CommandMap["d"] = deleteCmd

	completeCmd := NewCommand("complete", false, false, a.CompleteTodo)
	a.CommandMap["c"] = completeCmd
	a.CommandMap["complete"] = completeCmd

	uncompleteCmd := NewCommand("uncomplete", false, false, a.UncompleteTodo)
	a.CommandMap["uc"] = uncompleteCmd
	a.CommandMap["uncomplete"] = uncompleteCmd

	archiveCmd := NewCommand("archive", false, false, a.ArchiveTodo)
	a.CommandMap["ar"] = archiveCmd
	a.CommandMap["archive"] = archiveCmd

	unarchiveCmd := NewCommand("unarchive", false, false, a.UnarchiveTodo)
	a.CommandMap["uar"] = unarchiveCmd
	a.CommandMap["unarchive"] = unarchiveCmd

	archiveCompletedCmd := NewCommand("ac", false, false, a.ArchiveCompleted)
	a.CommandMap["ac"] = archiveCompletedCmd

	editCmd := NewCommand("edit", true, false, a.EditTodo)
	a.CommandMap["e"] = editCmd
	a.CommandMap["edit"] = editCmd

	garbageCollectCmd := NewCommand("gc", false, false, a.GarbageCollect)
	a.CommandMap["gc"] = garbageCollectCmd

	initCmd := NewCommand("init", false, false, a.InitializeRepo)
	a.CommandMap["init"] = initCmd

	viewCmd := NewCommand("view", true, false, a.SetView)
	a.CommandMap["view"] = viewCmd

	webCmd := NewCommand("web", false, false, a.NewWebApp)
	a.CommandMap["web"] = webCmd

	addNoteCmd := NewCommand("an", true, false, a.AddNote)
	a.CommandMap["an"] = addNoteCmd
	a.CommandMap["n"] = addNoteCmd

	deleteNoteCmd := NewCommand("dn", true, false, a.DeleteNote)
	a.CommandMap["dn"] = deleteNoteCmd

	editNoteCmd := NewCommand("en", true, false, a.EditNote)
	a.CommandMap["en"] = editNoteCmd

	orderTodosCmd := NewCommand("ord", true, false, a.OrderTodos)
	a.CommandMap["ord"] = orderTodosCmd
	a.CommandMap["order"] = orderTodosCmd
	a.CommandMap["reorder"] = orderTodosCmd

	projectsCmd := NewCommand("projects", false, false, a.ListProjects)
	a.CommandMap["projects"] = projectsCmd

	contextsCmd := NewCommand("contexts", false, false, a.ListContexts)
	a.CommandMap["contexts"] = contextsCmd

	printCmd := NewCommand("print", false, false, a.PrintTodoDetail)
	a.CommandMap["print"] = printCmd

	syncCmd := NewCommand("sync", true, false, a.Sync)
	a.CommandMap["sync"] = syncCmd

	printHelpCmd := NewCommand("help", false, true, a.PrintHelp)
	a.CommandMap["help"] = printHelpCmd

	openCmd := NewCommand("help", false, true, a.Open)
	a.CommandMap["open"] = openCmd

	statsCmd := NewCommand("stats", true, false, a.Stats)
	a.CommandMap["stats"] = statsCmd
}
