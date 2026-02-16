package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go-concept-trainer/concepts"
)

type Concept struct {
	Number         int        `json:"number"`
	ID             string     `json:"id"`
	Category       string     `json:"category"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Instruction    string     `json:"instruction"`
	Boilerplate    string     `json:"boilerplate"`
	Answer         string     `json:"answer"`
	ExpectedOutput string     `json:"expectedOutput"`
	TestCases      []TestCase `json:"testCases,omitempty"`
	Difficulty     string     `json:"difficulty"`
	Explanation    string     `json:"explanation"`
	Example        string     `json:"example"`
	UseCase        string     `json:"useCase"`
	Prerequisites  []string   `json:"prerequisites"`
	RelatedTopics  []string   `json:"relatedTopics"`
	DocsURL        string     `json:"docsUrl"`
}

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type RunRequest struct {
	Code      string `json:"code"`
	ConceptID string `json:"conceptId"`
}

type RunResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error"`
}

func getConcepts() []Concept {
	// Convert from concepts package format to main package format
	pkgConcepts := concepts.GetAll()
	result := make([]Concept, len(pkgConcepts))
	for i, c := range pkgConcepts {
		result[i] = Concept{
			Number:         c.Number,
			ID:             c.ID,
			Category:       c.Category,
			Name:           c.Name,
			Description:    c.Description,
			Instruction:    c.Instruction,
			Boilerplate:    c.Boilerplate,
			Answer:         c.Answer,
			ExpectedOutput: c.ExpectedOutput,
			TestCases:      convertTestCases(c.TestCases),
			Difficulty:     c.Difficulty,
			Explanation:    c.Explanation,
			Example:        c.Example,
			UseCase:        c.UseCase,
			Prerequisites:  c.Prerequisites,
			RelatedTopics:  c.RelatedTopics,
			DocsURL:        c.DocsURL,
		}
	}
	return result
}

func convertTestCases(tcs []concepts.TestCase) []TestCase {
	result := make([]TestCase, len(tcs))
	for i, tc := range tcs {
		result[i] = TestCase{
			Input:    tc.Input,
			Expected: tc.Expected,
		}
	}
	return result
}

func main() {
	allConcepts := getConcepts()
	fmt.Printf("Loaded %d concepts from concepts package\n", len(allConcepts))

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api/concepts", serveConcepts)
	http.HandleFunc("/api/run", serveRun)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func serveConcepts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getConcepts())
}

func serveRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	allConcepts := getConcepts()
	var concept *Concept
	for i := range allConcepts {
		if allConcepts[i].ID == req.ConceptID {
			concept = &allConcepts[i]
			break
		}
	}

	if concept == nil {
		http.Error(w, "Concept not found", http.StatusNotFound)
		return
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", req.Code, parser.ImportsOnly)
	if err != nil {
		sendError(w, "Compilation Error", fmt.Errorf("failed to parse code: %v", err))
		return
	}

	forbidden := map[string]bool{
		"os/exec":   true,
		"net":       true,
		"net/http":  true,
		"syscall":   true,
		"unsafe":    true,
		"os/signal": true,
	}

	for _, s := range f.Imports {
		path := strings.Trim(s.Path.Value, "\"")
		if forbidden[path] {
			sendError(w, "Security Violation", fmt.Errorf("package '%s' is not allowed", path))
			return
		}
	}

	tmpDir, err := os.MkdirTemp("", "gorun-*")
	if err != nil {
		sendError(w, "Failed to create temp directory", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	mainPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(req.Code), 0644); err != nil {
		sendError(w, "Failed to write code", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Dir = tmpDir
	output, execErr := cmd.CombinedOutput()

	outputStr := strings.TrimSpace(string(output))
	expected := strings.TrimSpace(concept.ExpectedOutput)

	success := execErr == nil && outputStr == expected

	resp := RunResponse{
		Success: success,
		Output:  outputStr,
	}

	if execErr != nil {
		resp.Error = execErr.Error()
	} else if !success {
		resp.Error = fmt.Sprintf("Expected: %q, Got: %q", expected, outputStr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, msg string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(RunResponse{
		Success: false,
		Error:   fmt.Sprintf("%s: %v", msg, err),
	})
}
