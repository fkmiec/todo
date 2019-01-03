package todolist

import (
	"fmt"
	"os"

	//"regexp"

	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/buger/goterm"
	"github.com/fatih/color"
)

type ScreenPrinter struct {
	Writer    *tabwriter.Writer
	fgGreen   func(a ...interface{}) string
	fgYellow  func(a ...interface{}) string
	fgRed     func(a ...interface{}) string
	fgWhite   func(a ...interface{}) string
	fgBlue    func(a ...interface{}) string
	fgMagenta func(a ...interface{}) string
	fgCyan    func(a ...interface{}) string
}

func NewScreenPrinter() *ScreenPrinter {
	w := new(tabwriter.Writer)
	w.Init(color.Output, 0, 8, 1, ' ', 0) //Changed from os.Stdout to color.Output when compiled on Windows.
	//w.Init(color.Output, 5, 0, 1, ' ', tabwriter.StripEscape)
	green := color.New(color.FgGreen).Add(color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).Add(color.Bold).SprintFunc()
	red := color.New(color.FgRed).Add(color.Bold).SprintFunc()
	white := color.New(color.FgWhite).Add(color.Bold).SprintFunc()
	blue := color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	magenta := color.New(color.FgMagenta).Add(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).Add(color.Bold).SprintFunc()
	formatter := &ScreenPrinter{Writer: w, fgGreen: green, fgYellow: yellow, fgRed: red, fgWhite: white, fgBlue: blue, fgMagenta: magenta, fgCyan: cyan,}
	return formatter
}

/*
func (f *ScreenPrinter) Print(groupedTodos *GroupedTodos, printNotes bool) {
	cyan := color.New(color.FgCyan).SprintFunc()
	var keys []string
	for key := range groupedTodos.Groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(f.Writer, "\n %s\n", cyan(key))
		for _, todo := range groupedTodos.Groups[key] {
			f.printTodo(todo)
			if printNotes {
				for nid, note := range todo.Notes {
					fmt.Fprintf(f.Writer, "   %s\t%s\t\n",
						cyan(strconv.Itoa(nid)), note)
				}
			}
		}
	}
	f.Writer.Flush()
}
*/
func (f *ScreenPrinter) printTodo(todo *Todo) {

	fmt.Fprintf(f.Writer, " %s\t%s\t%s\t%s\t%s\t%s\t\n",
		f.fgYellow(strconv.Itoa(todo.Id)),
		f.formatCompleted(todo.Completed),
		f.formatDue(todo.Due),
		f.formatContexts(todo.Contexts),
		f.formatProjects(todo.Projects),
		f.formatSubject(todo.Subject))
}

func (f *ScreenPrinter) PrintSetCounts(set string, m map[string]int) {
	setColor := f.fgBlue
	valColor := f.fgYellow
	if set == "Projects" {
		setColor = f.fgMagenta
	} else if set == "Contexts" {
		setColor = f.fgRed
	}
	set = f.fgGreen(set)
	fmt.Fprintf(f.Writer, "%s:\n", set)
	for k, v := range m {
		k = setColor(k)
		val := valColor(v)
		fmt.Fprintf(f.Writer, " %s\t%s\n", k, val)
	}
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintTodoDetail(todos []*Todo) {
	key := f.fgGreen
	val := f.fgYellow

	for i, todo := range todos {
		if i > 0 {
			fmt.Println("")
		}
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("ID:"), val(todo.Id))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("UUID:"), val(todo.Uuid))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Subject:"), val(todo.Subject))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Contexts:"), val(strings.Join(todo.Contexts, ",")))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Projects:"), val(strings.Join(todo.Projects, ",")))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Due:"), val(todo.Due))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Priority:"), val(todo.Priority))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Ordinals:"), val(fmt.Sprint(todo.Ordinals)[3:]))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("IsModified:"), val(todo.IsModified))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Completed:"), val(todo.Completed))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Status:"), val(todo.Status))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("CreatedDate:"), val(todo.CreatedDate))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("ModifiedDate:"), val(todo.ModifiedDate))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("CompletedDate:"), val(todo.CompletedDate))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Until:"), val(todo.Until))
		fmt.Fprintf(f.Writer, " %s\t%s\n", key("Wait:"), val(todo.Wait))
		notes := todo.Notes
		if len(notes) > 0 {
			//fmt.Fprintf(f.Writer, " %s\t%s\n", key("Notes:"), val(""))
			f.printNotes(notes)
		} else {
			fmt.Fprintf(f.Writer, " %s\t%s\n", key("Notes:"), val(""))
		}

		f.Writer.Flush()
	}
}

func (f *ScreenPrinter) printNotes(notes []string) {
	fmt.Fprintf(f.Writer, " %s\t%s\n", "Notes:", "")
	for _, note := range notes {
		fmt.Fprintf(f.Writer, " %s\t%s\n", "", note)
	}
}

func (f *ScreenPrinter) PrintReport(report *Report, todos []*Todo) {

	report.Sorter.Sort(todos)
	filtered := NewToDoFilter(todos).Filter(report.Filters)
	if len(filtered) == 0 {
		fmt.Println("No todos matching filter criteria.")
		return
	}
	consoleHeight := goterm.Height()
	//consoleHeight := 5
	rowNum := 0
	doGroups := report.Group != ""
	lastGroup := "none"
	for i, todo := range filtered {
		if doGroups {
			if report.Group == "project" {
				projects := strings.Join(todo.Projects, ",")
				if projects != lastGroup {
					if i > 0 {
						fmt.Fprintf(f.Writer, "%s\n", "")
					}
					fmt.Fprintln(f.Writer, f.fgYellow("[") + f.formatProjects(todo.Projects) + f.fgYellow("]"))
					lastGroup = projects
					rowNum = 0
				}
			} else {
				contexts := strings.Join(todo.Contexts, ",")
				if contexts != lastGroup {
					if i > 0 {
						fmt.Fprintf(f.Writer, "%s\n", "")
					}
					fmt.Fprintln(f.Writer, f.fgYellow("[") + f.formatContexts(todo.Contexts) + f.fgYellow("]"))
					lastGroup = contexts
					rowNum = 0
				}
			}
		}
		if rowNum%consoleHeight == 0 {

			if rowNum > 0 {
				fmt.Fprintf(f.Writer, "%s\n", "")
				f.Writer.Flush()
			}

			f.printColumnHeaders(report.Columns, report.Headers)
		}

		f.printCustomTodo(todo, report.Columns)
		if report.PrintNotes && len(todo.Notes) > 0 {
			f.printNotes(todo.Notes)
		}
		rowNum++
	}
	f.Writer.Flush()
}

func (f *ScreenPrinter) printColumnHeaders(cols []string, headers []string) {
	vals := []string{}
	for i, col := range cols {
		//Note the switch statement simply ensures cols are valid choices
		switch col {
		case "id":
			vals = append(vals, f.fgGreen(headers[i]))
		case "completed":
			vals = append(vals, f.fgGreen(headers[i]))
		case "age":
			vals = append(vals, f.fgGreen(headers[i]))
		case "due":
			vals = append(vals, f.fgGreen(headers[i]))
		case "priority":
			vals = append(vals, f.fgGreen(headers[i]))
		case "ord:all":
			vals = append(vals, f.fgGreen(headers[i]))
		case "ord:pro":
			vals = append(vals, f.fgGreen(headers[i]))
		case "ord:ctx":
			vals = append(vals, f.fgGreen(headers[i]))
		case "context":
			vals = append(vals, f.fgGreen(headers[i]))
		case "project":
			vals = append(vals, f.fgGreen(headers[i]))
		case "subject":
			vals = append(vals, f.fgGreen(headers[i]))
		}
	}
	f.PrintRow(vals)
}

func (f *ScreenPrinter) PrintRow(line []string) {
	fmt.Fprintln(f.Writer, strings.Join(line, "\t"))
}

func iface(list []string) []interface{} {
	vals := make([]interface{}, len(list))
	for i, v := range list {
		vals[i] = v
	}
	return vals
}

//Print todo with specific columns, order of columns, column headings, sort order
func (f *ScreenPrinter) printCustomTodo(todo *Todo, cols []string) {
	//PROBLEM: tabwriter chokes on the additional formatting codes for italic and bold. Throws off column alignment. Fix later (maybe new tabwriter)
	//if todo.IsPriority {
	//	yellow.Add(color.Bold, color.Italic)
	//}
	vals := []string{}
	for _, col := range cols {
		switch col {
		case "id":
			vals = append(vals, f.fgYellow(strconv.Itoa(todo.Id)))
		case "completed":
			vals = append(vals, f.formatCompleted(todo.Completed))
		case "age":
			vals = append(vals, f.formatAge(todo.CreatedDate))
		case "due":
			vals = append(vals, f.formatDue(todo.Due))
		case "priority":
			vals = append(vals, f.formatPriority(todo.Priority))
		case "ord:all":
			vals = append(vals, f.formatOrdinal(0, todo)) //0 = all
		case "ord:pro":
			vals = append(vals, f.formatOrdinal(1, todo)) //1 = pro
		case "ord:ctx":
			vals = append(vals, f.formatOrdinal(2, todo)) //2 = ctx
		case "context":
			vals = append(vals, f.formatContexts(todo.Contexts))
		case "project":
			vals = append(vals, f.formatProjects(todo.Projects))
		case "subject":
			vals = append(vals, f.formatSubject(todo.Subject))
		}
	}
	f.PrintRow(vals)
}

func (f *ScreenPrinter) formatDue(due string) string {

	if due == "" {
		return f.fgBlue(" ")
	}
	dueTime, err := time.Parse(time.RFC3339, due)

	if err != nil {
		fmt.Println(err)
		fmt.Println("This may due to the corruption of .todos.json file.")
		os.Exit(-1)
	}

	if isToday(dueTime) {
		return f.fgBlue("today")
	} else if isTomorrow(dueTime) {
		return f.fgBlue("tomorrow")
	} else if isPastDue(dueTime) {
		return f.fgRed(dueTime.Format("Mon Jan 2"))
	} else {
		return f.fgBlue(dueTime.Format("Mon Jan 2"))
	}
}

func (f *ScreenPrinter) formatProjects(projects []string) string {

	words := []string{}

	for _, word := range projects {
		words = append(words, word)
	}
	coloredWords := f.fgMagenta(strings.Join(words, ", "))
	return coloredWords
}

func (f *ScreenPrinter) formatContexts(contexts []string) string {
	words := []string{}
	for _, word := range contexts {
		words = append(words, word)
	}
	coloredWords := f.fgRed(strings.Join(words, ", "))
	return coloredWords
}

func (f *ScreenPrinter) formatSubject(subject string) string {
	return f.fgWhite(subject)
}

func (f *ScreenPrinter) formatPriority(p string) string {
	return f.fgRed(p)
}

func (f *ScreenPrinter) formatOrdinal(ordType int, todo *Todo) string {
	switch ordType {
	case 0:
		return f.fgYellow(todo.Ordinals["all"])
	case 1:
		if len(todo.Projects) > 0 {
			return f.fgYellow(todo.Ordinals["+"+todo.Projects[0]])
		}
	case 2:
		if len(todo.Contexts) > 0 {
			return f.fgYellow(todo.Ordinals["@"+todo.Contexts[0]])
		}
	}
	return f.fgYellow(" ")
}

func (f *ScreenPrinter) formatAge(createdDate string) string {
	days := 0
	if len(createdDate) > 0 {
		tmpTime, err := time.Parse(time.RFC3339, createdDate)
		if err == nil {
			createTime := tmpTime.Unix()
			now := time.Now().Unix()
			diff := now - createTime
			days = (int)(diff / (60 * 60 * 24))
		}
	}
	coloredWords := f.fgYellow(days, "d")
	return coloredWords
}

func (f *ScreenPrinter) formatCompleted(completed bool) string {
	if completed {
		return f.fgWhite("[x]")
	} else {
		return f.fgWhite("[ ]")
	}
}

func (f *ScreenPrinter) PrintOverallHelp() {
	
	tmp := []string{f.fgCyan("Todolist is a simple, command line based, GTD-style todo manager")}
	f.PrintRow(tmp)
	f.Writer.Flush()

	fmt.Println("")
	colors := []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors, "  Syntax:", "todo [filters] <command> [modifiers] [args]")
	f.Writer.Flush()

	fmt.Println("")
	colors = []func(a ...interface{}) string {f.fgGreen, f.fgGreen}
	f.printCols(colors, "  Command", "Description")
	colors = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors, "  help", "Print this help message. Pass specific command or 'config' as arg for more detail.")
	f.printCols(colors, "  init", "Initialize a new repository in local directory.")
	f.printCols(colors, "  add | a", "Add a new todo.")
	f.printCols(colors, "  done", "Add an already completed todo (for recording purposes)")
	f.printCols(colors, "  list | l", "List todos. Listed todos can be constrained by filters (see below).")
	f.printCols(colors, "  projects", "List all projects and count of todos for each.")
	f.printCols(colors, "  contexts", "List all contexts and count of todos for each.")
	f.printCols(colors, "  print", "Print all todo details. Select todos by filter (see below).")
	f.printCols(colors, "  edit | e", "Edit one or more todos. Todos edited are determined by filters (see below)")
	f.printCols(colors, "  delete | d", "Delete todos. Deleted todos can be constrained by filters (see below).")
	f.printCols(colors, "  order | ord | reorder", "Order todos in a set (all|+project|@context) relative to each other using ids.")
	f.printCols(colors, "  complete | c", "Mark one or more todos as completed. Select todos by filter (see below).")
	f.printCols(colors, "  archive | ar", "Archive one or more todos. Select todos by filter (see below).")
	f.printCols(colors, "  unarchive | uar", "Un-archive one ore more todos. Select todos by filter (see below).")
	f.printCols(colors, "  ac", "Archive all completed todos.")
	f.printCols(colors, "  sync", "Synchronize todos with another file location. See .todorc file for sample config.")
	f.printCols(colors, "  open", "Open a file or URL referenced in a todo note. The value 'notes' opens a todo-specific text file.")
	f.printCols(colors, "  an", "Add a note to one or more todos. Select todos by filter (see below).")
	f.printCols(colors, "  en", "Edit a note for one or more todos. Select todos by filter (see below).")
	f.printCols(colors, "  dn", "Delete a note for one or more todos. Select todos by filter (see below).")
	f.printCols(colors, "  view", "Set a view (ie. a default set of filters). A view is typically based on a context filter.")
	f.printCols(colors, "  gc", "Garbage collect (permanently delete) all archived todos.")
	f.Writer.Flush()

	println("")
	println(f.fgGreen("  Date specifiers used in filters and modifiers:"))
	f.printCols(colors, "    tod(ay)|tom(orrow)|yes(terday)|this_week|next_week|last_week", "Relative date.")
	f.printCols(colors, "    mon|tue|wed|thu|fri|sat|sun", "Day of the week.")
	f.printCols(colors, "    1d|1w|1m|1y", "Date calculated using relative duration from today.")
	f.printCols(colors, "    any", "Any date specified (e.g. filter for todos with any due date (ie. not blank)).")
	f.printCols(colors, "    none", "No date specified (e.g. filter for todos with no due date).")
	f.printCols(colors, "    overdue", "Past due todos.")
	f.Writer.Flush()

	println("")
	println(f.fgGreen("  Filters: "))
	f.printCols(colors, "    [id or id range]", "Filter for specific id (e.g. 4) or range of ids (e.g. 4-7).")
	f.printCols(colors, "    +[project name]", "Filter for todos with the specified project.")
	f.printCols(colors, "    -[project name]", "Filter for todos WITHOUT the specified project.")
	f.printCols(colors, "    @[context name]", "Filter for todos with the specified context.")
	f.printCols(colors, "    -@[context name]", "Filter for todos WITHOUT the specified context.")
	f.printCols(colors, "    due:[date][:end date]", "Filter for todos with due dates equal to date or within date range.")
	f.printCols(colors, "    pri:[priorities (comma-separated)]", "Filter for indicated priorities.")
	f.printCols(colors, "    age:[number or range of days]", "Filter for todos by age (e.g. age:1 or age:1-5).")
	f.printCols(colors, "    top:[pro|ctx]:[number]", "Filter for top N todos by project or context, based on sort chosen.")
	f.printCols(colors, "    waiting", "Filter for todos that are waiting.")
	f.printCols(colors, "    until", "Filter for todos that will expire.")
	f.printCols(colors, "    completed", "Filter for todos that are completed.")
	f.printCols(colors, "    archived", "Filter for todos that are archived.")
	f.printCols(colors, "    notes:[true or false]", "Filter for todos with notes (or without notes if false).")
	f.printCols(colors, "    [search words]", "Filter for todos with search words in the subject. Must not match other filters above.")
	f.Writer.Flush()

	println("")
	println(f.fgGreen("  Modifiers:"))
	f.printCols(colors, "    +[project name]", "Add a project.")
	f.printCols(colors, "    -[project name]", "Remove a project.")
	f.printCols(colors, "    @[context name]", "Add a context.")
	f.printCols(colors, "    -@[context name]", "Remove a context.")
	f.printCols(colors, "    due:[date]", "Add or change the due date.")
	f.printCols(colors, "    wait:[date specifier]", "Add or change the wait date.")
	f.printCols(colors, "    until:[date specifier]", "Add or change the until (expiry) date.")
	f.printCols(colors, "    pri:[priority specifier]", "Add or change the priority. Configurable. Default values are H,M,L.")
	f.Writer.Flush()

	println("")
	println(f.fgGreen("  Arguments (Generally only for list or report commands):"))
	f.printCols(colors, "    sort:[+|-][id|project|context|ord:[all|pro|ctx]|due|age|priority]", "Override sort for the todo list.")
	f.printCols(colors, "    filter:[+|-][see filters above]", "Override filters for todo list.")
	f.printCols(colors, "    group:[project | context]", "Group todos by project or context. Override group config for todo list.")
	f.printCols(colors, "    notes:[true or false]", "List of todos will include the notes for todos that have them.")
	f.Writer.Flush()
	
	println("")
	println(f.fgGreen("  For full documentation, please visit http://github.com/fkmiec/todolist"))
}

func (f *ScreenPrinter) printCols(colors []func(a ...interface{}) string, txt ...string) {
	vals := []string{}
	for i, val := range txt {
		vals = append(vals, colors[i](val))
	}
	f.PrintRow(vals)
}

func (f *ScreenPrinter) PrintAddHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Add a todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [add | a] [modifiers]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for adding a todo:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Add todo with project (BigProject), context (Bob) and due date using relative date.")
	f.printCols(colors2, "  Example:  ", "todo add Meeting with Bob. +BigProject @Bob due:today")
	f.printCols(colors1, "Add todo with project (HoneyDo), context (Home) and due within 1 week.")
	f.printCols(colors2, "  Example:  ", "todo a +HoneyDo @Home Fix the sink. due:1w")
	f.printCols(colors1, "Add todo with project (Lawn) and wait date of next Saturday (ie. will not show in list until Saturday).")
	f.printCols(colors2, "  Example:  ", "todo a +Lawn Mow and trim. wait:sat")
	f.printCols(colors1, "Add todo with context (Wife), priority (H), due date 2018-09-21 and expiration (ie. auto archive) after 2018-09-21.")
	f.printCols(colors2, "  Example:  ", "todo a due:2018-09-21 until:2018-09-22 Buy anniversary gift. @Wife pri:H")	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintListHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "List todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] [list | l | <blank>] [arguments]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for listing todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "List all pending todos using default report and default command.")
	f.printCols(colors2, "  Example:  ", "todo")
	f.printCols(colors1, "List todos using default report with ids 1 to 5.")
	f.printCols(colors2, "  Example:  ", "todo 1-5")
	f.printCols(colors1, "List todos using default report with project BigProject that are due this week.")
	f.printCols(colors2, "  Example:  ", "todo +BigProject due:this_week")
	f.printCols(colors1, "List High and Medium priority todos using configured report 'due' that sorts by project and due date ascending.")
	f.printCols(colors2, "  Example:  ", "todo pri:H,M due")
	f.printCols(colors1, "List todos due Monday.")
	f.printCols(colors2, "  Example:  ", "todo due:mon")
	f.printCols(colors1, "List todos due after today. Trailing colon indicates open ended end date.")
	f.printCols(colors2, "  Example:  ", "todo due:tod:")
	f.printCols(colors1, "List todos due before Sunday. Leading colon indicates open ended start date.")
	f.printCols(colors2, "  Example:  ", "todo due::sun")
	f.printCols(colors1, "List todos due the next two weeks.")
	f.printCols(colors2, "  Example:  ", "todo due:0w:2w")
	f.printCols(colors1, "List todos less than one week old.")
	f.printCols(colors2, "  Example:  ", "todo age:0-7")
	f.printCols(colors1, "List todos more than two weeks old.")
	f.printCols(colors2, "  Example:  ", "todo age:14-0")
	f.printCols(colors1, "List todos containing the word 'leadership approval'.")
	f.printCols(colors2, "  Example:  ", "todo leadship approval list")
	f.printCols(colors1, "List with filter for top 2 todos for each project and sort list by project.")
	f.printCols(colors2, "  Example:  ", "todo top:pro:2 list sort:project")
	f.printCols(colors1, "List todos sorted by project and due date ascending.")
	f.printCols(colors2, "  Example:  ", "todo l sort:+project,+due")
	f.printCols(colors1, "List completed todos.")
	f.printCols(colors2, "  Example:  ", "todo completed")	
	f.printCols(colors1, "List archived todos.")
	f.printCols(colors2, "  Example:  ", "todo archived")		
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintEditHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Edit todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] [edit | e] [modifiers]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for editing todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Edit the todo with id 2 to add a due date.")
	f.printCols(colors2, "  Example:  ", "todo 2 e due:sat")
	f.printCols(colors1, "Edit the todo with id 2 to remove a due date. Use due:none or due:<blank>.")
	f.printCols(colors2, "  Example:  ", "todo 2 e due:none")
	f.printCols(colors1, "Edit all todos for project BigProject to add expiration on Wednesday.")
	f.printCols(colors2, "  Example:  ", "todo +BigProject e until:wed")
	f.printCols(colors1, "Edit todos with ids 1,3,5,7,9 to add context OddNumbers and remove context EvenNumbers.")
	f.printCols(colors2, "  Example:  ", "todo 1,3,5,7,9 @OddNumbers -@EvenNumbers")
	f.printCols(colors1, "Edit todos with low priority to wait until next week.")
	f.printCols(colors2, "  Example:  ", "todo pri:L e wait:1w")
	f.printCols(colors1, "Edit todo with id 5 to change the subject.")
	f.printCols(colors2, "  Example:  ", "todo 5 e Do something else instead.")		
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintDeleteHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Delete todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] [delete | d]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for deleting todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Delete the todo with id 2.")
	f.printCols(colors2, "  Example:  ", "todo 2 delete")
	f.printCols(colors1, "Delete all todos for project BigProject. Warning: Be careful with bulk deletes!")
	f.printCols(colors2, "  Example:  ", "todo +BigProject d")
	f.printCols(colors1, "Delete all waiting todos.")
	f.printCols(colors2, "  Example:  ", "todo waiting d")		
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintOpenHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Open a file or URL referenced in a todo note.")
	f.printCols(colors1, "Uses system default program or configured command.")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo <filter> open <args>")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Create and open a notes file associated with a todo")
	f.printCols(colors2, "  Example:  ", "todo 1 an notes")
	f.printCols(colors2, "            ", "todo 1 open")
	f.printCols(colors1, "Open a web URL associated with a todo")
	f.printCols(colors2, "  Example:  ", "todo 2 an Find stuff using www.google.com")
	f.printCols(colors2, "            ", "todo 2 open")	
	f.printCols(colors1, "Open a .docx file associated with a todo")
	f.printCols(colors2, "  Example:  ", "todo 3 an Review C:/Documents/important.docx for the boss.")
	f.printCols(colors2, "            ", "todo 3 open")	
	f.printCols(colors1, "Open the URI associated with third note of todo with id 1. Notes indexing starts at 0.")
	f.printCols(colors2, "  Example:  ", "todo 1 open 2")
	f.printCols(colors1, "Open a URI for a note in verbose mode to help debug issues with configured regex.")
	f.printCols(colors2, "  Example:  ", "todo 1 open verbose")
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintProjectsHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Print list of projects with count of todos for each")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo projects")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Print list of projects")
	f.printCols(colors2, "  Example:  ", "todo projects")	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintContextsHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Print list of contexts with count of todos for each")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo contexts")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Print list of contexts")
	f.printCols(colors2, "  Example:  ", "todo contexts")	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintPrintTodoDetailHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Print details of todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] print")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Print details of todo with id 3")
	f.printCols(colors2, "  Example:  ", "todo 3 print")
	f.printCols(colors1, "Print details of todos for project BigProject")
	f.printCols(colors2, "  Example:  ", "todo +BigProject print")	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintGarbageCollectHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Garbage Collect all archived todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo gc")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for garbage collecting all archived todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Delete all archived todods")
	f.printCols(colors2, "  Example:  ", "todo gc")	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintDeleteNoteHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Delete a note to a todo")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] dn [note index]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for deleting notes to todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Delete first note of todo with id 3. Note that notes index starts with 0.")
	f.printCols(colors2, "  Example:  ", "todo 3 dn 0")	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintEditNoteHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Edit a note to a todo")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] en [note index] [new note text]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for editing notes to todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Edit second note of todo with id 1")
	f.printCols(colors2, "  Example:  ", "todo 1 en 1 This is the second note. List indexing starts at 0.")	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintAddNoteHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Add a note to a todo")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] an [note text]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for adding notes to todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Add a URL as a note to the todo with id 1")
	f.printCols(colors2, "  Example:  ", "todo 1 an http://google.com")
	f.printCols(colors1, "Add a note to todos 1,6,8")
	f.printCols(colors2, "  Example:  ", "todo 1,6,8 an Discuss with Bob at 1-1 meeting.")	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintDoneHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Add an already completed todo")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo done [modifiers]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for adding already completed todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Add a completed todo with due date today and priority High.")
	f.printCols(colors2, "  Example:  ", "todo done Meet with Bob due:tod pri:H")
	f.printCols(colors1, "Add a completed todo with project BigProject")
	f.printCols(colors2, "  Example:  ", "todo done +BigProject Meet with stakeholders.")	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintInitHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Initialize a todo repo")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo init")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Initialize a repo in the current folder.")
	f.printCols(colors1, "Creates config file .todorc, pending todos file .todos.json, archived todos file .todos_archive.json and backlog file .todos_backlog.json")
	f.printCols(colors2, "  Example:  ", "todo init")
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintViewHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Set a 'view' (aka a default filter. Typically a context, such as home or work)")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo view [filters]")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Set view to context home. See .todorc config file for example of defining a view.")
	f.printCols(colors2, "  Example:  ", "todo view home")
	f.printCols(colors1, "Change view to context work.")
	f.printCols(colors2, "  Example:  ", "todo view work")
	f.printCols(colors1, "Unset view (a.k.a. No default filtering applied)")
	f.printCols(colors2, "  Example:  ", "todo view")
	
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintSyncHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Sync todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo sync [verbose]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for syncing todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Sync todos to a configured alternate file location.")
	f.printCols(colors1, "See .todorc config file for configuring sync location and encryption.")
	f.printCols(colors2, "  Example:  ", "todo sync")
	f.printCols(colors1, "Sync todos and list each locally added, modified or deleted todo.")
	f.printCols(colors2, "  Example:  ", "todo sync verbose")
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintOrderTodosHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Order todos (relative to each other in groups 'all' or by project or by context).")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [order | ord | reorder] [all|<project>|<context>]:[ids]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for [re]ordering todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Move todo with id 8 behind todo with id 14. Todo 14 will keep its position.")
	f.printCols(colors1, "To see impact of re-ordering, you need to use a report that sorts by ord:all.")
	f.printCols(colors2, "  Example:  ", "todo order all:14,8")
	f.printCols(colors1, "Move todo with id 6 to top of the list for project BigProject. Note that id '0' represents the top of the list.")
	f.printCols(colors1, "To see impact of re-ordering, you need to use a report that sorts by ord:pro.")
	f.printCols(colors2, "  Example:  ", "todo ord +BigProject:0,6")
	f.printCols(colors1, "Move todos with ids 11,2,5,9 to follow id 4 in list sorted by context Home. Todo 4 will keep its position.")
	f.printCols(colors1, "To see impact of re-ordering, you need to use a report that sorts by ord:ctx.")
	f.printCols(colors2, "  Example:  ", "todo reorder all:0,6")
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintUnarchiveHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Unarchive todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] [unarchive | uar]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for un-archiving todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Unarchive todo with id 8.")
	f.printCols(colors2, "  Example:  ", "todo 8 uar")
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintArchiveCompletedHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Archive all completed todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo ac")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for archiving completed todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Archive all completed todos.")
	f.printCols(colors2, "  Example:  ", "todo ac")
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintArchiveHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Archive todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] [archive | ar]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for archiving todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Archive the todo with id 7.")
	f.printCols(colors2, "  Example:  ", "todo 7 ar")
	f.printCols(colors1, "Archive all waiting todos.")
	f.printCols(colors2, "  Example:  ", "todo waiting archive")		
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintCompleteHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Complete todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] [complete | c]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for completing todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Complete the todo with id 2.")
	f.printCols(colors2, "  Example:  ", "todo 2 complete")
	f.printCols(colors1, "Complete all todos for project BigProject.")
	f.printCols(colors2, "  Example:  ", "todo +BigProject c")		
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintUncompleteHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Uncomplete todos")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Syntax: ", "todo [filters] [uncomplete | uc]")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Examples for un-completing todos:"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Uncomplete the todo with id 2.")
	f.printCols(colors2, "  Example:  ", "todo 2 uncomplete")
	f.printCols(colors1, "Uncomplete all todos for project BigProject.")
	f.printCols(colors2, "  Example:  ", "todo +BigProject uc")		
	f.Writer.Flush()
}

func (f *ScreenPrinter) PrintConfigHelp() {
	colors1 := []func(a ...interface{}) string {f.fgYellow}
	f.printCols(colors1, "Configuration")
	f.Writer.Flush()

	fmt.Println("")
	colors1 = []func(a ...interface{}) string {f.fgCyan, f.fgYellow}
	f.printCols(colors1, "  Filename: ", ".todorc")
	f.Writer.Flush()

	fmt.Println("")
	fmt.Println(f.fgGreen("Configuration Attributes (key=value format):"))
	colors1 = []func(a ...interface{}) string {f.fgBlue, f.fgYellow}
	colors2 := []func(a ...interface{}) string {f.fgMagenta, f.fgYellow}
	f.printCols(colors1, "Configure a report (format for listing todos). Report name is an alias for 'list'. Report 'default' will be applied if no other report name matched.")
	f.printCols(colors2, "  report.<name>.description  ", "A description for this report.")
	f.printCols(colors2, "  report.<name>.columns  ", "Columns to display (comma-sep). [id|completed|age|due|context|project|ord:all|ord:pro|ord:ctx]")
	f.printCols(colors2, "  report.<name>.headers  ", "Display headers for columns (comma-sep). e.g. 'Id' for id, 'Age' for age.")
	f.printCols(colors2, "  report.<name>.sort  ", "Multi-sorting instructions (comma-sep). [+/-][id|age|due|context|project|ord:all|ord:pro|ord:ctx]")
	f.printCols(colors2, "  report.<name>.filter  ", "Filters (comma-sep). See main 'help' for details on filters.")
	f.printCols(colors2, "  report.<name>.group  ", "[project | context]")
	f.printCols(colors2, "  report.<name>.notes  ", "[true|false]")
	f.printCols(colors1, "Configure priority values. Default is H,M,L.")
	f.printCols(colors2, "  priority  ", "[comma-separated values] Order highest to lowest (e.g. H,M,L).")
	f.printCols(colors1, "Configure synchronization of todos to another file location.")
	f.printCols(colors2, "  sync.filepath  ", "[Path to file including filename. Directory must exist.]")
	f.printCols(colors2, "  sync.encrypt.passphrase  ", "[passphrase | * (prompt) | <blank> (don't encrypt)]")
	f.printCols(colors1, "Define aliases to save typing on common commands.")
	f.printCols(colors2, "  alias.<name>  ", "[command line to alias (after todo executable). E.g. list group:project]")
	f.printCols(colors1, "Define named view filters that can be applied by default and referenced by name.")
	f.printCols(colors2, "  view.<name>.filter  ", "[comma-separated filters. E.g. @home,due:any]")
	f.printCols(colors1, "Set the currently applied view filter from list of views defined. Change with 'view' command.")
	f.printCols(colors2, "  view.current  ", "[view name]")
	f.printCols(colors1, "Configure the open command")
	f.printCols(colors2, "  open.notes.folder  ", "[path to notes folder]")
	f.printCols(colors2, "  open.notes.ext  ", "[file extension for notes file]")
	f.printCols(colors2, "  open.notes.regex  ", "[regex to match to open notes file]")
	f.printCols(colors2, "  open.notes.cmd  ", "[command to open notes file]")
	f.printCols(colors1, "Configure file and URI types and commands to open them. Uses system default if not specified.")
	f.printCols(colors2, "  open.browser.regex  ", "[regex to match browser URLs]")
	f.printCols(colors2, "  open.browser.cmd  ", "[command to open browser URLs]")
	f.printCols(colors2, "  open.file.regex  ", "[regex to match file paths]")
	f.printCols(colors2, "  open.file.cmd  ", "[command to open file paths]")
	f.printCols(colors2, "  open.<type>.regex  ", "[regex to match <type> paths]")
	f.printCols(colors2, "  open.<type>.cmd  ", "[command to open <type> paths]")
	f.Writer.Flush()
}
