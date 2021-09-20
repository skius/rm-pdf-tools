package document

import (
	"archive/zip"
	"encoding/json"
	"io"
	"strings"
)

func getContentFromReader(r io.ReadCloser) Content {
	buf := getBytesFromReader(r)
	content := Content{}
	err := json.Unmarshal(buf, &content)
	if err != nil {
		panic(err)
	}
	return content
}

func getPagedataFromReader(r io.ReadCloser) []string {
	buf := getBytesFromReader(r)
	pagedataRaw := string(buf)
	return strings.Split(pagedataRaw[:len(pagedataRaw) - 1], "\n")
}

func getBytesFromReader(r io.ReadCloser) []byte {
	buf, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return buf
}

func writeToZip(w *zip.Writer, fileName string, data []byte) {
	fw, err := w.Create(fileName)
	if err != nil {
		panic(err)
	}
	_, err = fw.Write(data)
	if err != nil {
		panic(err)
	}
}
