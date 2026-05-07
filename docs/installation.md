# Installation

## Go binary

```bash
cd core
go build -o ./bin/drift ./cmd/drift
./bin/drift --help
```

## npm / bunx

```bash
npx @netpiedev/drift --help
bunx @netpiedev/drift --help
```

The wrapper auto-downloads the correct binary from GitHub Releases.  
Optional override:

```bash
DRIFT_BINARY_PATH=/absolute/path/to/drift npx @netpiedev/drift migrate up
```
