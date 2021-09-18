package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Action interface {
	Page() int
}

type Delete struct {
	Count int
	PageNo int
}
func (d Delete) Page() int {
	return d.PageNo
}

type Insert struct {
	Count int
	PageNo int
	InsertAfter bool
}
func (i Insert) Page() int {
	return i.PageNo
}

func ActionsFromString(s string) []Action {
	actions := make([]Action, 0)

	actionStrs := strings.Split(s, ",")

	for _, actionStr := range actionStrs {
		if strings.HasPrefix(actionStr, "-") {
			pageNo, err := strconv.Atoi(actionStr[1:])
			if err != nil {
				panic(err)
			}
			actions = append(actions, Delete{Count: 1, PageNo: pageNo})
		} else if strings.Contains(actionStr, "a") {
			// XaY => Insert{X, Y, true}
			args := strings.Split(actionStr, "a")
			count, err := strconv.Atoi(args[0])
			if err != nil {
				panic(err)
			}
			pageNo, err := strconv.Atoi(args[1])
			if err != nil {
				panic(err)
			}

			actions = append(actions, Insert{Count: count, PageNo: pageNo, InsertAfter: true})
		} else if strings.Contains(actionStr, "b") {
			// XbY => Insert{X, Y, false}
			args := strings.Split(actionStr, "b")
			count, err := strconv.Atoi(args[0])
			if err != nil {
				panic(err)
			}
			pageNo, err := strconv.Atoi(args[1])
			if err != nil {
				panic(err)
			}

			actions = append(actions, Insert{Count: count, PageNo: pageNo, InsertAfter: false})
		}
	}

	checkActions(actions)
	return actions
}

func checkActions(actions []Action) {
	seen := make(map[int]struct{})
	for _, a := range actions {
		if _, ok := seen[a.Page()]; ok {
			panic(fmt.Sprintf("Page %i occurs more than once!", a.Page()))
		}

		seen[a.Page()] = struct{}{}
	}
}

