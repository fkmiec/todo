package todolist

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"
)

type ConfigStore struct {
	FileLocation string
	Loaded       bool
}

type Config struct {
	Aliases                  map[string]string
	Reports                  map[string]map[string]string
	Views                    map[string][]string
	CurrentView              string
	SyncFilepath             string
	SyncEncryptionPassphrase string
	OpenNotesFolder          string
	OpenNotesExt             string
	OpenNotesRegex           string
	OpenNotesCmd             string
	OpenCustomRegex          map[string]string
	OpenCustomCmd            map[string]string
}

//Declare Priority global because need access in filter and sorter
var (
	Priority map[string]int
)

func NewConfigStore() *ConfigStore {
	return &ConfigStore{FileLocation: ".todorc", Loaded: false}
}

func (f *ConfigStore) Load() (*Config, error) {
	f.FileLocation = getConfigLocation()
	usr, _ := user.Current()
	notesDir := fmt.Sprintf("%s/.todo_notes", usr.HomeDir)

	// init with some defaults
	config := Config{
		Aliases:                  map[string]string{"alias.report": "list"},
		Reports:                  map[string]map[string]string{},
		Views:                    map[string][]string{},
		CurrentView:              "",
		SyncFilepath:             "",
		SyncEncryptionPassphrase: "",
		OpenNotesFolder:          notesDir,
		OpenNotesExt:             ".txt",
		OpenNotesRegex:           "notes",
		OpenNotesCmd:             "",
		OpenCustomRegex:          map[string]string{},
		OpenCustomCmd:            map[string]string{},
	}
	//Default regex for web URLs
	config.OpenCustomRegex["browser"] = "((((https?://)?(www.))|(https?://))\\S+)"
	//Default regex for files
	slash := string(os.PathSeparator)
	config.OpenCustomRegex["file"] = "((\\/|\\.\\/|~\\/|\\w:\\" + slash + ").+)"

	// default values for priority
	Priority = map[string]int{"H": 1, "M": 2, "L": 3}

	if len(f.FileLocation) == 0 {
		return &config, nil
	}
	file, err := os.Open(f.FileLocation)
	if err != nil {
		return &config, nil
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if strings.HasPrefix(line, "#") {
			continue
		}
		// check if the line has = sign
		// and process the line. Ignore the rest.
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				// assign the config map
				if strings.HasPrefix(key, "alias") {
					keys := strings.Split(key, ".")
					if len(keys) > 1 {
						config.Aliases[keys[1]] = value
					}
				} else if strings.HasPrefix(key, "report") {
					keys := strings.Split(key, ".")
					if len(keys) > 2 {
						rep, ok := config.Reports[keys[1]]
						if !ok {
							rep = map[string]string{}
							config.Reports[keys[1]] = rep
						}
						rep[keys[2]] = value
					}
				} else if strings.HasPrefix(key, "view") {
					keys := strings.Split(key, ".")
					if len(keys) > 2 {
						if keys[2] == "filter" {
							config.Views[keys[1]] = strings.Split(value, " ")
						}
					} else {
						if keys[1] == "current" {
							config.CurrentView = value
						}
					}
				} else if strings.HasPrefix(key, "priority") {
					Priority = map[string]int{} //replace default values
					v := strings.Split(strings.TrimSpace(value), ",")
					for i, p := range v {
						Priority[p] = i
					}
				} else if strings.HasPrefix(key, "sync.filepath") {
					config.SyncFilepath = strings.TrimSpace(value)
				} else if strings.HasPrefix(key, "sync.encrypt.passphrase") {
					config.SyncEncryptionPassphrase = strings.TrimSpace(value)
				} else if strings.HasPrefix(key, "open") {
					keys := strings.Split(key, ".")
					if len(keys) < 3 {
						continue
					}
					if keys[1] == "notes" {
						switch keys[2] {
						case "ext":
							config.OpenNotesExt = strings.TrimSpace(value)
						case "folder":
							config.OpenNotesFolder = strings.TrimSpace(value)
						case "cmd":
							config.OpenNotesCmd = strings.TrimSpace(value)
						case "regex":
							config.OpenNotesRegex = strings.TrimSpace(value)
						}
					} else {
						switch keys[2] {
						case "regex":
							config.OpenCustomRegex[strings.TrimSpace(keys[1])] = strings.TrimSpace(value)
						case "cmd":
							config.OpenCustomCmd[strings.TrimSpace(keys[1])] = strings.TrimSpace(value)
						}
					}
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return &config, nil
		}
	}

	f.Loaded = true
	return &config, nil
}

func (f *ConfigStore) SetConfigValue(attr string, attrValue string) error {

	var key string
	var line string
	var err error
	var file *os.File

	f.FileLocation = getConfigLocation()

	if len(f.FileLocation) == 0 {
		return nil
	}
	file, err = os.Open(f.FileLocation)
	if err != nil {
		return nil
	}

	reader := bufio.NewReader(file)
	modified := false
	todorc := []string{}

	for {
		line, err = reader.ReadString('\n')
		if len(line) < 1 {
			break
		}

		//Only modify non-comment lines
		if !strings.HasPrefix(line, "#") {

			// check if the line has = sign
			// and process the line. Ignore the rest.
			if equal := strings.Index(line, "="); equal >= 0 {
				key = strings.TrimSpace(line[:equal])
			}
			if key == attr {
				modified = true
				line = attr + "=" + attrValue + "\n"
				todorc = append(todorc, line)
			} else {
				todorc = append(todorc, line)
			}
			//put comment back in resulting file
		} else {
			todorc = append(todorc, line)
		}
	}

	//If no modification of existing attribute, then append as a new attribute.
	if !modified {
		line = attr + "=" + attrValue + "\n"
		todorc = append(todorc, line)
	}

	file.Close()

	file, err = os.OpenFile(f.FileLocation, os.O_RDWR, os.FileMode(int(0777)))
	if err != nil {
		return nil
	}
	defer file.Close()

	cnt := 0
	writer := bufio.NewWriter(file)
	for _, line := range todorc {
		num, err := writer.WriteString(line)
		if err != nil {
			return nil
		}
		cnt += num
	}
	err = writer.Flush()
	if err != nil {
		return nil
	}
	return nil
}

func CreateDefaultConfig() error {
	repoLoc := ".todorc"
	file, err := os.Create(repoLoc)
	if err != nil {
		println("Error writing .todorc file")
		return nil
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString("## Notes on reports and commands. Type 'todolist help' for details and examles.\n")
	_, err = writer.WriteString("## Columns: 'id' 'completed' 'age' 'due' 'context' 'project' 'ord:all' 'ord:pro' 'ord:ctx'\n")
	_, err = writer.WriteString("## Headers: Labels for the columns.\n")
	_, err = writer.WriteString("## Sort: '+/-' plus 'id' 'age' 'due' 'context' 'project' 'ord:all' 'ord:pro' 'ord:ctx'\n")
	_, err = writer.WriteString("## Filter: Show results matching projects, contexts, due dates, etc.\n")
	_, err = writer.WriteString("###### Exclusion: Prefix the filter with '-' to include todos that do NOT match that filter.\n")
	_, err = writer.WriteString("## Group: Show results grouped by 'project' or 'context'\n")
	_, err = writer.WriteString("## Notes: Show notes if true.\n")
	_, err = writer.WriteString("\n")
	_, err = writer.WriteString("## Define a default report format used to print tasks to terminal\n")
	_, err = writer.WriteString("report.default.description='Default report of pending todos'\n")
	_, err = writer.WriteString("report.default.columns=id,completed,age,due,context,project,subject\n")
	_, err = writer.WriteString("report.default.headers=Id,Status,Age,Due,Context,Project,Subject\n")
	_, err = writer.WriteString("report.default.sort=+project,+due\n")
	_, err = writer.WriteString("report.default.filter=\n")
	_, err = writer.WriteString("report.default.group=project\n")
	_, err = writer.WriteString("#report.default.notes=true\n")
	_, err = writer.WriteString("\n")
	_, err = writer.WriteString("## Define custom priorities. Default is H,M,L.\n")
	_, err = writer.WriteString("#priority=H,M,L\n")
	_, err = writer.WriteString("\n")
	_, err = writer.WriteString("## Define sync file path and encryption passphrase.\n")
	_, err = writer.WriteString("###### encrypt.passphrase options: actual passphrase, *=prompt, <blank>=do not encrypt.\n")
	_, err = writer.WriteString("###### filepath includes filename. Directory must exist.\n")
	_, err = writer.WriteString("sync.encrypt.passphrase=*\n")
	_, err = writer.WriteString("sync.filepath=./backup/todo_sync.json\n")
	_, err = writer.WriteString("\n")
	_, err = writer.WriteString("## Define aliases to save typing on common commands\n")
	_, err = writer.WriteString("#alias.top2=top:pro:2 list sort:+project,+due\n")
	_, err = writer.WriteString("\n")
	_, err = writer.WriteString("## Define named view filters that can be applied by default\n")
	_, err = writer.WriteString("#view.work.filter=@Work\n")
	_, err = writer.WriteString("#view.home.filter=@Home\n")
	_, err = writer.WriteString("\n")
	_, err = writer.WriteString("## Set the currently applied view filter\n")
	_, err = writer.WriteString("#view.current=home\n")
	_, err = writer.WriteString("\n")
	_, err = writer.WriteString("## Configure the open command. Below are all defaults. Uncomment and change to override.\n")
	_, err = writer.WriteString("## Notes folder. If you sync, use a location available to other computers. E.g. a cloud drive\n")
	_, err = writer.WriteString("#open.notes.folder=~/.todo_notes\n")
	_, err = writer.WriteString("# Extension for notes\n")
	_, err = writer.WriteString("#open.notes.ext=.txt\n")
	_, err = writer.WriteString("# Command that opens notes\n")
	_, err = writer.WriteString("#open.notes.cmd=mousepad\n")
	_, err = writer.WriteString("# Regular expression if matched opens a notes file for a todo\n")
	_, err = writer.WriteString("#open.notes.regex=notes\n")
	_, err = writer.WriteString("## Define regex and (optionally) commands for open command to open differnt URI types.\n")
	_, err = writer.WriteString("# Web URLs (www|http)\n")
	_, err = writer.WriteString("#open.browser.regex=((((https?://)?(www.))|(https?://))\\S+)\n")
	_, err = writer.WriteString("#open.browser.cmd=netsurf\n")
	_, err = writer.WriteString("# File paths\n")
	_, err = writer.WriteString("#open.file.regex=((\\/|\\.\\/|~\\/|\\w:\\/\\w)\\S+)\n")

	err = writer.Flush()
	if err != nil {
		return nil
	} else {
		return err
	}
	return nil
}

func getConfigLocation() string {
	localrepo := ".todorc"
	usr, _ := user.Current()
	homerepo := fmt.Sprintf("%s/.todorc", usr.HomeDir)
	_, ferr := os.Stat(localrepo)

	if ferr == nil {
		return localrepo
	} else {
		return homerepo
	}
}

type Report struct {
	Description string
	Filters     []string
	Columns     []string
	Headers     []string
	Sorter      *TodoSorter
	Group       string
	PrintNotes  bool
}

func (r *Report) Init(rc map[string]string) {
	/*
		report.default.description="Default report of pending todos"
		report.default.columns=id,completed,due,context,project,subject
		report.default.headers=Id,Status,Due,Context,Project,Subject
		report.default.sort=+project,-due
		report.default.filter=
		report.default.notes=false

		report.byid.description='List of tasks ordered by ID'
		report.byid.columns=id,project,subject
		report.byid.headers=ID,Proj,Desc
		report.byid.sort=+id
		report.byid.filter=

		report.t2p.description='List of top 2 tasks for each project'
		report.t2p.columns=id,project,subject
		report.t2p.headers=ID,Proj,Desc
		report.t2p.sort=+project,-due
		report.t2p.filter=top:pro:2
	*/
	//Set the description
	r.Description = rc["description"]
	//Create the Filter slice
	r.Filters = strings.Split(rc["filter"], ",")
	//Get group by (none, project or context)
	group := strings.ToLower(rc["group"])
	if group == "project" || group == "context" {
		r.Group = group
	}
	//Create the Sorter
	sorts := strings.Split(rc["sort"], ",")
	//if a group was specified, add group as first sort if not already there.
	//Will then print in groups in ScreenPrinter as needed
	if r.Group != "" {
		if !strings.Contains(sorts[0], r.Group) {
			sorts = append([]string{r.Group}, sorts...)
		}
	}
	r.Sorter = NewTodoSorter(sorts...)
	//Create the Header slice
	r.Headers = strings.Split(rc["headers"], ",")
	//Create the Columns slice
	r.Columns = strings.Split(rc["columns"], ",")
	//Include/exclude notes
	if tmp, ok := rc["notes"]; ok {
		doPrint, err := strconv.ParseBool(tmp)
		if err != nil {
			fmt.Println("Error parsing bool from report configuration: ", rc["notes"])
			os.Exit(1)
		}
		r.PrintNotes = doPrint
	}
}

func (c *Config) GetAlias(alias string) (string, bool) {
	command, ok := c.Aliases[alias]
	return command, ok
}

func (c *Config) GetReport(report string) (*Report, bool) {
	//Get the report configuration from Config (map of values for the report name (from nested map under map of reports))
	rc, ok := c.Reports[report]
	if ok {
		rep := Report{}
		rep.Init(rc)
		return &rep, true
	} else {
		return nil, false
	}
}
