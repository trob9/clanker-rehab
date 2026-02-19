# Stage 1: Build Yaegi WASM binary
FROM golang:1.22-alpine AS wasm-builder
WORKDIR /wasm
COPY wasm/go.mod wasm/go.sum* ./
RUN go mod download
COPY wasm/ ./
RUN GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o yaegi.wasm .
RUN cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" wasm_exec.js

# Stage 2: Build Go server binary
FROM golang:1.25-alpine AS server-builder
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY *.go ./
COPY concepts/ ./concepts/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server .

# Stage 3: Minimal runtime
FROM alpine:3.21
RUN adduser -D -u 1000 appuser
WORKDIR /app

COPY --from=server-builder /app/server .
COPY templates/ ./templates/
COPY static/ ./static/
COPY --from=wasm-builder /wasm/yaegi.wasm ./static/yaegi.wasm
COPY --from=wasm-builder /wasm/wasm_exec.js ./static/wasm_exec.js

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080
CMD ["./server"]
