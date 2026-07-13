//go:build wasm

package main

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/goflare-demo/modules/contact"
	"github.com/tinywasm/goflare-demo/routes"
	"github.com/tinywasm/goflare/d1"
	"github.com/tinywasm/goflare/edge"
	"github.com/tinywasm/goflare/files"
	"github.com/tinywasm/goflare/r2"
	"github.com/tinywasm/router"
)

func main() {
	db, err := d1.NewEdge("DB")
	if err != nil {
		fmt.Println("d1:", err)
		return
	}
	if err := db.CreateTable(&contact.Contact{}); err != nil {
		fmt.Println("migrate:", err)
		return
	}

	r := edge.NewRouter()
	routes.Register(r, db)

	bucket, err := r2.NewEdge("FILES")
	if err != nil {
		fmt.Println("r2:", err)
		return
	}
	store, err := files.New(bucket, "/api/files/")
	if err != nil {
		fmt.Println("files:", err)
		return
	}

	// Logging wrapper for Mount
	r.Use(func(ctx router.Context, next router.HandlerFunc) {
		if fmt.HasPrefix(ctx.Path(), "/api/files/") {
			if ctx.Method() == "PUT" {
				fmt.Println("demo: upload recibido", len(ctx.Body()), "bytes")
			} else if ctx.Method() == "GET" {
				key := ctx.Path()[len("/api/files/"):]
				if key != "" {
					fmt.Println("demo: sirviendo", key)
				}
			}
		}
		next(ctx)
	})

	store.Mount(r)

	edge.Serve(r)
}
