# redish

Quick, simple and small redis-cli written in GO

Inspired by [Redli](https://github.com/IBM-Cloud/redli)

## Installation

```bash
go install github.com/dansailer/redish@latest
```

Or build from source:

```bash
git clone https://github.com/dansailer/redish.git
cd redish
go build -o redish main.go
```

## Usage

### Basic Connection

```bash
# Connect to local Redis
./redish

# Connect to remote Redis
./redish -uri redis.example.com:6379

# Connect with authentication
./redish -uri redis.example.com:6379 -user myuser -password mypassword

# Connect with TLS
./redish -uri redis.example.com:6380 -tls

# Connect with insecure TLS (skip certificate validation)
./redish -uri redis.example.com:6380 -tls -insecure
```

### Command Line Execution

```bash
# Execute single command
./redish -commands "PING"

# Execute multiple commands
./redish -commands "SET key value;GET key;DEL key"
```

### Environment Variables

```bash
# Set password via environment variable
export REDIS_PASSWORD=mysecretpassword
./redish -uri redis.example.com:6379
```

## Testing

### Quick Local Test

```bash
# Run the test script (requires Docker or Podman)
./test-redis-connectivity.sh
```

## Development

### Building

```bash
go build -o redish main.go
```

## Releasing

This repository uses Goreleaser in CI to build cross-platform binaries, create archives, generate checksums, and publish a GitHub release.

CI behavior

- The GitHub Actions workflow triggers on pushed tags matching `v*`.
- A `test` job runs first (it starts a `redis` service for integration tests). If tests pass, the `build_release` job runs goreleaser to produce artifacts and publish the release.

Triggering a release (normal flow)

1. Create a signed or annotated tag locally, for example:

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

1. The Actions workflow will run on the pushed tag and publish artifacts to a GitHub release.

Running goreleaser locally (dry run)

- Install goreleaser: https://goreleaser.com/install/
- To test a release locally without publishing, run:

```bash
goreleaser release --snapshot --rm-dist
```

Publishing locally (be careful — this will attempt to publish to GitHub):

```bash
GITHUB_TOKEN=ghp_xxx goreleaser release --rm-dist
```

Notes

- The CI job runs a Redis container to satisfy integration tests. When running tests locally, make sure Redis is running on `127.0.0.1:6379` (or adapt tests).
- The goreleaser configuration is in `.goreleaser.yml`. Adjust it if you need different archive names, signatures, or additional artifacts.

## License

See [LICENSE](LICENSE) file for details.
