package actions

import (
	"fmt"
	"strconv"
	"strings"
)

func FromString(s string) []Action {
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

func runSlice(arr []interface{}, actions []Action, defaultCreator func() interface{}) []interface{} {
	Sort(actions)

	for i := len(actions)-1; i >= 0; i-- {
		currActionI := actions[i]
		switch currAction := currActionI.(type) {
		case Delete:
			pageIdx := currAction.PageNo - 1
			arr = append(arr[:pageIdx], arr[pageIdx + currAction.Count:]...)
		case Insert:
			pageIdx := currAction.PageNo - 1

			// make copy of arrPre, because appending causes issues due to aliasing
			arrPre := make([]interface{}, pageIdx+1)
			copy(arrPre, arr[:pageIdx+1])
			//arrPre := arr[:pageIdx+1]
			arrPost := arr[pageIdx+1:]
			if !currAction.InsertAfter {
				arrPre = make([]interface{}, pageIdx)
				copy(arrPre, arr[:pageIdx])
				//arrPre = arr[:pageIdx]
				arrPost = arr[pageIdx:]
			}

			arrMid := make([]interface{}, 0)

			//fmt.Println("Processing insert action, arr:", arr)
			for c := 0; c < currAction.Count; c++ {
				arrMid = append(arrMid, defaultCreator())
				//fmt.Println("arrMid", arrMid)
			}

			//fmt.Println("after loop: arrPre:", arrPre, "arrMid:", arrMid, "arrPost:", arrPost)
			arr = append(arrPre, arrMid...)
			arr = append(arr, arrPost...)
			//fmt.Println("arr final", arr)

		}
	}

	return arr
}

func anySliceToStringSlice(arrI []interface{}) []string {
	arr := make([]string, len(arrI))
	for i := range arrI {
		arr[i] = arrI[i].(string)
	}
	return arr
}

func stringSliceToAnySlice(arr []string) []interface{} {
	arrI := make([]interface{}, len(arr))
	for i := range arr {
		arrI[i] = arr[i]
	}
	return arrI
}

func getIdxFromFileName(name string) int {
	if strings.Contains(name, "-") {
		// xxx-metadata.json
		parts := strings.Split(name, "-")
		idx, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		return idx
	} else {
		// xxx.rm
		idxStr := name[:len(name)-3]
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			panic(err)
		}
		return idx
	}
}
