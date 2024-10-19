package main

import (
	"net/http"

	"github.com/unstoppablemango/pulumi-bun/pulumi-language-bun/proto/pulumi/pulumiconnect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	// Still very boiler-platy from https://connectrpc.com/docs/go/getting-started/
	lang := &LanguageServer{}
	mux := http.NewServeMux()
	path, handler := pulumiconnect.NewLanguageRuntimeHandler(lang)
	mux.Handle(path, handler)
	http.ListenAndServe(
		"localhost:8080",
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
}
