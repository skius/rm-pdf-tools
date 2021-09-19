package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"io"
	"sort"
	"strings"
)

func Sort(actions []Action) {
	sort.Slice(actions, func(i, j int) bool {
		iP := actions[i].Page()
		jP := actions[j].Page()

		return iP < jP
	})
}

// RunPdf takes a PDF as input and writes the resulting PDF after applying actions to outW.
func RunPdf(pdf io.ReadSeeker, outW io.Writer, actions []Action) {
	Sort(actions)

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

// RunLines takes a slice of all filenames in the uuid/ directory and computes their respective new index and whether they get deleted or not.
func RunLines(files []string, actions []Action) map[string]PageReplacement {
	Sort(actions)

	res := make(map[string]PageReplacement)
	for _, f := range files {
		idx := getIdxFromFileName(f)
		res[f] = PageReplacement{OriginalIdx: idx, NewIdx: idx, Deleted: false}
	}

	for i := len(actions)-1; i >= 0; i-- {
		currActionI := actions[i]
		switch currAction := currActionI.(type) {
		case Delete:
			pageIdx := currAction.PageNo - 1
			for k, pr := range res {
				if pr.OriginalIdx >= pageIdx && pr.OriginalIdx < pageIdx + currAction.Count {
					// This is one of the pages that the user wants to delete
					pr.Deleted = true
					res[k] = pr
				} else if pr.OriginalIdx >= pageIdx + currAction.Count {
					// we need to reduce that page's index
					pr.NewIdx -= currAction.Count
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
				if pr.OriginalIdx > pageIdx {
					// we need to increase that page's index by Count
					pr.NewIdx += currAction.Count
					res[k] = pr
				}
			}
		}
	}

	return res
}

// RunPagedata takes pagedata and returns the pagedata after applying actions.
func RunPagedata(pagedata string, actions []Action) string {
	Sort(actions)

	lines := strings.Split(pagedata, "\n")
	linesI := stringSliceToAnySlice(lines)
	linesProcI := runSlice(linesI, actions, func() interface{} { return "Blank" })
	linesProc := anySliceToStringSlice(linesProcI)

	return strings.Join(linesProc, "\n")
}

// RunContent takes a content JSON string and returns the content JSON string after applying actions.
func RunContent(contentStr string, actions []Action) string {
	Sort(actions)

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
	pages := content.Pages
	// Keeping old page UUIDs for now
	//for i := range pages {
	//	pages[i] = uuid.New().String()
	//}
	pagesI := stringSliceToAnySlice(pages)
	pagesProcI := runSlice(pagesI, actions, func() interface{} {
		randomUuid := uuid.New()
		return randomUuid.String()
	})
	pagesProc := anySliceToStringSlice(pagesProcI)


	//content.Pages = []string{}
	content.Pages = pagesProc

	res, err := json.Marshal(&content)
	if err != nil {
		panic(err)
	}
	return string(res)
}
