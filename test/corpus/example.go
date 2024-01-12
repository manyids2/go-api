// --- Package ---

// Corpus to test go-api documentation tool
package main

// --- Imports ---
// import "log"
import (
	// Format
	ff "fmt"
	"log" // Log
)

// Constants with comments, iota
const (
	// Const A
	ca = iota
	cb = 2 // Const B
)

// Constants on separate lines
const cc = 5

// --- Variables ---

// Variables with comments
var (
	// Variable A
	va = 5
	vb string // Variable B
)

// Variables on separate lines
var vc = 5

var vd []string // With comment on side

// --- Types ---

// Struct type
type StructA struct {
	// Integer A
	A int
	b float32 // Float b
}

// Stacked
type (
	tickMsg  struct{}
	frameMsg struct{}
)

// Interface type
type Interface interface {
	// Method A
	MA(a, b int, c string) (StructA, error)
	MB(a, b int, c string) (s StructA, e error) // Method B
}

// --- Functions ---

// SumAB
func SumAB(a, b int) int {
	// Local variables, constants
	var c string
	const d = ""
	ff.Println(c, d)
	return a + b
}

// Method
func (*StructA) MA(a, b int, c string) int {
	return a + b
}

// Main function
func main() {
	ff.Println("hello")
	log.Println("hello")
}
