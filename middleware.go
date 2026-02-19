package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// scannerUAs are known scanner/bot user agent substrings (lowercased).
var scannerUAs = []string{
	"nikto", "sqlmap", "masscan", "nmap", "zgrab", "nuclei",
	"gobuster", "dirb", "dirbuster", "wfuzz", "ffuf", "hydra",
	"metasploit", "acunetix", "nessus", "openvas", "skipfish",
	"appscan", "webinspect", "libwww-perl", "scrapy",
	"semrushbot", "ahrefsbot", "mj12bot", "dotbot", "petalbot",
	"bytespider", "gptbot",
}

// scannerPaths are paths commonly probed by scanners and exploit scripts.
var scannerPaths = []string{
	"/.env", "/.git/", "/wp-admin", "/wp-login.php", "/wp-content/",
	"/phpmyadmin", "/.htaccess", "/config.php", "/backup",
	"/.ds_store", "/etc/passwd", "/cgi-bin/", "/xmlrpc.php",
	"/actuator", "/.aws/", "/.ssh/", "/shell.php", "/cmd.php",
	"/eval.php", "/.bash_history", "/id_rsa", "/console",
	"/jmx-console", "/manager/html", "/solr/", "/jenkins",
	"/.well-known/security.txt",
}

// unusualMethods are HTTP methods beyond standard REST operations.
var unusualMethods = map[string]bool{
	"TRACE": true, "CONNECT": true, "DEBUG": true,
	"PROPFIND": true, "PROPPATCH": true, "MKCOL": true,
	"COPY": true, "MOVE": true, "LOCK": true, "UNLOCK": true,
	"SEARCH": true,
}

// clientIPMiddleware extracts the real client IP from X-Forwarded-For (set by Caddy).
func clientIPMiddleware(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i > 0 {
		return addr[:i]
	}
	return addr
}

// accessLog is an HTTP middleware that logs every request as structured JSON.
// It detects scanner patterns and flags suspicious requests for easy monitoring.
func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		latency := time.Since(start)

		ip := clientIPMiddleware(r)
		ua := r.UserAgent()
		path := r.URL.Path
		method := r.Method

		flags := detectFlags(method, path, ua, rw.status)

		entry := map[string]interface{}{
			"ts":         time.Now().UTC().Format(time.RFC3339),
			"method":     method,
			"path":       path,
			"status":     rw.status,
			"latency_ms": latency.Milliseconds(),
			"ip":         ip,
			"ua":         ua,
			"size":       rw.size,
		}
		if len(flags) > 0 {
			entry["flags"] = flags
		}
		if ref := r.Referer(); ref != "" {
			entry["referer"] = ref
		}

		b, _ := json.Marshal(entry)
		fmt.Println(string(b))
	})
}

func detectFlags(method, path, ua string, status int) []string {
	var flags []string

	uaLower := strings.ToLower(ua)
	for _, s := range scannerUAs {
		if strings.Contains(uaLower, s) {
			flags = append(flags, "scanner_ua")
			break
		}
	}

	pathLower := strings.ToLower(path)
	for _, p := range scannerPaths {
		if strings.HasPrefix(pathLower, p) || strings.Contains(pathLower, p) {
			flags = append(flags, "scanner_path")
			break
		}
	}

	if unusualMethods[strings.ToUpper(method)] {
		flags = append(flags, "unusual_method")
	}

	if status >= 400 && status < 500 {
		flags = append(flags, "client_error")
	}
	if status >= 500 {
		flags = append(flags, "server_error")
	}
	return flags
}
