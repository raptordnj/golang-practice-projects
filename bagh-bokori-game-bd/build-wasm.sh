#!/usr/bin/env bash
# Build the WebAssembly / WebGL version of বাঘবন্দী.
#
# The same Go program that builds the desktop game builds here: Fyne renders
# through OpenGL on the desktop and through WebGL in the browser, so the
# engine, the AI and the UI are all reused unchanged.
set -euo pipefail

out="${1:-build/web}"
version="${VERSION:-dev}"
mkdir -p "$out"

echo "building main.wasm…"
# -s -w strips the debug tables. It saves little on its own, but there is no
# debugger on the far side of a browser to need them.
GOOS=js GOARCH=wasm go build \
  -ldflags "-s -w -X main.version=$version" \
  -o "$out/main.wasm" ./cmd/bagh-bandi

# A Go wasm binary is tens of megabytes but compresses to roughly a third of
# that, so ship it pre-compressed; tools/serve hands this file straight to any
# browser that accepts gzip.
echo "compressing…"
gzip -9 -f -k -c "$out/main.wasm" > "$out/main.wasm.gz"

runtime="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [ ! -f "$runtime" ]; then
  runtime="$(go env GOROOT)/misc/wasm/wasm_exec.js" # Go < 1.24
fi
cp "$runtime" "$out/wasm_exec.js"
cp web/index.html "$out/index.html"

raw=$(du -h "$out/main.wasm" | cut -f1)
gz=$(du -h "$out/main.wasm.gz" | cut -f1)

echo
echo "built into $out  (main.wasm $raw, $gz over the wire)"
echo "serve it - the wasm MIME type matters - with:"
echo "  go run ./tools/serve $out          # http://localhost:8080"
echo "  go run ./tools/serve $out :8088    # ...or any other free port"
