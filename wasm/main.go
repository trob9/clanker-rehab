//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"syscall/js"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// safePackages is an explicit whitelist of stdlib packages user code may import.
// Dangerous packages (os, net, syscall, unsafe, os/exec, plugin, runtime) are
// intentionally excluded as defence-in-depth on top of the WASM sandbox.
var safePackages = map[string]bool{
	"bufio":         true,
	"bytes":         true,
	"context":       true,
	"crypto/sha256": true,
	"encoding/json": true,
	"errors":        true,
	"fmt":           true,
	"io":            true,
	"log":           true,
	"math":          true,
	"math/bits":     true,
	"math/rand":     true,
	"reflect":       true,
	"regexp":        true,
	"sort":          true,
	"strconv":       true,
	"strings":       true,
	"sync":          true,
	"sync/atomic":   true,
	"time":          true,
	"unicode":       true,
	"unicode/utf8":  true,
}

// safeSymbols returns a filtered interp.Exports containing only whitelisted packages.
func safeSymbols() interp.Exports {
	filtered := make(interp.Exports)
	for pkg, symbols := range stdlib.Symbols {
		if safePackages[pkg] {
			filtered[pkg] = symbols
		}
	}
	return filtered
}

type result struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

func main() {
	js.Global().Set("runGoCode", js.FuncOf(runGoCode))
	js.Global().Call("eval", "if(typeof onWasmReady==='function'){onWasmReady()}")
	<-make(chan struct{})
}

func runGoCode(this js.Value, args []js.Value) interface{} {
	code := args[0].String()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	i := interp.New(interp.Options{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err := i.Use(safeSymbols()); err != nil {
		return marshal(result{Error: "failed to load stdlib: " + err.Error()})
	}

	_, err := i.Eval(code)

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	r := result{Output: output}
	if err != nil {
		r.Error = err.Error()
	}

	return marshal(r)
}

func marshal(r result) string {
	b, _ := json.Marshal(r)
	return string(b)
}
