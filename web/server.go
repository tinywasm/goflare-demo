//go:build !wasm

package main

import (
	"os"

	"github.com/tinywasm/fmt"
	"github.com/tinywasm/goflare-demo/modules/contact"
	"github.com/tinywasm/goflare-demo/routes"
	"github.com/tinywasm/goflare/devserver"
	"github.com/tinywasm/server/httpd"
	"github.com/tinywasm/sqlite"
)

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
		port = "8080"
	}
	publicDir := lookupArg("server_public_dir")
	if publicDir == "" {
		publicDir = "web/public"
	}

	db, err := sqlite.Open(":memory:")
	if err != nil {
		fmt.Println("sqlite:", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.CreateTable(&contact.Contact{}); err != nil {
		fmt.Println("migrate:", err)
		os.Exit(1)
	}

	srv := devserver.New(httpd.Config{
		Port:      port,
		PublicDir: publicDir,
	})
	routes.Register(srv.Router(), db)

	fmt.Println("Dev server on :"+port+" — static:", publicDir, "API: /api/*")
	if err := srv.ListenAndServe(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
