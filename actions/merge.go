package actions

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/skius/rm-pdf-tools/document"
	"io"
	"strings"
)


// MergeFiles merges the documents stored in fileNames and writes the merged document to outFileName.
// TODO: Decide if this belongs in a different package
func MergeFiles(fileNames []string, uuids []string, outFileName string) {
	pdfDocs := getPdfDocsFromFiles(fileNames, uuids)

	for _, pdfDoc := range pdfDocs {
		fmt.Println(pdfDoc)
	}

	mergedDoc := &document.PdfDocument{}

	// Copy general information, then overwrite more specific information
	mergedDoc.Content = pdfDocs[0].Content

	mergedDoc.Uuid = uuid.New().String()

	totalPageCount := 0
	for _, pdfDoc := range pdfDocs {
		totalPageCount += pdfDoc.Content.PageCount
	}
	mergedDoc.Content.PageCount = totalPageCount

	allPages := make([][]string, totalPageCount)
	for i, pdfDoc := range pdfDocs {
		allPages[i] = pdfDoc.Content.Pages
	}
	mergedDoc.Content.Pages = mergeSlices(allPages)

	allPagedata := make([][]string, totalPageCount)
	for i, pdfDoc := range pdfDocs {
		allPagedata[i] = pdfDoc.Pagedata
	}
	mergedDoc.Pagedata = mergeSlices(allPagedata)

	allPdfs := make([][]byte, len(fileNames))
	for i, pdfDoc := range pdfDocs {
		//w, _ := os.Create(fmt.Sprint("Tempfile", i, ".pdf"))
		//w.Write(pdfDoc.Pdf)
		allPdfs[i] = pdfDoc.Pdf
	}
	mergedDoc.Pdf = mergePdfs(allPdfs)

	// To compute the new .rm filenames, simply keep track of a rolling page sum and run the "<pagesum>b1" action
	// to shift all the names back by <pagesum>.
	rollingPageCount := 0
	mergedDoc.RmFiles = make(map[string][]byte)
	for _, pdfDoc := range pdfDocs {
		rmFiles := pdfDoc.RmFiles
		rmFileNames := make([]string, 0, len(rmFiles))
		for fn, _ := range rmFiles {
			rmFileNames = append(rmFileNames, fn)
		}

		repls := RunLines(rmFileNames, []Action{Insert{
			Count: rollingPageCount,
			PageNo: 1,
			InsertAfter: false,
		}})

		for fn, content := range rmFiles {
			repl := repls[fn]
			newFn := strings.ReplaceAll(fn, fmt.Sprint(repl.OriginalIdx), fmt.Sprint(repl.NewIdx))
			mergedDoc.RmFiles[newFn] = content
		}

		rollingPageCount += pdfDoc.Content.PageCount
	}

	mergedDoc.WriteToFile(outFileName)
}

func getPdfDocsFromFiles(fileNames []string, uuids []string) []document.PdfDocument {
	pdfDocs := make([]document.PdfDocument, len(fileNames))

	for i, fileName := range fileNames {
		pdfDocs[i] = document.FromZipFilePdf(fileName, uuids[i])
	}

	return pdfDocs
}

func mergePdfs(pdfs [][]byte) []byte {
	readers := make([]io.ReadSeeker, len(pdfs))
	for i := range pdfs {
		//fmt.Println(len(pdfs[i]))
		readers[i] = bytes.NewReader(pdfs[i])
	}

	conf := pdfcpu.NewDefaultConfiguration()
	writer := new(bytes.Buffer)

	err := api.Merge(readers, writer, conf)
	if err != nil {
		panic(err)
	}

	return writer.Bytes()
}

func mergeSlices(slices [][]string) []string {
	res := make([]string, 0)
	for _, slice := range slices {
		res = append(res, slice...)
	}
	return res
}

