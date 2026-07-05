# CLAUDE.md - Technical Context

## 🚀 Commands

### Build & Run

- **Build CLI**: `make build` or `go build -o yfin ./cmd/yfin`
- **Install CLI**: `go install ./cmd/yfin`
- **Run CLI**: `./yfin --help` or `go run ./cmd/yfin`

### Test & Lint

- **Unit Tests**: `make test` or `go test ./...`
- **Integration Tests**: `go test -tags=integration ./...`
- **Coverage**: `make test-coverage`
- **Lint**: `make lint` or `golangci-lint run`
- **Format**: `make fmt` or `go fmt ./...`

---

## 🗂️ Project Structure

- `cmd/yfin`: Composition root and entry point for the `yfin` CLI.
- `facade/`: Publicly exported, reflection-free plain Go structs (e.g. `facade.Bar`, `facade.Quote`) for external packages (e.g., `stock`, `data`) to avoid importing `internal/`.
- `internal/`: Core implementation details:
    - `internal/httpx`: Resilient HTTP client with QPS rate limiting, exponential backoff, retry logic, and circuit breaking. (Note: Session rotation has been fully removed).
    - `internal/scrape`: Yahoo web scraping engine complying with robots.txt.
    - `internal/norm`: Data normalization logic (e.g., converting to `ampy-proto` or facade formats).
        - `internal/scrape`: Yahoo web scraping engine complying with robots.txt.
    - `internal/norm`: Data normalization logic (e.g., converting to `ampy-proto` or facade formats).
    - `internal/config`: YAML configuration parsing and structures.
- `svc/`: Specialized stock services (e.g., TWSE).
- `dashboards/`: Grafana dashboards, Prometheus alert rules, and operational runbooks.

---

## 🔧 Code Conventions & Decisions

1. **No Session Rotation**: Session rotation was removed to simplify HTTP connection reuse and state management. The HTTP client relies on a single shared `http.Client` with robust rate-limiting, retries, and circuit breakers.
2. **Facade boundary**: Other projects must import models from `facade/` to avoid using `internal/norm` types and `ScaledDecimal` reflection directly.
3. **Decimals**: Prices/amounts are internally stored as `ScaledDecimal` for precision. Expose them as `float64` via the `facade` package.
4. **Timezones**: All timestamps must be handled in UTC.
