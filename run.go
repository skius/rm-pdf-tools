package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/google/uuid"
	"io"
	"sort"
	"strings"
)

type PageReplacement struct {
	originalIdx int
	newIdx int
	deleted bool
}

func runActionsPdf(pdf io.ReadSeeker, outW io.Writer, actions []Action) {
	sort.Slice(actions, func(i, j int) bool {
		iP := actions[i].Page()
		jP := actions[j].Page()

		return iP < jP
	})
	fmt.Println(actions)

	conf := pdfcpu.NewDefaultConfiguration()

	currReader := pdf


	for i := len(actions)-1; i >= 0; i-- {
		currActionI := actions[i]

		switch currAction := currActionI.(type) {
		case Delete:
			writer := new(bytes.Buffer)
			err := api.RemovePages(currReader, writer, []string{ fmt.Sprintf("%d-%d", currAction.PageNo, currAction.PageNo + currAction.Count - 1) }, conf)
			if err != nil {
				panic(err)
			}
			currReader = bytes.NewReader(writer.Bytes())
		case Insert:
			for c := 0; c < currAction.Count; c++ {
				writer := new(bytes.Buffer)
				err := api.InsertPages(currReader, writer, []string{ fmt.Sprintf("%d", currAction.PageNo) }, !currAction.InsertAfter, conf)
				if err != nil {
					panic(err)
				}
				currReader = bytes.NewReader(writer.Bytes())
			}
		}
	}

	buf, err := io.ReadAll(currReader)
	if err != nil {
		panic(err)
	}

	_, err = outW.Write(buf)
	if err != nil {
		panic(err)
	}
}

// runActionsLines takes a slice of all filenames in the uuid/ directory and computes their respective new index and whether they get deleted or not.
func runActionsLines(files []string, actions []Action) map[string]PageReplacement {
	sort.Slice(actions, func(i, j int) bool {
		iP := actions[i].Page()
		jP := actions[j].Page()

		return iP < jP
	})
	fmt.Println(actions)

	res := make(map[string]PageReplacement)
	for _, f := range files {
		idx := getIdxFromFileName(f)
		res[f] = PageReplacement{originalIdx: idx, newIdx: idx, deleted: false}
	}

	for i := len(actions)-1; i >= 0; i-- {
		currActionI := actions[i]
		switch currAction := currActionI.(type) {
		case Delete:
			pageIdx := currAction.PageNo - 1
			for k, pr := range res {
				if pr.originalIdx >= pageIdx && pr.originalIdx < pageIdx + currAction.Count {
					// This is one of the pages that the user wants to delete
					pr.deleted = true
					res[k] = pr
				} else if pr.originalIdx >= pageIdx + currAction.Count {
					// we need to reduce that page's index
					pr.newIdx -= currAction.Count
					res[k] = pr
				}
			}

		case Insert:
			pageIdx := currAction.PageNo - 1
			if !currAction.InsertAfter {
				// Inserting after the previous page is equivalent to inserting before the page
				pageIdx--
			}
			for k, pr := range res {
				if pr.originalIdx > pageIdx {
					// we need to increase that page's index by Count
					pr.newIdx += currAction.Count
					res[k] = pr
				}
			}
		}
	}

	return res
}

func runActionsPagedata(pagedata string, actions []Action) string {
	sort.Slice(actions, func(i, j int) bool {
		iP := actions[i].Page()
		jP := actions[j].Page()

		return iP < jP
	})
	fmt.Println(actions)

	lines := strings.Split(pagedata, "\n")
	linesI := make([]interface{}, len(lines))
	for i, line := range lines {
		linesI[i] = line
	}
	fmt.Println(linesI)
	linesProcI := _runActionsSlice(linesI, actions, func() interface{} { return "Blank" })
	fmt.Println(linesProcI)
	linesProc := make([]string, len(linesProcI))
	for i, lineProcI := range linesProcI {
		linesProc[i] = lineProcI.(string)
	}

	return strings.Join(linesProc, "\n")
}

func runActionsContent(contentStr string, actions []Action) string {
	sort.Slice(actions, func(i, j int) bool {
		iP := actions[i].Page()
		jP := actions[j].Page()

		return iP < jP
	})
	fmt.Println(actions)

	content := Content{}
	err := json.Unmarshal([]byte(contentStr), &content)
	if err != nil {
		panic(err)
	}

	pageCnt := content.PageCount
	newPageCnt := pageCnt
	for i := len(actions)-1; i >= 0; i-- {
		currActionI := actions[i]
		switch currAction := currActionI.(type) {
		case Delete:
			newPageCnt -= currAction.Count
		case Insert:
			newPageCnt += currAction.Count
		}
	}

	content.PageCount = newPageCnt

	// Leaving out content.Pages seems to fix the issue with empty pages having annotations on them: https://github.com/juruen/rmapi/issues/201

	pages := content.Pages
	for i := range pages {
		pages[i] = uuid.New().String()
	}
	pagesI := make([]interface{}, len(pages))
	for i, p := range pages {
		pagesI[i] = p
	}
	pages = _runActionsSlice(pagesI, actions, func() interface{} {
		randomUuid := uuid.New()
		return randomUuid.String()
	})

	// content.Pages = []string{}
	content.Pages = pages

	res, err := json.Marshal(&content)
	if err != nil {
		panic(err)
	}
	return string(res)
}

func _runActionsSlice(arr []interface{}, actions []Action, defaultCreator func() interface{}) []interface{} {
	sort.Slice(actions, func(i, j int) bool {
		iP := actions[i].Page()
		jP := actions[j].Page()

		return iP < jP
	})
	fmt.Println(actions)

	for i := len(actions)-1; i >= 0; i-- {
		currActionI := actions[i]
		switch currAction := currActionI.(type) {
		case Delete:
			pageIdx := currAction.PageNo - 1
			arr = append(arr[:pageIdx], arr[pageIdx + currAction.Count:]...)
			//fmt.Println(arr)
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
			fmt.Println("arr final", arr)

		}
	}

	return arr
}

// Content represents the top-level UUID.content JSON
type Content struct {
	CoverPageNumber  int `json:"coverPageNumber"`
	DocumentMetadata struct {
	} `json:"documentMetadata"`
	DummyDocument bool `json:"dummyDocument"`
	ExtraMetadata struct {
		LastBallpointv2Color   string `json:"LastBallpointv2Color"`
		LastBallpointv2Size    string `json:"LastBallpointv2Size"`
		LastCalligraphyColor   string `json:"LastCalligraphyColor"`
		LastCalligraphySize    string `json:"LastCalligraphySize"`
		LastEraseSectionColor  string `json:"LastEraseSectionColor"`
		LastEraseSectionSize   string `json:"LastEraseSectionSize"`
		LastEraserColor        string `json:"LastEraserColor"`
		LastEraserSize         string `json:"LastEraserSize"`
		LastEraserTool         string `json:"LastEraserTool"`
		LastFinelinerv2Color   string `json:"LastFinelinerv2Color"`
		LastFinelinerv2Size    string `json:"LastFinelinerv2Size"`
		LastHighlighterv2Color string `json:"LastHighlighterv2Color"`
		LastHighlighterv2Size  string `json:"LastHighlighterv2Size"`
		LastPaintbrushv2Color  string `json:"LastPaintbrushv2Color"`
		LastPaintbrushv2Size   string `json:"LastPaintbrushv2Size"`
		LastPen                string `json:"LastPen"`
		LastPencilv2Color      string `json:"LastPencilv2Color"`
		LastPencilv2Size       string `json:"LastPencilv2Size"`
		LastSelectionToolColor string `json:"LastSelectionToolColor"`
		LastSelectionToolSize  string `json:"LastSelectionToolSize"`
		LastSharpPencilv2Color string `json:"LastSharpPencilv2Color"`
		LastSharpPencilv2Size  string `json:"LastSharpPencilv2Size"`
		LastTool               string `json:"LastTool"`
		LastUndefinedColor     string `json:"LastUndefinedColor"`
		LastUndefinedSize      string `json:"LastUndefinedSize"`
	} `json:"extraMetadata"`
	FileType      string   `json:"fileType"`
	FontName      string   `json:"fontName"`
	LineHeight    int      `json:"lineHeight"`
	Margins       int      `json:"margins"`
	Orientation   string   `json:"orientation"`
	PageCount     int      `json:"pageCount"`
	Pages         []interface{} `json:"pages"`
	TextAlignment string   `json:"textAlignment"`
	TextScale     int      `json:"textScale"`
}
