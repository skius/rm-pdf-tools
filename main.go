package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/juruen/rmapi/model"
	"github.com/skius/rm-pdf-tools/actions"
	"github.com/skius/rm-pdf-tools/cloud"
	"io"
	"os"
	"strconv"
	"strings"
)

const remoteWorkDir = "/pdf-tools/"
const remoteWatchDir = remoteWorkDir + "work/"
const remoteOriginalDir = remoteWorkDir + "original/"
const remoteProcessedDir = remoteWorkDir + "processed/"

func main() {
	c, err := cloud.New()
	if err != nil {
		panic(err)
	}

	files := c.FindNewFiles(remoteWatchDir)
	if len(files) == 0 {
		fmt.Println("No files to process found! Exiting...")
		os.Exit(0)
	}
	for _, f := range files {
		processDoc(c, f)
	}
}

// processDoc extracts actions, runs them, and uploads the new document for the given node.
func processDoc(c *cloud.Cloud, node *model.Node) {
	fmt.Println("Processing file:", node.Name())
	docName := node.Name()
	fileNameOriginal := docName + "_original.zip"
	docNameProcessed := docName + "_processed"
	fileNameProcessed := docNameProcessed + ".zip"

	acts := actions.FromString(node.Parent.Name())

	err := c.Download(node, fileNameOriginal)
	if err != nil {
		panic(err)
	}

	processFile(node.Id(), fileNameOriginal, fileNameProcessed, acts)

	_, err = c.Upload(fileNameProcessed, remoteProcessedDir)
	if err != nil {
		panic(err)
	}
	processed, err := c.FindFile(remoteProcessedDir + docNameProcessed)
	if err != nil {
		panic(err)
	}
	_, err = c.Move(processed, remoteProcessedDir, docName)
	if err != nil {
		panic(err)
	}

	_, err = c.Move(node, remoteOriginalDir, docName)
	if err != nil {
		panic(err)
	}

	err = os.Remove(fileNameOriginal)
	if err != nil {
		panic(err)
	}
	err = os.Remove(fileNameProcessed)
	if err != nil {
		panic(err)
	}
}

// processFile processes fileNameOriginal to fileNameProcessed after applying acts.
func processFile(uuidOriginal, fileNameOriginal, fileNameProcessed string, acts actions.T) {
	r, err := zip.OpenReader(fileNameOriginal)
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create(fileNameProcessed)
	if err != nil {
		panic(err)
	}
	w := zip.NewWriter(outFile)

	// Use a fresh UUID to avoid collisions when uploading the document
	uuidNew := uuid.New().String()
	innerFiles := []*zip.File{}
	innerFilesStrs := []string{}

	for _, f := range r.File {
		// Handle inner files (annotations)
		if strings.Contains(f.Name, "/") {
			innerFiles = append(innerFiles, f)
			innerFilesStrs = append(innerFilesStrs, f.FileInfo().Name())
			continue
		}

		// Handle files in top-level (should only be uuid.pagedata, uuid.content and uuid.pdf)
		newName := strings.ReplaceAll(f.Name, uuidOriginal, uuidNew)
		fw, err := w.Create(newName)
		if err != nil {
			panic(err)
		}

		fb := new(bytes.Buffer)
		rc, err := f.Open()
		_, err = io.Copy(fb, rc)
		if err != nil {
			panic(err)
		}
		err = rc.Close()
		if err != nil {
			panic(err)
		}

		var data []byte

		if strings.HasSuffix(f.Name, ".content") {
			newContent := actions.RunContent(string(fb.Bytes()), acts)
			data = []byte(newContent)
		} else if strings.HasSuffix(f.Name, ".pagedata") {
			newPagedata := actions.RunPagedata(string(fb.Bytes()), acts)
			data = []byte(newPagedata)
		} else if strings.HasSuffix(f.Name, ".pdf") {
			reader := bytes.NewReader(fb.Bytes())
			buf := new(bytes.Buffer)

			actions.RunPdf(reader, buf, acts)
			data = buf.Bytes()
		} else {
			panic("unexpected file: " + f.Name)
		}

		_, err = fw.Write(data)
		if err != nil {
			panic(err)
		}
	}

	// Handle all files in "uuid/*"
	repl := actions.RunLines(innerFilesStrs, acts)
	for _, f := range innerFiles {
		innerName := f.FileInfo().Name()
		pr := repl[innerName]

		fmt.Println("Processing replacement for:", innerName, "orig:", pr.OriginalIdx, "new:", pr.NewIdx, "deleted:", pr.Deleted)
		if pr.Deleted {
			continue
		}

		newName := uuidNew + "/" + strings.ReplaceAll(innerName, strconv.Itoa(pr.OriginalIdx), strconv.Itoa(pr.NewIdx))
		fw, err := w.Create(newName)
		if err != nil {
			panic(err)
		}

		fb := new(bytes.Buffer)
		rc, err := f.Open()
		_, err = io.Copy(fb, rc)
		if err != nil {
			panic(err)
		}
		err = rc.Close()
		if err != nil {
			panic(err)
		}

		_, err = fw.Write(fb.Bytes())
		if err != nil {
			panic(err)
		}
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}
	err = outFile.Close()
	if err != nil {
		panic(err)
	}
	err = r.Close()
	if err != nil {
		panic(err)
	}
}


