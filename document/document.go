package document

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"io"
	"os"
	"strings"
)

//go:embed empty.pdf
var emptyPdf []byte

type Document struct {
	Uuid string
	Content Content
	Pagedata []string
	RmFiles map[string][]byte
}

func (doc Document) String() string {
	fileNames := make([]string, 0)
	for fn, _ := range doc.RmFiles {
		fileNames = append(fileNames, fn)
	}
	return fmt.Sprintf("{uuid: %s, content: %v, pagedata: %s, rmfiles: %s}", doc.Uuid, doc.Content, doc.Pagedata, fileNames)
}

type PdfDocument struct {
	Document
	Pdf []byte
}

func (pdfDoc PdfDocument) String() string {
	return pdfDoc.Document.String()
}

// Content represents the top-level UUID.content JSON
type Content struct {
	CoverPageNumber  int `json:"coverPageNumber"`
	DocumentMetadata map[string]interface{} `json:"documentMetadata"`
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
	Pages         []string `json:"pages"`
	TextAlignment string   `json:"textAlignment"`
	TextScale     int      `json:"textScale"`
}

func (c Content) String() string {
	buf, _ := json.Marshal(c)
	return string(buf)
}

// ToPdfDoc converts the Document into a PdfDocument by creating the appropriate number of blank PDF pages.
func (doc Document) ToPdfDoc() PdfDocument {
	pdfDoc := PdfDocument{}
	pdfDoc.Document = doc
	pdfDoc.Content.FileType = "pdf"

	pageCount := pdfDoc.Content.PageCount
	// Insert pageCount - 1 empty pages
	reader := bytes.NewReader(emptyPdf)
	conf := pdfcpu.NewDefaultConfiguration()
	for c := 0; c < pageCount - 1; c++ {
		writer := new(bytes.Buffer)
		err := api.InsertPages(reader, writer, []string{"1"}, true, conf)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(writer.Bytes())
	}

	filledPdf, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	pdfDoc.Pdf = filledPdf

	return pdfDoc
}

// FromZipFilePdf deserializes a .zip'd PdfDocument from the .zip file.
func FromZipFilePdf(fileName, uuid string) PdfDocument {
	r, err := zip.OpenReader(fileName)
	if err != nil {
		panic(err)
	}

	pdfDoc := FromZipReaderPdf(r, uuid)

	err = r.Close()
	if err != nil {
		panic(err)
	}

	return pdfDoc
}

// FromZipReaderPdf deserializes a .zip'd PdfDocument from the .zip's Reader.
func FromZipReaderPdf(reader *zip.ReadCloser, uuid string) PdfDocument {
	pdfDoc := PdfDocument{}
	pdfDoc.Uuid = uuid
	pdfDoc.RmFiles = make(map[string][]byte)
	pdfDoc.Pagedata = make([]string, 0)
	for _, f := range reader.File {
		fr, err := f.Open()
		if err != nil {
			panic(err)
		}
		// 36 character UUID
		suffix := f.Name[36:]
		switch suffix {
		case ".content":
			pdfDoc.Content = getContentFromReader(fr)
		case ".pagedata":
			pdfDoc.Pagedata = getPagedataFromReader(fr)
		case ".pdf":
			pdfDoc.Pdf = getBytesFromReader(fr)
		default:
			// Must be inner file:
			fmt.Println("Inner file", f.Name)
			rmFileName := f.Name[37:]
			buf := getBytesFromReader(fr)
			pdfDoc.RmFiles[rmFileName] = buf
		}
		err = fr.Close()
		if err != nil {
			panic(err)
		}
	}

	if pdfDoc.Content.FileType != "pdf" {
		fmt.Println("Filetype is not PDF! PDF content is:", pdfDoc.Pdf)
		pdfDoc = pdfDoc.ToPdfDoc()
	}

	return pdfDoc
}

// WriteToFile writes the document as .zip to the given file.
func (pdfDoc PdfDocument) WriteToFile(fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	w := zip.NewWriter(file)

	writeToZip(w, pdfDoc.Uuid + ".pdf", pdfDoc.Pdf)

	contentData, err := json.Marshal(pdfDoc.Content)
	if err != nil {
		panic(err)
	}
	writeToZip(w, pdfDoc.Uuid + ".content", contentData)

	// Add newline
	pagedata := strings.Join(pdfDoc.Pagedata, "\n") + "\n"
	writeToZip(w, pdfDoc.Uuid + ".pagedata", []byte(pagedata))

	for fn, data := range pdfDoc.RmFiles {
		writeToZip(w, pdfDoc.Uuid + "/" + fn, data)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
}
