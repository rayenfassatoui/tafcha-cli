# Tafcha: CLI-First Plain-Text Publishing Service

## TL;DR

> **Quick Summary**: Build a Go CLI that pipes stdin to an API and returns a short URL. Opening the URL returns the exact plaintext. Default 3-day expiry with configurable duration.
>
> **Deliverables**:
> - `tafcha` CLI binary (stdin → URL)
> - `tafcha-server` API binary (REST API + PostgreSQL)
> - Docker deployment configuration
> - Comprehensive test suite (TDD)
>
> **Estimated Effort**: Large (2-3 weeks for production-ready)
> **Parallel Execution**: YES - 4 waves
> **Critical Path**: Project Setup → Storage Layer → API Core → CLI → Docker

---

## Context

### Original Request
Build a CLI-first plain-text publishing service called Tafcha. User pipes text via `echo "text" | tafcha` and receives a short, unguessable URL. Opening the URL returns exact plaintext. Implemented in Go with PostgreSQL (Neon) storage.

### Interview Summary
**Key Discussions**:
- **Build targets**: Two separate binaries (`tafcha` CLI, `tafcha-server` API)
- **Configuration**: Environment variables only (12-factor app)
- **Migrations**: Embedded SQL via Go embed directive
- **Deployment**: Docker container with reverse proxy for HTTPS
- **Testing**: TDD approach with unit + integration tests
- **Observability**: slog JSON structured logging
- **Cleanup**: Internal goroutine for expired snippet deletion

**Research Findings**:
- **Cobra**: Best CLI framework for Go (used by Docker, K8s)
- **chi router**: Lightweight, built-in middleware (Logger, Recoverer, RequestID)
- **httprate**: Per-IP rate limiting that integrates with chi
- **pgx/v5**: High-performance PostgreSQL driver with connection pooling
- **go-nanoid**: CSPRNG-based short ID generation

### Metis Review (Self-Analysis)
**Identified Gaps** (addressed in plan):
1. **ID collision handling**: Need retry logic with bounded attempts
2. **Content encoding**: UTF-8 validation vs raw bytes handling
3. **Cleanup interval**: Need configurable sweep frequency
4. **CLI timeout default**: 10s may be too short for large payloads
5. **Rate limit storage**: In-memory vs Redis (single instance = in-memory OK)
6. **Exit code consistency**: Must map all error types to correct codes
7. **Database connection retry**: Neon cold starts may need retry logic

---

## Work Objectives

### Core Objective
Deliver a production-ready CLI + API system where users can publish plaintext snippets via stdin and retrieve them via short URLs.

### Concrete Deliverables
- `cmd/tafcha/main.go` - CLI binary entry point
- `cmd/tafcha-server/main.go` - API server entry point
- `internal/api/` - HTTP handlers and middleware
- `internal/storage/` - PostgreSQL repository
- `internal/id/` - Secure ID generation
- `internal/config/` - Environment configuration
- `migrations/` - Embedded SQL schema
- `Dockerfile` - Multi-stage build for both binaries
- `docker-compose.yml` - Local development setup

### Definition of Done
- [ ] `echo "hello" | go run ./cmd/tafcha` returns valid URL (exit 0)
- [ ] `curl <URL>` returns exact "hello" text
- [ ] `go test ./...` passes with >80% coverage
- [ ] `docker build .` succeeds
- [ ] Rate limiting triggers 429 after threshold
- [ ] Expired snippets return 404

### Must Have
- Pipe-only stdin input (reject interactive)
- Expiry flag parsing (10m, 12h, 3d formats)
- Cryptographically secure short IDs (72+ bits entropy)
- Rate limiting per IP (POST: 30/min, GET: 300/min)
- Structured JSON logging
- Health check endpoints (/healthz, /readyz)
- Proper HTTP status codes and error responses

### Must NOT Have (Guardrails)
- **NO user authentication** - snippets are public, unguessable URLs are the security
- **NO editing/updating** - snippets are immutable
- **NO listing/browsing** - prevents enumeration
- **NO syntax highlighting** - pure text/plain only
- **NO binary uploads** - text only, reject non-UTF-8 or treat as raw bytes
- **NO file parameters** - stdin only, no `tafcha file.txt`
- **NO premature abstraction** - keep interfaces minimal
- **NO over-validation** - simple checks, trust the PRD limits
- **NO excessive comments** - self-documenting Go code

---

## Verification Strategy (MANDATORY)

### Test Decision
- **Infrastructure exists**: NO (greenfield project)
- **User wants tests**: TDD
- **Framework**: Go standard `testing` package + testify assertions

### TDD Workflow

Each TODO follows RED-GREEN-REFACTOR:

**Task Structure:**
1. **RED**: Write failing test first
   - Test file: `*_test.go` next to implementation
   - Test command: `go test ./internal/...`
   - Expected: FAIL (test exists, implementation doesn't)
2. **GREEN**: Implement minimum code to pass
   - Command: `go test ./internal/...`
   - Expected: PASS
3. **REFACTOR**: Clean up while keeping green
   - Command: `go test ./internal/...`
   - Expected: PASS (still)

### Test Infrastructure Setup (Task 0)
- [ ] 0. Initialize Go module and test infrastructure
  - Init: `go mod init github.com/rayen/tafcha` (or chosen module path)
  - Deps: `go get github.com/stretchr/testify`
  - Verify: `go test ./...` → 0 tests, no errors
  - Structure: Create placeholder test file

### Automated Verification Approach

| Type | Verification Tool | Procedure |
|------|------------------|-----------|
| API endpoints | curl via Bash | Send requests, validate JSON/text responses |
| CLI behavior | Bash pipe commands | `echo "x" \| ./tafcha`, check stdout/exit code |
| Database ops | go test + testcontainers | Real PostgreSQL in container |
| Rate limiting | curl loop | Exceed threshold, expect 429 |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately):
├── Task 1: Project scaffolding + go.mod
├── Task 2: Configuration module
└── Task 3: ID generation module

Wave 2 (After Wave 1):
├── Task 4: Database migrations + storage layer
├── Task 5: Expiry parsing module
└── Task 6: API router skeleton

Wave 3 (After Wave 2):
├── Task 7: Create snippet handler (POST /api/v1/snippets)
├── Task 8: Retrieve snippet handler (GET /{id})
├── Task 9: Health check handlers
└── Task 10: Rate limiting middleware

Wave 4 (After Wave 3):
├── Task 11: Expiry cleanup goroutine
├── Task 12: CLI implementation
├── Task 13: Docker configuration
└── Task 14: Integration tests

Critical Path: 1 → 4 → 7 → 12 → 14
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2,3,4,5,6 | None (must be first) |
| 2 | 1 | 4,6,7,11 | 3, 5 |
| 3 | 1 | 7 | 2, 5 |
| 4 | 1, 2 | 7, 8, 11 | 5, 6 |
| 5 | 1 | 7, 12 | 2, 3, 4, 6 |
| 6 | 1, 2 | 7, 8, 9, 10 | 4, 5 |
| 7 | 3, 4, 5, 6 | 12, 14 | 8, 9, 10 |
| 8 | 4, 6 | 12, 14 | 7, 9, 10 |
| 9 | 6 | 14 | 7, 8, 10, 11 |
| 10 | 6 | 14 | 7, 8, 9, 11 |
| 11 | 2, 4 | 14 | 9, 10 |
| 12 | 7, 8 | 14 | 13 |
| 13 | 1 | 14 | 12 |
| 14 | All | None | None (final) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Dispatch |
|------|-------|---------------------|
| 1 | 1 | Sequential (foundation) |
| 1 | 2, 3 | delegate_task(category="quick", parallel) |
| 2 | 4, 5, 6 | delegate_task(category="unspecified-high", parallel) |
| 3 | 7, 8, 9, 10, 11 | delegate_task(category="unspecified-high", parallel) |
| 4 | 12, 13, 14 | Sequential (integration) |

---

## TODOs

### Wave 1: Foundation

- [ ] 1. Project Scaffolding

  **What to do**:
  - Create directory structure: `cmd/tafcha/`, `cmd/tafcha-server/`, `internal/`, `migrations/`
  - Initialize Go module: `go mod init tafcha`
  - Add core dependencies: cobra, chi, pgx, nanoid, testify
  - Create placeholder main.go files
  - Create .gitignore for Go projects

  **Must NOT do**:
  - Don't add unnecessary directories (no pkg/, no vendor/ initially)
  - Don't add linting/CI config yet

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple file/directory creation, straightforward setup
  - **Skills**: []
    - No special skills needed for scaffolding
  - **Skills Evaluated but Omitted**:
    - `git-master`: Not needed until commit phase

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential (must be first)
  - **Blocks**: Tasks 2, 3, 4, 5, 6
  - **Blocked By**: None

  **References**:
  - External: https://github.com/golang-standards/project-layout - Standard Go layout
  - External: https://go.dev/doc/modules/layout - Official module layout

  **Acceptance Criteria**:

  ```bash
  # Agent runs:
  ls cmd/tafcha/main.go cmd/tafcha-server/main.go internal/ migrations/
  # Assert: All paths exist
  
  go mod tidy
  # Assert: Exit code 0, go.sum created
  
  go build ./...
  # Assert: Exit code 0 (compiles, even if binaries do nothing)
  ```

  **Evidence to Capture:**
  - [ ] `go mod tidy` output showing dependencies resolved
  - [ ] Directory listing showing structure

  **Commit**: YES
  - Message: `chore: initialize project structure and dependencies`
  - Files: `go.mod`, `go.sum`, `cmd/`, `internal/`, `migrations/`, `.gitignore`
  - Pre-commit: `go build ./...`

---

- [ ] 2. Configuration Module

  **What to do**:
  - Create `internal/config/config.go`
  - Define Config struct with all env vars:
    ```go
    type Config struct {
        DatabaseURL     string        // DATABASE_URL
        Port            int           // PORT (default: 8080)
        BaseURL         string        // BASE_URL (default: http://localhost:8080)
        MaxContentSize  int64         // MAX_CONTENT_SIZE (default: 1MB)
        DefaultExpiry   time.Duration // DEFAULT_EXPIRY (default: 72h)
        MinExpiry       time.Duration // MIN_EXPIRY (default: 10m)
        MaxExpiry       time.Duration // MAX_EXPIRY (default: 720h/30d)
        CleanupInterval time.Duration // CLEANUP_INTERVAL (default: 1h)
        RateLimitCreate int           // RATE_LIMIT_CREATE (default: 30/min)
        RateLimitRead   int           // RATE_LIMIT_READ (default: 300/min)
    }
    ```
  - Implement `Load() (*Config, error)` function
  - Validate required fields (DATABASE_URL must be set)
  - Write tests first (TDD)

  **Must NOT do**:
  - Don't use viper or other config libraries (env vars only)
  - Don't read from files

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single module, well-defined interface
  - **Skills**: []
    - Standard Go patterns
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 3, 5)
  - **Blocks**: Tasks 4, 6, 7, 11
  - **Blocked By**: Task 1

  **References**:
  - External: https://pkg.go.dev/os#Getenv - os.Getenv for env vars
  - External: https://pkg.go.dev/strconv - strconv.Atoi for parsing

  **Acceptance Criteria**:

  **TDD - RED:**
  - [ ] Test file: `internal/config/config_test.go`
  - [ ] Test: `TestLoad_RequiresDatabaseURL` - missing DATABASE_URL returns error
  - [ ] Test: `TestLoad_DefaultValues` - unset optional vars use defaults
  - [ ] `go test ./internal/config` → FAIL (no implementation)

  **TDD - GREEN:**
  - [ ] Implement `Load()` function
  - [ ] `go test ./internal/config` → PASS

  ```bash
  # Agent runs:
  DATABASE_URL=postgres://test go test ./internal/config -v
  # Assert: All tests pass
  
  go test ./internal/config -v 2>&1 | grep -c "PASS"
  # Assert: Output >= 2 (at least 2 test cases)
  ```

  **Commit**: YES
  - Message: `feat(config): add environment configuration loader`
  - Files: `internal/config/config.go`, `internal/config/config_test.go`
  - Pre-commit: `go test ./internal/config`

---

- [ ] 3. ID Generation Module

  **What to do**:
  - Create `internal/id/generator.go`
  - Implement `Generate() (string, error)` using go-nanoid
  - Use URL-safe alphabet (A-Za-z0-9)
  - Generate 12-character IDs (≈71 bits entropy with base62)
  - Ensure CSPRNG usage (crypto/rand via nanoid)
  - Write tests first (TDD)

  **Must NOT do**:
  - Don't use sequential IDs
  - Don't use timestamps or hashes alone
  - Don't make IDs shorter than 10 chars

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single function, well-defined output
  - **Skills**: []
    - Standard Go + nanoid library
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 5)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:
  - External: https://github.com/matoous/go-nanoid - go-nanoid v2 docs
  - External: https://pkg.go.dev/github.com/matoous/go-nanoid/v2 - API reference

  **Acceptance Criteria**:

  **TDD - RED:**
  - [ ] Test file: `internal/id/generator_test.go`
  - [ ] Test: `TestGenerate_ReturnsNonEmpty` - ID is not empty
  - [ ] Test: `TestGenerate_CorrectLength` - ID is 12 chars
  - [ ] Test: `TestGenerate_URLSafe` - ID matches `^[A-Za-z0-9]+$`
  - [ ] Test: `TestGenerate_Unique` - 1000 IDs are all unique
  - [ ] `go test ./internal/id` → FAIL

  **TDD - GREEN:**
  - [ ] `go test ./internal/id` → PASS

  ```bash
  # Agent runs:
  go test ./internal/id -v -run TestGenerate
  # Assert: All 4 tests pass
  
  go test ./internal/id -bench=. 2>&1 | head -5
  # Assert: Benchmark output shows reasonable ns/op
  ```

  **Commit**: YES
  - Message: `feat(id): add secure short ID generator using nanoid`
  - Files: `internal/id/generator.go`, `internal/id/generator_test.go`
  - Pre-commit: `go test ./internal/id`

---

### Wave 2: Core Infrastructure

- [ ] 4. Database Migrations + Storage Layer

  **What to do**:
  - Create `migrations/001_create_snippets.sql`:
    ```sql
    CREATE TABLE IF NOT EXISTS snippets (
        id VARCHAR(20) PRIMARY KEY,
        content BYTEA NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        expires_at TIMESTAMPTZ NOT NULL
    );
    CREATE INDEX idx_snippets_expires_at ON snippets(expires_at);
    ```
  - Embed migration using `//go:embed migrations/*.sql`
  - Create `internal/storage/postgres.go`:
    - `type Repository struct { pool *pgxpool.Pool }`
    - `New(ctx, databaseURL) (*Repository, error)` - creates pool
    - `Create(ctx, id, content, expiresAt) error`
    - `Get(ctx, id) (*Snippet, error)` - returns nil if not found/expired
    - `DeleteExpired(ctx) (int64, error)` - deletes and returns count
    - `Close()`
    - `Ping(ctx) error` - for health checks
  - Run migration on `New()` call
  - Write tests with testcontainers-go (real PostgreSQL)

  **Must NOT do**:
  - Don't use ORM (raw SQL with pgx)
  - Don't return expired snippets from Get()
  - Don't log content (security)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Database layer is critical path, needs careful implementation
  - **Skills**: []
    - pgx patterns from research
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 5, 6)
  - **Blocks**: Tasks 7, 8, 11
  - **Blocked By**: Tasks 1, 2

  **References**:
  - PRD Section 7: Data Model & Storage
  - Research: pgxpool connection management pattern
  - External: https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool - Pool config
  - External: https://golang.testcontainers.org/modules/postgres/ - Test containers

  **Acceptance Criteria**:

  **TDD - RED:**
  - [ ] Test file: `internal/storage/postgres_test.go`
  - [ ] Test: `TestCreate_StoresSnippet`
  - [ ] Test: `TestGet_ReturnsSnippet`
  - [ ] Test: `TestGet_ReturnsNilForExpired`
  - [ ] Test: `TestGet_ReturnsNilForNotFound`
  - [ ] Test: `TestDeleteExpired_RemovesOldSnippets`
  - [ ] `go test ./internal/storage` → FAIL

  **TDD - GREEN:**
  - [ ] `go test ./internal/storage` → PASS

  ```bash
  # Agent runs (requires Docker for testcontainers):
  go test ./internal/storage -v -timeout 60s
  # Assert: All tests pass
  
  # Verify migration embedded:
  go build -o /dev/null ./internal/storage 2>&1
  # Assert: No embed errors
  ```

  **Commit**: YES
  - Message: `feat(storage): add PostgreSQL repository with embedded migrations`
  - Files: `migrations/001_create_snippets.sql`, `internal/storage/postgres.go`, `internal/storage/postgres_test.go`, `internal/storage/types.go`
  - Pre-commit: `go test ./internal/storage`

---

- [ ] 5. Expiry Parsing Module

  **What to do**:
  - Create `internal/expiry/parser.go`
  - Implement `Parse(input string) (time.Duration, error)`:
    - Accepts: `10m`, `12h`, `3d`, `30d`
    - Rejects: empty, negative, invalid format
  - Implement `Validate(d time.Duration, min, max time.Duration) error`:
    - Returns descriptive error if out of bounds
  - Write tests first (TDD)

  **Must NOT do**:
  - Don't accept seconds (security: too granular)
  - Don't parse complex formats like "1d12h" (keep simple)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Pure function, well-defined input/output
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 6)
  - **Blocks**: Tasks 7, 12
  - **Blocked By**: Task 1

  **References**:
  - PRD Section 5.2: CLI Flags - expiry formats

  **Acceptance Criteria**:

  **TDD - RED:**
  - [ ] Test: `TestParse_Minutes` - "30m" → 30*time.Minute
  - [ ] Test: `TestParse_Hours` - "12h" → 12*time.Hour
  - [ ] Test: `TestParse_Days` - "3d" → 72*time.Hour
  - [ ] Test: `TestParse_InvalidFormat` - "abc" → error
  - [ ] Test: `TestValidate_InBounds` - 1h with min=10m,max=30d → nil
  - [ ] Test: `TestValidate_BelowMin` - 5m with min=10m → error
  - [ ] Test: `TestValidate_AboveMax` - 60d with max=30d → error

  **TDD - GREEN:**
  - [ ] `go test ./internal/expiry` → PASS

  ```bash
  go test ./internal/expiry -v
  # Assert: 7+ tests pass
  ```

  **Commit**: YES
  - Message: `feat(expiry): add duration parser for m/h/d formats`
  - Files: `internal/expiry/parser.go`, `internal/expiry/parser_test.go`
  - Pre-commit: `go test ./internal/expiry`

---

- [ ] 6. API Router Skeleton

  **What to do**:
  - Create `internal/api/server.go`:
    - `type Server struct { router chi.Router; repo *storage.Repository; config *config.Config }`
    - `New(config, repo) *Server`
    - `ServeHTTP(w, r)` - implements http.Handler
    - Setup middleware: RequestID, RealIP, Logger, Recoverer, Timeout
  - Create `internal/api/middleware.go`:
    - `ContentType(allowed ...string)` middleware for POST validation
    - `MaxBodySize(n int64)` middleware using http.MaxBytesReader
    - `SecurityHeaders()` middleware for X-Content-Type-Options, Cache-Control
  - Wire up routes (handlers as stubs returning 501):
    - `POST /api/v1/snippets`
    - `GET /{id}`
    - `GET /healthz`
    - `GET /readyz`
  - Write basic router tests

  **Must NOT do**:
  - Don't implement handlers yet (stubs only)
  - Don't add rate limiting yet (separate task)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Router setup affects all API work
  - **Skills**: []
    - chi patterns from research
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Not applicable (API only)

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5)
  - **Blocks**: Tasks 7, 8, 9, 10
  - **Blocked By**: Tasks 1, 2

  **References**:
  - Research: chi middleware stack example
  - External: https://pkg.go.dev/github.com/go-chi/chi/v5/middleware

  **Acceptance Criteria**:

  **TDD - RED:**
  - [ ] Test: `TestRoutes_Exist` - all routes respond (even if 501)
  - [ ] Test: `TestMiddleware_AddsRequestID` - response has X-Request-ID
  - [ ] Test: `TestMiddleware_MaxBodySize` - large body returns 413

  **TDD - GREEN:**
  - [ ] `go test ./internal/api` → PASS

  ```bash
  go test ./internal/api -v -run TestRoutes
  # Assert: Routes exist test passes
  ```

  **Commit**: YES
  - Message: `feat(api): add chi router skeleton with middleware`
  - Files: `internal/api/server.go`, `internal/api/middleware.go`, `internal/api/server_test.go`
  - Pre-commit: `go test ./internal/api`

---

### Wave 3: API Handlers

- [ ] 7. Create Snippet Handler

  **What to do**:
  - Implement `POST /api/v1/snippets` handler in `internal/api/handlers.go`:
    - Validate Content-Type: text/plain
    - Read body (respect MaxBytesReader)
    - Reject empty body → 400 EMPTY_BODY
    - Parse X-Expiry-Seconds header (optional)
    - Validate expiry bounds → 400 EXPIRY_OUT_OF_RANGE
    - Generate ID using id.Generate()
    - Store in database with retry on collision (max 3 attempts)
    - Return 201 with JSON: `{id, url, expires_at}`
  - Create `internal/api/errors.go`:
    - Define error codes: EMPTY_BODY, INVALID_EXPIRY, EXPIRY_OUT_OF_RANGE, PAYLOAD_TOO_LARGE
    - `WriteError(w, status, code, message)` helper

  **Must NOT do**:
  - Don't log snippet content
  - Don't accept content-type other than text/plain
  - Don't process body if Content-Length exceeds limit

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Core business logic, critical path
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 10, 11)
  - **Blocks**: Tasks 12, 14
  - **Blocked By**: Tasks 3, 4, 5, 6

  **References**:
  - PRD Section 6.1.1: Create Snippet endpoint spec
  - PRD Section 9.3: Error cases for create

  **Acceptance Criteria**:

  **TDD - RED:**
  - [ ] Test: `TestCreateSnippet_Success` - valid request returns 201
  - [ ] Test: `TestCreateSnippet_EmptyBody` - returns 400 EMPTY_BODY
  - [ ] Test: `TestCreateSnippet_WrongContentType` - returns 415
  - [ ] Test: `TestCreateSnippet_ExpiryTooShort` - returns 400
  - [ ] Test: `TestCreateSnippet_ExpiryTooLong` - returns 400
  - [ ] Test: `TestCreateSnippet_LargeBody` - returns 413

  **TDD - GREEN:**
  - [ ] `go test ./internal/api -run TestCreateSnippet` → PASS

  ```bash
  go test ./internal/api -v -run TestCreateSnippet
  # Assert: All 6 tests pass
  ```

  **Commit**: YES
  - Message: `feat(api): implement create snippet endpoint`
  - Files: `internal/api/handlers.go`, `internal/api/errors.go`, `internal/api/handlers_test.go`
  - Pre-commit: `go test ./internal/api`

---

- [ ] 8. Retrieve Snippet Handler

  **What to do**:
  - Implement `GET /{id}` handler:
    - Extract ID from URL path
    - Validate ID format (alphanumeric, length 10-20) - silent 404 on invalid
    - Query database
    - If not found or expired → 404 (no details)
    - Return 200 with text/plain body
    - Add headers: Cache-Control: no-store, X-Content-Type-Options: nosniff

  **Must NOT do**:
  - Don't reveal whether ID was invalid vs expired vs not found (always 404)
  - Don't cache responses
  - Don't transform content

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Core retrieval path
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 7, 9, 10, 11)
  - **Blocks**: Tasks 12, 14
  - **Blocked By**: Tasks 4, 6

  **References**:
  - PRD Section 6.1.2: Retrieve Snippet endpoint spec
  - PRD Section 9.4: Error cases for retrieve

  **Acceptance Criteria**:

  **TDD - RED:**
  - [ ] Test: `TestGetSnippet_Success` - existing snippet returns 200 + content
  - [ ] Test: `TestGetSnippet_NotFound` - missing ID returns 404
  - [ ] Test: `TestGetSnippet_Expired` - expired snippet returns 404
  - [ ] Test: `TestGetSnippet_InvalidID` - invalid format returns 404
  - [ ] Test: `TestGetSnippet_Headers` - response has correct headers

  **TDD - GREEN:**
  - [ ] `go test ./internal/api -run TestGetSnippet` → PASS

  ```bash
  go test ./internal/api -v -run TestGetSnippet
  # Assert: All 5 tests pass
  ```

  **Commit**: YES
  - Message: `feat(api): implement retrieve snippet endpoint`
  - Files: `internal/api/handlers.go`, `internal/api/handlers_test.go`
  - Pre-commit: `go test ./internal/api`

---

- [ ] 9. Health Check Handlers

  **What to do**:
  - Implement `GET /healthz`:
    - Always return 200 OK (liveness)
    - Body: `{"status": "ok"}`
  - Implement `GET /readyz`:
    - Check database connection via repo.Ping()
    - If OK → 200 `{"status": "ready"}`
    - If fail → 503 `{"status": "not ready", "error": "..."}`

  **Must NOT do**:
  - Don't add complex health checks to /healthz (keep it fast)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple endpoints, standard pattern
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 7, 8, 10, 11)
  - **Blocks**: Task 14
  - **Blocked By**: Task 6

  **References**:
  - PRD Section 6.1.3: Health endpoints

  **Acceptance Criteria**:

  **TDD - RED/GREEN:**
  - [ ] Test: `TestHealthz_Returns200`
  - [ ] Test: `TestReadyz_DBUp_Returns200`
  - [ ] Test: `TestReadyz_DBDown_Returns503`
  - [ ] `go test ./internal/api -run TestHealth` → PASS

  ```bash
  go test ./internal/api -v -run "TestHealth|TestReadyz"
  # Assert: 3 tests pass
  ```

  **Commit**: YES
  - Message: `feat(api): add health check endpoints`
  - Files: `internal/api/handlers.go`, `internal/api/handlers_test.go`
  - Pre-commit: `go test ./internal/api`

---

- [ ] 10. Rate Limiting Middleware

  **What to do**:
  - Add go-chi/httprate dependency
  - Create rate limiting middleware in `internal/api/middleware.go`:
    - Separate limits for create vs read endpoints
    - Create: `httprate.LimitByIP(config.RateLimitCreate, time.Minute)`
    - Read: `httprate.LimitByIP(config.RateLimitRead, time.Minute)`
  - Return 429 with Retry-After header when exceeded
  - Return JSON error body: `{error: {code: "RATE_LIMITED", message: "..."}}`
  - Wire into router with appropriate route groups

  **Must NOT do**:
  - Don't use Redis (in-memory OK for single instance)
  - Don't rate limit health endpoints

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: httprate does heavy lifting
  - **Skills**: []
    - httprate patterns from research
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 7, 8, 9, 11)
  - **Blocks**: Task 14
  - **Blocked By**: Task 6

  **References**:
  - PRD Section 11: Rate Limiting Requirements
  - Research: httprate.LimitByIP example

  **Acceptance Criteria**:

  **TDD - RED/GREEN:**
  - [ ] Test: `TestRateLimit_Create_Exceeds` - 31st request in 1min returns 429
  - [ ] Test: `TestRateLimit_Read_Exceeds` - 301st request in 1min returns 429
  - [ ] Test: `TestRateLimit_HasRetryAfter` - 429 response has Retry-After header
  - [ ] `go test ./internal/api -run TestRateLimit` → PASS

  ```bash
  go test ./internal/api -v -run TestRateLimit
  # Assert: 3 tests pass
  ```

  **Commit**: YES
  - Message: `feat(api): add per-IP rate limiting`
  - Files: `internal/api/middleware.go`, `internal/api/server.go`, `internal/api/middleware_test.go`
  - Pre-commit: `go test ./internal/api`

---

- [ ] 11. Expiry Cleanup Goroutine

  **What to do**:
  - Create `internal/api/cleanup.go`:
    - `StartCleanup(ctx, repo, interval time.Duration)`
    - Runs in background goroutine
    - Calls `repo.DeleteExpired()` every interval
    - Logs number of deleted snippets
    - Respects context cancellation for graceful shutdown
  - Integrate into server startup

  **Must NOT do**:
  - Don't delete too aggressively (batch size limits if needed)
  - Don't panic on DB errors (log and continue)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple goroutine, clear pattern
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 7, 8, 9, 10)
  - **Blocks**: Task 14
  - **Blocked By**: Tasks 2, 4

  **References**:
  - PRD Section 7.2: Expiry deletion strategy

  **Acceptance Criteria**:

  **TDD - RED/GREEN:**
  - [ ] Test: `TestCleanup_DeletesExpired` - expired snippets removed after interval
  - [ ] Test: `TestCleanup_RespectsContext` - stops when context cancelled
  - [ ] `go test ./internal/api -run TestCleanup` → PASS

  ```bash
  go test ./internal/api -v -run TestCleanup
  # Assert: 2 tests pass
  ```

  **Commit**: YES
  - Message: `feat(api): add background cleanup for expired snippets`
  - Files: `internal/api/cleanup.go`, `internal/api/cleanup_test.go`
  - Pre-commit: `go test ./internal/api`

---

### Wave 4: CLI + Integration

- [ ] 12. CLI Implementation

  **What to do**:
  - Implement `cmd/tafcha/main.go` using Cobra:
    - Root command reads from stdin
    - Flags: `--expiry`, `--api`, `--timeout`, `--quiet`, `--version`
    - Detect pipe: `os.Stdin.Stat()` check for ModeCharDevice
    - Read stdin fully: `io.ReadAll(os.Stdin)`
    - Validate non-empty content
    - Parse and validate expiry
    - Make HTTP POST to API
    - Print URL to stdout on success
    - Print errors to stderr
    - Exit codes: 0=success, 1=input, 2=network, 3=server error
  - Create `internal/cli/client.go`:
    - HTTP client with configurable timeout
    - `Post(content []byte, expirySeconds int) (*Response, error)`
    - Parse JSON response

  **Must NOT do**:
  - Don't accept file paths as arguments
  - Don't read from stdin if not piped
  - Don't retry on failure (simple single attempt)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: User-facing component, UX critical
  - **Skills**: []
    - Cobra patterns from research
    - stdin detection pattern from GitHub
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Task 13)
  - **Blocks**: Task 14
  - **Blocked By**: Tasks 7, 8

  **References**:
  - PRD Section 5: CLI Requirements
  - Research: os.Stdin.Stat() pattern from rclone, nomad
  - Research: Cobra flag patterns

  **Acceptance Criteria**:

  **TDD - RED/GREEN:**
  - [ ] Test: `TestCLI_PipeInput` - piped input works
  - [ ] Test: `TestCLI_NoInput_ExitsError` - no pipe exits 1
  - [ ] Test: `TestCLI_EmptyInput_ExitsError` - empty stdin exits 1
  - [ ] Test: `TestCLI_ExpiryParsing` - --expiry flag parsed
  - [ ] Test: `TestClient_Post_Success` - HTTP client works
  - [ ] `go test ./internal/cli ./cmd/tafcha` → PASS

  ```bash
  # Unit tests:
  go test ./internal/cli -v
  # Assert: Client tests pass
  
  # Build CLI:
  go build -o tafcha ./cmd/tafcha
  # Assert: Binary created
  
  # Manual verification (with server running):
  echo "hello" | ./tafcha --api http://localhost:8080
  # Assert: Prints URL to stdout, exit 0
  ```

  **Commit**: YES
  - Message: `feat(cli): implement tafcha command-line tool`
  - Files: `cmd/tafcha/main.go`, `internal/cli/client.go`, `internal/cli/client_test.go`
  - Pre-commit: `go test ./internal/cli`

---

- [ ] 13. Docker Configuration

  **What to do**:
  - Create `Dockerfile`:
    - Multi-stage build
    - Stage 1: Build both binaries with Go
    - Stage 2: Minimal runtime (scratch or distroless)
    - Copy binaries
    - Default CMD: tafcha-server
  - Create `docker-compose.yml`:
    - Service: tafcha-server
    - Service: postgres (for local dev, not for production)
    - Environment variables configured
    - Health checks defined
  - Create `.dockerignore`:
    - Exclude .git, .sisyphus, etc.

  **Must NOT do**:
  - Don't include dev tools in final image
  - Don't hardcode secrets

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard Docker patterns
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - None needed for standard Dockerfile

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Task 12)
  - **Blocks**: Task 14
  - **Blocked By**: Task 1

  **References**:
  - External: https://docs.docker.com/language/golang/build-images/ - Go Docker best practices

  **Acceptance Criteria**:

  ```bash
  # Build image:
  docker build -t tafcha .
  # Assert: Build succeeds
  
  # Check image size:
  docker images tafcha --format "{{.Size}}"
  # Assert: Size < 50MB (minimal image)
  
  # Run with compose:
  docker compose up -d
  docker compose ps
  # Assert: Services running
  
  # Cleanup:
  docker compose down
  ```

  **Commit**: YES
  - Message: `build: add Docker configuration for deployment`
  - Files: `Dockerfile`, `docker-compose.yml`, `.dockerignore`
  - Pre-commit: `docker build -t tafcha .`

---

- [ ] 14. Integration Tests

  **What to do**:
  - Create `tests/integration_test.go`:
    - Use testcontainers for PostgreSQL
    - Test full flow: CLI → API → Database → Retrieval
    - Test expiry behavior
    - Test rate limiting
    - Test error cases end-to-end
  - Ensure tests can run with `go test ./tests/...`

  **Must NOT do**:
  - Don't mock database (real PostgreSQL via testcontainers)
  - Don't skip any error paths

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Final validation, comprehensive testing
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - None needed

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential (final task)
  - **Blocks**: None (final)
  - **Blocked By**: All previous tasks

  **References**:
  - All PRD acceptance criteria
  - External: https://golang.testcontainers.org/ - testcontainers-go

  **Acceptance Criteria**:

  ```bash
  # Run integration tests:
  go test ./tests/... -v -timeout 120s
  # Assert: All tests pass
  
  # Coverage report:
  go test ./... -coverprofile=coverage.out
  go tool cover -func=coverage.out | tail -1
  # Assert: Total coverage >= 80%
  ```

  **Commit**: YES
  - Message: `test: add integration tests for full system flow`
  - Files: `tests/integration_test.go`
  - Pre-commit: `go test ./tests/...`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `chore: initialize project structure and dependencies` | go.mod, cmd/, internal/ | `go build ./...` |
| 2 | `feat(config): add environment configuration loader` | internal/config/ | `go test ./internal/config` |
| 3 | `feat(id): add secure short ID generator using nanoid` | internal/id/ | `go test ./internal/id` |
| 4 | `feat(storage): add PostgreSQL repository with embedded migrations` | migrations/, internal/storage/ | `go test ./internal/storage` |
| 5 | `feat(expiry): add duration parser for m/h/d formats` | internal/expiry/ | `go test ./internal/expiry` |
| 6 | `feat(api): add chi router skeleton with middleware` | internal/api/ | `go test ./internal/api` |
| 7 | `feat(api): implement create snippet endpoint` | internal/api/ | `go test ./internal/api` |
| 8 | `feat(api): implement retrieve snippet endpoint` | internal/api/ | `go test ./internal/api` |
| 9 | `feat(api): add health check endpoints` | internal/api/ | `go test ./internal/api` |
| 10 | `feat(api): add per-IP rate limiting` | internal/api/ | `go test ./internal/api` |
| 11 | `feat(api): add background cleanup for expired snippets` | internal/api/ | `go test ./internal/api` |
| 12 | `feat(cli): implement tafcha command-line tool` | cmd/tafcha/, internal/cli/ | `go test ./internal/cli` |
| 13 | `build: add Docker configuration for deployment` | Dockerfile, docker-compose.yml | `docker build .` |
| 14 | `test: add integration tests for full system flow` | tests/ | `go test ./tests/...` |

---

## Success Criteria

### Verification Commands
```bash
# Build both binaries
go build -o bin/tafcha ./cmd/tafcha
go build -o bin/tafcha-server ./cmd/tafcha-server
# Expected: Both binaries created in bin/

# Run all tests
go test ./... -v
# Expected: All tests pass

# Coverage check
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1
# Expected: total: (statements) >= 80%

# Start server with Neon
DATABASE_URL="postgresql://..." PORT=8080 ./bin/tafcha-server &

# Test create
echo "hello world" | ./bin/tafcha --api http://localhost:8080
# Expected: URL printed, e.g., http://localhost:8080/AbC123def456

# Test retrieve
curl http://localhost:8080/AbC123def456
# Expected: "hello world"

# Test expiry flag
echo "test" | ./bin/tafcha --expiry 10m --api http://localhost:8080
# Expected: URL with 10-minute expiry

# Test rate limiting
for i in {1..35}; do echo "x" | curl -s -X POST -H "Content-Type: text/plain" --data-binary @- http://localhost:8080/api/v1/snippets; done
# Expected: 429 after 30 requests

# Docker build
docker build -t tafcha:latest .
# Expected: Image built successfully
```

### Final Checklist
- [ ] All "Must Have" from PRD present
- [ ] All "Must NOT Have" guardrails enforced
- [ ] All tests pass with >= 80% coverage
- [ ] Docker image builds and runs
- [ ] CLI works end-to-end with live API
- [ ] Rate limiting triggers correctly
- [ ] Expired snippets return 404
- [ ] Health endpoints respond correctly
