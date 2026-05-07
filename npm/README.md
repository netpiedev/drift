# @netpiedev/drift

Node.js/Bun wrapper for the Drift Go migration binary.

## Usage

```bash
npx @netpiedev/drift migrate up
bunx @netpiedev/drift migrate up
```

The wrapper automatically downloads the right Drift binary from GitHub Releases and caches it locally.

Optional override:

```bash
DRIFT_BINARY_PATH=/absolute/path/to/drift npx @netpiedev/drift migrate up
```
