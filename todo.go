package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fkmiec/todo/todolist"
)

const (
	VERSION = "1.0"
)

func main() {
	if len(os.Args) <= 1 {
		os.Args = append(os.Args, "l")
	}

	input := strings.Join(os.Args[1:], " ")
	app := todolist.NewApp()
	command := app.ProcessCmdLine(input)

	//Protect against mass edit or delete
	cmd := command.GetCmd()
	if (cmd == "edit" || cmd == "archive" || cmd == "delete" || cmd == "complete") && len(command.GetFilters()) == 0 {
		fmt.Println("Destructive operations require a filter. None specified. Aborting.")
		return
	}
	//Apply the view (set of filters applied by default)
	if len(app.Cfg.CurrentView) > 0 {
		viewFilters := app.Cfg.Views[app.Cfg.CurrentView]
		command.SetFilters(append(viewFilters, command.GetFilters()...))
	}

	command.Exec(app)

}
