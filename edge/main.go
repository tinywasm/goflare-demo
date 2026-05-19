//go:build wasm

package main

import (
	"github.com/tinywasm/goflare/pages"
	"github.com/tinywasm/goflare-demo/routes"
)

func main() {
	r := pages.NewRouter()
	routes.Register(r)
	pages.Serve(r)
}
