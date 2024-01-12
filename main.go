package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

type FileStructure struct {
	Path  string
	Width int
	// TODO: Extract properly instead of just *Node
	// i.e. function name, args, import identifier and string, etc.
	Package    []*sitter.Node
	Imports    []*sitter.Node
	Constants  []*sitter.Node
	Variables  []*sitter.Node
	Structs    []*sitter.Node
	Interfaces []*sitter.Node
	Functions  []*sitter.Node
	Methods    []*sitter.Node
}

func (f FileStructure) String() string {
	return fmt.Sprintf("%s\n"+
		"  Package    : %d\n"+
		"  Imports    : %d\n"+
		"  Constants  : %d\n"+
		"  Variables  : %d\n"+
		"  Interfaces : %d\n"+
		"  Structs    : %d\n"+
		"  Functions  : %d\n"+
		"  Methods    : %d\n",
		f.Path,
		len(f.Package),
		len(f.Imports),
		len(f.Constants),
		len(f.Variables),
		len(f.Interfaces),
		len(f.Structs),
		len(f.Functions),
		len(f.Methods))
}

func (f FileStructure) PrintPackage(data []byte) string {
	line := ""
	if len(f.Package) == 1 {
		line += fmt.Sprintf("// --- Package ---\n%s\n", f.Package[0].Content(data))
	} else {
		if len(f.Package) == 0 {
			line += fmt.Sprintf("// --- Package declaration not found\n")
		} else {
			line += fmt.Sprintf("// --- Packages ---\n")
			for _, n := range f.Package {
				name := fmt.Sprintf("package %s\n", n.Content(data))
				line += name
			}
		}
	}
	return line
}

func (f FileStructure) PrintNodes(title string, nodes []*sitter.Node, select_fn func(*sitter.Node) *sitter.Node, data []byte) string {
	if len(nodes) == 0 {
		return ""
	}
	line := fmt.Sprintf("// --- %s ---\n", title)
	for _, n := range nodes {
		line += fmt.Sprintf("%s\n", select_fn(n).Content(data))
	}
	return line
}

func (f FileStructure) PrintNestedNodes(title string, nodes []*sitter.Node, print_block bool, select_fn func(*sitter.Node) *sitter.Node, data []byte) string {
	if len(nodes) == 0 {
		return ""
	}
	line := fmt.Sprintf("// --- %s ---\n", title)
	for _, n := range nodes {
		lline := ""
		max := int(n.NamedChildCount() - 1)
		if print_block {
			max = int(n.NamedChildCount())
		}
		for ci := 0; ci < max; ci++ {
			c := n.NamedChild(ci)
			lline += select_fn(c).Content(data) + " "
		}
		line += fmt.Sprintf("%s\n", lline)
	}
	return line
}

func getNode(n *sitter.Node) *sitter.Node {
	return n
}

func getFirstChild(n *sitter.Node) *sitter.Node {
	return n.Child(0)
}

func getFirstNamedChild(n *sitter.Node) *sitter.Node {
	return n.NamedChild(0)
}

func (f FileStructure) Print(format string, n *sitter.Node, lang *sitter.Language, data []byte) string {
	// Special syntax for path
	lines := ""
	if string(format[0]) == "/" {
		lines += fmt.Sprintf("// %s\n", f.Path)
	}

	// Rest are just first letter, except interface which is n
	for _, s := range format {
		switch string(s) {
		case "p", "P":
			f.Package = NodesFromQuery([]byte(queryPackage), n, lang, data)
			lines += f.PrintPackage(data)
		case "i", "I":
			f.Imports = NodesFromQuery([]byte(queryImports), n, lang, data)
			lines += f.PrintNodes("Imports", f.Imports, getNode, data)
		case "c", "C":
			f.Constants = NodesFromQuery([]byte(queryConstants), n, lang, data)
			lines += f.PrintNodes("Constants", f.Constants, getFirstNamedChild, data)
		case "v", "V":
			f.Variables = NodesFromQuery([]byte(queryVariables), n, lang, data)
			lines += f.PrintNodes("Variables", f.Variables, getFirstNamedChild, data)
		case "n":
			f.Interfaces = NodesFromQuery([]byte(queryTypeInterfaces), n, lang, data)
			lines += f.PrintNestedNodes("Interfaces", f.Interfaces, false, getNode, data)
		case "N":
			f.Interfaces = NodesFromQuery([]byte(queryTypeInterfaces), n, lang, data)
			lines += f.PrintNestedNodes("Interfaces", f.Interfaces, true, getNode, data)
		case "s":
			f.Structs = NodesFromQuery([]byte(queryTypeStructs), n, lang, data)
			lines += f.PrintNestedNodes("Structs", f.Structs, false, getNode, data)
		case "S":
			f.Structs = NodesFromQuery([]byte(queryTypeStructs), n, lang, data)
			lines += f.PrintNestedNodes("Structs", f.Structs, true, getNode, data)
		case "m":
			f.Methods = NodesFromQuery([]byte(queryMethods), n, lang, data)
			lines += f.PrintNestedNodes("Methods", f.Methods, false, getNode, data)
		case "M":
			f.Methods = NodesFromQuery([]byte(queryMethods), n, lang, data)
			lines += f.PrintNestedNodes("Methods", f.Methods, true, getNode, data)
		case "f":
			f.Functions = NodesFromQuery([]byte(queryFunctions), n, lang, data)
			lines += f.PrintNestedNodes("Functions", f.Functions, false, getNode, data)
		case "F":
			f.Functions = NodesFromQuery([]byte(queryFunctions), n, lang, data)
			lines += f.PrintNestedNodes("Functions", f.Functions, true, getNode, data)
		}
	}

	return lines
}

// GoFiles recursively finds all .go files in a given directory.
func GoFiles(path string) ([]string, error) {
	goFiles := []string{}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	return goFiles, err
}

// NodesFromQuery runs query string and returns matching nodes
func NodesFromQuery(query []byte, root *sitter.Node, lang *sitter.Language, data []byte) []*sitter.Node {
	// Execute the query
	q, err := sitter.NewQuery(query, lang)
	if err != nil {
		log.Fatalln("Invalid query:", string(query), err)
	}
	qc := sitter.NewQueryCursor()
	qc.Exec(q, root)

	// Iterate over query results and assert only one match
	nodes := []*sitter.Node{}
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		m = qc.FilterPredicates(m, data)
		for _, c := range m.Captures {
			nodes = append(nodes, c.Node)
		}
	}
	return nodes
}

func HighlightOutput(text string) (string, error) {
	tmp, err := os.CreateTemp("/tmp", "temp*.go")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())

	_, err = tmp.WriteString(text)
	if err != nil {
		return "", err
	}
	err = tmp.Close()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("tree-sitter", "highlight", tmp.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), err
}

func main() {
	// Define boolean flags for each component
	path := flag.String("path", "", "Path to .go file")
	format := flag.String("format", "picvnsmf", "Apis to print")
	nohighlight := flag.Bool("nohighlight", false, "Highlight syntax with tree-sitter")

	// Parse the flags
	flag.Parse()

	// Read entire file
	f := FileStructure{Path: *path}
	data, err := os.ReadFile(*path)
	if err != nil {
		log.Fatalln("FAILED os.ReadFile: ", path, err)
	}

	// Parse with tree-sitter
	ctx := context.Background()
	lang := golang.GetLanguage()
	n, err := sitter.ParseCtx(ctx, data, lang)
	if err != nil {
		log.Fatalln("FAILED sitter.ParseCtx: ", path, err)
	}

	// Highlight with tree-sitter if needed
	if *nohighlight {
		fmt.Println(f.Print(*format, n, lang, data))
	} else {
		text := f.Print(*format, n, lang, data)
		text, err = HighlightOutput(text)
		if err != nil {
			log.Println("// WARNING Highlight failed: ", err)
		}
		fmt.Println(text)
	}
}
