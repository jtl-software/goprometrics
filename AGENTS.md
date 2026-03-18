# AGENTS.md

Guidance for agentic coding agents operating in this repository.

## Project Overview

**goprometrics** is an HTTP-based Prometheus metrics aggregator for ephemeral processes (e.g. PHP scripts).
It exposes two ports:
- `9111` — REST API (PUT endpoints to push metrics)
- `9112` — Prometheus `/metrics` scrape endpoint

Module name: `goprometrics` (see `go.mod`). Minimum Go version: **1.26**.

---

## Build / Run / Test Commands

```bash
# Build
go build

# Run (with defaults: 0.0.0.0:9111 API, 0.0.0.0:9112 metrics)
./goprometrics

# Run with explicit flags
./goprometrics -host 0.0.0.0 -port 9111 -hostm 0.0.0.0 -portm 9112

# Run all tests
go test ./...

# Run tests for a single package
go test ./src/store/...
go test ./src/api/...

# Run a single test by name (use -run with a regex matching the test function name)
go test ./src/store/... -run TestCanAppendNew
go test ./src/api/...  -run Test_createPrometheusMetricOpts

# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...

# Docker
docker build ./
docker-compose up -d
```

No Makefile or Taskfile exists. CI runs `go test ./...` across Go versions 1.24–1.26 on every push
(see `.github/workflows/gotest.yml`).

---

## Project Structure

```
goprometrics/
├── main.go                 # Entry point, version var injected via ldflags
├── go.mod / go.sum
├── Dockerfile
├── docker-compose.yml
├── src/
│   ├── api/
│   │   ├── config.go       # CLI flag parsing, HostConfig
│   │   ├── http.go         # HTTP adapter, route registration, request handling
│   │   └── http_test.go
│   └── store/
│       ├── store.go        # Store interface + 4 concrete implementations
│       ├── manager.go      # Thread-safe Append() with panic recovery
│       ├── param.go        # MetricOpts and ConstLabel data types
│       ├── store_test.go
│       ├── manager_test.go
│       └── param_test.go
├── prometheus/
│   └── prometheus.yml      # Prometheus scrape config for local dev
├── docs/
│   └── api.yaml            # OpenAPI 3.0 spec
└── examples/
    ├── counter.http
    ├── gauge.http
    ├── histogram.http
    └── summary.http
```

---

## Code Style

### Formatting

- Standard `gofmt` formatting — always run `gofmt` before committing.
- No `.golangci.yml` or linter configuration exists; only the Go toolchain is used.
- No external assertion libraries — use `t.Errorf` and `reflect.DeepEqual` for test assertions.

### Imports

Imports use a single `import (...)` block. Order: stdlib, then third-party, then internal.
Internal packages are imported using the module prefix:

```go
import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "github.com/prometheus/client_golang/prometheus"

    "goprometrics/src/store"
)
```

### Naming Conventions

| Construct | Convention | Example |
|---|---|---|
| Exported types | `PascalCase` | `MetricOpts`, `HostConfig` |
| Unexported types | `camelCase` | `counterStore`, `adapter` |
| Interfaces | Single `PascalCase` word | `Store` |
| Constructors | `NewTypeName()` | `NewCounterStore`, `NewAdapter` |
| Exported functions / methods | `PascalCase` | `Append`, `RequestHandler` |
| Unexported functions | `camelCase` | `parseBuckets`, `handleBadRequestError` |
| Exported struct fields | `PascalCase` | `Ns`, `Name`, `HistogramBuckets` |
| Files | lowercase, no separators | `store.go`, `manager.go` |
| Test files | `<source>_test.go` | `store_test.go` |
| Package-level Prometheus vars | `PascalCase` | `AppendCounter`, `ErrorCounter` |

### Types and Grouping

Group related type definitions with a single `type (...)` block:

```go
type (
    Store interface { ... }

    counterStore  struct { store map[string]*prometheus.CounterVec }
    gaugeStore    struct { store map[string]*prometheus.GaugeVec }
    // ...
)
```

Prefer unexported structs for internal implementations; return them as their interface type from constructors.

---

## Patterns and Architecture

### Store Interface

The `Store` interface (`src/store/store.go`) is the central abstraction:

```go
type Store interface {
    Append(opts MetricOpts)
    Inc(opts MetricOpts, value float64)
    Has(opts MetricOpts) bool
}
```

Four concrete implementations (`counterStore`, `gaugeStore`, `histogramStore`, `summaryStore`) are all unexported
and returned as `Store` from their respective `New*` constructors.

### Thread-Safe Double-Checked Locking

The `Append` function in `manager.go` uses double-checked locking to safely register new metrics:

```go
if !s.Has(opts) {
    mutex.Lock()
    defer mutex.Unlock()
    if !s.Has(opts) {
        s.Append(opts)
    }
}
```

### Adapter Pattern (HTTP Layer)

`src/api/http.go` wraps `gorilla/mux` and `http.Server` in an unexported `adapter` struct, exposing typed
route-registration methods (`CounterHandleFunc`, `SummaryHandleFunc`, etc.) and lifecycle methods.

### Closure-Based Handler Factory

`adapter.RequestHandler(s store.Store)` returns `func(http.ResponseWriter, *http.Request)` — the store is
captured in the closure, binding a concrete store type to a specific HTTP route.

### Self-Monitoring via promauto

Package-level `var` blocks in `http.go` and `manager.go` register Prometheus metrics using `promauto`
(auto-registered on `init`). GoProMetrics monitors itself this way.

---

## Error Handling

- **HTTP boundary errors:** Use `handleBadRequestError(err, w)` — marshals `{"message": "..."}` as JSON with HTTP 400.
- **Unrecoverable startup errors:** Log with `slog.Error(...)` and then terminate explicitly with `os.Exit(1)`.
- **Panic recovery:** `manager.go` uses a `defer` + `recover()` wrapper around calls into the Prometheus library
  (e.g. when a metric is re-registered with conflicting options). The recovered value is converted to an `error`,
  logged via `slog.Error(...)`, the `ErrorCounter` metric is incremented, and the resulting error is returned to
  the caller.
- **Silently ignored errors:** Several non-critical errors are explicitly discarded with `_`
  (e.g. response write errors, form parse errors, invalid bucket values). This is an established pattern here —
  when adding new code, only discard errors that are genuinely non-actionable; log or return others.

---

## Testing Conventions

- Tests use the **same package** as source (white-box testing; no `_test` package suffix).
- **Table-driven tests** are used consistently everywhere:

```go
tests := []struct {
    name    string
    args    args
    wantErr bool
}{
    {name: "valid input", ...},
    {name: "invalid input", wantErr: true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

- Test function naming (mixed, follow the existing file's style):
  - `TestTypeName` for constructor / exported function tests
  - `Test_functionName` for unexported function tests
  - `Test_TypeName_MethodName` for method tests
  - `TestDescriptiveBehavior` for integration-style scenario tests

- Use `reflect.DeepEqual` for struct/slice/map comparison; no assertion libraries.
- Use `testutil.ToFloat64` and `testutil.CollectAndCount` from
  `github.com/prometheus/client_golang/prometheus/testutil` to assert metric values.
- Define test fakes (implementing `Store`) inline in the test file instead of mocks.
- Name the system under test variable `sut` or a short, descriptive name matching the type.

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/gorilla/mux v1.7.4` | HTTP routing |
| `github.com/prometheus/client_golang v1.5.1` | Prometheus metrics SDK |

Use `go get` to add dependencies; keep `go.sum` committed.

---

## Docker / Local Environment

```bash
# Build image
docker build ./

# Start full local stack (goprometrics + Prometheus)
docker-compose up -d

# Ports after compose up:
#   9111  — goprometrics REST API
#   9112  — goprometrics /metrics
#   9090  — Prometheus UI
```

The `Dockerfile` uses a multi-stage build: `golang:1.26-alpine` builder stage, `alpine:3.21` final image.
