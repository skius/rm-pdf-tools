package main

import (
	"fmt"
	"github.com/juruen/rmapi/model"
	"github.com/skius/rm-pdf-tools/actions"
	"github.com/skius/rm-pdf-tools/cloud"
	"os"
	"sort"
)

const remoteWorkDir = "/pdf-tools/"
const remoteWatchDir = remoteWorkDir + "work/"
const remoteMergeDir = remoteWorkDir + "merge/"
const remoteOriginalDir = remoteWorkDir + "original/"
const remoteProcessedDir = remoteWorkDir + "processed/"

func main() {
	c, err := cloud.New()
	if err != nil {
		panic(err)
	}

	docsToEdit := c.FindNewFilesEdit(remoteWatchDir)
	if len(docsToEdit) == 0 {
		fmt.Println("No docs to edit found!")
	} else {
		for _, f := range docsToEdit {
			processDoc(c, f)
		}
	}

	docsToMerge := c.FindNewFilesMerge(remoteMergeDir)
	if len(docsToMerge) == 0 {
		fmt.Println("No docs to merge found!")
	} else {
		mergeDocs(c, docsToMerge)
	}
}

// mergeDocs merges the given documents and uploads the resulting document (merge order is alphabetical in their names).
func mergeDocs(c *cloud.Cloud, nodes []*model.Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name() < nodes[j].Name()
	})
	mkFileName := func(i int, uuid string) string { return fmt.Sprintf("doc-%d-%s.zip", i, uuid) }

	fileNamesToMerge := make([]string, len(nodes))
	uuids := make([]string, len(nodes))
	for i, node := range nodes {
		fn := mkFileName(i, node.Id())
		err := c.Download(node, fn)
		if err != nil {
			panic(err)
		}
		fileNamesToMerge[i] = fn
		uuids[i] = node.Id()
	}

	outDocName := "merged"
	outFileName := outDocName + ".zip"
	actions.MergeFiles(fileNamesToMerge, uuids, outFileName)

	_, err := c.Upload(outFileName, remoteProcessedDir)
	if err != nil {
		panic(err)
	}
	for _, node := range nodes {
		_, err = c.Move(node, remoteOriginalDir, node.Name())
		if err != nil {
			panic(err)
		}
	}

	err = os.Remove(outFileName)
	if err != nil {
		panic(err)
	}

	for _, fn := range fileNamesToMerge {
		err = os.Remove(fn)
		if err != nil {
			panic(err)
		}
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

	actions.RunFile(node.Id(), fileNameOriginal, fileNameProcessed, acts)

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

