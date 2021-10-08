package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/projectdiscovery/gologger"
)

// FileInfo is a struct created from os.FileInfo interface for serialization.
type FileInfo struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"mod_time"`
	IsDir   bool        `json:"is_dir"`
}

// Helper function to create a local FileInfo struct from os.FileInfo interface.
func fileInfoFromInterface(v os.FileInfo) *FileInfo {
	return &FileInfo{v.Name(), v.Size(), v.Mode(), v.ModTime(), v.IsDir()}
}

// Node represents a node in a directory tree.
type Node struct {
	FullPath   string    `json:"path"`
	Name       string    `json:"name"`
	Size       int       `json:"size"`
	Children   []*Node   `json:"children"`
	ParentName string    `json:"parent"`
	Info       *FileInfo `json:"-"`
	Parent     *Node     `json:"-"`
}

// Create directory hierarchy.
func NewTree(root string) (result *Node, err error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return
	}

	parents := make(map[string]*Node)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// if !isExcluded(info.Name(), excusions) {
		inf := fileInfoFromInterface(info)
		parents[path] = &Node{
			FullPath: path,
			Name:     inf.Name,
			Info:     inf,
			Children: make([]*Node, 0),
		}
		// }
		return nil
	}
	if err = filepath.Walk(absRoot, walkFunc); err != nil {
		return
	}
	for path, node := range parents {
		parentPath := filepath.Dir(path)
		parent, exists := parents[parentPath]

		if !exists { // If a parent does not exist, this is the root.
			result = node
		} else {
			node.Parent = parent
			node.ParentName = parent.Name
			parent.Children = append(parent.Children, node)

			if !node.Info.IsDir {
				node.Size = int(node.Info.Size)
			}
		}
	}
	return
}

type Options struct {
	Repository string `long:"repository" description:"full path of the repository" default:"Unknown"`
	OutputFile string `long:"o" description:"full path of the output file" default:"Unknown"`
}

func handleError(err error) {
	if err != nil {
		gologger.Error().Msg(err.Error())
		os.Exit(1)
	}
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	tree, err := NewTree(opts.Repository)
	handleError(err)

	dat, err := json.Marshal(tree)
	handleError(err)

	err = ioutil.WriteFile(opts.OutputFile, dat, 0644)
	handleError(err)
}
