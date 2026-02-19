package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
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

// ipLimiter is a simple fixed-window per-IP rate limiter.
type ipLimiter struct {
	mu      sync.Mutex
	windows map[string]*ipWindow
}

type ipWindow struct {
	count int
	reset time.Time
}

const (
	rateLimit  = 120           // requests per window
	rateWindow = time.Minute   // window duration
	maxBodyLen = 1 << 20       // 1 MB
)

func newIPLimiter() *ipLimiter {
	l := &ipLimiter{windows: make(map[string]*ipWindow)}
	go func() {
		for range time.Tick(5 * time.Minute) {
			l.mu.Lock()
			now := time.Now()
			for ip, w := range l.windows {
				if now.After(w.reset) {
					delete(l.windows, ip)
				}
			}
			l.mu.Unlock()
		}
	}()
	return l
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	w, ok := l.windows[ip]
	if !ok || now.After(w.reset) {
		l.windows[ip] = &ipWindow{count: 1, reset: now.Add(rateWindow)}
		return true
	}
	if w.count >= rateLimit {
		return false
	}
	w.count++
	return true
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; worker-src 'self'; base-uri 'self'; form-action 'self'")
		next.ServeHTTP(w, r)
	})
}

func rateLimitMiddleware(limiter *ipLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use X-Forwarded-For to get the real client IP behind Caddy.
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
		} else if idx := strings.Index(ip, ","); idx != -1 {
			ip = strings.TrimSpace(ip[:idx])
		}
		if !limiter.allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxBodyLen {
			http.Error(w, "Request Entity Too Large", http.StatusRequestEntityTooLarge)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyLen)
		next.ServeHTTP(w, r)
	})
}

func main() {
	allConcepts := getConcepts()
	fmt.Printf("Loaded %d concepts from concepts package\n", len(allConcepts))

	conceptsJSON, err := json.Marshal(allConcepts)
	if err != nil {
		log.Fatalf("Failed to marshal concepts: %v", err)
	}

	indexTmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	limiter := newIPLimiter()

	mux := http.NewServeMux()

	// Go 1.22+ method+path routing: non-matching methods get 405 automatically.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		if err := indexTmpl.Execute(w, nil); err != nil {
			log.Printf("Template execute error: %v", err)
		}
	})
	mux.HandleFunc("GET /api/concepts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(conceptsJSON)
	})
	mux.HandleFunc("POST /api/log-run", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ExitCode    int `json:"exit_code"`
			DurationMs  int `json:"duration_ms"`
			OutputBytes int `json:"output_bytes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		entry := map[string]interface{}{
			"ts":           time.Now().UTC().Format(time.RFC3339),
			"event":        "code_execute",
			"lang":         "go",
			"exit_code":    body.ExitCode,
			"duration_ms":  body.DurationMs,
			"output_bytes": body.OutputBytes,
		}
		b, _ := json.Marshal(entry)
		fmt.Println(string(b))
		w.WriteHeader(http.StatusNoContent)
	})
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || r.URL.Path[len(r.URL.Path)-1] == '/' {
			http.NotFound(w, r)
			return
		}
		http.FileServer(http.Dir("static")).ServeHTTP(w, r)
	})))

	handler := accessLog(securityHeaders(rateLimitMiddleware(limiter, requestLimiter(mux))))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
