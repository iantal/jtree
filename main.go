package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	Name       string    `json:"name,omitempty"`
	Size       int       `json:"value,omitempty"`
	FullPath   string    `json:"path"`
	Color      string    `json:"color"`
	Children   []*Node   `json:"children,omitempty"`
	ParentName string    `json:"-"`
	Info       *FileInfo `json:"-"`
	Parent     *Node     `json:"-"`
}

func isExcluded(name string) bool {
	exclusions := []string{".git", "build/", "bin/", "gradle/", "libs/", ".gradle/", "buildSrc/", ".ci/"}
	for _, ex := range exclusions {
		if strings.HasPrefix(name, ex) {
			return true
		}
	}
	return false
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
		inf := fileInfoFromInterface(info)
		parents[path] = &Node{
			FullPath: path,
			Name:     inf.Name,
			Info:     inf,
			Children: make([]*Node, 0),
		}

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
			// node.Parent = parent
			
			if !node.Info.IsDir {
				node.Size = int(node.Info.Size)
			}

			if node.Name == ".git" {
				fmt.Println(path)
				continue
			}
			parent.Children = append(parent.Children, node)
		}
	}

	return
}

func colorTree(node *Node, intensity int) {
	intensity += 4
	node.Color = fmt.Sprint("hsl(0, 0%, ", 100 - intensity, "%)")

	for _, node := range node.Children {
		colorTree(node, intensity)

	}
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

	colorTree(tree, 0)

	dat, err := json.Marshal(tree)
	handleError(err)

	err = ioutil.WriteFile(opts.OutputFile, dat, 0644)
	handleError(err)
}
