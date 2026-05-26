//go:build !wasm

package main

import (
	"log"
	"os"

	"github.com/tinywasm/fmt" // NO usar stdlib strings — convención tinywasm
	"github.com/tinywasm/goflare/devserver"
	"github.com/tinywasm/goflare-demo/routes"
)

// lookupArg lee -key=value o -key value de os.Args. Usa tinywasm/fmt, no strings.
func lookupArg(key string) string {
	prefix := "-" + key + "="
	args := os.Args[1:]
	for i, arg := range args {
		if fmt.HasPrefix(arg, prefix) {
			return fmt.Convert(arg).TrimPrefix(prefix).String()
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
		port = "6060"
	}
	publicDir := lookupArg("server_public_dir")
	if publicDir == "" {
		publicDir = "web/public"
	}

	r := devserver.NewRouter()
	routes.Register(r) // MISMAS rutas/handlers que el edge (edge/main.go)

	log.Printf("Dev server on :%s — static: %s, API: /api/*", port, publicDir)
	if err := devserver.ListenAndServe(":"+port, r, publicDir); err != nil {
		log.Fatal(err)
	}
}
