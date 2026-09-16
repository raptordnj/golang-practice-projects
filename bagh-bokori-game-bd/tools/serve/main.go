// Command serve is a small static file server for the WebAssembly build.
//
// It exists for two reasons a general-purpose static server usually gets
// wrong: browsers refuse to instantiate a .wasm file served with the wrong
// MIME type, and a Go wasm binary is large enough that serving the
// pre-compressed copy matters.
package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := "build/web"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	addr := ":8080"
	if len(os.Args) > 2 {
		addr = os.Args[2]
	}

	if err := mime.AddExtensionType(".wasm", "application/wasm"); err != nil {
		log.Fatalf("registering the wasm MIME type: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		log.Fatalf("nothing to serve: %v (run ./build-wasm.sh first)", err)
	}

	fmt.Printf("serving %s at http://localhost%s\n", dir, addr)
	if err := http.ListenAndServe(addr, gzipWasm(dir, http.FileServer(http.Dir(dir)))); err != nil {
		log.Fatal(err)
	}
}

// gzipWasm serves main.wasm.gz in place of main.wasm when the browser accepts
// gzip and the compressed file exists, which turns a ~37 MB download into
// roughly a third of that. Everything else falls through untouched.
func gzipWasm(dir string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".wasm") ||
			!strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz := filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/"))+".gz")
		f, err := os.Open(gz)
		if err != nil {
			next.ServeHTTP(w, r) // no compressed copy: serve the plain one
			return
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/wasm")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		http.ServeContent(w, r, "", info.ModTime(), f)
	})
}
