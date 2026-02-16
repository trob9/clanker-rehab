package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

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

func getConcepts() []Concept {
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

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; worker-src 'self'")
		next.ServeHTTP(w, r)
	})
}

func requestLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
		next.ServeHTTP(w, r)
	})
}

func main() {
	allConcepts := getConcepts()
	fmt.Printf("Loaded %d concepts from concepts package\n", len(allConcepts))

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/api/concepts", serveConcepts)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	handler := securityHeaders(requestLimiter(mux))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func serveConcepts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getConcepts())
}
