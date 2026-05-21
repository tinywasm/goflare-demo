//go:build !wasm

package main

import (
	"log"
	"os"
	"strings"

	"github.com/tinywasm/goflare/pages/devserver"
	"github.com/tinywasm/goflare-demo/routes"
)

// lookupArg returns the value for -key=value or -key value in os.Args.
func lookupArg(key string) string {
	prefix := "-" + key + "="
	args := os.Args[1:]
	for i, arg := range args {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
		if arg == "-"+key && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func main() {
	port := lookupArg("server_port")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	publicDir := lookupArg("server_public_dir")
	if publicDir == "" {
		publicDir = "web/public"
	}

	log.Printf("Serving static files from: %s on port %s", publicDir, port)

	r := devserver.NewRouter()
	routes.Register(r)

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	if err := devserver.ListenAndServe(port, r, publicDir); err != nil {
		log.Fatal(err)
	}
}
