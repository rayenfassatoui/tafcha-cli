# Tafcha

A CLI-first plain-text publishing service. Pipe text, get a URL.

```bash
echo "hello world" | tafcha
# https://tafcha.dev/AlNqaGNP4POi
```

Opening the URL returns the exact plain text. No accounts, no formatting, no bloat.

## Installation

### From Source

```bash
go install github.com/rayenfassatoui/tafcha-cli/cmd/tafcha@latest
go install github.com/rayenfassatoui/tafcha-cli/cmd/tafcha-server@latest
```

### From Binary

Download from [Releases](https://github.com/rayen/tafcha/releases).

## CLI Usage

```bash
# Basic usage - pipe any text
echo "hello" | tafcha

# From a file
cat script.sh | tafcha

# Custom expiry (10m, 12h, 3d, 1w)
echo "temporary" | tafcha --expiry 1h

# Quiet mode - only output URL
echo "secret" | tafcha -q

# Custom API server
echo "local" | tafcha --api http://localhost:8080
```

### CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--api` | `-a` | `https://tafcha.dev` | API server URL |
| `--expiry` | `-e` | `3d` | Expiry duration |
| `--timeout` | `-t` | `30s` | Request timeout |
| `--quiet` | `-q` | `false` | Only output URL |

## Server

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | *required* | PostgreSQL connection string |
| `PORT` | `8080` | Server port |
| `HOST` | `0.0.0.0` | Server host |
| `BASE_URL` | `http://localhost:8080` | Public URL for generated links |
| `MAX_CONTENT_SIZE` | `1048576` | Max content size (1 MiB) |
| `DEFAULT_EXPIRY` | `72h` | Default expiry (3 days) |
| `MIN_EXPIRY` | `10m` | Minimum expiry |
| `MAX_EXPIRY` | `720h` | Maximum expiry (30 days) |
| `POST_RATE_LIMIT` | `30` | POST requests per minute per IP |
| `GET_RATE_LIMIT` | `300` | GET requests per minute per IP |

### Running

```bash
export DATABASE_URL="postgresql://user:pass@host/db?sslmode=require"
./tafcha-server
```

### Docker

```bash
docker run -e DATABASE_URL="..." -p 8080:8080 ghcr.io/rayen/tafcha
```

## API

### Create Snippet

```bash
curl -X POST https://tafcha.dev -d "your content here"
curl -X POST "https://tafcha.dev?expiry=1d" -d "expires in 1 day"
```

Response:
```json
{
  "id": "AlNqaGNP4POi",
  "url": "https://tafcha.dev/AlNqaGNP4POi",
  "expires_at": "2026-01-31T22:39:46Z"
}
```

### Get Snippet

```bash
curl https://tafcha.dev/AlNqaGNP4POi
# Returns plain text content
```

### Health Checks

```bash
curl https://tafcha.dev/healthz  # Liveness
curl https://tafcha.dev/readyz   # Readiness (includes DB check)
```

## Technical Details

- **IDs**: 12-character base62 (A-Z, a-z, 0-9) with ~71 bits of entropy
- **Storage**: PostgreSQL with automatic expired snippet cleanup
- **Rate Limiting**: Per-IP limits on POST (30/min) and GET (300/min)
- **Content Limit**: 1 MiB maximum

## Project Structure

```
tafcha/
├── cmd/
│   ├── tafcha/           # CLI binary
│   └── tafcha-server/    # Server binary
├── internal/
│   ├── api/              # HTTP handlers, middleware, cleanup worker
│   ├── cli/              # HTTP client for CLI
│   ├── config/           # Environment configuration
│   ├── expiry/           # Duration parsing (10m, 12h, 3d)
│   ├── id/               # Nanoid generation
│   └── storage/          # PostgreSQL repository
└── tests/                # Integration tests
```

## Development

```bash
# Run tests
go test ./...

# Build binaries
go build -o tafcha ./cmd/tafcha
go build -o tafcha-server ./cmd/tafcha-server

# Run server locally
export DATABASE_URL="postgresql://localhost/tafcha_dev"
go run ./cmd/tafcha-server
```

## License

MIT
