package concepts

// 009. Type Conversion
var Concept009 = Concept{
	Number:      9,
	ID:          "type-conversion",
	Category:    "Core Syntax",
	Name:        "9. Type Conversion",
	Description: "Convert between types",
	Instruction: "Convert the floating-point value 3.7 to an integer and print the result, which should be 3",
	Boilerplate: `package main

import "fmt"

func main() {
	// Your code here
}`,
	Answer: `package main

import "fmt"

func main() {
	f := 3.7
	x := int(f)
	fmt.Println(x)
}`,
	ExpectedOutput: "3",
	Difficulty:     "beginner",
	Explanation:    "Go requires explicit type conversion - there's no automatic type coercion. Use T(v) syntax to convert value v to type T. Be aware that conversions can lose precision (float to int) or fail (string to int requires strconv).",
	Example:        "var i int = 42\nf := float64(i)  // int to float64\nx := int(3.99)   // 3 (truncates)\n// strconv for strings:\nn, _ := strconv.Atoi(\"123\")",
	UseCase:        "Use type conversion when mixing numeric types in expressions, assigning to variables of different types, or when APIs require specific types. Always be mindful of precision loss and range overflow.",
	Prerequisites:  []string{"var-declaration"},
	RelatedTopics:  []string{"strconv-atoi", "strconv-itoa", "type-assertion"},
	DocsURL:        "https://go.dev/tour/basics/13",
}

func init() {
	Register(Concept009)
}
