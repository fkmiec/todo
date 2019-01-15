package todolist

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Parser struct{}

func (p *Parser) ParseNewTodo(mods []string, todolist *TodoList) *Todo {
	if len(mods) == 0 {
		return nil
	}

	todo := NewTodo()

	p.ParseInput(mods, todo, todolist)
	todolist.AddOrdinal("all", todo)

	return todo
}

func (p *Parser) ParseInput(mods []string, todo *Todo, todolist *TodoList) {
	subj := []string{}
	for i, part := range mods {
		if strings.HasPrefix(part, "+") {
			tmp := part[1:]
			todolist.AddProject(tmp, todo)
		} else if strings.HasPrefix(part, "@") {
			tmp := part[1:]
			todolist.AddContext(tmp, todo)
		} else if strings.HasPrefix(part, "-") {
			if strings.Index(part, "@") == 1 {
				tmp := part[2:]
				todolist.RemoveContext(tmp, todo)
			} else {
				tmp := part[1:]
				todolist.RemoveProject(tmp, todo)
			}
		} else if strings.HasPrefix(part, "due:") {
			tmp := part[4:]
			todo.Due = p.FormatDateTime(tmp, Now)
		} else if strings.HasPrefix(part, "wait:") {
			tmp := part[5:]
			todo.Wait = p.FormatDateTime(tmp, Now)
		} else if strings.HasPrefix(part, "until:") {
			tmp := part[6:]
			todo.Until = p.FormatDateTime(tmp, Now)
		} else if strings.HasPrefix(part, "pri:") {
			tmp := part[4:]
			todo.Priority = tmp
		} else {
			subj = append(subj, mods[i])
		}
	}
	if len(subj) > 0 {
		s := strings.Join(subj, " ")
		if strings.HasPrefix(s, "pre:") {
			s = s[4:]
			todo.Subject = s + todo.Subject
		} else if strings.HasPrefix(s, "app:") {
			s = s[4:]
			todo.Subject = todo.Subject + s
		} else {
			todo.Subject = s
		}
	}
	return
}

func (p *Parser) ParseEditTodo(todo *Todo, mods []string, todolist *TodoList) bool {

	if len(mods) == 0 {
		return false
	}

	p.ParseInput(mods, todo, todolist)
	return true
}

func (p *Parser) Projects(filters []string) []string {
	projects := []string{}
	for _, filter := range filters {
		if strings.HasPrefix(filter, "+") {
			projects = append(projects, filter[1:])
		}
	}
	return projects
}

func (p *Parser) Contexts(filters []string) []string {
	contexts := []string{}
	for _, filter := range filters {
		if strings.HasPrefix(filter, "@") {
			contexts = append(contexts, filter[1:])
		}
	}
	return contexts
}

func (p *Parser) ParseAddNote(todo *Todo, mods []string) bool {
	note := strings.Join(mods, " ")
	todo.Notes = append(todo.Notes, note)
	return true
}

func (p *Parser) ParseDeleteNote(todo *Todo, mods []string) bool {
	for _, part := range mods {
		rmid, err := p.getNoteID(part)
		if err != nil {
			return false
		}

		for id, _ := range todo.Notes {
			if id == rmid {
				todo.Notes = append(todo.Notes[:rmid], todo.Notes[rmid+1:]...)
				return true
			}
		}
	}
	return false
}

func (p *Parser) ParseEditNote(todo *Todo, mods []string) bool {
	if len(mods) < 2 {
		return false
	}

	rmid, err := p.getNoteID(mods[0])
	if err != nil {
		return false
	}

	for id, _ := range todo.Notes {
		if id == rmid {
			todo.Notes[id] = strings.Join(mods[1:], " ")
			return true
		}
	}
	return false
}

func (p *Parser) getNoteID(input string) (int, error) {
	ret, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("wrong note id")
		return -1, err
	}
	return ret, nil
}

func (p *Parser) FormatDateTime(input string, relativeTime time.Time) string {
	return timeToString(p.ParseDateTime(input, relativeTime))
}

func (p *Parser) ParseDateTime(input string, relativeTime time.Time) time.Time {

	tmp := strings.ToLower(input)
	//Check for relative date
	r, _ := regexp.Compile(`(-?\d+)([dwmyh]{1})`)
	if r.MatchString(tmp) {
		matches := r.FindStringSubmatch(tmp)
		unit := strings.ToLower(matches[2])
		cnt, err := strconv.Atoi(matches[1])
		if err != nil {
			fmt.Println("Could not parse date: ", input, ":", err)
			os.Exit(1)
		}
		targetDate := Now
		if unit == "d" {
			targetDate = targetDate.AddDate(0, 0, 1*cnt)
		} else if unit == "w" {
			targetDate = targetDate.AddDate(0, 0, 7*cnt)
		} else if unit == "m" {
			targetDate = targetDate.AddDate(0, 1*cnt, 0)
		} else if unit == "y" {
			targetDate = targetDate.AddDate(1*cnt, 0, 0)
		} else if unit == "h" {
			targetDate = targetDate.Add(time.Duration(cnt) * time.Hour)
		}
		return bod(targetDate)
	}

	//support look back a week as well as look forward
	forward := true
	if strings.HasPrefix(tmp, "-") {
		forward = false
		tmp = tmp[1:]
	}
	switch {
	case strings.HasPrefix(tmp, "non"):
		return bod(relativeTime)
	case strings.HasPrefix(tmp, "tod"):
		return bod(relativeTime)
	case strings.HasPrefix(tmp, "tom"):
		return bod(relativeTime).AddDate(0, 0, 1)
	case strings.HasPrefix(tmp, "yes"):
		return bod(relativeTime).AddDate(0, 0, -1)
	case strings.HasPrefix(tmp, "mon"):
		return monday(relativeTime, forward)
	case strings.HasPrefix(tmp, "tue"):
		return tuesday(relativeTime, forward)
	case strings.HasPrefix(tmp, "wed"):
		return wednesday(relativeTime, forward)
	case strings.HasPrefix(tmp, "thu"):
		return thursday(relativeTime, forward)
	case strings.HasPrefix(tmp, "fri"):
		return friday(relativeTime, forward)
	case strings.HasPrefix(tmp, "sat"):
		return saturday(relativeTime, forward)
	case strings.HasPrefix(tmp, "sun"):
		return sunday(relativeTime, forward)
	case tmp == "last_week":
		n := bod(relativeTime)
		return mostRecentMonday(n).AddDate(0, 0, -7)
	case tmp == "this_week":
		n := bod(relativeTime)
		return mostRecentMonday(n)
	case tmp == "next_week":
		n := bod(relativeTime)
		return mostRecentMonday(n).AddDate(0, 0, 7)
	}
	return p.parseArbitraryDate(tmp)
}

func (p *Parser) parseArbitraryDate(_date string) time.Time {

	if date, err := time.Parse("2006-01-02", _date); err == nil {
		return date
	}

	if date, err := time.Parse("20060102", _date); err == nil {
		return date
	}

	fmt.Printf("Could not parse the date you gave me: %s\n", _date)
	fmt.Println("I'm expecting a date like \"yyyy-MM-dd\" or \"yyyyMMdd\".")
	os.Exit(-1)
	return Now
}

/*
func (p *Parser) matchWords(input string, r *regexp.Regexp) []string {
	results := r.FindAllString(input, -1)
	ret := []string{}

	for _, val := range results {
		ret = append(ret, val[1:])
	}
	return ret
}
*/
