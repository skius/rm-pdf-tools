package cloud

import (
	"errors"
	"fmt"
	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
)

// Cloud is used to interact with the remarkable cloud.
type Cloud struct {
	api *api.ApiCtx
}

// New creates a new authenticated Cloud using rmapi's authentication.
func New() (*Cloud, error) {
	var rm *api.ApiCtx
	var err error

	// needed for rmapi
	log.InitLog()

	rm, err = api.CreateApiCtx(api.AuthHttpCtx(true, false))
	if err != nil {
		panic(err)
	}

	if rm.Filetree == nil {
		return nil, errors.New("failed to build remarkable documents tree")
	}

	return &Cloud{api: rm}, nil
}

// Download downloads node to the file dst.
func (r *Cloud) Download(node *model.Node, dst string) error {
	return r.api.FetchDocument(node.Id(), dst)
}

// Upload uploads the file src to folder dstPath in the cloud.
func (r *Cloud) Upload(src string, dstPath string) (*model.Document, error) {
	dstNode, err := r.api.Filetree.NodeByPath(dstPath, r.api.Filetree.Root())
	if err != nil {
		return nil, err
	}
	doc, err := r.api.UploadDocument(dstNode.Id(), src)
	if err != nil {
		return nil, err
	}
	r.api.Filetree.AddDocument(*doc)
	return doc, nil
}

// Move moves node to `dstPath/dstName`.
func (r *Cloud) Move(node *model.Node, dstPath, dstName string) (*model.Node, error) {
	dstNode, err := r.api.Filetree.NodeByPath(dstPath, r.api.Filetree.Root())
	if err != nil {
		return nil, err
	}
	return r.api.MoveEntry(node, dstNode, dstName)
}

// FindFile finds a file in the cloud by path and returns the associated Node.
func (r *Cloud) FindFile(path string) (*model.Node, error) {
	return r.api.Filetree.NodeByPath(path, r.api.Filetree.Root())
}

// FindNewFilesEdit provides all files in subdirectories of the provided directory.
// e.g. FindNewFilesEdit("pdf-tools") -> ["pdf-tools/sub1/file1", "pdf-tools/sub2/file2"]
func (r *Cloud) FindNewFilesEdit(dir string) []*model.Node {
	files := make([]*model.Node, 0)

	dirNode, err := r.api.Filetree.NodeByPath(dir, r.api.Filetree.Root())
	if err != nil {
		panic(err)
	}

	for uuid, node := range dirNode.Children {
		fmt.Println("UUID:", uuid, "name:", node.Name(),"node:", node)
		for _, child := range node.Children {
			fmt.Println("Inside", node.Name(), "child:", child.Name())
			files = append(files, child)
		}
		fmt.Println()
	}

	return files
}

// FindNewFilesMerge returns all files in the provided directory.
func (r *Cloud) FindNewFilesMerge(dir string) []*model.Node {
	files := make([]*model.Node, 0)

	dirNode, err := r.api.Filetree.NodeByPath(dir, r.api.Filetree.Root())
	if err != nil {
		return nil
	}

	for _, node := range dirNode.Children {
		files = append(files, node)
	}

	return files
}
