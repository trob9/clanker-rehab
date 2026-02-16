// Web Worker for Yaegi WASM Go code execution
let wasmReady = false;
let pendingResolve = null;

// Called by WASM main() when Yaegi is initialized
self.onWasmReady = function () {
    wasmReady = true;
    self.postMessage({ type: 'ready' });
};

// Load wasm_exec.js (Go WASM support) and instantiate the WASM binary
importScripts('/static/wasm_exec.js');

const go = new Go();
WebAssembly.instantiateStreaming(fetch('/static/yaegi.wasm'), go.importObject)
    .then(wasmResult => {
        go.run(wasmResult.instance);
    })
    .catch(err => {
        self.postMessage({ type: 'error', error: 'Failed to load WASM: ' + err.message });
    });

// Handle messages from main thread
self.onmessage = function (e) {
    if (e.data.type === 'run') {
        if (!wasmReady) {
            self.postMessage({ type: 'result', error: 'WASM interpreter not ready yet' });
            return;
        }

        try {
            const jsonResult = self.runGoCode(e.data.code);
            const result = JSON.parse(jsonResult);
            self.postMessage({ type: 'result', output: result.output || '', error: result.error || '' });
        } catch (err) {
            self.postMessage({ type: 'result', output: '', error: 'Execution error: ' + err.message });
        }
    }
};
