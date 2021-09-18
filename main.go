package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/juruen/rmapi/model"
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
		processFile(c, f)
	}
}

func processFile(c *cloud.Cloud, node *model.Node) {
	fmt.Println("Processing file:", node.Name())
	docName := node.Name()
	fileNameOriginal := docName + "_original.zip"
	docNameProcessed := docName + "_processed"
	fileNameProcessed := docNameProcessed + ".zip"

	actions := ActionsFromString(node.Parent.Name())


	err := c.Download(node, fileNameOriginal)
	if err != nil {
		panic(err)
	}

	r, err := zip.OpenReader(fileNameOriginal)
	if err != nil {
		panic(err)
	}
	//bw := new(bytes.Buffer)
	outFile, err := os.Create(fileNameProcessed)
	if err != nil {
		panic(err)
	}

	w := zip.NewWriter(outFile)

	newUuid := uuid.New().String()
	innerFiles := []*zip.File{}
	innerFilesStr := []string{}


	for _, f := range r.File {
		//fmt.Println("In zip we have:", f.FileInfo().Name())
		fmt.Println("In zip we have:", f.Name)


		if strings.Contains(f.Name, "/") {
			innerFiles = append(innerFiles, f)
			innerFilesStr = append(innerFilesStr, f.FileInfo().Name())
			continue
		}

		// Handle files in top-level (should only be uuid.pagedata and uuid.content and uuid.pdf)

		newName := strings.ReplaceAll(f.Name, node.Id(), newUuid)
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
			newContent := runActionsContent(string(fb.Bytes()), actions)
			data = []byte(newContent)
		} else if strings.HasSuffix(f.Name, ".pagedata") {
			newPagedata := runActionsPagedata(string(fb.Bytes()), actions)
			data = []byte(newPagedata)
		} else if strings.HasSuffix(f.Name, ".pdf") {
			reader := bytes.NewReader(fb.Bytes())
			buf := new(bytes.Buffer)

			runActionsPdf(reader, buf, actions)
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
	repl := runActionsLines(innerFilesStr, actions)
	for _, f := range innerFiles {
		innerName := f.FileInfo().Name()
		pr := repl[innerName]

		fmt.Println("Processing replacement for:", innerName, "orig:", pr.originalIdx, "new:", pr.newIdx, "deleted:", pr.deleted)
		if pr.deleted {
			continue
		}

		newName := newUuid + "/" + strings.ReplaceAll(innerName, strconv.Itoa(pr.originalIdx), strconv.Itoa(pr.newIdx))
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
