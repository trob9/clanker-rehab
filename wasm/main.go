//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"syscall/js"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

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

	if err := i.Use(stdlib.Symbols); err != nil {
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
